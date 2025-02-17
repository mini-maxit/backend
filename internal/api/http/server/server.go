package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mini-maxit/backend/internal/api/http/middleware"
	"github.com/mini-maxit/backend/internal/api/http/routes"
	"github.com/mini-maxit/backend/internal/initialization"
	"github.com/mini-maxit/backend/package/utils"
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

func NewServer(init *initialization.Initialization, log *zap.SugaredLogger) *Server {
	mux := http.NewServeMux()
	apiPrefix := fmt.Sprintf("/api/%s", ApiVersion)

	// Auth routes
	authMux := http.NewServeMux()
	authMux.HandleFunc("/login", init.AuthRoute.Login)
	authMux.HandleFunc("/register", init.AuthRoute.Register)

	// Task routes
	taskMux := http.NewServeMux()
	routes.RegisterTaskRoutes(taskMux, init.TaskRoute)

	// User routes
	userMux := http.NewServeMux()
	userMux.HandleFunc("/", init.UserRoute.GetAllUsers)
	userMux.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			init.UserRoute.GetUserById(w, r)
		} else if r.Method == http.MethodPut {
			init.UserRoute.EditUser(w, r)
		}
	},
	)
	userMux.HandleFunc("/user", init.UserRoute.GetAllUsers)
	userMux.HandleFunc("/email", init.UserRoute.GetUserByEmail)

	// Submission routes
	subbmissionMux := http.NewServeMux()
	subbmissionMux.HandleFunc("/", init.SubmissionRoute.GetAll)
	subbmissionMux.HandleFunc("/{id}", init.SubmissionRoute.GetById)
	subbmissionMux.HandleFunc("/user/{id}", init.SubmissionRoute.GetAllForUser)
	subbmissionMux.HandleFunc("/group/{id}", init.SubmissionRoute.GetAllForGroup)
	subbmissionMux.HandleFunc("/task/{id}", init.SubmissionRoute.GetAllForTask)
	subbmissionMux.HandleFunc("/submit", init.SubmissionRoute.SubmitSolution)
	subbmissionMux.HandleFunc("/languages", init.SubmissionRoute.GetAvailableLanguages)

	// Group routes
	groupMux := http.NewServeMux()
	routes.RegisterGroupRoutes(groupMux, init.GroupRoute)

	// Session routes
	sessionMux := http.NewServeMux()
	sessionMux.HandleFunc("/", init.SessionRoute.CreateSession)
	sessionMux.HandleFunc("/validate", init.SessionRoute.ValidateSession)
	sessionMux.HandleFunc("/invalidate", init.SessionRoute.InvalidateSession)

	// Secure routes (require authentication)
	secureMux := http.NewServeMux()
	secureMux.Handle("/task/", http.StripPrefix("/task", taskMux))
	secureMux.Handle("/session/", http.StripPrefix("/session", sessionMux))
	secureMux.Handle("/submission/", http.StripPrefix("/submission", subbmissionMux))
	secureMux.Handle("/user/", http.StripPrefix("/user", userMux))
	secureMux.Handle("/group/", http.StripPrefix("/group", groupMux))

	// API routes
	apiMux := http.NewServeMux()
	apiMux.Handle("/auth/", http.StripPrefix("/auth", authMux))
	apiMux.Handle("/", middleware.SessionValidationMiddleware(secureMux, init.Db, init.SessionService))
	apiMux.Handle("/docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir("docs"))))

	// Logging middleware
	httpLoger := utils.NewHttpLogger()
	loggingMux := http.NewServeMux()
	loggingMux.Handle("/", middleware.LoggingMiddleware(apiMux, httpLoger))
	// Add the API prefix to all routes
	mux.Handle(apiPrefix+"/", http.StripPrefix(apiPrefix, middleware.RecoveryMiddleware(middleware.DatabaseMiddleware(loggingMux, init.Db), log)))
	return &Server{mux: mux, port: init.Cfg.Api.Port, logger: log}
}
