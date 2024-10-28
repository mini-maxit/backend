package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type TaskRepository interface {
	// CreateEmpty creates a new empty task and returns the task ID
	CreateEmpty(tx *gorm.DB) (int64, error)
}

type TaskRepositoryImpl struct {
}

func (tr *TaskRepositoryImpl) CreateEmpty(tx *gorm.DB) (int64, error) {
	emptyTask := &models.Task{}
	err := tx.Model(&models.Task{}).Create(emptyTask).Error
	if err != nil {
		return 0, err
	}
	return emptyTask.ID, nil
}

func NewTaskRepository(db *gorm.DB) (TaskRepository, error) {
	if !db.Migrator().HasTable(&models.Task{}) {
		err := db.Migrator().CreateTable(&models.Task{})
		if err != nil {
			return nil, err
		}
	}

	return &TaskRepositoryImpl{}, nil
}
