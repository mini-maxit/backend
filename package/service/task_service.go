package service

import (
	"fmt"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TaskService interface {
	// Create creates a new empty task and returns the task ID
	Create(tx *gorm.DB, current_user schemas.User, task *schemas.Task) (int64, error)
	GetAll(tx *gorm.DB, current_user schemas.User, queryParams map[string]string) ([]schemas.Task, error)
	GetAllForGroup(tx *gorm.DB, current_user schemas.User, groupId int64, queryParams map[string]string) ([]schemas.Task, error)
	GetTask(tx *gorm.DB, current_user schemas.User, taskId int64) (*schemas.TaskDetailed, error)
	GetTaskByTitle(tx *gorm.DB, title string) (*schemas.Task, error)
	GetAllAssignedTasks(tx *gorm.DB, current_user schemas.User, queryParams map[string]string) ([]schemas.Task, error)
	GetAllCreatedTasks(tx *gorm.DB, current_user schemas.User, queryParams map[string]string) ([]schemas.Task, error)
	UpdateTask(tx *gorm.DB, taskId int64, updateInfo schemas.UpdateTask) error
	AssignTaskToUsers(tx *gorm.DB, current_user schemas.User, taskId int64, userIds []int64) error
	AssignTaskToGroups(tx *gorm.DB, current_user schemas.User, taskId int64, groupIds []int64) error
	UnAssignTaskFromUsers(tx *gorm.DB, current_user schemas.User, taskId int64, userId []int64) error
	UnAssignTaskFromGroups(tx *gorm.DB, current_user schemas.User, taskId int64, groupIds []int64) error
	DeleteTask(tx *gorm.DB, current_user schemas.User, taskId int64) error
	modelToSchema(model *models.Task) *schemas.Task
}

type taskService struct {
	fileStorageUrl  string
	taskRepository  repository.TaskRepository
	userRepository  repository.UserRepository
	groupRepository repository.GroupRepository
	logger          *zap.SugaredLogger
}

func (ts *taskService) Create(tx *gorm.DB, current_user schemas.User, task *schemas.Task) (int64, error) {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return 0, err
	}
	// Create a new task
	_, err = ts.GetTaskByTitle(tx, task.Title)
	if err != nil && err != errors.ErrTaskNotFound {
		ts.logger.Errorf("Error getting task by title: %v", err.Error())
		return 0, err
	} else if err == nil {
		return 0, errors.ErrTaskExists
	}

	author, err := ts.userRepository.GetUser(tx, current_user.Id)
	if err != nil {
		ts.logger.Errorf("Error getting user: %v", err.Error())
		return 0, err
	}

	model := models.Task{
		Title:     task.Title,
		CreatedBy: task.CreatedBy,
		Author:    *author,
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
			return nil, errors.ErrTaskNotFound
		}
		ts.logger.Errorf("Error getting task by title: %v", err.Error())
		return nil, err
	}

	result := ts.modelToSchema(task)

	return result, nil
}

func (ts *taskService) GetAll(tx *gorm.DB, current_user schemas.User, queryParams map[string]string) ([]schemas.Task, error) {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return nil, err
	}

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

func (ts *taskService) GetAllForGroup(tx *gorm.DB, current_user schemas.User, groupId int64, queryParams map[string]string) ([]schemas.Task, error) {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return nil, err
	}

	limit := queryParams["limit"]
	offset := queryParams["offset"]
	sort := queryParams["sort"]

	group, err := ts.groupRepository.GetGroup(tx, groupId)
	if err != nil {
		ts.logger.Errorf("Error getting group: %v", err.Error())
		return nil, err
	}

	if current_user.Role == types.UserRoleTeacher && group.CreatedBy != current_user.Id {
		return nil, errors.ErrNotAuthorized
	}

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

func (ts *taskService) GetTask(tx *gorm.DB, current_user schemas.User, taskId int64) (*schemas.TaskDetailed, error) {
	// Get the task
	task, err := ts.taskRepository.GetTask(tx, taskId)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return nil, err
	}

	switch types.UserRole(current_user.Role) {
	case types.UserRoleStudent:
		// Check if the task is assigned to the user
		isAssigned, err := ts.taskRepository.IsTaskAssignedToUser(tx, taskId, current_user.Id)
		if err != nil {
			ts.logger.Errorf("Error checking if task is assigned to user: %v", err.Error())
			return nil, err
		}
		if !isAssigned {
			return nil, errors.ErrNotAuthorized
		}
	case types.UserRoleTeacher:
		// Check if the task is created by the user
		if task.CreatedBy != current_user.Id {
			return nil, errors.ErrNotAuthorized
		}
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

func (ts *taskService) GetAllAssignedTasks(tx *gorm.DB, current_user schemas.User, queryParams map[string]string) ([]schemas.Task, error) {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleStudent})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return nil, err
	}

	limit := queryParams["limit"]
	offset := queryParams["offset"]
	sort := queryParams["sort"]

	tasks, err := ts.taskRepository.GetAllAssignedTasks(tx, current_user.Id, limit, offset, sort)
	if err != nil {
		ts.logger.Errorf("Error getting all assigned tasks: %v", err.Error())
		return nil, err
	}

	var result []schemas.Task

	for task := range tasks {
		result = append(result, *ts.modelToSchema(&tasks[task]))
	}

	return result, nil
}

func (ts *taskService) GetAllCreatedTasks(tx *gorm.DB, current_user schemas.User, queryParams map[string]string) ([]schemas.Task, error) {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleTeacher})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return nil, err
	}

	limit := queryParams["limit"]
	offset := queryParams["offset"]
	sort := queryParams["sort"]

	tasks, err := ts.taskRepository.GetAllCreatedTasks(tx, current_user.Id, limit, offset, sort)
	if err != nil {
		ts.logger.Errorf("Error getting all created tasks: %v", err.Error())
		return nil, err
	}

	var result []schemas.Task

	for task := range tasks {
		result = append(result, *ts.modelToSchema(&tasks[task]))
	}

	return result, nil
}

func (ts *taskService) AssignTaskToUsers(tx *gorm.DB, current_user schemas.User, taskId int64, userIds []int64) error {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return err
	}

	task, err := ts.taskRepository.GetTask(tx, taskId)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return err
	}
	if current_user.Role == types.UserRoleTeacher && task.CreatedBy != current_user.Id {
		return errors.ErrNotAuthorized
	}

	for _, userId := range userIds {
		_, err := ts.userRepository.GetUser(tx, userId)
		if err != nil {
			ts.logger.Errorf("Error getting user: %v", err.Error())
			return err
		}

		isAssigned, err := ts.taskRepository.IsTaskAssignedToUser(tx, taskId, userId)
		if err != nil {
			ts.logger.Errorf("Error checking if task is assigned to user: %v", err.Error())
			return err
		}
		if isAssigned {
			ts.logger.Errorf("Error task already assigned to user %d", userId)
			continue
		}

		err = ts.taskRepository.AssignTaskToUser(tx, taskId, userId)
		if err != nil {
			ts.logger.Errorf("Error assigning task to user: %v", err.Error())
			return err
		}
	}

	return nil
}

func (ts *taskService) AssignTaskToGroups(tx *gorm.DB, current_user schemas.User, taskId int64, groupIds []int64) error {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return err
	}

	task, err := ts.taskRepository.GetTask(tx, taskId)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return err
	}
	if current_user.Role == types.UserRoleTeacher && task.CreatedBy != current_user.Id {
		return errors.ErrNotAuthorized
	}

	for _, groupId := range groupIds {
		_, err := ts.groupRepository.GetGroup(tx, groupId)
		if err != nil {
			ts.logger.Errorf("Error getting group: %v", err.Error())
			return err
		}

		isAssigned, err := ts.taskRepository.IsTaskAssignedToGroup(tx, taskId, groupId)
		if err != nil {
			ts.logger.Errorf("Error checking if task is assigned to group: %v", err.Error())
			return err
		}
		if isAssigned {
			ts.logger.Errorf("Error task already assigned to group %d", groupId)
			continue
		}

		err = ts.taskRepository.AssignTaskToGroup(tx, taskId, groupId)
		if err != nil {
			ts.logger.Errorf("Error assigning task to group: %v", err.Error())
			return err
		}
	}

	return nil
}

func (ts *taskService) UnAssignTaskFromUsers(tx *gorm.DB, current_user schemas.User, taskId int64, userId []int64) error {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return err
	}

	task, err := ts.taskRepository.GetTask(tx, taskId)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return err
	}
	if current_user.Role == types.UserRoleTeacher && task.CreatedBy != current_user.Id {
		return errors.ErrNotAuthorized
	}

	for _, userId := range userId {

		isAssigned, err := ts.taskRepository.IsTaskAssignedToUser(tx, taskId, userId)
		if err != nil {
			ts.logger.Errorf("Error checking if task is assigned to user: %v", err.Error())
			return err
		}
		if !isAssigned {
			return errors.ErrTaskNotAssignedToUser
		}

		err = ts.taskRepository.UnAssignTaskFromUser(tx, taskId, userId)
		if err != nil {
			ts.logger.Errorf("Error unassigning task from user: %v", err.Error())
			return err
		}
	}

	return nil
}

func (ts *taskService) UnAssignTaskFromGroups(tx *gorm.DB, current_user schemas.User, taskId int64, groupIds []int64) error {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return err
	}

	task, err := ts.taskRepository.GetTask(tx, taskId)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return err
	}
	if current_user.Role == types.UserRoleTeacher && task.CreatedBy != current_user.Id {
		return errors.ErrNotAuthorized
	}

	for _, groupId := range groupIds {

		isAssigned, err := ts.taskRepository.IsTaskAssignedToGroup(tx, taskId, groupId)
		if err != nil {
			ts.logger.Errorf("Error checking if task is assigned to group: %v", err.Error())
			return err
		}
		if !isAssigned {
			return errors.ErrTaskNotAssignedToGroup
		}

		err = ts.taskRepository.UnAssignTaskFromGroup(tx, taskId, groupId)
		if err != nil {
			ts.logger.Errorf("Error unassigning task from group: %v", err.Error())
			return err
		}
	}

	return nil
}

func (ts *taskService) DeleteTask(tx *gorm.DB, current_user schemas.User, taskId int64) error {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin})
	if err != nil {
		ts.logger.Errorf("Error validating user role: %v", err.Error())
		return err
	}

	task, err := ts.taskRepository.GetTask(tx, taskId)
	if err != nil {
		ts.logger.Errorf("Error getting task: %v", err.Error())
		return err
	}
	if current_user.Role == types.UserRoleTeacher && task.CreatedBy != current_user.Id {
		return errors.ErrNotAuthorized
	}

	err = ts.taskRepository.DeleteTask(tx, taskId)
	if err != nil {
		ts.logger.Errorf("Error deleting task: %v", err.Error())
		return err
	}

	return nil
}

func (ts *taskService) UpdateTask(tx *gorm.DB, taskId int64, updateInfo schemas.UpdateTask) error {
	currentTask, err := ts.taskRepository.GetTask(tx, taskId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrTaskNotFound
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

func NewTaskService(fileStorageUrl string, taskRepository repository.TaskRepository, userRepository repository.UserRepository, groupRepository repository.GroupRepository) TaskService {
	log := utils.NewNamedLogger("task_service")
	return &taskService{
		fileStorageUrl:  fileStorageUrl,
		taskRepository:  taskRepository,
		userRepository:  userRepository,
		groupRepository: groupRepository,
		logger:          log,
	}
}
