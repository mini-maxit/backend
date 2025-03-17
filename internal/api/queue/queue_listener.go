package queue

import (
	"encoding/json"
	"runtime/debug"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueListener interface {
	// Start starts the queue listener
	Start() error
	listen() error
	Shutdown() error
}

const (
	Success = iota + 1
	Failed
	InternalError
)

const MessageTypeTask = "task"
const MessageTypeHandshake = "handshake"
const MessageTypeStatus = "status"

type QueueListenerImpl struct {
	// Service that handles task-related operations
	database          database.Database
	taskService       service.TaskService
	queueService      service.QueueService
	submissionService service.SubmissionService
	languageService   service.LanguageService
	// RabbitMQ connection and channel
	conn    *amqp.Connection
	channel *amqp.Channel
	done    chan error
	// Queue name
	queueName string
	// Logger
	logger *zap.SugaredLogger
}

func NewQueueListener(conn *amqp.Connection, channel *amqp.Channel, db database.Database, taskService service.TaskService, queueService service.QueueService, submissionService service.SubmissionService, langService service.LanguageService, queueName string) (*QueueListenerImpl, error) {
	// Declare the queue
	_, err := channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, err
	}

	log := utils.NewNamedLogger("queue_listener")

	return &QueueListenerImpl{
		database:          db,
		taskService:       taskService,
		queueService:      queueService,
		submissionService: submissionService,
		languageService:   langService,
		conn:              conn,
		channel:           channel,
		done:              make(chan error),
		queueName:         queueName,
		logger:            log,
	}, nil
}

func (ql *QueueListenerImpl) Start() error {
	// Start the queue listener with a cancelable context
	if err := ql.listen(); err != nil {
		ql.logger.Error("Error listening to queue:", err.Error())
	}

	return nil
}

func (ql *QueueListenerImpl) Shutdown() error {
	ql.logger.Info("Shutting down the queue listener...")
	if err := ql.channel.Close(); err != nil {
		ql.logger.Errorf("Failed to close the channel: %s", err.Error())
	}
	if err := ql.conn.Close(); err != nil {
		ql.logger.Errorf("Failed to close the connection: %s", err.Error())
	}
	return <-ql.done
}

func (ql *QueueListenerImpl) listen() error {
	// Start consuming messages from the queue
	msgs, err := ql.channel.Consume(
		ql.queueName, // queue name
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return err
	}

	// Process messages in a goroutine
	go func() {
		defer func() {
			ql.logger.Info("Closing the message listener...")
			ql.done <- nil
		}()
		ql.logger.Info("Starting the message listener...")
		for msg := range msgs {
			// Call the processMessage function with each message
			ql.processMessage(msg)
		}
	}()

	return nil
}

// TODO Implement better requeue mechanism, to try for a few times before dropping the message and marking as dropped
func (ql *QueueListenerImpl) processMessage(msg amqp.Delivery) {
	// Placeholder for processing the message
	ql.logger.Info("Processing message...")
	defer func() {
		if r := recover(); r != nil {
			ql.logger.Errorf("Recovered from panic: %s\n%s", r, string(debug.Stack()))
			err := msg.Reject(true)
			if err != nil {
				ql.logger.Errorf("Failed to reject and requeue message: %s", err.Error())
			}
		}
	}()

	// Unmarshal the message body
	queueMessage := schemas.QueueResponseMessage{}
	err := json.Unmarshal(msg.Body, &queueMessage)
	if err != nil {
		ql.logger.Error("Failed to unmarshal the message:", err.Error())
		err := msg.Reject(true)
		if err != nil {
			ql.logger.Errorf("Failed to reject and requeue message: %s", err.Error())
		}
		return
	}
	ql.logger.Infof("Received message: %s", queueMessage.MessageId)

	session := ql.database.NewSession()
	tx, err := session.BeginTransaction()
	if err != nil {
		ql.logger.Errorf("Failed to connect to database: %s", err)
		err := msg.Reject(true)
		if err != nil {
			ql.logger.Errorf("Failed to reject and requeue message: %s", err.Error())
		}
		return
	}
	
	switch queueMessage.Type {
	case MessageTypeTask:
		submissionId, err := ql.queueService.GetSubmissionId(tx, queueMessage.MessageId)
		if err != nil {
			ql.logger.Errorf("Failed to get submission id: %s", err.Error())
			tx.Rollback()
			err := msg.Reject(true)
			if err != nil {
				ql.logger.Errorf("Failed to reject and requeue message: %s", err.Error())
			}
			return
		}
		taskResponse := schemas.TaskResponsePayload{}
		err = json.Unmarshal(queueMessage.Payload, &taskResponse)
		if err != nil {
			ql.logger.Errorf("Failed to unmarshal task response: %s", err.Error())
			tx.Rollback()
			err := msg.Reject(true)
			if err != nil {
				ql.logger.Errorf("Failed to reject and requeue message: %s", err.Error())
			}
			return
		}
		if taskResponse.StatusCode == InternalError {
			err = ql.submissionService.MarkSubmissionFailed(tx, submissionId, taskResponse.Message)
			if err != nil {
				tx.Rollback()
				err := msg.Reject(true)
				if err != nil {
					ql.logger.Errorf("Failed to reject and requeue message: %s", err.Error())
				}
			}
			return
		}

		err = ql.submissionService.MarkSubmissionComplete(tx, submissionId)
		if err != nil {
			ql.logger.Errorf("Failed to mark submission as complete: %s", err.Error())
			tx.Rollback()
			err := msg.Reject(true)
			if err != nil {
				ql.logger.Errorf("Failed to reject and requeue message: %s", err.Error())
			}
			return
		}
		_, err = ql.submissionService.CreateSubmissionResult(tx, submissionId, queueMessage)
		if err != nil {
			ql.logger.Errorf("Failed to create user solution result: %s", err.Error())
			tx.Rollback()
			err := msg.Reject(true)
			if err != nil {
				ql.logger.Errorf("Failed to reject and requeue message: %s", err.Error())
			}
			return
		}
		tx.Commit()
		ql.logger.Infof("Succesfuly processed message: %s", queueMessage.MessageId)
	case MessageTypeHandshake:
		ql.logger.Info("Handshake message received")

		handshakeResponse := schemas.HandShakeResponsePayload{}
		err = json.Unmarshal(queueMessage.Payload, &handshakeResponse)
		if err != nil {
			ql.logger.Errorf("Failed to unmarshal handshake response: %s", err.Error())
			tx.Rollback()
			err := msg.Reject(true)
			if err != nil {
				ql.logger.Errorf("Failed to reject and requeue message: %s", err.Error())
			}
			return
		}

		err := ql.languageService.InitLanguages(tx, handshakeResponse)
		if err != nil {
			ql.logger.Errorf("Failed to initialize languages: %s", err.Error())
			tx.Rollback()
			err := msg.Reject(true)
			if err != nil {
				ql.logger.Errorf("Failed to reject and requeue message: %s", err.Error())
			}
			return
		}
		tx.Commit()
		ql.logger.Info("Languages initialized")
		return
	case MessageTypeStatus:
		ql.logger.Info("Status message received")

		statusResponse := schemas.StatusResponsePayload{}
		err = json.Unmarshal(queueMessage.Payload, &statusResponse)
		if err != nil {
			ql.logger.Errorf("Failed to unmarshal status response: %s", err.Error())
			tx.Rollback()
			err := msg.Reject(true)
			if err != nil {
				ql.logger.Errorf("Failed to reject and requeue message: %s", err.Error())
			}
			return
		}

		err := ql.queueService.UpdateWorkerStatus(tx, statusResponse)
		if err != nil {
			ql.logger.Errorf("Failed to update worker status: %s", err.Error())
			tx.Rollback()
			err := msg.Reject(true)
			if err != nil {
				ql.logger.Errorf("Failed to reject and requeue message: %s", err.Error())
			}
			return
		}
		tx.Commit()
		ql.logger.Info("Worker status updated")
		return
	default:
		ql.logger.Errorf("Unknown message type: %s", queueMessage.Type)
		tx.Rollback()
		err := msg.Reject(true)
		if err != nil {
			ql.logger.Errorf("Failed to reject and requeue message: %s", err.Error())
		}
		return
	}
}
