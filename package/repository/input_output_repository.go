package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type InputOutputRepository interface {
	// Create creates a new input output record in the database
	Create(tx *gorm.DB, inputOutput *models.InputOutput) error
	// DeleteAll deletes all input output records for given task
	DeleteAll(tx *gorm.DB, taskId int64) error
	// GetInputOutputId returns the id of the input output record with the given task id and order
	GetInputOutputId(db *gorm.DB, taskId int64, order int64) (int64, error)
}

type inputOutputRepository struct{}

func (i *inputOutputRepository) Create(tx *gorm.DB, inputOutput *models.InputOutput) error {
	err := tx.Create(inputOutput).Error
	return err
}

func (i *inputOutputRepository) GetInputOutputId(tx *gorm.DB, taskId int64, order int64) (int64, error) {
	var input_output_id int64
	err := tx.Table("input_outputs").Select("id").Where("task_id = ? AND \"order\" = ?", taskId, order).Scan(&input_output_id).Error
	return input_output_id, err
}

func (i *inputOutputRepository) DeleteAll(tx *gorm.DB, taskId int64) error {
	err := tx.Where("task_id = ?", taskId).Delete(&models.InputOutput{}).Error
	return err
}

func NewInputOutputRepository(db *gorm.DB) (InputOutputRepository, error) {
	if !db.Migrator().HasTable(&models.InputOutput{}) {
		err := db.Migrator().CreateTable(&models.InputOutput{})
		if err != nil {
			return nil, err
		}
	}
	return &inputOutputRepository{}, nil
}
