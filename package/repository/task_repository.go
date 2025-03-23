package repository

import (
	"time"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/utils"
	"gorm.io/gorm"
)

type TaskRepository interface {
	// AssignToGroup assigns a task to a group.
	AssignToGroup(tx *gorm.DB, taskId, groupId int64) error
	// AssignToUser assigns a task to a used.
	AssignToUser(tx *gorm.DB, taskId, userId int64) error
	// Create creates a new empty task and returns the task Id.
	Create(tx *gorm.DB, task *models.Task) (int64, error)
	// Delete deletes a task. It does not actually delete the task from the database, but performs a soft delete.
	Delete(tx *gorm.DB, taskId int64) error
	// Edit edits a task, by setting the fields of the task to the fields of the function argument.
	Edit(tx *gorm.DB, taskId int64, task *models.Task) error
	// GetAllAssigned returns all tasks assigned to a user, either directly or through a group. The tasks are paginated.
	GetAllAssigned(tx *gorm.DB, userId int64, limit, offset int, sort string) ([]models.Task, error)
	// GetAllCreated returns all tasks created by a user. The tasks are paginated.
	GetAllCreated(tx *gorm.DB, userId int64, limit, offset int, sort string) ([]models.Task, error)
	// GetAllForGroup returns all tasks assigned to a group. The tasks are paginated.
	GetAllForGroup(tx *gorm.DB, groupId int64, limit, offset int, sort string) ([]models.Task, error)
	// GetAll returns all tasks. The tasks are paginated.
	GetAll(tx *gorm.DB, limit, offset int, sort string) ([]models.Task, error)
	// Get returns a task by its Id.
	Get(tx *gorm.DB, taskId int64) (*models.Task, error)
	// GetByTitle returns a task by its title.
	GetByTitle(tx *gorm.DB, title string) (*models.Task, error)
	// GetMemoryLimits returns the memory limits of a task.
	GetMemoryLimits(tx *gorm.DB, taskId int64) ([]int64, error)
	// GetTimeLimits returns the time limits of a task.
	GetTimeLimits(tx *gorm.DB, taskId int64) ([]int64, error)
	// IsAssignedToGroup checks if a task is assigned to a group.
	IsAssignedToGroup(tx *gorm.DB, taskId, groupId int64) (bool, error)
	// IsAssignedToUser checks if a task is assigned to a user.
	IsAssignedToUser(tx *gorm.DB, taskId, userId int64) (bool, error)
	// UnassignFromGroup unassigns a task from a group.
	UnassignFromGroup(tx *gorm.DB, taskId, groupId int64) error
	// UnassignFromUser unassigns a task from a user.
	UnassignFromUser(tx *gorm.DB, taskId, userId int64) error
}

type taskRepository struct {
}

func (tr *taskRepository) Create(tx *gorm.DB, task *models.Task) (int64, error) {
	err := tx.Model(models.Task{}).Create(&task).Error
	if err != nil {
		return 0, err
	}
	return task.Id, nil
}

func (tr *taskRepository) GetByTitle(tx *gorm.DB, title string) (*models.Task, error) {
	task := &models.Task{}
	err := tx.Model(&models.Task{}).Where("title = ?", title).First(task).Error
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (tr *taskRepository) Get(tx *gorm.DB, taskId int64) (*models.Task, error) {
	task := &models.Task{}
	err := tx.Preload("Author").Preload("Groups").Model(&models.Task{}).Where("id = ? AND deleted_at IS NULL", taskId).First(task).Error
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (tr *taskRepository) GetAllAssigned(tx *gorm.DB, userId int64, limit, offset int, sort string) ([]models.Task, error) {
	var tasks []models.Task
	tx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, err
	}

	err = tx.Model(&models.Task{}).
		Joins("LEFT JOIN task_users ON task_users.task_id = tasks.id").
		Joins("LEFT JOIN task_groups ON task_groups.task_id = tasks.id").
		Joins("LEFT JOIN user_groups ON user_groups.group_id = task_groups.group_id").
		Where("(task_users.user_id = ? OR user_groups.user_id = ?) AND tasks.deleted_at IS NULL", userId, userId).
		Distinct().
		Find(&tasks).Error

	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (tr *taskRepository) AssignToUser(tx *gorm.DB, taskId, userId int64) error {
	taskUser := &models.TaskUser{
		TaskId: taskId,
		UserId: userId,
	}
	err := tx.Model(&models.TaskUser{}).Create(&taskUser).Error
	if err != nil {
		return err
	}
	return nil
}

func (tr *taskRepository) AssignToGroup(tx *gorm.DB, taskId, groupId int64) error {
	taskGroup := &models.TaskGroup{
		TaskId:  taskId,
		GroupId: groupId,
	}
	err := tx.Model(&models.TaskGroup{}).Create(&taskGroup).Error
	if err != nil {
		return err
	}
	return nil
}

func (tr *taskRepository) UnassignFromUser(tx *gorm.DB, taskId, userId int64) error {
	err := tx.Model(&models.TaskUser{}).Where("task_id = ? AND user_id = ?", taskId, userId).Delete(&models.TaskUser{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (tr *taskRepository) UnassignFromGroup(tx *gorm.DB, taskId, groupId int64) error {
	err := tx.Model(&models.TaskGroup{}).Where("task_id = ? AND group_id = ?", taskId, groupId).Delete(&models.TaskGroup{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (tr *taskRepository) GetAllCreated(tx *gorm.DB, userId int64, limit, offset int, sort string) ([]models.Task, error) {
	var tasks []models.Task
	tx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, err
	}

	err = tx.Model(&models.Task{}).Where("created_by_id = ? AND deleted_at IS NULL", userId).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (tr *taskRepository) IsAssignedToUser(tx *gorm.DB, taskId, userId int64) (bool, error) {
	var count int64

	err := tx.Model(&models.Task{}).
		Joins("LEFT JOIN task_users ON task_users.task_id = tasks.id").
		Joins("LEFT JOIN task_groups ON task_groups.task_id = tasks.id").
		Joins("LEFT JOIN user_groups ON user_groups.group_id = task_groups.group_id").
		Where("(task_users.task_id = ? AND task_users.user_id = ? OR task_groups.task_id = ? AND user_groups.user_id = ?) AND tasks.deleted_at IS NULL", taskId, userId, taskId, userId).
		Distinct().
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (tr *taskRepository) IsAssignedToGroup(tx *gorm.DB, taskId, groupId int64) (bool, error) {
	var count int64
	err := tx.Model(&models.Task{}).
		Joins("JOIN task_groups ON task_groups.task_id = tasks.id").
		Where("task_groups.task_id = ? AND task_groups.group_id = ? AND tasks.deleted_at IS NULL", taskId, groupId).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (tr *taskRepository) GetAll(tx *gorm.DB, limit, offset int, sort string) ([]models.Task, error) {
	tasks := []models.Task{}
	tx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, err
	}
	err = tx.Model(&models.Task{}).Where("deleted_at IS NULL").Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (tr *taskRepository) GetAllForGroup(tx *gorm.DB, groupId int64, limit, offset int, sort string) ([]models.Task, error) {
	var tasks []models.Task
	tx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, err
	}

	err = tx.Model(&models.Task{}).Joins("JOIN task_groups ON task_groups.task_id = tasks.id").
		Where("task_groups.group_id = ? AND tasks.deleted_at IS NULL", groupId).
		Find(&tasks).Error

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (tr *taskRepository) GetTimeLimits(tx *gorm.DB, taskId int64) ([]int64, error) {
	input_outputs := []models.InputOutput{}
	err := tx.Model(&models.InputOutput{}).Where("task_id = ?", taskId).Find(&input_outputs).Error
	if err != nil {
		return nil, err
	}
	// Sort by order
	timeLimits := make([]int64, len(input_outputs))
	for _, input_output := range input_outputs {
		timeLimits[input_output.Order-1] = input_output.TimeLimit
	}
	return timeLimits, nil
}

func (tr *taskRepository) GetMemoryLimits(tx *gorm.DB, taskId int64) ([]int64, error) {
	input_outputs := []models.InputOutput{}
	err := tx.Model(&models.InputOutput{}).Where("task_id = ?", taskId).Find(&input_outputs).Error
	if err != nil {
		return nil, err
	}
	// Sort by order
	memoryLimits := make([]int64, len(input_outputs))
	for _, input_output := range input_outputs {
		memoryLimits[input_output.Order-1] = input_output.MemoryLimit
	}
	return memoryLimits, nil
}

func (tr *taskRepository) Edit(tx *gorm.DB, taskId int64, task *models.Task) error {
	err := tx.Model(&models.Task{}).Where("id = ?", taskId).Updates(task).Error
	if err != nil {
		return err
	}
	return nil
}

func (tr *taskRepository) Delete(tx *gorm.DB, taskId int64) error {
	err := tx.Model(&models.Task{}).Where("id = ?", taskId).Update("deleted_at", time.Now()).Error
	if err != nil {
		return err
	}
	return nil
}

func NewTaskRepository(db *gorm.DB) (TaskRepository, error) {
	tables := []any{&models.Task{}, &models.InputOutput{}, &models.TaskUser{}}
	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			err := db.Migrator().CreateTable(table)
			if err != nil {
				return nil, err
			}
		}
	}

	return &taskRepository{}, nil
}
