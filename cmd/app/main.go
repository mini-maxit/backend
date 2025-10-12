package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/mini-maxit/backend/internal/api/http/server"
	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/internal/initialization"
	"github.com/mini-maxit/backend/package/utils"
)

// @title Mini-Maxit API
// @version 1.0.0
// @BasePath		/api/v1.
func main() {
	utils.InitializeLogger()
	log := utils.NewNamedLogger("server")

	if _, ok := os.LookupEnv("DEBUG"); ok {
		err := godotenv.Load("././.env")
		if err != nil {
			panic(err)
		}
	}
	cfg := config.NewConfig()

	init := initialization.NewInitialization(cfg)

	queueListener := init.QueueListener
	log.Info("Starting queue listener...")
	err := queueListener.Start()
	if err != nil {
		log.Errorf("failed to start queue listener: %v", err.Error())
		os.Exit(1)
	}

	server := server.NewServer(init, log)
	err = server.Start()
	if err != nil {
		err2 := queueListener.Shutdown()
		if err2 != nil {
			log.Errorf("failed to shutdown queue listener: %v", err2.Error())
		}
		log.Errorf("failed to start server: %v", err.Error())
		os.Exit(1)
	}

	err = queueListener.Shutdown()
	if err != nil {
		log.Errorf("failed to shutdown queue listener: %v", err.Error())
	}
}
