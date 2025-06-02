package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type SubmissionResultRepository interface {
	// Create creates a new submission result in the database
	Create(tx *gorm.DB, solutionResult models.SubmissionResult) (int64, error)
}

type submissionResultRepository struct{}

func (usr *submissionResultRepository) Create(tx *gorm.DB, submissionResult models.SubmissionResult) (int64, error) {
	if err := tx.Create(&submissionResult).Error; err != nil {
		return 0, err
	}
	return submissionResult.ID, nil
}

func NewSubmissionResultRepository(db *gorm.DB) (SubmissionResultRepository, error) {
	if !db.Migrator().HasTable(&models.SubmissionResult{}) {
		if err := db.Migrator().CreateTable(&models.SubmissionResult{}); err != nil {
			return nil, err
		}
	}
	return &submissionResultRepository{}, nil
}
