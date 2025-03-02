package testutils

import (
	"fmt"
	"testing"

	"github.com/mini-maxit/backend/internal/config"
	"github.com/rabbitmq/amqp091-go"
)

func NewTestConfig() *config.Config {
	return &config.Config{
		DB: config.DBConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			Name:     "test-maxit",
		},
		Api: config.ApiConfig{
			Port: 8080,
		},
		BrokerConfig: config.BrokerConfig{
			QueueName:         "test_worker_queue",
			ResponseQueueName: "test_worker_response_queue",
			Host:              "localhost",
			Port:              5672,
			User:              "guest",
			Password:          "guest",
		},
	}
}

func NewTestChannel(t *testing.T) (*amqp091.Connection, *amqp091.Channel) {
	cfg := NewTestConfig()
	conn, err := amqp091.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.BrokerConfig.User, cfg.BrokerConfig.Password, cfg.BrokerConfig.Host, cfg.BrokerConfig.Port))
	if err != nil {
		t.Fatalf("failed to create a new amqp connection: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("failed to create a new amqp channel: %v", err)
	}

	return conn, ch
}
