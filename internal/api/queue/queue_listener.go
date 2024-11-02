package queue

import (
	"context"
	"encoding/json"

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
	taskService       service.TaskService
	queueService      service.QueueService
	submissionService service.SubmissionService
	// RabbitMQ connection and channel
	conn    *amqp.Connection
	channel *amqp.Channel
	// Queue name
	queueName string
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

	return &QueueListenerImpl{
		taskService: taskService,
		conn:        conn,
		channel:     channel,
		queueName:   queueName,
	}, nil
}

func (ql *QueueListenerImpl) Start() (context.CancelFunc, error) {
	// Start the queue listener with a cancelable context
	ctx, cancel := context.WithCancel(context.Background())
	if err := ql.listen(ctx); err != nil {
		logrus.Printf("Error listening to queue: %v", err)
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
		logrus.Info("Listening for messages...")
		for {
			select {
			case <-ctx.Done():
				logrus.Println("Stopping the message listener...")
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
	logrus.Info("Received a message")

	// Unmarshal the message body
	queueMessage := schemas.ResponseMessage{}
	err := json.Unmarshal(msg.Body, &queueMessage)
	if err != nil {
		logrus.Errorf("Failed to unmarshal the message: %s", err)
		return
	}
	logrus.Infof("Received message: %v", queueMessage.MessageId)
	submissionId, err := ql.queueService.GetSubmissionId(queueMessage.MessageId)
	if err != nil {
		logrus.Errorf("Failed to get submission id: %s", err)
		return
	}
	if queueMessage.Result.StatusCode == InternalError {
		ql.submissionService.MarkSubmissionFailed(submissionId, queueMessage.Result.Message)
		return
	}

	ql.submissionService.MarkSubmissionComplete(submissionId)
	_, err = ql.submissionService.CreateSubmissionResult(submissionId, queueMessage)
	if err != nil {
		logrus.Errorf("Failed to create user solution result: %s", err)
		return
	}
	logrus.Infof("Succesfuly processed message: %v", queueMessage.MessageId)
}
