package repository

import (
	"fmt"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type TestRepository interface {
	// Create creates a new test result in the database
	Create(tx *gorm.DB, testResult *models.TestResult) error
	Put(tx *gorm.DB, testResult *models.TestResult) error
	GetBySubmissionAndOrder(tx *gorm.DB, submissionID int64, order int) (*models.TestResult, error)
}

type testResultRepository struct{}

func (tr *testResultRepository) Create(tx *gorm.DB, testResult *models.TestResult) error {
	err := tx.Create(&testResult).Error
	return err
}

func (tr *testResultRepository) GetBySubmissionAndOrder(tx *gorm.DB, submissionID int64, order int) (*models.TestResult, error) {
	testResult := &models.TestResult{}
	err := tx.Model(&models.TestResult{}).
		Joins(fmt.Sprintf("LEFT JOIN %s ON test_results.submission_result_id = submission_results.id", database.ResolveTableName(tx, &models.SubmissionResult{}))).
		Joins(fmt.Sprintf("LEFT JOIN %s ON test_results.test_case_id = test_cases.id", database.ResolveTableName(tx, &models.TestCase{}))).
		Where("submission_results.submission_id = ? AND test_cases.order = ?", submissionID, order).
		First(testResult).Error
	if err != nil {
		return nil, err
	}
	return testResult, nil
}

func (tr *testResultRepository) Put(t *gorm.DB, testResult *models.TestResult) error {
	err := t.Model(&models.TestResult{}).Where("id = ?", testResult.ID).Save(&testResult).Error
	return err
}

func NewTestResultRepository() TestRepository {
	return &testResultRepository{}
}
