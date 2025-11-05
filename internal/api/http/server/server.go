package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
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
	baseMux := mux.NewRouter()
	apiPrefix := "/api/" + APIVersion

	// Auth routes
	authMux := mux.NewRouter()
	routes.RegisterAuthRoute(authMux, init.AuthRoute)

	// Task routes
	taskMux := mux.NewRouter()
	routes.RegisterTaskRoutes(taskMux, init.TaskRoute)

	tasksManagementMux := mux.NewRouter()
	routes.RegisterTasksManagementRoutes(tasksManagementMux, init.TaskManagementRoute)

	// User routes
	userMux := mux.NewRouter()
	routes.RegisterUserRoutes(userMux, init.UserRoute)

	// Submission routes
	subbmissionMux := mux.NewRouter()
	routes.RegisterSubmissionRoutes(subbmissionMux, init.SubmissionRoute)

	// Group routes
	groupMux := mux.NewRouter()
	routes.RegisterGroupRoutes(groupMux, init.GroupRoute)

	// Contest routes
	contestMux := mux.NewRouter()
	routes.RegisterContestRoutes(contestMux, init.ContestRoute)

	contestManagementMux := mux.NewRouter()
	routes.RegistertContestsManagementRoute(contestManagementMux, init.ContestManagementRoute)

	// Worker routes
	workerMux := mux.NewRouter()
	routes.RegisterWorkerRoutes(workerMux, init.WorkerRoute)

	// Secure routes (require authentication with JWT)
	secureMux := mux.NewRouter()
	secureMux.PathPrefix("/tasks-management/").Handler(http.StripPrefix("/tasks-management", tasksManagementMux))
	secureMux.PathPrefix("/tasks/").Handler(taskMux)
	secureMux.PathPrefix("/submissions/").Handler(http.StripPrefix("/submissions", subbmissionMux))
	secureMux.PathPrefix("/users/").Handler(http.StripPrefix("/users", userMux))
	secureMux.PathPrefix("/groups/").Handler(http.StripPrefix("/groups", groupMux))
	secureMux.PathPrefix("/contests-management/").Handler(http.StripPrefix("/contests-management", contestManagementMux))
	secureMux.PathPrefix("/contests").Handler(contestMux)
	secureMux.PathPrefix("/workers/").Handler(http.StripPrefix("/workers", workerMux))

	// API routes
	apiMux := mux.NewRouter()
	apiMux.PathPrefix("/auth/").Handler(http.StripPrefix("/auth", authMux))
	apiMux.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(http.Dir("docs"))))
	apiMux.PathPrefix("/").Handler(middleware.JWTValidationMiddleware(secureMux, init.DB, init.JWTService))

	// Logging middleware
	httpLoger := utils.NewNamedLogger("http")
	loggingMux := mux.NewRouter()
	loggingMux.PathPrefix("/").Handler(middleware.LoggingMiddleware(apiMux, httpLoger))
	// Add the API prefix to all routes
	httpLoger.Infof("Query params middleware")
	baseMux.PathPrefix(apiPrefix + "/").Handler(http.StripPrefix(
		apiPrefix, middleware.QueryParamsMiddleware(
			middleware.DatabaseMiddleware(
				middleware.RecoveryMiddleware(loggingMux, log), init.DB,
			),
		),
	))

	baseMux.NotFoundHandler = http.HandlerFunc(httputils.NotFoundHandler)
	return &Server{mux: baseMux, port: init.Cfg.API.Port, logger: log}
}
