package service

import (
	"fmt"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
)

var ErrDatabaseConnection = fmt.Errorf("failed to connect to the database")

type TaskService interface {
	// Create creates a new empty task and returns the task ID
	Create(task schemas.Task) (int64, error)
	UpdateTask(taskId int64, updateInfo schemas.UpdateTask) error
	CreateSubmission(taskId int64, userId int64, languageId int64, order int64) (int64, error)
}

type TaskServiceImpl struct {
	database             database.Database
	taskRepository       repository.TaskRepository
	submissionRepository repository.SubmissionRepository
}

func (ts *TaskServiceImpl) Create(task schemas.Task) (int64, error) {
	// Connect to the database and start a transaction
	db := ts.database.Connect()
	if db == nil {
		return 0, ErrDatabaseConnection
	}
	tx := db.Begin()

	// Create a new task
	model := models.Task{
		Title:     task.Title,
		CreatedBy: task.CreatedBy,
	}
	taskId, err := ts.taskRepository.Create(tx, model)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// Commit the transaction and return the task ID
	tx.Commit()
	return taskId, nil
}

func (ts *TaskServiceImpl) UpdateTask(taskId int64, updateInfo schemas.UpdateTask) error {
	// Connect to the database and start a transaction
	db := ts.database.Connect()
	if db == nil {
		return ErrDatabaseConnection
	}
	tx := db.Begin()

	currentTask, err := ts.taskRepository.GetTask(tx, taskId)
	if err != nil {
		tx.Rollback()
		return err
	}

	ts.updateModel(currentTask, &updateInfo)

	// Update the task
	err = ts.taskRepository.UpdateTask(tx, taskId, currentTask)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	tx.Commit()
	return nil
}

func (ts *TaskServiceImpl) CreateSubmission(taskId int64, userId int64, languageId int64, order int64) (int64, error) {
	// Connect to the database and start a transaction
	db := ts.database.Connect()
	if db == nil {
		return 0, ErrDatabaseConnection
	}
	tx := db.Begin()

	// Create a new submission
	submission := models.Submission{
		TaskId:     taskId,
		UserId:     userId,
		Order:      order,
		LanguageId: languageId,
		Status:     "received",
		CheckedAt:  nil,
	}
	submissionId, err := ts.submissionRepository.CreateSubmission(tx, submission)

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// Commit the transaction
	tx.Commit()
	return submissionId, nil
}

func (ts *TaskServiceImpl) updateModel(currentModel *models.Task, updateInfo *schemas.UpdateTask) {
	if updateInfo.Title != "" {
		currentModel.Title = updateInfo.Title
	}
}

func NewTaskService(db database.Database, taskRepository repository.TaskRepository, submissionRepository repository.SubmissionRepository) TaskService {
	return &TaskServiceImpl{
		database:             db,
		taskRepository:       taskRepository,
		submissionRepository: submissionRepository,
	}
}
