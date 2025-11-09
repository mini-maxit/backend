package main

import (
	"errors"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/mini-maxit/backend/internal/api/http/server"
	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/internal/initialization"
	"github.com/mini-maxit/backend/package/utils"
)

// @title		Mini-Maxit API
// @version	1.0.0
// @BasePath	/api/v1
func main() {
	log := utils.NewNamedLogger("server")

	if _, ok := os.LookupEnv("DEBUG"); ok {
		err := godotenv.Load("././.env")
		if err != nil {
			panic(err)
		}
	}
	cfg := config.NewConfig()

	init := initialization.NewInitialization(cfg)

	// Queue listener is always created and manages its own connection
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
		if !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("failed to start server: %v", err.Error())
		} else {
			log.Info("Server stopped gracefully")
		}
	}
}
