package main

import (
	"os"

	"github.com/joho/godotenv"
	_ "github.com/mini-maxit/backend/docs"
	"github.com/mini-maxit/backend/internal/api/http/initialization"
	"github.com/mini-maxit/backend/internal/api/http/server"
	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/internal/logger"
)

// @title Mini Maxit API Documentation testing the workflow
// @version 1.0
// @description This is the API documentation for the Mini Maxit API.
// @host localhost:8080
// @BasePath /api/v1
func main() {
	if _, ok := os.LookupEnv("DEBUG"); ok {
		err := godotenv.Load("././.env")
		if err != nil {
			panic(err)
		}
	}
	cfg := config.NewConfig()

	initialization := initialization.NewInitialization(cfg)

	logger.InitializeLogger()
	server_logger := logger.NewNamedLogger("server")

	queueListener := initialization.QueueListener
	cancel, err := queueListener.Start()
	if err != nil {
		logger.Log(&server_logger, "failed to start queue listener:", err.Error(), logger.Error)
		os.Exit(1)
	}

	server := server.NewServer(initialization, &server_logger)
	err = server.Start()
	if err != nil {
		cancel() // Stop the queue listener
		logger.Log(&server_logger, "failed to start server:", err.Error(), logger.Error)
		os.Exit(1)

	}

	cancel() // Stop the queue listener on graceful shutdown
}
