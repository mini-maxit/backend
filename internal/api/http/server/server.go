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

const APIVersion = "v1"
const shudownTimeout = 5
const ReadHeaderTimeout = 3

type Server struct {
	mux    http.Handler
	port   uint16
	logger *zap.SugaredLogger
}

func (s *Server) Start() error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, os.Interrupt)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", s.port),
		Handler:           s.mux,
		ReadHeaderTimeout: ReadHeaderTimeout * time.Second,
	}
	ctx := context.Background()
	go func() {
		<-sigChan
		s.logger.Info("Shutting down server...")

		// Create a context with timeout to allow graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shudownTimeout*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Error shutting down server:", err.Error())
		}
	}()

	s.logger.Infof("Starting server on port %d", s.port)
	return server.ListenAndServe()
}

func NewServer(init *initialization.Initialization, log *zap.SugaredLogger) *Server {
	mux := http.NewServeMux()
	apiPrefix := "/api/" + APIVersion

	// Auth routes
	authMux := http.NewServeMux()
	authMux.HandleFunc("/login", init.AuthRoute.Login)
	authMux.HandleFunc("/register", init.AuthRoute.Register)
	authMux.HandleFunc("/refresh", init.AuthRoute.RefreshToken)
	authMux.HandleFunc("/validate", init.AuthRoute.Validate)

	// Task routes
	taskMux := http.NewServeMux()
	routes.RegisterTaskRoutes(taskMux, init.TaskRoute)

	// User routes
	userMux := http.NewServeMux()
	routes.RegisterUserRoutes(userMux, init.UserRoute)

	// Submission routes
	subbmissionMux := http.NewServeMux()
	routes.RegisterSubmissionRoutes(subbmissionMux, init.SubmissionRoute)

	// Group routes
	groupMux := http.NewServeMux()
	routes.RegisterGroupRoutes(groupMux, init.GroupRoute)

	// Worker routes
	workerMux := http.NewServeMux()
	routes.RegisterWorkerRoutes(workerMux, init.WorkerRoute)

	// Secure routes (require authentication with JWT)
	secureMux := http.NewServeMux()
	secureMux.Handle("/task/", http.StripPrefix("/task", taskMux))
	secureMux.Handle("/submission/", http.StripPrefix("/submission", subbmissionMux))
	secureMux.Handle("/user/", http.StripPrefix("/user", userMux))
	secureMux.Handle("/group/", http.StripPrefix("/group", groupMux))
	secureMux.Handle("/worker/", http.StripPrefix("/worker", workerMux))

	// API routes
	apiMux := http.NewServeMux()
	apiMux.Handle("/auth/", http.StripPrefix("/auth", authMux))
	apiMux.Handle("/", middleware.JWTValidationMiddleware(secureMux, init.DB, init.JWTService))
	apiMux.Handle("/docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir("docs"))))

	// Logging middleware
	httpLoger := utils.NewHTTPLogger()
	loggingMux := http.NewServeMux()
	loggingMux.Handle("/", middleware.LoggingMiddleware(apiMux, httpLoger))
	// Add the API prefix to all routes
	httpLoger.Infof("Query params middleware")
	mux.Handle(apiPrefix+"/", http.StripPrefix(
		apiPrefix, middleware.RecoveryMiddleware(
			middleware.QueryParamsMiddleware(
				middleware.DatabaseMiddleware(loggingMux, init.DB)), log,
		),
	))
	return &Server{mux: mux, port: init.Cfg.API.Port, logger: log}
}
