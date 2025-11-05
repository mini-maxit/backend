package initialization

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mini-maxit/backend/internal/api/http/routes"
	"github.com/mini-maxit/backend/internal/api/queue"
	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/filestorage"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"

	amqp "github.com/rabbitmq/amqp091-go"
)

const numTries = 10

type Initialization struct {
	Cfg *config.Config
	DB  database.Database

	TaskService service.TaskService
	JWTService  service.JWTService

	AuthRoute              routes.AuthRoute
	ContestRoute           routes.ContestRoute
	ContestManagementRoute routes.ContestsManagementRoute
	GroupRoute             routes.GroupRoute
	SubmissionRoute        routes.SubmissionRoutes
	TaskRoute              routes.TaskRoute
	UserRoute              routes.UserRoute
	WorkerRoute            routes.WorkerRoute

	QueueListener queue.Listener

	Dump func(w http.ResponseWriter, r *http.Request)
}

func connectToBroker(cfg *config.Config) (*amqp.Connection, *amqp.Channel) {
	log := utils.NewNamedLogger("connect_to_broker")

	var err error
	var conn *amqp.Connection
	brokerURL := fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.Broker.User, cfg.Broker.Password, cfg.Broker.Host, cfg.Broker.Port)
	for v := range numTries {
		conn, err = amqp.Dial(brokerURL)
		if err != nil {
			log.Warnf("Failed to connect to RabbitMQ: %s", err.Error())
			time.Sleep(2 * time.Second * time.Duration(v))
			continue
		}
	}
	log.Info("Connected to RabbitMQ")

	if err != nil {
		log.Panicf("Failed to connect to RabbitMQ: %s", err.Error())
	}
	channel, err := conn.Channel()
	if err != nil {
		log.Panicf("Failed to create channel: %s", err.Error())
	}

	return conn, channel
}

func NewInitialization(cfg *config.Config) *Initialization {
	log := utils.NewNamedLogger("initialization")
	conn, channel := connectToBroker(cfg)
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
	queueService, err := service.NewQueueService(taskRepository,
		submissionRepository,
		submissionResultRepository,
		queueRepository,
		conn,
		channel,
		cfg.Broker.QueueName,
		cfg.Broker.ResponseQueueName,
	)
	if err != nil {
		log.Panicf("Failed to create queue service: %s", err.Error())
	}
	err = queueService.PublishHandshake()
	if err != nil {
		log.Panicf("Failed to publish handshake: %s", err.Error())
	}
	jwtService := service.NewJWTService(userRepository, cfg.JWTSecretKey)
	authService := service.NewAuthService(userRepository, jwtService)
	contestService := service.NewContestService(contestRepository, userRepository, submissionRepository, taskRepository)
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
	workerService := service.NewWorkerService(queueService)

	// Routes
	authRoute := routes.NewAuthRoute(userService, authService, cfg.API.RefreshTokenPath)
	contestRoute := routes.NewContestRoute(contestService, submissionService)
	contestManagementRoute := routes.NewContestsManagementRoute(contestService, submissionService)
	groupRoute := routes.NewGroupRoute(groupService)
	submissionRoute := routes.NewSubmissionRoutes(submissionService, cfg.FileStorageURL, queueService, taskService)
	taskRoute := routes.NewTaskRoute(cfg.FileStorageURL, taskService)
	userRoute := routes.NewUserRoute(userService, contestService)
	workerRoute := routes.NewWorkerRoute(workerService)

	// Queue listener
	queueListener, err := queue.NewListener(
		conn,
		channel,
		db,
		taskService,
		queueService,
		submissionService,
		langService,
		cfg.Broker.ResponseQueueName,
	)
	if err != nil {
		log.Panicf("Failed to create queue listener: %s", err.Error())
	}
	if cfg.Dump {
		dump(db, log, authService, userRepository)
	}
	return &Initialization{
		Cfg: cfg,
		DB:  db,

		TaskService: taskService,
		JWTService:  jwtService,

		AuthRoute:              authRoute,
		ContestRoute:           contestRoute,
		ContestManagementRoute: contestManagementRoute,
		GroupRoute:             groupRoute,
		SubmissionRoute:        submissionRoute,
		TaskRoute:              taskRoute,
		UserRoute:              userRoute,
		WorkerRoute:            workerRoute,

		QueueListener: queueListener,
	}
}
