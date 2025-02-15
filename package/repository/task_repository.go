package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/utils"
	"gorm.io/gorm"
)

type TaskRepository interface {
	// Create creates a new empty task and returns the task ID
	Create(tx *gorm.DB, task *models.Task) (int64, error)
	GetTask(tx *gorm.DB, taskId int64) (*models.Task, error)
	GetAllAssignedTasks(tx *gorm.DB, userId int64, limit, offset, sort string) ([]models.Task, error)
	GetAllCreatedTasks(tx *gorm.DB, userId int64, limit, offset, sort string) ([]models.Task, error)
	IsTaskAssignedToUser(tx *gorm.DB, taskId, userId int64) (bool, error)
	IsTaskAssignedToGroup(tx *gorm.DB, taskId, groupId int64) (bool, error)
	AssignTaskToUser(tx *gorm.DB, taskId, userId int64) error
	AssignTaskToGroup(tx *gorm.DB, taskId, groupId int64) error
	UnAssignTaskFromUser(tx *gorm.DB, taskId, userId int64) error
	UnAssignTaskFromGroup(tx *gorm.DB, taskId, groupId int64) error
	GetAllTasks(tx *gorm.DB, limit, offset, sort string) ([]models.Task, error)
	GetAllForGroup(tx *gorm.DB, groupId int64, limit, offset, sort string) ([]models.Task, error)
	GetTaskByTitle(tx *gorm.DB, title string) (*models.Task, error)
	GetTaskTimeLimits(tx *gorm.DB, taskId int64) ([]float64, error)
	GetTaskMemoryLimits(tx *gorm.DB, taskId int64) ([]float64, error)
	UpdateTask(tx *gorm.DB, taskId int64, task *models.Task) error
	DeleteTask(tx *gorm.DB, taskId int64) error
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

func (tr *taskRepository) GetTaskByTitle(tx *gorm.DB, title string) (*models.Task, error) {
	task := &models.Task{}
	err := tx.Model(&models.Task{}).Where("title = ?", title).First(task).Error
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (tr *taskRepository) GetTask(tx *gorm.DB, taskId int64) (*models.Task, error) {
	task := &models.Task{}
	err := tx.Preload("Author").Model(&models.Task{}).Where("id = ?", taskId).First(task).Error
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (tr *taskRepository) GetAllAssignedTasks(tx *gorm.DB, userId int64, limit, offset, sort string) ([]models.Task, error) {
	var tasks []models.Task
	tx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, err
	}

	err = tx.Model(&models.Task{}).
		Joins("JOIN task_user ON task_user.task_id = tasks.id").
		Where("task_user.user_id = ?", userId).
		Find(&tasks).Error

	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (tr *taskRepository) AssignTaskToUser(tx *gorm.DB, taskId, userId int64) error {
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

func (tr *taskRepository) AssignTaskToGroup(tx *gorm.DB, taskId, groupId int64) error {
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

func (tr *taskRepository) UnAssignTaskFromUser(tx *gorm.DB, taskId, userId int64) error {
	err := tx.Model(&models.TaskUser{}).Where("task_id = ? AND user_id = ?", taskId, userId).Delete(&models.TaskUser{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (tr *taskRepository) UnAssignTaskFromGroup(tx *gorm.DB, taskId, groupId int64) error {
	err := tx.Model(&models.TaskGroup{}).Where("task_id = ? AND group_id = ?", taskId, groupId).Delete(&models.TaskGroup{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (tr *taskRepository) GetAllCreatedTasks(tx *gorm.DB, userId int64, limit, offset, sort string) ([]models.Task, error) {
	var tasks []models.Task
	tx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, err
	}

	err = tx.Model(&models.Task{}).Where("created_by_id = ?", userId).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (tr *taskRepository) IsTaskAssignedToUser(tx *gorm.DB, taskId, userId int64) (bool, error) {
	var count int64
	err := tx.Model(&models.Task{}).
		Joins("JOIN task_user ON task_user.task_id = tasks.id").
		Where("task_user.task_id = ? AND task_user.user_id = ?", taskId, userId).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (tr *taskRepository) IsTaskAssignedToGroup(tx *gorm.DB, taskId, groupId int64) (bool, error) {
	var count int64
	err := tx.Model(&models.Task{}).
		Joins("JOIN task_group ON task_group.task_id = tasks.id").
		Where("task_group.task_id = ? AND task_group.group_id = ?", taskId, groupId).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (tr *taskRepository) GetAllTasks(tx *gorm.DB, limit, offset, sort string) ([]models.Task, error) {
	tasks := []models.Task{}
	tx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, err
	}
	err = tx.Model(&models.Task{}).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (tr *taskRepository) GetAllForGroup(tx *gorm.DB, groupId int64, limit, offset, sort string) ([]models.Task, error) {
	var tasks []models.Task
	tx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, err
	}

	err = tx.Joins("JOIN task_groups ON task_groups.task_id = tasks.id").
		Where("task_groups.group_id = ?", groupId).
		Find(&tasks).Error

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (tr *taskRepository) GetTaskTimeLimits(tx *gorm.DB, taskId int64) ([]float64, error) {
	input_outputs := []models.InputOutput{}
	err := tx.Model(&models.InputOutput{}).Where("id = ?", taskId).Find(&input_outputs).Error
	if err != nil {
		return nil, err
	}
	// Sort by order
	timeLimits := make([]float64, len(input_outputs))
	for _, input_output := range input_outputs {
		timeLimits[input_output.Order] = input_output.TimeLimit
	}
	return timeLimits, nil
}

func (tr *taskRepository) GetTaskMemoryLimits(tx *gorm.DB, taskId int64) ([]float64, error) {
	input_outputs := []models.InputOutput{}
	err := tx.Model(&models.InputOutput{}).Where("id = ?", taskId).Find(&input_outputs).Error
	if err != nil {
		return nil, err
	}
	// Sort by order
	memoryLimits := make([]float64, len(input_outputs))
	for _, input_output := range input_outputs {
		memoryLimits[input_output.Order] = input_output.MemoryLimit
	}
	return memoryLimits, nil
}

func (tr *taskRepository) UpdateTask(tx *gorm.DB, taskId int64, task *models.Task) error {
	err := tx.Model(&models.Task{}).Where("id = ?", taskId).Updates(task).Error
	if err != nil {
		return err
	}
	return nil
}

func (tr *taskRepository) DeleteTask(tx *gorm.DB, taskId int64) error {
	err := tx.Model(&models.Task{}).Where("id = ?", taskId).Delete(&models.Task{}).Error
	if err != nil {
		return err
	}
	return nil
}

func NewTaskRepository(db *gorm.DB) (TaskRepository, error) {
	tables := []interface{}{&models.Task{}, &models.InputOutput{}, &models.TaskUser{}}
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
