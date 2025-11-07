package initialization

import (
	"net/http"

	"github.com/mini-maxit/backend/internal/api/http/routes"
	"github.com/mini-maxit/backend/internal/api/queue"
	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/filestorage"
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

	// Services
	filestorage, err := filestorage.NewFileStorageService(cfg.FileStorageURL)
	if err != nil {
		log.Panicf("Failed to create file storage service: %s", err.Error())
	}
	userService := service.NewUserService(userRepository)
	taskService := service.NewTaskService(
		filestorage,
		fileRepository,
		taskRepository,
		testCaseRepository,
		userRepository,
		groupRepository,
	)
	queueService := service.NewQueueService(
		taskRepository,
		submissionRepository,
		submissionResultRepository,
		queueRepository,
		cfg.Broker.QueueName,
		cfg.Broker.ResponseQueueName,
	)
	jwtService := service.NewJWTService(userRepository, cfg.JWTSecretKey)
	authService := service.NewAuthService(userRepository, jwtService)
	contestService := service.NewContestService(contestRepository, userRepository, submissionRepository, taskRepository, taskService)
	groupService := service.NewGroupService(groupRepository, userRepository, userService)
	langService := service.NewLanguageService(langRepository)
	submissionService := service.NewSubmissionService(
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
	submissionRoute := routes.NewSubmissionRoutes(submissionService, cfg.FileStorageURL, queueService, taskService)
	taskRoute := routes.NewTaskRoute(taskService)
	tasksManagementRoute := routes.NewTasksManagementRoute(taskService)
	userRoute := routes.NewUserRoute(userService)
	workerRoute := routes.NewWorkerRoute(workerService)

	// Queue listener - always created, manages its own connection internally
	queueListener := queue.NewListener(
		db,
		taskService,
		queueService,
		submissionService,
		langService,
		cfg.Broker,
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
		UserRoute:              userRoute,
		WorkerRoute:            workerRoute,

		QueueListener: queueListener,
	}
}
