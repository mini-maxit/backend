package service

import (
	"fmt"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var ErrDatabaseConnection = fmt.Errorf("failed to connect to the database")
var ErrTaskExists = fmt.Errorf("task with this title already exists")
var ErrTaskNotFound = fmt.Errorf("task not found")

type TaskService interface {
	// Create creates a new empty task and returns the task ID
	Create(tx *gorm.DB, task *schemas.Task) (int64, error)
	GetAll(tx *gorm.DB, queryParams map[string]string) ([]schemas.Task, error)
	GetAllForUser(tx *gorm.DB, userId int64, queryParams map[string]string) ([]schemas.Task, error)
	GetAllForGroup(tx *gorm.DB, groupId int64, queryParams map[string]string) ([]schemas.Task, error)
	GetTask(tx *gorm.DB, taskId int64) (*schemas.TaskDetailed, error)
	GetTaskByTitle(tx *gorm.DB, title string) (*schemas.Task, error)
	UpdateTask(tx *gorm.DB, taskId int64, updateInfo schemas.UpdateTask) error
	modelToSchema(model *models.Task) *schemas.Task
}

type taskService struct {
	fileStorageUrl string
	taskRepository repository.TaskRepository
	logger         *zap.SugaredLogger
}

func (ts *taskService) Create(tx *gorm.DB, task *schemas.Task) (int64, error) {
	// Create a new task
	_, err := ts.GetTaskByTitle(tx, task.Title)
	if err != nil && err != ErrTaskNotFound {
		ts.logger.Errorf("Error getting task by title: %v", err.Error())
		return 0, err
	} else if err == nil {
		return 0, ErrTaskExists
	}

	model := models.Task{
		Title:     task.Title,
		CreatedBy: task.CreatedBy,
	}
	taskId, err := ts.taskRepository.Create(tx, &model)
	if err != nil {
		ts.logger.Errorf("Error creating task: %v", err.Error())
		return 0, err
	}

	return taskId, nil
}

func (ts *taskService) GetTaskByTitle(tx *gorm.DB, title string) (*schemas.Task, error) {
	task, err := ts.taskRepository.GetTaskByTitle(tx, title)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTaskNotFound
		}
		ts.logger.Errorf("Error getting task by title: %v", err.Error())
		return nil, err
	}

	result := ts.modelToSchema(task)

	return result, nil
}

func (ts *taskService) GetAll(tx *gorm.DB, queryParams map[string]string) ([]schemas.Task, error) {

	limit := queryParams["limit"]
	offset := queryParams["offset"]
	sort := queryParams["sort"]

	// Get all tasks
	tasks, err := ts.taskRepository.GetAllTasks(tx, limit, offset, sort)
	if err != nil {
		ts.logger.Errorf("Error getting all tasks: %v", err.Error())
		return nil, err
	}

	// Convert the models to schemas
	var result []schemas.Task
	for _, task := range tasks {
		result = append(result, *ts.modelToSchema(&task))
	}

	return result, nil
}

func (ts *taskService) GetAllForUser(tx *gorm.DB, userId int64, queryParams map[string]string) ([]schemas.Task, error) {
	limit := queryParams["limit"]
	offset := queryParams["offset"]
	sort := queryParams["sort"]

	// Get all tasks
	tasks, err := ts.taskRepository.GetAllForUser(tx, userId, limit, offset, sort)
	if err != nil {
		ts.logger.Errorf("Error getting all tasks for user: %v", err.Error())
		return nil, err
	}

	// Convert the models to schemas
	var result []schemas.Task
	for _, task := range tasks {
		result = append(result, *ts.modelToSchema(&task))
	}

	return result, nil
}

func (ts *taskService) GetAllForGroup(tx *gorm.DB, groupId int64, queryParams map[string]string) ([]schemas.Task, error) {
	limit := queryParams["limit"]
	offset := queryParams["offset"]
	sort := queryParams["sort"]

	// Get all tasks
	tasks, err := ts.taskRepository.GetAllForGroup(tx, groupId, limit, offset, sort)
	if err != nil {
		ts.logger.Error("Error getting all tasks for group")
		return nil, err
	}

	// Convert the models to schemas
	var result []schemas.Task
	for _, task := range tasks {
		result = append(result, *ts.modelToSchema(&task))
	}

	return result, nil
}

func (ts *taskService) GetTask(tx *gorm.DB, taskId int64) (*schemas.TaskDetailed, error) {
	// Get the task
	task, err := ts.taskRepository.GetTask(tx, taskId)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return nil, err
	}

	// Convert the model to schema
	result := &schemas.TaskDetailed{
		Id:             task.Id,
		Title:          task.Title,
		DescriptionURL: fmt.Sprintf("%s/getTaskDescription?taskID=%d", ts.fileStorageUrl, task.Id),
		CreatedBy:      task.CreatedBy,
		CreatedByName:  task.Author.Name,
		CreatedAt:      task.CreatedAt,
	}

	return result, nil
}

func (ts *taskService) UpdateTask(tx *gorm.DB, taskId int64, updateInfo schemas.UpdateTask) error {
	currentTask, err := ts.taskRepository.GetTask(tx, taskId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrTaskNotFound
		}
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return err
	}

	ts.updateModel(currentTask, &updateInfo)

	// Update the task
	err = ts.taskRepository.UpdateTask(tx, taskId, currentTask)
	if err != nil {
		ts.logger.Errorf("Error updating task: %v", err.Error())
		return err
	}
	return nil
}

func (ts *taskService) updateModel(currentModel *models.Task, updateInfo *schemas.UpdateTask) {
	if updateInfo.Title != "" {
		currentModel.Title = updateInfo.Title
	}
}

func (ts *taskService) modelToSchema(model *models.Task) *schemas.Task {
	return &schemas.Task{
		Id:        model.Id,
		Title:     model.Title,
		CreatedBy: model.CreatedBy,
		CreatedAt: model.CreatedAt,
	}
}

func NewTaskService(fileStorageUrl string, taskRepository repository.TaskRepository) TaskService {
	log := utils.NewNamedLogger("task_service")
	return &taskService{
		fileStorageUrl: fileStorageUrl,
		taskRepository: taskRepository,
		logger:         log,
	}
}
