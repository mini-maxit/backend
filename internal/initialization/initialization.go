package initialization

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mini-maxit/backend/internal/api/http/routes"
	"github.com/mini-maxit/backend/internal/api/queue"
	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"

	amqp "github.com/rabbitmq/amqp091-go"
)

const numTries = 10

type Initialization struct {
	Cfg *config.Config
	DB  database.Database

	TaskService service.TaskService
	JWTService  service.JWTService

	AuthRoute       routes.AuthRoute
	GroupRoute      routes.GroupRoute
	SubmissionRoute routes.SubmissionRoutes
	TaskRoute       routes.TaskRoute
	UserRoute       routes.UserRoute
	WorkerRoute     routes.WorkerRoute

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
	tx, err := db.BeginTransaction()

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
	groupRepository, err := repository.NewGroupRepository(tx)
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
	inputOutputRepository, err := repository.NewInputOutputRepository(tx)
	if err != nil {
		log.Panicf("Failed to create input output repository: %s", err.Error())
	}
	testResultRepository, err := repository.NewTestResultRepository(tx)
	if err != nil {
		log.Panicf("Failed to create test result repository: %s", err.Error())
	}
	queueRepository, err := repository.NewQueueMessageRepository(tx)
	if err != nil {
		log.Panicf("Failed to create queue repository: %s", err.Error())
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
	taskService := service.NewTaskService(cfg.FileStorageURL,
		taskRepository,
		inputOutputRepository,
		userRepository,
		groupRepository,
	)
	queueService, err := service.NewQueueService(taskRepository,
		submissionRepository,
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
	groupService := service.NewGroupService(groupRepository, userRepository, userService)
	langService := service.NewLanguageService(langRepository)
	submissionService := service.NewSubmissionService(
		submissionRepository,
		submissionResultRepository,
		inputOutputRepository,
		testResultRepository,
		groupRepository,
		taskRepository,
		langService,
		taskService,
		userService,
	)
	workerService := service.NewWorkerService(queueService)

	// Routes
	authRoute := routes.NewAuthRoute(userService, authService)
	groupRoute := routes.NewGroupRoute(groupService)
	submissionRoute := routes.NewSubmissionRoutes(submissionService, cfg.FileStorageURL, queueService, taskService)
	taskRoute := routes.NewTaskRoute(cfg.FileStorageURL, taskService)
	userRoute := routes.NewUserRoute(userService)
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
		dump(db, log, authService)
	}
	return &Initialization{
		Cfg: cfg,
		DB:  db,

		TaskService: taskService,
		JWTService:  jwtService,

		AuthRoute:       authRoute,
		GroupRoute:      groupRoute,
		SubmissionRoute: submissionRoute,
		TaskRoute:       taskRoute,
		UserRoute:       userRoute,
		WorkerRoute:     workerRoute,

		QueueListener: queueListener,
	}
}

func dump(db *database.PostgresDB, log *zap.SugaredLogger, authService service.AuthService) {
	tx, err := db.BeginTransaction()
	if err != nil {
		log.Warnf("Failed to connect to database to init dump: %s", err.Error())
	}

	_, err = authService.Register(tx, schemas.UserRegisterRequest{
		Name:     "AdminName",
		Surname:  "AdminSurname",
		Email:    "admin@admin.com",
		Username: "admin",
		Password: "adminadmin",
	})
	if err != nil {
		log.Warnf("Failed to create admin: %s", err.Error())
	}
	_, err = authService.Register(tx, schemas.UserRegisterRequest{
		Name:     "TeacherName",
		Surname:  "TeacherSurname",
		Email:    "teacher@teacher.com",
		Username: "teacher",
		Password: "teacherteacher",
	})
	if err != nil {
		log.Warnf("Failed to create teacher: %s", err.Error())
	}
	_, err = authService.Register(tx, schemas.UserRegisterRequest{
		Name:     "StudentName",
		Surname:  "StudentSurname",
		Email:    "student@student.com",
		Username: "student",
		Password: "studentstudent",
	})
	if err != nil {
		log.Warnf("Failed to create student: %s", err.Error())
	}

	err = db.Commit()
	if err != nil {
		log.Warnf("Failed to commit transaction after dump: %s", err.Error())
	}
}
