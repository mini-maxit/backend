package queue

import (
	"encoding/json"
	"errors"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/queue"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const (
	Success = iota + 1
	Failed
	InternalError
)

const MessageTypeTask = "task"
const MessageTypeHandshake = "handshake"
const MessageTypeStatus = "status"

type Listener interface {
	// Start starts the queue listener
	Start() error
	// Shutdown stops the queue listener
	Shutdown() error
}

type listener struct {
	// Services
	database          database.Database
	taskService       service.TaskService
	queueService      service.QueueService
	submissionService service.SubmissionService
	languageService   service.LanguageService

	// Queue client
	queueClient       queue.Client
	responseQueueName string

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
	queueClient queue.Client,
	responseQueueName string,
) Listener {
	log := utils.NewNamedLogger("queue_listener")

	return &listener{
		database:          db,
		taskService:       taskService,
		queueService:      queueService,
		submissionService: submissionService,
		languageService:   langService,
		queueClient:       queueClient,
		responseQueueName: responseQueueName,
		done:              make(chan struct{}),
		shutdown:          make(chan struct{}),
		logger:            log,
	}
}

func (ql *listener) Start() error {
	ql.logger.Info("Starting queue listener...")

	// Register callback to send handshake on reconnection
	ql.queueClient.OnReconnect(func() {
		ql.logger.Info("Queue reconnected, sending handshake...")
		if err := ql.queueService.PublishHandshake(); err != nil {
			ql.logger.Warnf("Failed to publish handshake after reconnect: %v", err)
		}
	})

	// Start listening in background
	go ql.run()

	return nil
}

func (ql *listener) Shutdown() error {
	ql.logger.Info("Shutting down the queue listener...")
	close(ql.shutdown)
	<-ql.done
	ql.logger.Info("Queue listener shut down successfully")
	return nil
}

func (ql *listener) run() {
	defer close(ql.done)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ql.shutdown:
			ql.logger.Info("Shutdown signal received")
			return
		case <-ticker.C:
			if ql.queueClient.IsConnected() {
				// Try to start listening (will block until error or shutdown)
				if err := ql.listen(); err != nil {
					ql.logger.Warnf("Listening failed: %v. Will retry...", err)
				}
			}
		}
	}
}

func (ql *listener) listen() error {
	msgs, err := ql.queueClient.Consume(ql.responseQueueName)
	if err != nil {
		return err
	}

	ql.logger.Info("Started listening for messages...")

	for {
		select {
		case <-ql.shutdown:
			return nil
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
	ql.logger.Info("Processing message...")
	defer func() {
		if r := recover(); r != nil {
			ql.logger.Errorf("Recovered from panic: %s\n%s", r, string(debug.Stack()))
			if err := msg.Nack(false, false); err != nil {
				ql.logger.Errorf("Failed to nack message after panic: %s", err)
			}
		}
	}()

	queueMessage := schemas.QueueResponseMessage{}
	err := json.Unmarshal(msg.Body, &queueMessage)
	if err != nil {
		ql.logger.Error("Failed to unmarshal message:", err)
		_ = msg.Nack(false, false)
		return
	}
	ql.logger.Infof("Received message: %s of type %s", queueMessage.MessageID, queueMessage.Type)

	session := ql.database.NewSession()
	_, err = session.BeginTransaction()
	if err != nil {
		ql.logger.Errorf("Failed to begin transaction: %s", err)
		_ = msg.Nack(false, false)
		return
	}

	switch queueMessage.Type {
	case MessageTypeTask:
		ql.handleTaskMessage(session, msg, queueMessage)
	case MessageTypeHandshake:
		ql.handleHandshakeMessage(session, msg, queueMessage)
	case MessageTypeStatus:
		ql.handleStatusMessage(session, msg, queueMessage)
	default:
		ql.logger.Errorf("Unknown message type: %s", queueMessage.Type)
		session.Rollback()
		_ = msg.Nack(false, false)
	}
}

func (ql *listener) handleTaskMessage(session database.Database, msg amqp.Delivery, queueMessage schemas.QueueResponseMessage) {
	submissionID, err := strconv.ParseInt(queueMessage.MessageID, 10, 64)
	if err != nil {
		ql.logger.Errorf("Failed to parse submission ID: %s", err)
		session.Rollback()
		_ = msg.Nack(false, false)
		return
	}
	ql.logger.Infof("Received task message for submission %d", submissionID)

	_, err = ql.submissionService.CreateSubmissionResult(session, submissionID, queueMessage)
	if err != nil {
		ql.logger.Errorf("Failed to create submission result: %s", err)
		session.Rollback()
		_ = msg.Nack(false, false)
		return
	}

	ql.logger.Infof("Submission %d result created", submissionID)
	session.Commit()
	_ = msg.Ack(false)
	ql.logger.Infof("Successfully processed message: %s", queueMessage.MessageID)
}

func (ql *listener) handleHandshakeMessage(session database.Database, msg amqp.Delivery, queueMessage schemas.QueueResponseMessage) {
	ql.logger.Info("Handshake message received")

	handshakeResponse := schemas.HandShakeResponsePayload{}
	err := json.Unmarshal(queueMessage.Payload, &handshakeResponse)
	if err != nil {
		ql.logger.Errorf("Failed to unmarshal handshake response: %s", err)
		session.Rollback()
		_ = msg.Nack(false, false)
		return
	}

	err = ql.languageService.Init(session, handshakeResponse)
	if err != nil {
		ql.logger.Errorf("Failed to initialize languages: %s", err)
		session.Rollback()
		_ = msg.Nack(false, false)
		return
	}

	session.Commit()
	_ = msg.Ack(false)
	ql.logger.Info("Languages initialized")
}

func (ql *listener) handleStatusMessage(session database.Database, msg amqp.Delivery, queueMessage schemas.QueueResponseMessage) {
	ql.logger.Info("Status message received")

	statusResponse := schemas.StatusResponsePayload{}
	err := json.Unmarshal(queueMessage.Payload, &statusResponse)
	if err != nil {
		ql.logger.Errorf("Failed to unmarshal status response: %s", err)
		session.Rollback()
		_ = msg.Nack(false, false)
		return
	}

	err = ql.queueService.UpdateWorkerStatus(statusResponse)
	if err != nil {
		ql.logger.Errorf("Failed to update worker status: %s", err)
		session.Rollback()
		_ = msg.Nack(false, false)
		return
	}

	session.Commit()
	_ = msg.Ack(false)
	ql.logger.Info("Worker status updated")
}
