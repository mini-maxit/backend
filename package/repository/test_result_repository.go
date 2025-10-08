package repository

import (
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
		Where("submission_id = ? AND order = ?", submissionID, order).
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

func NewTestResultRepository(db *gorm.DB) (TestRepository, error) {
	if !db.Migrator().HasTable(&models.TestResult{}) {
		err := db.Migrator().CreateTable(&models.TestResult{})
		if err != nil {
			return nil, err
		}
	} else {
		err := db.Migrator().AutoMigrate(&models.TestResult{})
		if err != nil {
			return nil, err
		}
	}
	return &testResultRepository{}, nil
}
