package repository

import (
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
)

type TestCaseRepository interface {
	// Create creates a new input output record in the database
	Create(tx database.Database, testCase *models.TestCase) error
	// DeleteAll deletes all input output records for given task
	DeleteAll(tx database.Database, taskID int64) error
	// GetTestCaseID returns the ID of the input output record with the given task ID and order
	GetTestCaseID(db database.Database, taskID int64, order int) (int64, error)
	// GetByTask returns all input/output for given task
	GetByTask(db database.Database, taskID int64) ([]models.TestCase, error)
	// Get returns input/ouput with given ID
	Get(tx database.Database, ioID int64) (*models.TestCase, error)
	// Put updates input/output with given ID
	Put(tx database.Database, testCase *models.TestCase) error
}

type testCaseRepository struct{}

func (i *testCaseRepository) Create(tx database.Database, testCase *models.TestCase) error {
	err := tx.Create(testCase).Error()
	return err
}

func (i *testCaseRepository) GetTestCaseID(tx database.Database, taskID int64, order int) (int64, error) {
	var testCaseID int64
	err := tx.Model(&models.TestCase{}).Select("id").Where(
		`task_id = ? AND "order" = ?`,
		taskID,
		order,
	).Scan(&testCaseID).Error()
	return testCaseID, err
}

func (i *testCaseRepository) DeleteAll(tx database.Database, taskID int64) error {
	err := tx.Where("task_id = ?", taskID).Delete(&models.TestCase{}).Error()
	return err
}

func (i *testCaseRepository) GetByTask(tx database.Database, taskID int64) ([]models.TestCase, error) {
	testCase := []models.TestCase{}
	err := tx.Model(&models.TestCase{}).Where("task_id = ?", taskID).Find(&testCase).Error()
	if err != nil {
		return nil, err
	}

	return testCase, nil
}

func (i *testCaseRepository) Get(tx database.Database, ioID int64) (*models.TestCase, error) {
	testCase := &models.TestCase{}
	err := tx.Model(&models.TestCase{}).Where("id = ?", ioID).First(testCase).Error()
	if err != nil {
		return nil, err
	}

	return testCase, nil
}

func (i *testCaseRepository) Put(tx database.Database, testCase *models.TestCase) error {
	err := tx.Save(testCase).Error()
	return err
}

func NewTestCaseRepository() TestCaseRepository {
	return &testCaseRepository{}
}
