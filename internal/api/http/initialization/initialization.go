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
	"github.com/mini-maxit/backend/package/utils"
	"github.com/sirupsen/logrus"

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
			logger.Log(&broker_logger, "Failed to connect to RabbitMQ:", err.Error(), logger.Warn)
			time.Sleep(2*time.Second + time.Duration(v))
			continue
		}
	}
	logrus.Printf("Connected to RabbitMQ")

	if err != nil {
		logger.Log(&broker_logger, "Failed to connect to RabbitMQ:", err.Error(), logger.Panic)
	}
	channel, err := conn.Channel()
	if err != nil {
		logger.Log(&broker_logger, "Failed to create channel:", err.Error(), logger.Panic)
	}

	return conn, channel
}

func NewInitialization(cfg *config.Config) *Initialization {
	init_logger := logger.NewNamedLogger("initialization")
	conn, channel := connectToBroker(cfg)
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		logger.Log(&init_logger, "Failed to connect to database:", err.Error(), logger.Panic)
	}
	tx, err := db.Connect()

	defer utils.TransactionPanicRecover(tx)

	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}
	// Repositories
	_, err = repository.NewLanguageRepository(tx)
	if err != nil {
		logger.Log(&init_logger, "Failed to create language repository:", err.Error(), logger.Panic)
	}
	userRepository, err := repository.NewUserRepository(tx)
	if err != nil {
		logger.Log(&init_logger, "Failed to create user repository:", err.Error(), logger.Panic)
	}
	taskRepository, err := repository.NewTaskRepository(tx)
	if err != nil {
		logger.Log(&init_logger, "Failed to create task repository:", err.Error(), logger.Panic)
	}
	_, err = repository.NewGroupRepository(tx)
	if err != nil {
		logger.Log(&init_logger, "Failed to create group repository:", err.Error(), logger.Panic)
	}
	submissionRepository, err := repository.NewSubmissionRepository(tx)
	if err != nil {
		logger.Log(&init_logger, "Failed to create submission repository:", err.Error(), logger.Panic)
	}
	_, err = repository.NewSubmissionResultRepository(tx)
	if err != nil {
		logger.Log(&init_logger, "Failed to create submission result repository:", err.Error(), logger.Panic)
	}
	queueRepository, err := repository.NewQueueMessageRepository(tx)
	if err != nil {
		logger.Log(&init_logger, "Failed to create queue repository:", err.Error(), logger.Panic)
	}
	sessionRepository, err := repository.NewSessionRepository(tx)
	if err != nil {
		logger.Log(&init_logger, "Failed to create session repository:", err.Error(), logger.Panic)
	}

	if err := db.Commit(); err != nil {
		panic(fmt.Errorf("failed to commit transaction: %v", err))
	}

	// Services
	userService := service.NewUserService(userRepository)
	taskService := service.NewTaskService(cfg, taskRepository, submissionRepository)
	queueService, err := service.NewQueueService(taskRepository, submissionRepository, queueRepository, conn, channel, cfg.BrokerConfig.QueueName, cfg.BrokerConfig.ResponseQueueName)
	if err != nil {
		logger.Log(&init_logger, "Failed to create queue service:", err.Error(), logger.Panic)
	}
	sessionService := service.NewSessionService(sessionRepository, userRepository)
	authService := service.NewAuthService(userRepository, sessionService)

	// Routes
	taskRoute := routes.NewTaskRoute(cfg.FileStorageUrl, taskService, queueService)
	sessionRoute := routes.NewSessionRoute(sessionService)
	authRoute := routes.NewAuthRoute(userService, authService)
	swaggerRoute := routes.NewSwaggerRoute()
	userRoute := routes.NewUserRoute(userService)

	// Queue listener
	queueListener, err := queue.NewQueueListener(conn, channel, taskService, cfg.BrokerConfig.ResponseQueueName)
	if err != nil {
		logger.Log(&init_logger, "Failed to create queue listener:", err.Error(), logger.Panic)
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
