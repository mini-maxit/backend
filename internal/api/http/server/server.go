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
	"github.com/mini-maxit/backend/internal/logger"
	"go.uber.org/zap"
)

const ApiVersion = "v1"

type Server struct {
	mux    http.Handler
	port   uint16
	logger *zap.SugaredLogger
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
		s.logger.Info("Shutting down server...")

		// Create a context with timeout to allow graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Error shutting down server:", err.Error())
		}
	}()

	s.logger.Infof("Starting server on port %d", s.port)
	return http.ListenAndServe(server.Addr, server.Handler)
}

func NewServer(initialization *initialization.Initialization, log *zap.SugaredLogger) *Server {
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

	// User routes
	userMux := http.NewServeMux()
	userMux.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			initialization.UserRoute.GetUserById(w, r)
		} else if r.Method == http.MethodPut {
			initialization.UserRoute.EditUser(w, r)
		}
	},
	)
	userMux.HandleFunc("/", initialization.UserRoute.GetAllUsers)
	userMux.HandleFunc("/email", initialization.UserRoute.GetUserByEmail)
	userMux.HandleFunc("/{id}/task", initialization.TaskRoute.GetAllForUser)

	// Group routes
	groupMux := http.NewServeMux()
	groupMux.HandleFunc("/{id}/task", initialization.TaskRoute.GetAllForGroup)

	// Session routes
	sessionMux := http.NewServeMux()
	sessionMux.HandleFunc("/", initialization.SessionRoute.CreateSession)
	sessionMux.HandleFunc("/validate", initialization.SessionRoute.ValidateSession)
	sessionMux.HandleFunc("/invalidate", initialization.SessionRoute.InvalidateSession)

	// Secure routes (require authentication)
	secureMux := http.NewServeMux()
	secureMux.Handle("/task/", http.StripPrefix("/task", taskMux))
	secureMux.Handle("/session/", http.StripPrefix("/session", sessionMux))
	secureMux.Handle("/user/", http.StripPrefix("/user", userMux))
	secureMux.Handle("/group/", http.StripPrefix("/group", groupMux))

	// API routes
	apiMux := http.NewServeMux()
	apiMux.Handle("/auth/", http.StripPrefix("/auth", authMux))
	apiMux.Handle("/", middleware.SessionValidationMiddleware(secureMux, initialization.Db, initialization.SessionService))
	apiMux.Handle("/docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir("docs"))))

	// Logging middleware
	httpLoger := logger.NewHttpLogger()
	loggingMux := http.NewServeMux()
	loggingMux.Handle("/", middleware.LoggingMiddleware(apiMux, httpLoger))
	// Add the API prefix to all routes
	mux.Handle(apiPrefix+"/", http.StripPrefix(apiPrefix, middleware.RecoveryMiddleware(middleware.DatabaseMiddleware(loggingMux, initialization.Db), log)))
	return &Server{mux: mux, port: initialization.Cfg.App.Port, logger: log}
}
