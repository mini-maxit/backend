package queue

import (
	"context"
	"encoding/json"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/internal/logger"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/service"
	"github.com/sirupsen/logrus"

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
	queue_logger *logger.ServiceLogger
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

	queue_logger := logger.NewNamedLogger("queue_listener")

	return &QueueListenerImpl{
		taskService:  taskService,
		conn:         conn,
		channel:      channel,
		queueName:    queueName,
		queue_logger: &queue_logger,
	}, nil
}

func (ql *QueueListenerImpl) Start() (context.CancelFunc, error) {
	// Start the queue listener with a cancelable context
	ctx, cancel := context.WithCancel(context.Background())
	if err := ql.listen(ctx); err != nil {
		logger.Log(ql.queue_logger, "Error listening to queue:", err.Error(), logger.Error)
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
		logger.Log(ql.queue_logger, "Starting the message listener...", "", logger.Info)
		for {
			select {
			case <-ctx.Done():
				logger.Log(ql.queue_logger, "Stopping the message listener...", "", logger.Info)
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
	logger.Log(ql.queue_logger, "Received a message", "", logger.Info)

	// Unmarshal the message body
	queueMessage := schemas.ResponseMessage{}
	err := json.Unmarshal(msg.Body, &queueMessage)
	if err != nil {
		logger.Log(ql.queue_logger, "Failed to unmarshal the message:", err.Error(), logger.Error)
		return
	}
	logger.Log(ql.queue_logger, "Received message: "+queueMessage.MessageId, "", logger.Info)

	tx, err := ql.database.Connect()
	if err != nil {
		logrus.Errorf("Failed to connect to database: %s", err)
		return
	}
	submissionId, err := ql.queueService.GetSubmissionId(tx, queueMessage.MessageId)
	if err != nil {
		logger.Log(ql.queue_logger, "Failed to get submission id:", err.Error(), logger.Error)
		return
	}
	if queueMessage.Result.StatusCode == InternalError {
		ql.submissionService.MarkSubmissionFailed(tx, submissionId, queueMessage.Result.Message)
		return
	}

	err = ql.submissionService.MarkSubmissionComplete(tx, submissionId)
	if err != nil {
		tx.Rollback()
		logger.Log(ql.queue_logger, "Failed to mark submission as complete:", err.Error(), logger.Error)
		return
	}
	_, err = ql.submissionService.CreateSubmissionResult(tx, submissionId, queueMessage)
	if err != nil {
		tx.Rollback()
		logger.Log(ql.queue_logger, "Failed to create user solution result:", err.Error(), logger.Error)
		return
	}
	logger.Log(ql.queue_logger, "Succesfuly processed message: "+queueMessage.MessageId, "", logger.Info)
}
