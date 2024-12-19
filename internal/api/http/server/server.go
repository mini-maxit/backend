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
	"github.com/mini-maxit/backend/internal/api/http/middleware"
	"github.com/sirupsen/logrus"
)

const ApiVersion = "v1"

type Server struct {
	mux  http.Handler
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
	apiPrefix := fmt.Sprintf("/api/%s", ApiVersion)

	// Auth routes
	authMux := http.NewServeMux()
	authMux.HandleFunc("/login", initialization.AuthRoute.Login)
	authMux.HandleFunc("/register", initialization.AuthRoute.Register)

	// Task routes
	taskMux := http.NewServeMux()
	taskMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			initialization.TaskRoute.UploadTask(w, r)
		} else if r.Method == "GET" {
			initialization.TaskRoute.GetAllTasks(w, r)
		}
	},
	)
	taskMux.HandleFunc("/{id}", initialization.TaskRoute.GetTask)
	taskMux.HandleFunc("/submit", initialization.TaskRoute.SubmitSolution)

	// Session routes
	sessionMux := http.NewServeMux()
	sessionMux.HandleFunc("/", initialization.SessionRoute.CreateSession)
	sessionMux.HandleFunc("/validate", initialization.SessionRoute.ValidateSession)
	sessionMux.HandleFunc("/invalidate", initialization.SessionRoute.InvalidateSession)

	// Secure routes (require authentication)
	secureMux := http.NewServeMux()
	secureMux.Handle("/task/", http.StripPrefix("/task", taskMux))
	secureMux.Handle("/session/", http.StripPrefix("/session", sessionMux))

	// API routes
	apiMux := http.NewServeMux()
	apiMux.Handle("/auth/", http.StripPrefix("/auth", authMux))
	apiMux.Handle("/", middleware.SessionValidationMiddleware(secureMux, initialization.SessionService))

	loggingMux := http.NewServeMux()
	loggingMux.Handle("/", middleware.LoggingMiddleware(apiMux))
	// Add the API prefix to all routes
	mux.Handle(apiPrefix+"/", http.StripPrefix(apiPrefix, loggingMux))

	return &Server{mux: mux, port: initialization.Cfg.App.Port}
}
