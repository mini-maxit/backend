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
	UserRoute    routes.UserRoute
	SubmissionRoute routes.SubmissionRoutes

	QueueListener queue.QueueListener
}

func connectToBroker(cfg *config.Config) (*amqp.Connection, *amqp.Channel) {
	log := logger.NewNamedLogger("connect_to_broker")

	var err error
	var conn *amqp.Connection
	for v := range 5 {
		conn, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.BrokerConfig.User, cfg.BrokerConfig.Password, cfg.BrokerConfig.Host, cfg.BrokerConfig.Port))
		if err != nil {
			log.Warnf("Failed to connect to RabbitMQ: %s", err.Error())
			time.Sleep(2*time.Second + time.Duration(v))
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
	log := logger.NewNamedLogger("initialization")
	conn, channel := connectToBroker(cfg)
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Panicf("Failed to connect to database: %s", err.Error())
	}
	tx, err := db.Connect()

	defer utils.TransactionPanicRecover(tx)

	if err != nil {
		log.Panicf("Failed to connect to database: %s", err.Error())
	}
	// Repositories
	langRepository, err := repository.NewLanguageRepository(tx)
	if err != nil {
		log.Panicf("Failed to create language repository: %s", err.Error())
	}
	userRepository, err := repository.NewUserRepository(tx)
	if err != nil {
		log.Panicf("Failed to create user repository: %s", err.Error())
	}
	taskRepository, err := repository.NewTaskRepository(tx)
	if err != nil {
		log.Panicf("Failed to create task repository: %s", err.Error())
	}
	_, err = repository.NewGroupRepository(tx)
	if err != nil {
		log.Panicf("Failed to create group repository: %s", err.Error())
	}
	submissionRepository, err := repository.NewSubmissionRepository(tx)
	if err != nil {
		log.Panicf("Failed to create submission repository: %s", err.Error())
	}
	_, err = repository.NewSubmissionResultRepository(tx)
	if err != nil {
		log.Panicf("Failed to create submission result repository: %s", err.Error())
	}
	queueRepository, err := repository.NewQueueMessageRepository(tx)
	if err != nil {
		log.Panicf("Failed to create queue repository: %s", err.Error())
	}
	sessionRepository, err := repository.NewSessionRepository(tx)
	if err != nil {
		log.Panicf("Failed to create session repository: %s", err.Error())
	}
	submissionResultRepository, err := repository.NewSubmissionResultRepository(tx)
	if err != nil {
		log.Panicf("Failed to create submission result repository: %s", err.Error())
	}

	if err := db.Commit(); err != nil {
		log.Panicf("Failed to commit transaction: %s", err.Error())
	}

	// Services
	userService := service.NewUserService(userRepository)
	taskService := service.NewTaskService(cfg, taskRepository, submissionRepository)
	queueService, err := service.NewQueueService(taskRepository, submissionRepository, queueRepository, conn, channel, cfg.BrokerConfig.QueueName, cfg.BrokerConfig.ResponseQueueName)
	if err != nil {
		log.Panicf("Failed to create queue service: %s", err.Error())
	}
	sessionService := service.NewSessionService(sessionRepository, userRepository)
	authService := service.NewAuthService(userRepository, sessionService)
	languageService := service.NewLanguageService(langRepository)
	submissionService := service.NewSubmissionService(submissionRepository, submissionResultRepository, languageService, taskService, userService)
	tx, err = db.Connect()
	if err != nil {
		log.Panicf("Failed to connect to database to init languages: %s", err.Error())
	}
	err = languageService.InitLanguages(tx)
	if err != nil {
		log.Panicf("Failed to init languages: %s", err.Error())
		tx.Rollback()
	}
	err = db.Commit()
	if err != nil {
		log.Panicf("Failed to commit transaction after lang init: %s", err.Error())
	}

	// Routes
	taskRoute := routes.NewTaskRoute(cfg.FileStorageUrl, taskService)
	sessionRoute := routes.NewSessionRoute(sessionService)
	authRoute := routes.NewAuthRoute(userService, authService)
	userRoute := routes.NewUserRoute(userService)
	submissionRoute := routes.NewSubmissionRoutes(submissionService, cfg.FileStorageUrl, queueService)

	// Queue listener
	queueListener, err := queue.NewQueueListener(conn, channel, taskService, cfg.BrokerConfig.ResponseQueueName)
	if err != nil {
		log.Panicf("Failed to create queue listener: %s", err.Error())
	}

	return &Initialization{
		Cfg:            cfg,
		Db:             db,
		QueueListener:  queueListener,
		TaskService:    taskService,
		SessionService: sessionService,
		AuthRoute:      authRoute,
		SessionRoute:   sessionRoute,
		TaskRoute:      taskRoute,
		UserRoute:      userRoute,
		SubmissionRoute: submissionRoute,
	}
}
