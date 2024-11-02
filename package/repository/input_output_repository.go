package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type InputOutput interface {
	GetInputOutputId(db *gorm.DB, taskId int64, order int64) (int64, error)
}

type InputOutputRepository struct{}

func (i *InputOutputRepository) GetInputOutputId(tx *gorm.DB, taskId int64, order int64) (int64, error) {
	var input_output_id int64
	err := tx.Table("input_outputs").Select("id").Where("task_id = ? AND input_output_order = ?", taskId, order).Scan(&input_output_id).Error
	return input_output_id, err
}

func NewInputOutputRepository(db *gorm.DB) (InputOutput, error) {
	if !db.Migrator().HasTable(&models.InputOutput{}) {
		err := db.Migrator().CreateTable(&models.InputOutput{})
		if err != nil {
			return nil, err
		}
	}
	return &InputOutputRepository{}, nil
}
