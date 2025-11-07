package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Listener interface {
	// Start starts the queue listener
	Start() error
	// Shutdown stops the queue listener
	Shutdown() error
	// IsConnected returns true if the listener is connected to the queue
	IsConnected() bool
}

const (
	Success = iota + 1
	Failed
	InternalError
)

const MessageTypeTask = "task"
const MessageTypeHandshake = "handshake"
const MessageTypeStatus = "status"

const MaxQueuePriority = 3

const (
	numConnectTries     = 10
	reconnectBackoff    = 1 * time.Second
	maxReconnectBackoff = 60 * time.Second
)

type listener struct {
	// Service that handles task-related operations
	database          database.Database
	taskService       service.TaskService
	queueService      service.QueueService
	submissionService service.SubmissionService
	languageService   service.LanguageService

	// Broker configuration
	brokerConfig config.BrokerConfig

	// RabbitMQ connection and channel
	connMux sync.RWMutex
	conn    *amqp.Connection
	channel *amqp.Channel

	// Control channels
	done     chan struct{}
	shutdown chan struct{}

	// Logger
	logger *zap.SugaredLogger
}

func NewListener(
	db database.Database,
	taskService service.TaskService,
	queueService service.QueueService,
	submissionService service.SubmissionService,
	langService service.LanguageService,
	brokerConfig config.BrokerConfig,
) Listener {
	log := utils.NewNamedLogger("queue_listener")

	return &listener{
		database:          db,
		taskService:       taskService,
		queueService:      queueService,
		submissionService: submissionService,
		languageService:   langService,
		brokerConfig:      brokerConfig,
		done:              make(chan struct{}),
		shutdown:          make(chan struct{}),
		logger:            log,
	}
}

func (ql *listener) IsConnected() bool {
	ql.connMux.RLock()
	defer ql.connMux.RUnlock()
	return ql.channel != nil && !ql.channel.IsClosed()
}

func (ql *listener) Start() error {
	ql.logger.Info("Starting queue listener...")

	// Try initial connection
	if err := ql.connect(); err != nil {
		ql.logger.Warnf("Initial connection failed: %v. Will retry in background...", err)
	}

	// Start background goroutine to manage connection
	go ql.run()

	return nil
}

func (ql *listener) Shutdown() error {
	ql.logger.Info("Shutting down the queue listener...")
	close(ql.shutdown)
	<-ql.done

	ql.connMux.Lock()
	defer ql.connMux.Unlock()

	if ql.channel != nil {
		if err := ql.channel.Close(); err != nil {
			ql.logger.Errorf("Failed to close the channel: %s", err.Error())
		}
	}
	if ql.conn != nil {
		if err := ql.conn.Close(); err != nil {
			ql.logger.Errorf("Failed to close the connection: %s", err.Error())
		}
	}

	ql.logger.Info("Queue listener shut down successfully")
	return nil
}

// connect establishes connection to RabbitMQ
func (ql *listener) connect() error {
	ql.connMux.Lock()
	defer ql.connMux.Unlock()

	brokerURL := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		ql.brokerConfig.User,
		ql.brokerConfig.Password,
		ql.brokerConfig.Host,
		ql.brokerConfig.Port,
	)

	var err error
	var conn *amqp.Connection

	// Try to connect with retries
	for i := range numConnectTries {
		conn, err = amqp.Dial(brokerURL)
		if err == nil {
			break
		}
		if i < numConnectTries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to connect after %d tries: %w", numConnectTries, err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create channel: %w", err)
	}

	// Declare the response queue
	args := make(amqp.Table)
	args["x-max-priority"] = MaxQueuePriority
	_, err = channel.QueueDeclare(
		ql.brokerConfig.ResponseQueueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		args,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	ql.conn = conn
	ql.channel = channel

	// Notify queue service of new connection
	if err := ql.queueService.SetConnection(conn, channel); err != nil {
		ql.logger.Warnf("Failed to set queue service connection: %v", err)
		channel.Close()
		conn.Close()
		ql.conn = nil
		ql.channel = nil
		return fmt.Errorf("failed to set queue service connection: %w", err)
	}

	ql.logger.Info("Successfully connected to RabbitMQ")
	return nil
}

// run manages the connection lifecycle
func (ql *listener) run() {
	defer close(ql.done)

	for {
		select {
		case <-ql.shutdown:
			ql.logger.Info("Shutdown signal received")
			return
		default:
		}

		// Check if we're connected
		if !ql.IsConnected() {
			ql.logger.Info("Not connected, attempting to connect...")
			if err := ql.connect(); err != nil {
				ql.logger.Warnf("Connection failed: %v. Retrying in %v...", err, reconnectBackoff)
				time.Sleep(reconnectBackoff)
				continue
			}
		}

		// Start listening
		if err := ql.listen(); err != nil {
			ql.logger.Warnf("Listening failed: %v. Will reconnect...", err)
			ql.closeConnection()

			// Wait before reconnecting
			select {
			case <-ql.shutdown:
				return
			case <-time.After(reconnectBackoff):
				// Reconnect after backoff
			}
		}
	}
}

// closeConnection safely closes the current connection
func (ql *listener) closeConnection() {
	ql.connMux.Lock()
	defer ql.connMux.Unlock()

	// Notify queue service that connection is being closed
	_ = ql.queueService.SetConnection(nil, nil)

	if ql.channel != nil {
		_ = ql.channel.Close()
		ql.channel = nil
	}
	if ql.conn != nil {
		_ = ql.conn.Close()
		ql.conn = nil
	}
}

func (ql *listener) listen() error {
	ql.connMux.RLock()
	channel := ql.channel
	queueName := ql.brokerConfig.ResponseQueueName
	ql.connMux.RUnlock()

	if channel == nil {
		return errors.New("channel is nil")
	}

	// Start consuming messages from the queue
	msgs, err := channel.Consume(
		queueName, // queue name
		"",        // consumer
		false,     // auto-ack -> set to false to use manual ack/nack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	// Monitor connection close
	ql.connMux.RLock()
	conn := ql.conn
	ql.connMux.RUnlock()

	connCloseChan := make(chan *amqp.Error, 1)
	chanCloseChan := make(chan *amqp.Error, 1)

	conn.NotifyClose(connCloseChan)
	channel.NotifyClose(chanCloseChan)

	ql.logger.Info("Started listening for messages...")

	// Process messages
	for {
		select {
		case <-ql.shutdown:
			return nil
		case err := <-connCloseChan:
			if err != nil {
				return fmt.Errorf("connection closed: %w", err)
			}
			return errors.New("connection closed")
		case err := <-chanCloseChan:
			if err != nil {
				return fmt.Errorf("channel closed: %w", err)
			}
			return errors.New("channel closed")
		case msg, ok := <-msgs:
			if !ok {
				return errors.New("message channel closed")
			}
			ql.processMessage(msg)
		}
	}
}

// TODO Implement better requeue mechanism, to try for a few times before dropping the message and marking as dropped.
func (ql *listener) processMessage(msg amqp.Delivery) {
	// Placeholder for processing the message
	ql.logger.Info("Processing message...")
	defer func() {
		if r := recover(); r != nil {
			ql.logger.Errorf("Recovered from panic: %s\n%s", r, string(debug.Stack()))
			// Attempt to Nack (requeue) the message. If this fails, log the error.
			if err := msg.Nack(false, false); err != nil {
				ql.logger.Errorf("Failed to nack and requeue message after panic: %s", err.Error())
			}
		}
	}()

	// Unmarshal the message body
	queueMessage := schemas.QueueResponseMessage{}
	err := json.Unmarshal(msg.Body, &queueMessage)
	if err != nil {
		ql.logger.Error("Failed to unmarshal the message:", err.Error())
		if err := msg.Nack(false, false); err != nil {
			ql.logger.Errorf("Failed to nack and requeue message: %s", err.Error())
		}
		return
	}
	ql.logger.Infof("Received message: %s of type %s", queueMessage.MessageID, queueMessage.Type)

	session := ql.database.NewSession()
	tx, err := session.BeginTransaction()
	if err != nil {
		ql.logger.Errorf("Failed to connect to database: %s", err)
		if err := msg.Nack(false, false); err != nil {
			ql.logger.Errorf("Failed to nack and requeue message: %s", err.Error())
		}
		return
	}

	switch queueMessage.Type {
	case MessageTypeTask:
		submissionID, err := strconv.ParseInt(queueMessage.MessageID, 10, 63)
		if err != nil {
			ql.logger.Errorf("Failed to get submission id: %s", err.Error())
			tx.Rollback()
			if err := msg.Nack(false, false); err != nil {
				ql.logger.Errorf("Failed to nack and requeue message: %s", err.Error())
			}
			return
		}
		ql.logger.Infof("Received task message for submission %d", submissionID)

		_, err = ql.submissionService.CreateSubmissionResult(tx, submissionID, queueMessage)
		if err != nil {
			ql.logger.Errorf("Failed to create user solution result: %s", err.Error())
			tx.Rollback()
			if err := msg.Nack(false, false); err != nil {
				ql.logger.Errorf("Failed to nack and requeue message: %s", err.Error())
			}
			return
		}
		ql.logger.Infof("Submission %d result created", submissionID)
		tx.Commit()
		// Ack the message to remove it from the queue
		if err := msg.Ack(false); err != nil {
			ql.logger.Errorf("Failed to ack message %s: %s", queueMessage.MessageID, err.Error())
		}
		ql.logger.Infof("Succesfuly processed message: %s", queueMessage.MessageID)

	case MessageTypeHandshake:
		ql.logger.Info("Handshake message received")

		handshakeResponse := schemas.HandShakeResponsePayload{}
		err = json.Unmarshal(queueMessage.Payload, &handshakeResponse)
		if err != nil {
			ql.logger.Errorf("Failed to unmarshal handshake response: %s", err.Error())
			tx.Rollback()
			if err := msg.Nack(false, false); err != nil {
				ql.logger.Errorf("Failed to nack and requeue message: %s", err.Error())
			}
			return
		}

		err := ql.languageService.Init(tx, handshakeResponse)
		if err != nil {
			ql.logger.Errorf("Failed to initialize languages: %s", err.Error())
			tx.Rollback()
			if err := msg.Nack(false, false); err != nil {
				ql.logger.Errorf("Failed to nack and requeue message: %s", err.Error())
			}
			return
		}
		tx.Commit()
		if err := msg.Ack(false); err != nil {
			ql.logger.Errorf("Failed to ack handshake message: %s", err.Error())
		}
		ql.logger.Info("Languages initialized")
		return

	case MessageTypeStatus:
		ql.logger.Info("Status message received")

		statusResponse := schemas.StatusResponsePayload{}
		err = json.Unmarshal(queueMessage.Payload, &statusResponse)
		if err != nil {
			ql.logger.Errorf("Failed to unmarshal status response: %s", err.Error())
			tx.Rollback()
			if err := msg.Nack(false, false); err != nil {
				ql.logger.Errorf("Failed to nack and requeue message: %s", err.Error())
			}
			return
		}

		err := ql.queueService.UpdateWorkerStatus(statusResponse)
		if err != nil {
			ql.logger.Errorf("Failed to update worker status: %s", err.Error())
			tx.Rollback()
			if err := msg.Nack(false, false); err != nil {
				ql.logger.Errorf("Failed to nack and requeue message: %s", err.Error())
			}
			return
		}
		tx.Commit()
		if err := msg.Ack(false); err != nil {
			ql.logger.Errorf("Failed to ack status message: %s", err.Error())
		}
		ql.logger.Info("Worker status updated")
		return
	default:
		ql.logger.Errorf("Unknown message type: %s", queueMessage.Type)
		tx.Rollback()
		if err := msg.Nack(false, false); err != nil {
			ql.logger.Errorf("Failed to nack and requeue message: %s", err.Error())
		}
		return
	}
}
