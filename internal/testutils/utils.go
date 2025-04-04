package testutils

import (
	"fmt"
	"testing"

	"github.com/mini-maxit/backend/internal/config"
	"github.com/rabbitmq/amqp091-go"
)

func NewTestConfig() *config.Config {
	dbPort := 5432
	apiPort := 8080
	brokerPort := 5672
	return &config.Config{
		DB: config.DBConfig{
			Host:     "localhost",
			Port:     uint16(dbPort),
			User:     "postgres",
			Password: "postgres",
			Name:     "test-maxit",
		},
		API: config.APIConfig{
			Port: uint16(apiPort),
		},
		Broker: config.BrokerConfig{
			QueueName:         "test_worker_queue",
			ResponseQueueName: "test_worker_response_queue",
			Host:              "localhost",
			Port:              uint16(brokerPort),
			User:              "guest",
			Password:          "guest",
		},
	}
}

func NewTestChannel(t *testing.T) (*amqp091.Connection, *amqp091.Channel) {
	cfg := NewTestConfig()
	brokerURL := fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.Broker.User, cfg.Broker.Password, cfg.Broker.Host, cfg.Broker.Port)
	conn, err := amqp091.Dial(brokerURL)
	if err != nil {
		t.Fatalf("failed to create a new amqp connection: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("failed to create a new amqp channel: %v", err)
	}

	return conn, ch
}
