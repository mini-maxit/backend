package queue

import (
	"context"
	"encoding/json"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/internal/logger"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/service"
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

	log := logger.NewNamedLogger("queue_listener")

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

func (ql *QueueListenerImpl) processMessage(msg amqp.Delivery) {
	// Placeholder for processing the message
	ql.logger.Info("Processing message...")

	// Unmarshal the message body
	queueMessage := schemas.ResponseMessage{}
	err := json.Unmarshal(msg.Body, &queueMessage)
	if err != nil {
		ql.logger.Error("Failed to unmarshal the message:", err.Error())
		return
	}
	ql.logger.Infof("Received message: %s", queueMessage.MessageId)

	tx, err := ql.database.Connect()
	if err != nil {
		ql.logger.Errorf("Failed to connect to database: %s", err)
		return
	}
	submissionId, err := ql.queueService.GetSubmissionId(tx, queueMessage.MessageId)
	if err != nil {
		ql.logger.Errorf("Failed to get submission id: %s", err.Error())
		return
	}
	if queueMessage.Result.StatusCode == InternalError {
		ql.submissionService.MarkSubmissionFailed(tx, submissionId, queueMessage.Result.Message)
		return
	}

	err = ql.submissionService.MarkSubmissionComplete(tx, submissionId)
	if err != nil {
		tx.Rollback()
		ql.logger.Errorf("Failed to mark submission as complete: %s", err.Error())
		return
	}
	_, err = ql.submissionService.CreateSubmissionResult(tx, submissionId, queueMessage)
	if err != nil {
		tx.Rollback()
		ql.logger.Errorf("Failed to create user solution result: %s", err.Error())
		return
	}
	ql.logger.Infof("Succesfuly processed message: %s", queueMessage.MessageId)
}
