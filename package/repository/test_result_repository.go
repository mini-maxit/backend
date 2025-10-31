package repository

import (
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
)

type TestRepository interface {
	// Create creates a new test result in the database
	Create(tx *database.DB, testResult *models.TestResult) error
	Put(tx *database.DB, testResult *models.TestResult) error
	GetBySubmissionAndOrder(tx *database.DB, submissionID int64, order int) (*models.TestResult, error)
}

type testResultRepository struct{}

func (tr *testResultRepository) Create(tx *database.DB, testResult *models.TestResult) error {
	err := tx.Create(&testResult).Error()
	return err
}

func (tr *testResultRepository) GetBySubmissionAndOrder(tx *database.DB, submissionID int64, order int) (*models.TestResult, error) {
	testResult := &models.TestResult{}
	err := tx.Model(&models.TestResult{}).
		Joins("LEFT JOIN submission_results ON test_results.submission_result_id = submission_results.id").
		Joins("LEFT JOIN test_cases ON test_results.test_case_id = test_cases.id").
		Where("submission_results.submission_id = ? AND test_cases.order = ?", submissionID, order).
		First(testResult).Error()
	if err != nil {
		return nil, err
	}
	return testResult, nil
}

func (tr *testResultRepository) Put(t *database.DB, testResult *models.TestResult) error {
	err := t.Model(&models.TestResult{}).Where("id = ?", testResult.ID).Save(&testResult).Error()
	return err
}

func NewTestResultRepository() TestRepository {
	return &testResultRepository{}
}
