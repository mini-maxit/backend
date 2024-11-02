package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/mini-maxit/backend/internal/api/http/initialization"
	"github.com/mini-maxit/backend/internal/api/http/server"
	"github.com/mini-maxit/backend/internal/config"
	"github.com/sirupsen/logrus"
)

func main() {
	if _, ok := os.LookupEnv("DEBUG"); ok {
		err := godotenv.Load("././.env")
		if err != nil {
			panic(err)
		}
	}
	cfg := config.NewConfig()

	initialization := initialization.NewInitialization(cfg)

	queueListener := initialization.QueueListener
	cancel, err := queueListener.Start()
	if err != nil {
		logrus.Errorf("failed to start queue listener: %v", err)
		os.Exit(1)
	}

	server := server.NewServer(initialization)
	err = server.Start()
	if err != nil {
		cancel() // Stop the queue listener
		logrus.Errorf("server exited with error: %v", err)
		os.Exit(1)

	}

	cancel() // Stop the queue listener on graceful shutdown
}
