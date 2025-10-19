package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type TestCaseRepository interface {
	// Create creates a new input output record in the database
	Create(tx *gorm.DB, testCase *models.TestCase) error
	// DeleteAll deletes all input output records for given task
	DeleteAll(tx *gorm.DB, taskID int64) error
	// GetTestCaseID returns the ID of the input output record with the given task ID and order
	GetTestCaseID(db *gorm.DB, taskID int64, order int) (int64, error)
	// GetByTask returns all input/output for given task
	GetByTask(db *gorm.DB, taskID int64) ([]models.TestCase, error)
	// Get returns input/ouput with given ID
	Get(tx *gorm.DB, ioID int64) (*models.TestCase, error)
	// Put updates input/output with given ID
	Put(tx *gorm.DB, testCase *models.TestCase) error
}

type testCaseRepository struct{}

func (i *testCaseRepository) Create(tx *gorm.DB, testCase *models.TestCase) error {
	err := tx.Create(testCase).Error
	return err
}

func (i *testCaseRepository) GetTestCaseID(tx *gorm.DB, taskID int64, order int) (int64, error) {
	var testCaseID int64
	err := tx.Model(&models.TestCase{}).Select("id").Where(
		`task_id = ? AND "order" = ?`,
		taskID,
		order,
	).Scan(&testCaseID).Error
	return testCaseID, err
}

func (i *testCaseRepository) DeleteAll(tx *gorm.DB, taskID int64) error {
	err := tx.Where("task_id = ?", taskID).Delete(&models.TestCase{}).Error
	return err
}

func (i *testCaseRepository) GetByTask(tx *gorm.DB, taskID int64) ([]models.TestCase, error) {
	testCase := []models.TestCase{}
	err := tx.Model(&models.TestCase{}).Where("task_id = ?", taskID).Find(&testCase).Error
	if err != nil {
		return nil, err
	}

	return testCase, nil
}

func (i *testCaseRepository) Get(tx *gorm.DB, ioID int64) (*models.TestCase, error) {
	testCase := &models.TestCase{}
	err := tx.Model(&models.TestCase{}).Where("id = ?", ioID).First(testCase).Error
	if err != nil {
		return nil, err
	}

	return testCase, nil
}

func (i *testCaseRepository) Put(tx *gorm.DB, testCase *models.TestCase) error {
	err := tx.Save(testCase).Error
	return err
}

func NewTestCaseRepository(db *gorm.DB) (TestCaseRepository, error) {
	if !db.Migrator().HasTable(&models.TestCase{}) {
		err := db.Migrator().CreateTable(&models.TestCase{})
		if err != nil {
			return nil, err
		}
	}
	return &testCaseRepository{}, nil
}
