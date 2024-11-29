package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mini-maxit/backend/internal/api/http/initialization"
	"github.com/sirupsen/logrus"
)

type Server struct {
	mux  *http.ServeMux
	port uint16
}

func (s *Server) Start() error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, os.Interrupt)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.mux,
	}
	ctx := context.Background()
	go func() {
		<-sigChan
		logrus.Info("Shutting down server...")

		// Create a context with timeout to allow graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logrus.Errorf("Server forced to shutdown: %v", err)
		}
	}()

	logrus.Info("Starting server on port ", s.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.mux)
}

func NewServer(initialization *initialization.Initialization) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Swagger route
	mux.HandleFunc("/api/v1/swagger", initialization.SwaggerRoute.Docs)

	// Auth routes
	mux.HandleFunc("/api/v1/auth/login", initialization.AuthRoute.Login)
	mux.HandleFunc("/api/v1/auth/register", initialization.AuthRoute.Register)

	// Task routes
	mux.HandleFunc("/api/v1/task", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			initialization.TaskRoute.UploadTask(w, r)
		} else if r.Method == "GET" {
			initialization.TaskRoute.GetAllTasks(w, r)
		}
	},
	)
	mux.HandleFunc("/api/v1/task/{id}", initialization.TaskRoute.GetTask)
	mux.HandleFunc("/api/v1/task/submit", initialization.TaskRoute.SubmitSolution)

	// Session routes
	mux.HandleFunc("/api/v1/session", initialization.SessionRoute.CreateSession)
	mux.HandleFunc("/api/v1/session/validate", initialization.SessionRoute.ValidateSession)
	mux.HandleFunc("/api/v1/session/invalidate", initialization.SessionRoute.InvalidateSession)

	return &Server{mux: mux, port: initialization.Cfg.App.Port}
}
