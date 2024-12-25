package initialization

import (
	"fmt"
	"time"

	"github.com/mini-maxit/backend/internal/api/http/routes"
	"github.com/mini-maxit/backend/internal/api/queue"
	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/internal/database"
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
	var err error
	var conn *amqp.Connection
	for v := range 5 {
		conn, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.BrokerConfig.User, cfg.BrokerConfig.Password, cfg.BrokerConfig.Host, cfg.BrokerConfig.Port))
		if err != nil {
			logrus.Printf("Failed to connect to RabbitMQ: %v\nRetrying...", err)
			time.Sleep(2*time.Second + time.Duration(v))
			continue
		}
	}
	logrus.Printf("Connected to RabbitMQ")

	if err != nil {
		panic(fmt.Errorf("failed to connect to RabbitMQ: %w", err))
	}
	channel, err := conn.Channel()
	if err != nil {
		panic(fmt.Errorf("failed to create channel: %w", err))
	}

	return conn, channel
}

func NewInitialization(cfg *config.Config) *Initialization {
	conn, channel := connectToBroker(cfg)
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}
	tx, err := db.Connect()

	defer utils.TransactionPanicRecover(tx)

	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}
	// Repositories
	_, err = repository.NewLanguageRepository(tx)
	if err != nil {
		panic(fmt.Errorf("failed to create language repository: %w", err))
	}
	userRepository, err := repository.NewUserRepository(tx)
	if err != nil {
		panic(fmt.Errorf("failed to create user repository: %w", err))
	}
	taskRepository, err := repository.NewTaskRepository(tx)
	if err != nil {
		panic(fmt.Errorf("failed to create task repository: %w", err))
	}
	_, err = repository.NewGroupRepository(tx)
	if err != nil {
		panic(fmt.Errorf("failed to create group repository: %w", err))
	}
	submissionRepository, err := repository.NewSubmissionRepository(tx)
	if err != nil {
		panic(fmt.Errorf("failed to create submission repository: %w", err))
	}
	_, err = repository.NewSubmissionResultRepository(tx)
	if err != nil {
		panic(fmt.Errorf("failed to create submission result repository: %w", err))
	}
	queueRepository, err := repository.NewQueueMessageRepository(tx)
	if err != nil {
		panic(fmt.Errorf("failed to create queue repository: %w", err))
	}
	sessionRepository, err := repository.NewSessionRepository(tx)
	if err != nil {
		panic(fmt.Errorf("failed to create session repository: %w", err))
	}

	if err := db.Commit(); err != nil {
		panic(fmt.Errorf("failed to commit transaction: %v", err))
	}

	// Services
	userService := service.NewUserService(userRepository)
	taskService := service.NewTaskService(cfg, taskRepository, submissionRepository)
	queueService, err := service.NewQueueService(taskRepository, submissionRepository, queueRepository, conn, channel, cfg.BrokerConfig.QueueName, cfg.BrokerConfig.ResponseQueueName)
	if err != nil {
		panic(fmt.Errorf("failed to create queue service: %w", err))
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
		panic(fmt.Errorf("failed to create queue listener: %w", err))
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
