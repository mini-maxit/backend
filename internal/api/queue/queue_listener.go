package queue

import (
	"context"
	"encoding/json"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueListener interface {
	// Start starts the queue listener
	Start() (context.CancelFunc, error)
	listen(ctx context.Context) error
}

const (
	Success = iota + 1
	Failed
	InternalError
)

type QueueListenerImpl struct {
	// Service that handles task-related operations
	database          database.Database
	taskService       service.TaskService
	queueService      service.QueueService
	submissionService service.SubmissionService
	// RabbitMQ connection and channel
	conn    *amqp.Connection
	channel *amqp.Channel
	// Queue name
	queueName string
	// Logger
	logger *zap.SugaredLogger
}

func NewQueueListener(conn *amqp.Connection, channel *amqp.Channel, taskService service.TaskService, queueName string) (*QueueListenerImpl, error) {
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
		taskService: taskService,
		conn:        conn,
		channel:     channel,
		queueName:   queueName,
		logger:      log,
	}, nil
}

func (ql *QueueListenerImpl) Start() (context.CancelFunc, error) {
	// Start the queue listener with a cancelable context
	ctx, cancel := context.WithCancel(context.Background())
	if err := ql.listen(ctx); err != nil {
		ql.logger.Error("Error listening to queue:", err.Error())
	}

	return cancel, nil
}

func (ql *QueueListenerImpl) listen(ctx context.Context) error {
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
		ql.logger.Info("Starting the message listener...")
		for {
			select {
			case <-ctx.Done():
				ql.logger.Info("Stopping the message listener...")
				return
			case msg := <-msgs:
				// Call the processMessage function with each message
				ql.processMessage(msg)
			}
		}
	}()

	return nil
}

// TODO Implement better requeue mechanism, to try for a few times before dropping the message and marking as dropped
func (ql *QueueListenerImpl) processMessage(msg amqp.Delivery) {
	// Placeholder for processing the message
	ql.logger.Info("Processing message...")

	// Unmarshal the message body
	queueMessage := schemas.ResponseMessage{}
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

	tx, err := ql.database.Connect()
	if err != nil {
		ql.logger.Errorf("Failed to connect to database: %s", err)
		err := msg.Reject(true)
		if err != nil {
			ql.logger.Errorf("Failed to reject and requeue message: %s", err.Error())
		}
		return
	}
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
	if queueMessage.Result.StatusCode == InternalError {
		err = ql.submissionService.MarkSubmissionFailed(tx, submissionId, queueMessage.Result.Message)
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
}
