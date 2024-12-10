package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type TaskRepository interface {
	// Create creates a new empty task and returns the task ID
	Create(tx *gorm.DB, task models.Task) (int64, error)
	GetTask(tx *gorm.DB, taskId int64) (*models.Task, error)
	GetAllTasks(tx *gorm.DB) ([]models.Task, error)
	GetAllForUser(tx *gorm.DB, userId int64) ([]models.Task, error)
	GetAllForGroup(tx *gorm.DB, groupId int64) ([]models.Task, error)
	GetTaskTimeLimits(tx *gorm.DB, taskId int64) ([]float64, error)
	GetTaskMemoryLimits(tx *gorm.DB, taskId int64) ([]float64, error)
	UpdateTask(tx *gorm.DB, taskId int64, task *models.Task) error
}

type TaskRepositoryImpl struct {
}

func (tr *TaskRepositoryImpl) Create(tx *gorm.DB, task models.Task) (int64, error) {
	err := tx.Model(&models.Task{}).Create(&task).Error
	if err != nil {
		return 0, err
	}
	return task.Id, nil
}

func (tr *TaskRepositoryImpl) GetTask(tx *gorm.DB, taskId int64) (*models.Task, error) {
	task := &models.Task{}
	err := tx.Preload("Author").Model(&models.Task{}).Where("id = ?", taskId).First(task).Error
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (tr *TaskRepositoryImpl) GetAllTasks(tx *gorm.DB) ([]models.Task, error) {
	tasks := []models.Task{}
	err := tx.Model(&models.Task{}).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (tr *TaskRepositoryImpl) GetAllForUser(tx *gorm.DB, userId int64) ([]models.Task, error) {
	var tasks []models.Task

	err := tx.Model(&models.Task{}).
		Joins("LEFT JOIN task_users ON task_users.task_id = tasks.id").
		Joins("LEFT JOIN task_groups ON task_groups.task_id = tasks.id").
		Joins("LEFT JOIN user_groups ON user_groups.group_id = task_groups.group_id").
		Where("task_users.user_id = ? OR user_groups.user_id = ?", userId, userId).
		Distinct().
		Find(&tasks).Error

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (tr * TaskRepositoryImpl) GetAllForGroup(tx *gorm.DB, groupId int64) ([]models.Task, error) {
	var tasks []models.Task

	err := tx.Joins("JOIN task_groups ON task_groups.task_id = tasks.id").
		Where("task_groups.group_id = ?", groupId).
		Find(&tasks).Error

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (tr *TaskRepositoryImpl) GetTaskTimeLimits(tx *gorm.DB, taskId int64) ([]float64, error) {
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

func (tr *TaskRepositoryImpl) GetTaskMemoryLimits(tx *gorm.DB, taskId int64) ([]float64, error) {
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

func (tr *TaskRepositoryImpl) UpdateTask(tx *gorm.DB, taskId int64, task *models.Task) error {
	err := tx.Model(&models.Task{}).Where("id = ?", taskId).Updates(task).Error
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

	return &TaskRepositoryImpl{}, nil
}
