package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type TaskRepository interface {
	// CreateEmpty creates a new empty task and returns the task ID
	CreateEmpty(tx *gorm.DB, userId int64) (int64, error)
	GetTask(tx *gorm.DB, taskId int64) (*models.Task, error)
	GetTaskTimeLimits(tx *gorm.DB, taskId int64) ([]float64, error)
	GetTaskMemoryLimits(tx *gorm.DB, taskId int64) ([]float64, error)
	UpdateTask(tx *gorm.DB, taskId int64, task *models.Task) error
}

type TaskRepositoryImpl struct {
}

func (tr *TaskRepositoryImpl) CreateEmpty(tx *gorm.DB, userId int64) (int64, error) {
	emptyTask := &models.Task{CreatedBy: userId}
	err := tx.Model(&models.Task{}).Create(emptyTask).Error
	if err != nil {
		return 0, err
	}
	return emptyTask.Id, nil
}

func (tr *TaskRepositoryImpl) GetTask(tx *gorm.DB, taskId int64) (*models.Task, error) {
	task := &models.Task{}
	err := tx.Model(&models.Task{}).Where("id = ?", taskId).First(task).Error
	if err != nil {
		return nil, err
	}
	return task, nil
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
	tables := []interface{}{&models.Task{}, &models.InputOutput{}}
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
