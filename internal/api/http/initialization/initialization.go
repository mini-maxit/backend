package initialization

import (
	"fmt"

	"github.com/mini-maxit/backend/internal/api/http/routes"
	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/service"
)

type Initialization struct {
	Cfg *config.Config
	Db  database.Database

	TaskService service.TaskService
	TaskRoute   routes.TaskRoute
}

func NewInitialization(cfg *config.Config) *Initialization {
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}

	taskRepository, err := repository.NewTaskRepository(db.Connect())
	if err != nil {
		panic(fmt.Errorf("failed to create task repository: %w", err))
	}
	taskService := service.NewTaskService(taskRepository)
	taskRoute := routes.NewTaskRoute(cfg.FileStorageUrl, taskService)

	return &Initialization{Cfg: cfg, Db: db, TaskService: taskService, TaskRoute: taskRoute}
}
