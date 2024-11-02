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
	TaskRoute   routes.TaskRoute

	QueueListener queue.QueueListener
}

func connectToBrocker(cfg *config.Config) (*amqp.Connection, *amqp.Channel) {
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
	conn, channel := connectToBrocker(cfg)
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}

	// Repositories
	_, err = repository.NewUserRepository(db.Connect())
	if err != nil {
		panic(fmt.Errorf("failed to create user repository: %w", err))
	}
	taskRepository, err := repository.NewTaskRepository(db.Connect())
	if err != nil {
		panic(fmt.Errorf("failed to create task repository: %w", err))
	}
	userSolutionRepository, err := repository.NewUserSolutionRepository(db.Connect())
	if err != nil {
		panic(fmt.Errorf("failed to create submission repository: %w", err))
	}
	queueRepository, err := repository.NewQueueMessageRepository(db.Connect())
	if err != nil {
		panic(fmt.Errorf("failed to create queue repository: %w", err))
	}

	// Services
	taskService := service.NewTaskService(db, taskRepository, userSolutionRepository)
	queueService, err := service.NewQueueService(db, taskRepository, userSolutionRepository, queueRepository, conn, channel, cfg.BrokerConfig.QueueName, cfg.BrokerConfig.ResponseQueueName)
	if err != nil {
		panic(fmt.Errorf("failed to create queue service: %w", err))
	}

	// Routes
	taskRoute := routes.NewTaskRoute(cfg.FileStorageUrl, taskService, queueService)

	// Queue listener
	queueListener, err := queue.NewQueueListener(conn, channel, taskService, cfg.BrokerConfig.ResponseQueueName)
	if err != nil {
		panic(fmt.Errorf("failed to create queue listener: %w", err))
	}

	return &Initialization{Cfg: cfg, Db: db, TaskService: taskService, TaskRoute: taskRoute, QueueListener: queueListener}
}
