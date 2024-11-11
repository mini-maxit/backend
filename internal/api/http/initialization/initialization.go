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

	amqp "github.com/rabbitmq/amqp091-go"
)

type Initialization struct {
	Cfg *config.Config
	Db  database.Database

	TaskService service.TaskService

	TaskRoute    routes.TaskRoute
	SessionRoute routes.SessionRoute

	QueueListener queue.QueueListener
}

func connectToBroker(cfg *config.Config) (*amqp.Connection, *amqp.Channel) {
	var err error
	var conn *amqp.Connection
	for v := range 5 {
		conn, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.BrokerConfig.User, cfg.BrokerConfig.Password, cfg.BrokerConfig.Host, cfg.BrokerConfig.Port))
		if err != nil {
			fmt.Printf("Failed to connect to RabbitMQ: %v\n", err)
			time.Sleep(2*time.Second + time.Duration(v))
			continue
		}
	}

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

	// Repositories
	_, err = repository.NewLanguageRepository(db.Connect())
	if err != nil {
		panic(fmt.Errorf("failed to create language repository: %w", err))
	}
	userRepository, err := repository.NewUserRepository(db.Connect())
	if err != nil {
		panic(fmt.Errorf("failed to create user repository: %w", err))
	}
	taskRepository, err := repository.NewTaskRepository(db.Connect())
	if err != nil {
		panic(fmt.Errorf("failed to create task repository: %w", err))
	}
	_, err = repository.NewGroupRepository(db.Connect())
	if err != nil {
		panic(fmt.Errorf("failed to create group repository: %w", err))
	}
	submissionRepository, err := repository.NewSubmissionRepository(db.Connect())
	if err != nil {
		panic(fmt.Errorf("failed to create submission repository: %w", err))
	}
	_, err = repository.NewSubmissionResultRepository(db.Connect())
	if err != nil {
		panic(fmt.Errorf("failed to create submission result repository: %w", err))
	}
	queueRepository, err := repository.NewQueueMessageRepository(db.Connect())
	if err != nil {
		panic(fmt.Errorf("failed to create queue repository: %w", err))
	}
	sessionRepository, err := repository.NewSessionRepository(db.Connect())
	if err != nil {
		panic(fmt.Errorf("failed to create session repository: %w", err))
	}

	// Services
	taskService := service.NewTaskService(db, taskRepository, submissionRepository)
	queueService, err := service.NewQueueService(db, taskRepository, submissionRepository, queueRepository, conn, channel, cfg.BrokerConfig.QueueName, cfg.BrokerConfig.ResponseQueueName)
	if err != nil {
		panic(fmt.Errorf("failed to create queue service: %w", err))
	}
	sessionService := service.NewSessionService(db, sessionRepository, userRepository)

	// Routes
	taskRoute := routes.NewTaskRoute(cfg.FileStorageUrl, taskService, queueService)
	sessionRoute := routes.NewSessionRoute(sessionService)

	// Queue listener
	queueListener, err := queue.NewQueueListener(conn, channel, taskService, cfg.BrokerConfig.ResponseQueueName)
	if err != nil {
		panic(fmt.Errorf("failed to create queue listener: %w", err))
	}

	return &Initialization{Cfg: cfg, Db: db, TaskService: taskService, TaskRoute: taskRoute, QueueListener: queueListener, SessionRoute: sessionRoute}
}
