package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type TestCaseRepository interface {
	// Create creates a new input output record in the database
	Create(tx *gorm.DB, inputOutput *models.TestCase) error
	// DeleteAll deletes all input output records for given task
	DeleteAll(tx *gorm.DB, taskID int64) error
	// GetInputOutputID returns the ID of the input output record with the given task ID and order
	GetInputOutputID(db *gorm.DB, taskID int64, order int) (int64, error)
	// GetByTask returns all input/output for given task
	GetByTask(db *gorm.DB, taskID int64) ([]models.TestCase, error)
	// Get returns input/ouput with given ID
	Get(tx *gorm.DB, ioID int64) (*models.TestCase, error)
	// Put updates input/output with given ID
	Put(tx *gorm.DB, inputOutput *models.TestCase) error
}

type testCaseRepository struct{}

func (i *testCaseRepository) Create(tx *gorm.DB, inputOutput *models.TestCase) error {
	err := tx.Create(inputOutput).Error
	return err
}

func (i *testCaseRepository) GetInputOutputID(tx *gorm.DB, taskID int64, order int) (int64, error) {
	var inputOutputID int64
	err := tx.Model(&models.TestCase{}).Select("id").Where(
		`task_id = ? AND "order" = ?`,
		taskID,
		order,
	).Scan(&inputOutputID).Error
	return inputOutputID, err
}

func (i *testCaseRepository) DeleteAll(tx *gorm.DB, taskID int64) error {
	err := tx.Where("task_id = ?", taskID).Delete(&models.TestCase{}).Error
	return err
}

func (i *testCaseRepository) GetByTask(tx *gorm.DB, taskID int64) ([]models.TestCase, error) {
	inputOutput := []models.TestCase{}
	err := tx.Model(&models.TestCase{}).Where("task_id = ?", taskID).Find(&inputOutput).Error
	if err != nil {
		return nil, err
	}

	return inputOutput, nil
}

func (i *testCaseRepository) Get(tx *gorm.DB, ioID int64) (*models.TestCase, error) {
	inputOutput := &models.TestCase{}
	err := tx.Model(&models.TestCase{}).Where("id = ?", ioID).First(inputOutput).Error
	if err != nil {
		return nil, err
	}

	return inputOutput, nil
}

func (i *testCaseRepository) Put(tx *gorm.DB, inputOutput *models.TestCase) error {
	err := tx.Save(inputOutput).Error
	return err
}

func NewInputOutputRepository(db *gorm.DB) (TestCaseRepository, error) {
	if !db.Migrator().HasTable(&models.TestCase{}) {
		err := db.Migrator().CreateTable(&models.TestCase{})
		if err != nil {
			return nil, err
		}
	}
	return &testCaseRepository{}, nil
}
