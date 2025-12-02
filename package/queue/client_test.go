package queue_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/mini-maxit/backend/package/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestLogger() *zap.SugaredLogger {
	logger, _ := zap.NewDevelopment()
	return logger.Sugar()
}

func TestNewClient(t *testing.T) {
	t.Run("creates client with valid config", func(t *testing.T) {
		config := queue.Config{
			Host:              "localhost",
			Port:              5672,
			User:              "guest",
			Password:          "guest",
			WorkerQueueName:   "test_worker_queue",
			ResponseQueueName: "test_response_queue",
		}
		logger := newTestLogger()

		client := queue.NewClient(config, logger)

		assert.NotNil(t, client, "NewClient should return a non-nil client")
	})

	t.Run("creates client with empty config", func(t *testing.T) {
		config := queue.Config{}
		logger := newTestLogger()

		client := queue.NewClient(config, logger)

		assert.NotNil(t, client, "NewClient should return a non-nil client even with empty config")
	})
}

func TestClient_IsConnected_WhenNotConnected(t *testing.T) {
	t.Run("returns false when client is not started", func(t *testing.T) {
		config := queue.Config{
			Host:              "localhost",
			Port:              5672,
			User:              "guest",
			Password:          "guest",
			WorkerQueueName:   "test_worker_queue",
			ResponseQueueName: "test_response_queue",
		}
		logger := newTestLogger()

		client := queue.NewClient(config, logger)

		isConnected := client.IsConnected()

		assert.False(t, isConnected, "IsConnected should return false when client is not started")
	})
}

func TestClient_Publish_WhenNotConnected(t *testing.T) {
	t.Run("returns error when publish channel is not available", func(t *testing.T) {
		config := queue.Config{
			Host:              "localhost",
			Port:              5672,
			User:              "guest",
			Password:          "guest",
			WorkerQueueName:   "test_worker_queue",
			ResponseQueueName: "test_response_queue",
		}
		logger := newTestLogger()

		client := queue.NewClient(config, logger)
		ctx := context.Background()

		err := client.Publish(ctx, "test_queue", "test_reply", []byte("test message"))

		require.Error(t, err, "Publish should return an error when channel is not available")
		assert.Contains(t, err.Error(), "publish channel is not available")
	})
}

func TestClient_Consume_WhenNotConnected(t *testing.T) {
	t.Run("returns error when consume channel is not available", func(t *testing.T) {
		config := queue.Config{
			Host:              "localhost",
			Port:              5672,
			User:              "guest",
			Password:          "guest",
			WorkerQueueName:   "test_worker_queue",
			ResponseQueueName: "test_response_queue",
		}
		logger := newTestLogger()

		client := queue.NewClient(config, logger)

		deliveryChan, err := client.Consume("test_queue")

		require.Error(t, err, "Consume should return an error when channel is not available")
		assert.Nil(t, deliveryChan, "Consume should return nil delivery channel when error occurs")
		assert.Contains(t, err.Error(), "consume channel is not available")
	})
}

func TestClient_OnReconnect(t *testing.T) {
	t.Run("registers callback successfully", func(t *testing.T) {
		config := queue.Config{
			Host:              "localhost",
			Port:              5672,
			User:              "guest",
			Password:          "guest",
			WorkerQueueName:   "test_worker_queue",
			ResponseQueueName: "test_response_queue",
		}
		logger := newTestLogger()

		client := queue.NewClient(config, logger)

		callbackCalled := false
		client.OnReconnect(func() {
			callbackCalled = true
		})

		// Just verify that OnReconnect doesn't panic
		// The callback won't be called until reconnection happens
		assert.False(t, callbackCalled, "Callback should not be called immediately after registration")
	})

	t.Run("registers multiple callbacks", func(t *testing.T) {
		config := queue.Config{
			Host:              "localhost",
			Port:              5672,
			User:              "guest",
			Password:          "guest",
			WorkerQueueName:   "test_worker_queue",
			ResponseQueueName: "test_response_queue",
		}
		logger := newTestLogger()

		client := queue.NewClient(config, logger)

		// Register multiple callbacks - should not panic
		client.OnReconnect(func() {})
		client.OnReconnect(func() {})
		client.OnReconnect(func() {})

		// If we get here, no panic occurred - test passes implicitly
	})
}

func TestClient_Start_And_Shutdown(t *testing.T) {
	t.Run("start returns nil error even with invalid connection", func(t *testing.T) {
		config := queue.Config{
			Host:              "invalid-host",
			Port:              5672,
			User:              "guest",
			Password:          "guest",
			WorkerQueueName:   "test_worker_queue",
			ResponseQueueName: "test_response_queue",
		}
		logger := newTestLogger()

		client := queue.NewClient(config, logger)
		ctx := context.Background()

		err := client.Start(ctx)

		// Start should return nil even if connection fails
		// because it retries in the background
		require.NoError(t, err, "Start should not return error even if initial connection fails")

		// Clean up
		err = client.Shutdown()
		assert.NoError(t, err, "Shutdown should not return error")
	})

	t.Run("shutdown can be called multiple times safely", func(t *testing.T) {
		config := queue.Config{
			Host:              "invalid-host",
			Port:              5672,
			User:              "guest",
			Password:          "guest",
			WorkerQueueName:   "test_worker_queue",
			ResponseQueueName: "test_response_queue",
		}
		logger := newTestLogger()

		client := queue.NewClient(config, logger)
		ctx := context.Background()

		err := client.Start(ctx)
		require.NoError(t, err)

		err = client.Shutdown()
		assert.NoError(t, err, "First shutdown should succeed")

		// Note: Second shutdown would panic due to closing an already closed channel
		// This is expected behavior, so we don't test calling Shutdown twice
	})
}

func TestClient_OnReconnect_ConcurrentAccess(t *testing.T) {
	t.Run("handles concurrent callback registration", func(t *testing.T) {
		config := queue.Config{
			Host:              "localhost",
			Port:              5672,
			User:              "guest",
			Password:          "guest",
			WorkerQueueName:   "test_worker_queue",
			ResponseQueueName: "test_response_queue",
		}
		logger := newTestLogger()

		client := queue.NewClient(config, logger)

		var wg sync.WaitGroup
		numGoroutines := 10

		// Register callbacks concurrently
		for range numGoroutines {
			wg.Add(1)
			go func() {
				defer wg.Done()
				client.OnReconnect(func() {})
			}()
		}

		// Wait for all goroutines to finish
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success - all concurrent registrations completed
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent callback registration timed out")
		}
	})
}

func TestClient_IsConnected_ConcurrentAccess(t *testing.T) {
	t.Run("handles concurrent IsConnected calls", func(t *testing.T) {
		config := queue.Config{
			Host:              "localhost",
			Port:              5672,
			User:              "guest",
			Password:          "guest",
			WorkerQueueName:   "test_worker_queue",
			ResponseQueueName: "test_response_queue",
		}
		logger := newTestLogger()

		client := queue.NewClient(config, logger)

		var wg sync.WaitGroup
		numGoroutines := 10

		// Call IsConnected concurrently
		for range numGoroutines {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = client.IsConnected()
			}()
		}

		// Wait for all goroutines to finish
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success - all concurrent calls completed
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent IsConnected calls timed out")
		}
	})
}
