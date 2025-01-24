package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type SubmissionResultRepository interface {
	CreateSubmissionResult(tx *gorm.DB, solutionResult models.SubmissionResult) (int64, error)
}

type submissionResultRepository struct{}

// Store the result of the solution in the database
func (usr *submissionResultRepository) CreateSubmissionResult(tx *gorm.DB, submissionResult models.SubmissionResult) (int64, error) {
	if err := tx.Create(&submissionResult).Error; err != nil {
		return 0, err
	}
	return submissionResult.Id, nil
}

func NewSubmissionResultRepository(db *gorm.DB) (SubmissionResultRepository, error) {
	if !db.Migrator().HasTable(&models.SubmissionResult{}) {
		if err := db.Migrator().CreateTable(&models.SubmissionResult{}); err != nil {
			return nil, err
		}
	}
	return &submissionResultRepository{}, nil

}
