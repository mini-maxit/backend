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
	"github.com/mini-maxit/backend/package/domain/types"
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

	AuthRoute       routes.AuthRoute
	GroupRoute      routes.GroupRoute
	SessionRoute    routes.SessionRoute
	SubmissionRoute routes.SubmissionRoutes
	TaskRoute       routes.TaskRoute
	UserRoute       routes.UserRoute

	QueueListener queue.QueueListener

	Dump func(w http.ResponseWriter, r *http.Request)
}

func connectToBroker(cfg *config.Config) (*amqp.Connection, *amqp.Channel) {
	log := utils.NewNamedLogger("connect_to_broker")

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
	taskService := service.NewTaskService(cfg.FileStorageUrl, taskRepository, inputOutputRepository, userRepository, groupRepository)
	queueService, err := service.NewQueueService(taskRepository, submissionRepository, queueRepository, conn, channel, cfg.BrokerConfig.QueueName, cfg.BrokerConfig.ResponseQueueName)
	if err != nil {
		log.Panicf("Failed to create queue service: %s", err.Error())
	}
	sessionService := service.NewSessionService(sessionRepository, userRepository)
	authService := service.NewAuthService(userRepository, sessionService)
	groupService := service.NewGroupService(groupRepository, userRepository, userService)
	langService := service.NewLanguageService(langRepository)
	submissionService := service.NewSubmissionService(submissionRepository, submissionResultRepository, inputOutputRepository, testResultRepository, langService, taskService, userService)
	tx, err = db.BeginTransaction()
	if err != nil {
		log.Panicf("Failed to connect to database to init languages: %s", err.Error())
	}
	err = langService.InitLanguages(tx, cfg.EnabledLanguages)
	if err != nil {
		log.Panicf("Failed to init languages: %s", err.Error())
		tx.Rollback()
	}
	err = db.Commit()
	if err != nil {
		log.Panicf("Failed to commit transaction after lang init: %s", err.Error())
	}

	// Routes
	authRoute := routes.NewAuthRoute(userService, authService)
	groupRoute := routes.NewGroupRoute(groupService)
	sessionRoute := routes.NewSessionRoute(sessionService)
	submissionRoute := routes.NewSubmissionRoutes(submissionService, cfg.FileStorageUrl, queueService)
	taskRoute := routes.NewTaskRoute(cfg.FileStorageUrl, taskService)
	userRoute := routes.NewUserRoute(userService)

	// Queue listener
	queueListener, err := queue.NewQueueListener(conn, channel, db, taskService, queueService, submissionService, cfg.BrokerConfig.ResponseQueueName)
	if err != nil {
		log.Panicf("Failed to create queue listener: %s", err.Error())
	}

	if cfg.Dump {
		tx, err := db.BeginTransaction()
		if err != nil {
			log.Warnf("Failed to connect to database to init dump: %s", err.Error())
		}
		fakeUser := schemas.User{
			Name:     "FakeName",
			Surname:  "FakeSurname",
			Email:    "asd@asdf.com",
			Username: "fake",
			Role:     types.UserRoleAdmin,
		}
		session, err := authService.Register(tx, schemas.UserRegisterRequest{
			Name:     "AdminName",
			Surname:  "AdminSurname",
			Email:    "admin@admin.com",
			Username: "admin",
			Password: "adminadmin",
		})
		if err != nil {
			log.Warnf("Failed to create admin: %s", err.Error())
		} else {
			err = userService.ChangeRole(tx, fakeUser, session.UserId, types.UserRoleAdmin)
			if err != nil {
				log.Warnf("Failed to change admin role: %s", err.Error())
			}
		}
		session, err = authService.Register(tx, schemas.UserRegisterRequest{
			Name:     "TeacherName",
			Surname:  "TeacherSurname",
			Email:    "teacher@teacher.com",
			Username: "teacher",
			Password: "teacherteacher",
		})
		if err != nil {
			log.Warnf("Failed to create teacher: %s", err.Error())
		} else {
			err = userService.ChangeRole(tx, fakeUser, session.UserId, types.UserRoleTeacher)
			if err != nil {
				log.Warnf("Failed to change teacher role: %s", err.Error())
			}
		}
		session, err = authService.Register(tx, schemas.UserRegisterRequest{
			Name:     "StudentName",
			Surname:  "StudentSurname",
			Email:    "student@student.com",
			Username: "student",
			Password: "studentstudent",
		})
		if err != nil {
			log.Warnf("Failed to create student: %s", err.Error())
		} else {
			err = userService.ChangeRole(tx, fakeUser, session.UserId, types.UserRoleStudent)
			if err != nil {
				log.Warnf("Failed to change student role: %s", err.Error())
			}
		}
		err = db.Commit()
		if err != nil {
			log.Warnf("Failed to commit transaction after dump: %s", err.Error())
		}

	}

	return &Initialization{
		Cfg: cfg,
		Db:  db,

		TaskService:    taskService,
		SessionService: sessionService,

		AuthRoute:       authRoute,
		GroupRoute:      groupRoute,
		SessionRoute:    sessionRoute,
		SubmissionRoute: submissionRoute,
		TaskRoute:       taskRoute,
		UserRoute:       userRoute,

		QueueListener: queueListener,
	}
}
