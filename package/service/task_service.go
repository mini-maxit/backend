package service

import (
	"fmt"

	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/internal/logger"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"gorm.io/gorm"
)

var ErrDatabaseConnection = fmt.Errorf("failed to connect to the database")

type TaskService interface {
	// Create creates a new empty task and returns the task ID
	Create(tx *gorm.DB, task schemas.Task) (int64, error)
	GetAll(tx *gorm.DB, limit, offset int64) ([]schemas.Task, error)
	GetAllForUser(tx *gorm.DB, userId, limit, offset int64) ([]schemas.Task, error)
	GetAllForGroup(tx *gorm.DB, groupId, limit, offset int64) ([]schemas.Task, error)
	GetTask(tx *gorm.DB, taskId int64) (*schemas.TaskDetailed, error)
	UpdateTask(tx *gorm.DB, taskId int64, updateInfo schemas.UpdateTask) error
	CreateSubmission(tx *gorm.DB, taskId int64, userId int64, languageId int64, order int64) (int64, error)
}

type TaskServiceImpl struct {
	cfg                  *config.Config
	taskRepository       repository.TaskRepository
	submissionRepository repository.SubmissionRepository
	task_logger          *logger.ServiceLogger
}

func (ts *TaskServiceImpl) Create(tx *gorm.DB, task schemas.Task) (int64, error) {
	// Create a new task
	model := models.Task{
		Title:     task.Title,
		CreatedBy: task.CreatedBy,
	}
	taskId, err := ts.taskRepository.Create(tx, model)
	if err != nil {
		logger.Log(ts.task_logger, "Error creating task:", err.Error(), logger.Error)
		return 0, err
	}

	return taskId, nil
}

func (ts *TaskServiceImpl) GetAll(tx *gorm.DB, limit, offset int64) ([]schemas.Task, error) {
	// Get all tasks
	tasks, err := ts.taskRepository.GetAllTasks(tx)
	if err != nil {
		logger.Log(ts.task_logger, "Error getting all tasks:", err.Error(), logger.Error)
		return nil, err
	}

	// Convert the models to schemas
	var result []schemas.Task
	for _, task := range tasks {
		result = append(result, ts.modelToSchema(task))
	}

	// Handle pagination
	if offset >= int64(len(result)) {
		return []schemas.Task{}, nil
	}

	end := offset + limit
	if end > int64(len(result)) {
		end = int64(len(result))
	}

	return result[offset:end], nil
}

func (ts *TaskServiceImpl) GetAllForUser(tx *gorm.DB, userId, limit, offset int64) ([]schemas.Task, error) {
	// Get all tasks
	tasks, err := ts.taskRepository.GetAllForUser(tx, userId)
	if err != nil {
		logger.Log(ts.task_logger, "Error getting all tasks for user:", err.Error(), logger.Error)
		return nil, err
	}

	// Convert the models to schemas
	var result []schemas.Task
	for _, task := range tasks {
		result = append(result, ts.modelToSchema(task))
	}

	// Handle pagination
	if offset >= int64(len(result)) {
		return []schemas.Task{}, nil
	}

	end := offset + limit
	if end > int64(len(result)) {
		end = int64(len(result))
	}

	return result[offset:end], nil
}

func (ts *TaskServiceImpl) GetAllForGroup(tx *gorm.DB, groupId, limit, offset int64) ([]schemas.Task, error) {
	// Get all tasks
	tasks, err := ts.taskRepository.GetAllForGroup(tx, groupId)
	if err != nil {
		logger.Log(ts.task_logger, "Error getting all tasks for group:", err.Error(), logger.Error)
		return nil, err
	}

	// Convert the models to schemas
	var result []schemas.Task
	for _, task := range tasks {
		result = append(result, ts.modelToSchema(task))
	}

	// Handle pagination
	if offset >= int64(len(result)) {
		return []schemas.Task{}, nil
	}

	end := offset + limit
	if end > int64(len(result)) {
		end = int64(len(result))
	}

	return result[offset:end], nil
}

func (ts *TaskServiceImpl) GetTask(tx *gorm.DB, taskId int64) (*schemas.TaskDetailed, error) {
	// Get the task
	task, err := ts.taskRepository.GetTask(tx, taskId)
	if err != nil {
		logger.Log(ts.task_logger, "Error getting task:", err.Error(), logger.Error)
		return nil, err
	}

	// Convert the model to schema
	result := &schemas.TaskDetailed{
		Id:             task.Id,
		Title:          task.Title,
		DescriptionURL: fmt.Sprintf("%s/getTaskDescription?taskID=%d", ts.cfg.FileStorageUrl, task.Id),
		CreatedBy:      task.CreatedBy,
		CreatedByName:  task.Author.Name,
		CreatedAt:      task.CreatedAt,
	}

	return result, nil
}

func (ts *TaskServiceImpl) UpdateTask(tx *gorm.DB, taskId int64, updateInfo schemas.UpdateTask) error {
	currentTask, err := ts.taskRepository.GetTask(tx, taskId)
	if err != nil {
		logger.Log(ts.task_logger, "Error getting task:", err.Error(), logger.Error)
		return err
	}

	ts.updateModel(currentTask, &updateInfo)

	// Update the task
	err = ts.taskRepository.UpdateTask(tx, taskId, currentTask)
	if err != nil {
		logger.Log(ts.task_logger, "Error updating task:", err.Error(), logger.Error)
		return err
	}

	return nil
}

func (ts *TaskServiceImpl) CreateSubmission(tx *gorm.DB, taskId int64, userId int64, languageId int64, order int64) (int64, error) {
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
		logger.Log(ts.task_logger, "Error creating submission:", err.Error(), logger.Error)
		return 0, err
	}

	return submissionId, nil
}

func (ts *TaskServiceImpl) updateModel(currentModel *models.Task, updateInfo *schemas.UpdateTask) {
	if updateInfo.Title != "" {
		currentModel.Title = updateInfo.Title
	}
}

func (ts *TaskServiceImpl) modelToSchema(model models.Task) schemas.Task {
	return schemas.Task{
		Id:        model.Id,
		Title:     model.Title,
		CreatedBy: model.CreatedBy,
		CreatedAt: model.CreatedAt,
	}
}

func NewTaskService(cfg *config.Config, taskRepository repository.TaskRepository, submissionRepository repository.SubmissionRepository) TaskService {
	task_logger := logger.NewNamedLogger("task_service")
	return &TaskServiceImpl{
		cfg:                  cfg,
		taskRepository:       taskRepository,
		submissionRepository: submissionRepository,
		task_logger:          &task_logger,
	}
}
