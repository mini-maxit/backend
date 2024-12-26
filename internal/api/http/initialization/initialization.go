package initialization

import (
	"fmt"
	"time"

	"github.com/mini-maxit/backend/internal/api/http/routes"
	"github.com/mini-maxit/backend/internal/api/queue"
	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/internal/logger"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/service"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Initialization struct {
	Cfg *config.Config
	Db  database.Database

	TaskService    service.TaskService
	SessionService service.SessionService

	AuthRoute    routes.AuthRoute
	TaskRoute    routes.TaskRoute
	SessionRoute routes.SessionRoute
	SwaggerRoute routes.SwaggerRoute
	UserRoute    routes.UserRoute

	QueueListener queue.QueueListener
}

func connectToBroker(cfg *config.Config) (*amqp.Connection, *amqp.Channel) {
	broker_logger := logger.NewNamedLogger("connect_to_broker")

	var err error
	var conn *amqp.Connection
	for v := range 5 {
		conn, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.BrokerConfig.User, cfg.BrokerConfig.Password, cfg.BrokerConfig.Host, cfg.BrokerConfig.Port))
		if err != nil {
			broker_logger.Warnf("Failed to connect to RabbitMQ: %s", err.Error())
			time.Sleep(2*time.Second + time.Duration(v))
			continue
		}
	}

	if err != nil {
		broker_logger.Panicf("Failed to connect to RabbitMQ: %s", err.Error())
	}
	channel, err := conn.Channel()
	if err != nil {
		broker_logger.Panicf("Failed to create channel: %s", err.Error())
	}

	return conn, channel
}

func NewInitialization(cfg *config.Config) *Initialization {
	init_logger := logger.NewNamedLogger("initialization")
	conn, channel := connectToBroker(cfg)
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		init_logger.Panicf("Failed to connect to database: %s", err.Error())
	}

	// Repositories
	_, err = repository.NewLanguageRepository(db.Connect())
	if err != nil {
		init_logger.Panicf("Failed to create language repository: %s", err.Error())
	}
	userRepository, err := repository.NewUserRepository(db.Connect())
	if err != nil {
		init_logger.Panicf("Failed to create user repository: %s", err.Error())
	}
	taskRepository, err := repository.NewTaskRepository(db.Connect())
	if err != nil {
		init_logger.Panicf("Failed to create task repository: %s", err.Error())
	}
	_, err = repository.NewGroupRepository(db.Connect())
	if err != nil {
		init_logger.Panicf("Failed to create group repository: %s", err.Error())
	}
	submissionRepository, err := repository.NewSubmissionRepository(db.Connect())
	if err != nil {
		init_logger.Panicf("Failed to create submission repository: %s", err.Error())
	}
	_, err = repository.NewSubmissionResultRepository(db.Connect())
	if err != nil {
		init_logger.Panicf("Failed to create submission result repository: %s", err.Error())
	}
	queueRepository, err := repository.NewQueueMessageRepository(db.Connect())
	if err != nil {
		init_logger.Panicf("Failed to create queue repository: %s", err.Error())
	}
	sessionRepository, err := repository.NewSessionRepository(db.Connect())
	if err != nil {
		init_logger.Panicf("Failed to create session repository: %s", err.Error())
	}

	// Services
	userService := service.NewUserService(db, userRepository)
	taskService := service.NewTaskService(db, cfg, taskRepository, submissionRepository)
	queueService, err := service.NewQueueService(db, taskRepository, submissionRepository, queueRepository, conn, channel, cfg.BrokerConfig.QueueName, cfg.BrokerConfig.ResponseQueueName)
	if err != nil {
		init_logger.Panicf("Failed to create queue service: %s", err.Error())
	}
	sessionService := service.NewSessionService(db, sessionRepository, userRepository)
	authService := service.NewAuthService(db, userRepository, sessionService)

	// Routes
	taskRoute := routes.NewTaskRoute(cfg.FileStorageUrl, taskService, queueService)
	sessionRoute := routes.NewSessionRoute(sessionService)
	authRoute := routes.NewAuthRoute(userService, authService)
	swaggerRoute := routes.NewSwaggerRoute()
	userRoute := routes.NewUserRoute(userService)

	// Queue listener
	queueListener, err := queue.NewQueueListener(conn, channel, taskService, cfg.BrokerConfig.ResponseQueueName)
	if err != nil {
		init_logger.Panicf("Failed to create queue listener: %s", err.Error())
	}


	return &Initialization{Cfg: cfg,
		QueueListener:  queueListener,
		Db:             db,
		TaskService:    taskService,
		SessionService: sessionService,
		AuthRoute:      authRoute,
		SessionRoute:   sessionRoute,
		SwaggerRoute:   swaggerRoute,
		TaskRoute:      taskRoute,
		UserRoute:      userRoute}
}
