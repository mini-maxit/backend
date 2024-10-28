package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mini-maxit/backend/internal/api/http/initialization"
	"github.com/mini-maxit/backend/package/service"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type QueueListener interface {
	// Start starts the queue listener
	Start() (context.CancelFunc, error)
	listen(ctx context.Context) error
}

type QueueListenerImpl struct {
	// Service that handles task-related operations
	taskService service.TaskService
	// RabbitMQ connection and channel
	conn    *amqp.Connection
	channel *amqp.Channel
	// Queue name
	queueName string
}

func NewQueueListener(initialization *initialization.Initialization) (*QueueListenerImpl, error) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/", initialization.Cfg.BrokerConfig.User, initialization.Cfg.BrokerConfig.Password, initialization.Cfg.BrokerConfig.Host, initialization.Cfg.BrokerConfig.Port))
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Declare the queue
	_, err = channel.QueueDeclare(
		initialization.Cfg.BrokerConfig.ResponseQueueName, // name of the queue
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}

	return &QueueListenerImpl{
		taskService: initialization.TaskService,
		conn:        conn,
		channel:     channel,
		queueName:   initialization.Cfg.BrokerConfig.ResponseQueueName,
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

type ResponseMessage struct {
	MessageID      string `json:"message_id"`
	TaskID         int64  `json:"task_id"`
	UserID         int64  `json:"user_id"`
	UserSolutionID int64  `json:"user_solution_id"`
	Result         Result `json:"result"`
}

type Result struct {
	Success     bool         `json:"Success"`
	StatusCode  int64        `json:"StatusCode"`
	Code        string       `json:"Code"`
	Message     string       `json:"Message"`
	TestResults []TestResult `json:"TestResults"`
}

type TestResult struct {
	InputFile    string `json:"InputFile"`
	ExpectedFile string `json:"ExpectedFile"`
	ActualFile   string `json:"ActualFile"`
	Passed       bool   `json:"Passed"`
	ErrorMessage string `json:"ErrorMessage"`
	Order        int64  `json:"Order"`
}

func (ql *QueueListenerImpl) processMessage(msg amqp.Delivery) {
	// Placeholder for processing the message
	logrus.Info("Received a message")

	// Unmarshal the message body
	queueMessage := ResponseMessage{}
	err := json.Unmarshal(msg.Body, &queueMessage)
	if err != nil {
		logrus.Errorf("Failed to unmarshal the message: %s", err)
	}
	logrus.Infof("Received message: %v", queueMessage)
	// You could implement task-specific processing here using ql.taskService
}
