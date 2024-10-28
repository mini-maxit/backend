package service

import (
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/repository"
)

type TaskService interface {
	// CreateEmpty creates a new empty task and returns the task ID
	CreateEmpty() (int64, error)
}

type TaskServiceImpl struct {
	database       database.Database
	taskRepository repository.TaskRepository
}

func (ts *TaskServiceImpl) CreateEmpty() (int64, error) {
	// Connect to the database and start a transaction
	db := ts.database.Connect()
	tx := db.Begin()

	// Create a new task
	taskId, err := ts.taskRepository.CreateEmpty(tx)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// Commit the transaction and return the task ID
	tx.Commit()
	return taskId, nil
}

func NewTaskService(taskRepository repository.TaskRepository) TaskService {
	return &TaskServiceImpl{
		taskRepository: taskRepository,
	}
}
