package queue

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const (
	maxQueuePriority = 3
	numConnectTries  = 10
	reconnectBackoff = 60 * time.Second
)

// Publisher interface for publishing messages
type Publisher interface {
	Publish(ctx context.Context, queueName string, replyTo string, body []byte) error
	IsConnected() bool
}

// Consumer interface for consuming messages
type Consumer interface {
	Consume(queueName string) (<-chan amqp.Delivery, error)
	IsConnected() bool
}

// Client manages RabbitMQ connection, channels, and queue declarations
type Client interface {
	Publisher
	Consumer
	// Start initializes connection (non-blocking, retries in background)
	Start(ctx context.Context) error
	// Shutdown closes all connections gracefully
	Shutdown() error
	// OnReconnect registers a callback for reconnection events
	OnReconnect(callback func())
}

// Config holds queue client configuration
type Config struct {
	Host              string
	Port              uint16
	User              string
	Password          string
	WorkerQueueName   string
	ResponseQueueName string
}

type client struct {
	config Config
	logger *zap.SugaredLogger

	// Connection management
	connMux sync.RWMutex
	conn    *amqp.Connection
	pubChan *amqp.Channel // Channel for publishing
	conChan *amqp.Channel // Channel for consuming

	// Declared queues
	workerQueue   amqp.Queue
	responseQueue amqp.Queue

	// Reconnection callbacks
	reconnectCallbacks []func()
	callbackMux        sync.RWMutex

	// Control
	ctx      context.Context
	cancel   context.CancelFunc
	done     chan struct{}
	shutdown chan struct{}
}

// NewClient creates a new queue client
func NewClient(config Config, logger *zap.SugaredLogger) Client {
	return &client{
		config:             config,
		logger:             logger,
		reconnectCallbacks: make([]func(), 0),
		done:               make(chan struct{}),
		shutdown:           make(chan struct{}),
	}
}

func (c *client) Start(ctx context.Context) error {
	c.ctx, c.cancel = context.WithCancel(ctx)

	// Try initial connection
	if err := c.connect(); err != nil {
		c.logger.Warnf("Initial connection failed: %v. Will retry in background...", err)
	}

	// Start background reconnection manager
	go c.manageConnection()

	return nil
}

func (c *client) Shutdown() error {
	c.logger.Info("Shutting down queue client...")
	close(c.shutdown)
	c.cancel()
	<-c.done

	c.connMux.Lock()
	defer c.connMux.Unlock()

	if c.pubChan != nil {
		_ = c.pubChan.Close()
	}
	if c.conChan != nil {
		_ = c.conChan.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}

	c.logger.Info("Queue client shut down successfully")
	return nil
}

func (c *client) IsConnected() bool {
	c.connMux.RLock()
	defer c.connMux.RUnlock()
	return c.conn != nil && !c.conn.IsClosed() &&
		c.pubChan != nil && !c.pubChan.IsClosed() &&
		c.conChan != nil && !c.conChan.IsClosed()
}

func (c *client) Publish(ctx context.Context, queueName string, replyTo string, body []byte) error {
	c.connMux.RLock()
	pubChan := c.pubChan
	c.connMux.RUnlock()

	if pubChan == nil || pubChan.IsClosed() {
		return errors.New("publish channel is not available")
	}

	return pubChan.PublishWithContext(ctx, "", queueName, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
		ReplyTo:     replyTo,
	})
}

func (c *client) Consume(queueName string) (<-chan amqp.Delivery, error) {
	c.connMux.RLock()
	conChan := c.conChan
	c.connMux.RUnlock()

	if conChan == nil || conChan.IsClosed() {
		return nil, errors.New("consume channel is not available")
	}

	return conChan.Consume(
		queueName,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
}

func (c *client) OnReconnect(callback func()) {
	c.callbackMux.Lock()
	defer c.callbackMux.Unlock()
	c.reconnectCallbacks = append(c.reconnectCallbacks, callback)
}

// manageConnection runs in background and manages connection lifecycle
func (c *client) manageConnection() {
	defer close(c.done)

	ticker := time.NewTicker(reconnectBackoff)
	defer ticker.Stop()

	for {
		select {
		case <-c.shutdown:
			c.logger.Info("Connection manager shutting down")
			return
		case <-c.ctx.Done():
			c.logger.Info("Context cancelled, connection manager exiting")
			return
		case <-ticker.C:
			if !c.IsConnected() {
				c.logger.Debug("Connection lost, attempting to reconnect...")
				if err := c.connect(); err != nil {
					c.logger.Warnf("Reconnection failed: %v", err)
				} else {
					c.notifyReconnect()
				}
			}
		}
	}
}

// connect establishes connection to RabbitMQ and declares queues
func (c *client) connect() error {
	c.connMux.Lock()
	defer c.connMux.Unlock()

	// Close existing connections if any
	c.closeUnsafe()

	brokerURL := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		c.config.User,
		c.config.Password,
		c.config.Host,
		c.config.Port,
	)

	var err error
	var conn *amqp.Connection

	conn, err = amqp.Dial(brokerURL)
	if err != nil {
		return fmt.Errorf("failed to connect after %d tries: %w", numConnectTries, err)
	}

	// Create separate channels for publishing and consuming (best practice)
	pubChan, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create publish channel: %w", err)
	}

	conChan, err := conn.Channel()
	if err != nil {
		pubChan.Close()
		conn.Close()
		return fmt.Errorf("failed to create consume channel: %w", err)
	}

	// Declare queues
	args := amqp.Table{"x-max-priority": maxQueuePriority}

	workerQueue, err := pubChan.QueueDeclare(
		c.config.WorkerQueueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		args,
	)
	if err != nil {
		conChan.Close()
		pubChan.Close()
		conn.Close()
		return fmt.Errorf("failed to declare worker queue: %w", err)
	}

	responseQueue, err := conChan.QueueDeclare(
		c.config.ResponseQueueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		args,
	)
	if err != nil {
		conChan.Close()
		pubChan.Close()
		conn.Close()
		return fmt.Errorf("failed to declare response queue: %w", err)
	}

	c.conn = conn
	c.pubChan = pubChan
	c.conChan = conChan
	c.workerQueue = workerQueue
	c.responseQueue = responseQueue

	c.logger.Info("Successfully connected to RabbitMQ and declared queues")
	return nil
}

// closeUnsafe closes connection without locking (must be called with lock held)
func (c *client) closeUnsafe() {
	if c.pubChan != nil {
		_ = c.pubChan.Close()
		c.pubChan = nil
	}
	if c.conChan != nil {
		_ = c.conChan.Close()
		c.conChan = nil
	}
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
}

func (c *client) notifyReconnect() {
	c.callbackMux.RLock()
	callbacks := make([]func(), len(c.reconnectCallbacks))
	copy(callbacks, c.reconnectCallbacks)
	c.callbackMux.RUnlock()

	for _, callback := range callbacks {
		go callback()
	}
}
