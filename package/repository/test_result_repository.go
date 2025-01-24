package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type TestRepository interface {
	CreateTestResults(tx *gorm.DB, testResult models.TestResult) error
}

type testResultRepository struct{}

func (tr *testResultRepository) CreateTestResults(tx *gorm.DB, testResult models.TestResult) error {
	err := tx.Create(&testResult).Error
	return err
}

func NewTestResultRepository(db *gorm.DB) (TestRepository, error) {
	if !db.Migrator().HasTable(&models.TestResult{}) {
		err := db.Migrator().CreateTable(&models.TestResult{})
		if err != nil {
			return nil, err
		}
	}
	return &testResultRepository{}, nil
}
