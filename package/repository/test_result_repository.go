package repository

import (
	"fmt"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
)

type TestRepository interface {
	// Create creates a new test result in the database
	Create(db database.Database, testResult *models.TestResult) error
	Put(db database.Database, testResult *models.TestResult) error
	GetBySubmissionAndOrder(db database.Database, submissionID int64, order int) (*models.TestResult, error)
}

type testResultRepository struct{}

func (tr *testResultRepository) Create(db database.Database, testResult *models.TestResult) error {
	tx := db.GetInstance()
	err := tx.Create(&testResult).Error
	return err
}

func (tr *testResultRepository) GetBySubmissionAndOrder(db database.Database, submissionID int64, order int) (*models.TestResult, error) {
	tx := db.GetInstance()
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

func (tr *testResultRepository) Put(db database.Database, testResult *models.TestResult) error {
	tx := db.GetInstance()
	err := tx.Model(&models.TestResult{}).Where("id = ?", testResult.ID).Save(&testResult).Error
	return err
}

func NewTestResultRepository() TestRepository {
	return &testResultRepository{}
}
