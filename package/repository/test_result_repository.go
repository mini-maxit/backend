package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type TestResult interface {
	CreateTestResults(tx *gorm.DB, testResult models.TestResult) error
}

type TestResultRepository struct{}

func (tr *TestResultRepository) CreateTestResults(tx *gorm.DB, testResult models.TestResult) error {
	err := tx.Create(&testResult).Error
	return err
}

func NewTestResultRepository(db *gorm.DB) (TestResult, error) {
	if !db.Migrator().HasTable(&models.TestResult{}) {
		err := db.Migrator().CreateTable(&models.TestResult{})
		if err != nil {
			return nil, err
		}
	}
	return &TestResultRepository{}, nil
}
