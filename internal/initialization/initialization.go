package initialization

import (
	"context"
	"net/http"

	"github.com/mini-maxit/backend/internal/api/http/routes"
	"github.com/mini-maxit/backend/internal/api/queue"
	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/filestorage"
	pkgqueue "github.com/mini-maxit/backend/package/queue"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
)

type Initialization struct {
	Cfg *config.Config
	DB  database.Database

	TaskService  service.TaskService
	JWTService   service.JWTService
	QueueService service.QueueService

	AuthRoute              routes.AuthRoute
	ContestRoute           routes.ContestRoute
	ContestManagementRoute routes.ContestsManagementRoute
	GroupRoute             routes.GroupRoute
	SubmissionRoute        routes.SubmissionRoutes
	TaskRoute              routes.TaskRoute
	TaskManagementRoute    routes.TasksManagementRoute
	AccessControlRoute     routes.AccessControlRoute
	UserRoute              routes.UserRoute
	WorkerRoute            routes.WorkerRoute

	QueueListener queue.Listener

	Dump func(w http.ResponseWriter, r *http.Request)
}

func NewInitialization(cfg *config.Config) *Initialization {
	log := utils.NewNamedLogger("initialization")

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Panicf("Failed to connect to database: %s", err.Error())
	}

	// Repositories
	langRepository := repository.NewLanguageRepository()
	userRepository := repository.NewUserRepository()
	fileRepository := repository.NewFileRepository()
	taskRepository := repository.NewTaskRepository()
	groupRepository := repository.NewGroupRepository()
	submissionRepository := repository.NewSubmissionRepository()
	testCaseRepository := repository.NewTestCaseRepository()
	testResultRepository := repository.NewTestResultRepository()
	queueRepository := repository.NewQueueMessageRepository()
	submissionResultRepository := repository.NewSubmissionResultRepository()
	contestRepository := repository.NewContestRepository()
	accessControlRepository := repository.NewAccessControlRepository()

	// Services
	filestorage, err := filestorage.NewFileStorageService(cfg.FileStorageURL)
	if err != nil {
		log.Panicf("Failed to create file storage service: %s", err.Error())
	}

	// Create queue client (shared by both service and listener)
	queueClient := pkgqueue.NewClient(
		pkgqueue.Config{
			Host:              cfg.Broker.Host,
			Port:              cfg.Broker.Port,
			User:              cfg.Broker.User,
			Password:          cfg.Broker.Password,
			WorkerQueueName:   cfg.Broker.QueueName,
			ResponseQueueName: cfg.Broker.ResponseQueueName,
		},
		utils.NewNamedLogger("queue_client"),
	)

	// Start queue client
	if err := queueClient.Start(context.Background()); err != nil {
		log.Warnf("Failed to start queue client: %v", err)
	}

	queueService := service.NewQueueService(
		taskRepository,
		submissionRepository,
		submissionResultRepository,
		queueRepository,
		queueClient,
		cfg.Broker.QueueName,
		cfg.Broker.ResponseQueueName,
	)
	jwtService := service.NewJWTService(userRepository, cfg.JWTSecretKey)
	authService := service.NewAuthService(userRepository, jwtService)

	// Create AccessControlService first
	accessControlService := service.NewAccessControlService(accessControlRepository, userRepository, taskRepository, contestRepository)

	// Create TaskService (needs AccessControlService)
	taskService := service.NewTaskService(
		filestorage,
		fileRepository,
		taskRepository,
		testCaseRepository,
		userRepository,
		groupRepository,
		submissionRepository,
		contestRepository,
		accessControlService,
	)

	contestService := service.NewContestService(contestRepository, userRepository, submissionRepository, taskRepository, accessControlService, taskService)
	userService := service.NewUserService(userRepository, contestService)
	groupService := service.NewGroupService(groupRepository, userRepository, userService)
	langService := service.NewLanguageService(langRepository)
	submissionService := service.NewSubmissionService(
		accessControlService,
		contestService,
		filestorage,
		fileRepository,
		submissionRepository,
		submissionResultRepository,
		testCaseRepository,
		testResultRepository,
		groupRepository,
		taskRepository,
		langService,
		taskService,
		userService,
		queueService,
	)
	workerService := service.NewWorkerService(queueService, submissionRepository, db.DB())

	// Routes
	authRoute := routes.NewAuthRoute(userService, authService, cfg.API.RefreshTokenPath)
	contestRoute := routes.NewContestRoute(contestService, submissionService)
	contestManagementRoute := routes.NewContestsManagementRoute(contestService, submissionService)
	groupRoute := routes.NewGroupRoute(groupService)
	submissionRoute := routes.NewSubmissionRoutes(submissionService, queueService, taskService)
	taskRoute := routes.NewTaskRoute(taskService)
	tasksManagementRoute := routes.NewTasksManagementRoute(taskService)
	accessControlRoute := routes.NewAccessControlRoute(accessControlService)
	userRoute := routes.NewUserRoute(userService)
	workerRoute := routes.NewWorkerRoute(workerService)

	// Queue listener - uses the same queue client as queue service
	queueListener := queue.NewListener(
		db,
		taskService,
		queueService,
		submissionService,
		langService,
		queueClient,
		cfg.Broker.ResponseQueueName,
	)

	if cfg.Dump {
		dump(db, log, authService, userRepository)
	}

	return &Initialization{
		Cfg: cfg,
		DB:  db,

		TaskService:  taskService,
		JWTService:   jwtService,
		QueueService: queueService,

		AuthRoute:              authRoute,
		ContestRoute:           contestRoute,
		ContestManagementRoute: contestManagementRoute,
		GroupRoute:             groupRoute,
		SubmissionRoute:        submissionRoute,
		TaskRoute:              taskRoute,
		TaskManagementRoute:    tasksManagementRoute,
		AccessControlRoute:     accessControlRoute,
		UserRoute:              userRoute,
		WorkerRoute:            workerRoute,

		QueueListener: queueListener,
	}
}
