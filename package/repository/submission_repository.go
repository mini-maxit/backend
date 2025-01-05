package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type SubmissionRepository interface {
	GetSubmission(tx *gorm.DB, submissionId int64) (*models.Submission, error)
	CreateSubmission(tx *gorm.DB, submission *models.Submission) (int64, error)
	MarkSubmissionProcessing(tx *gorm.DB, submissionId int64) error
	MarkSubmissionComplete(tx *gorm.DB, submissionId int64) error
	MarkSubmissionFailed(db *gorm.DB, submissionId int64, errorMsg string) error
}

type SubmissionRepositoryImpl struct{}

func (us *SubmissionRepositoryImpl) GetSubmission(tx *gorm.DB, submissionId int64) (*models.Submission, error) {
	var submission models.Submission
	err := tx.Where("id = ?", submissionId).First(&submission).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

func (us *SubmissionRepositoryImpl) CreateSubmission(tx *gorm.DB, submission *models.Submission) (int64, error) {
	err := tx.Create(&submission).Error
	if err != nil {
		return 0, err
	}
	return submission.Id, nil
}

func (us *SubmissionRepositoryImpl) MarkSubmissionProcessing(tx *gorm.DB, submissionId int64) error {
	err := tx.Model(&models.Submission{}).Where("id = ?", submissionId).Update("status", "processing").Error
	return err
}

func (us *SubmissionRepositoryImpl) MarkSubmissionComplete(tx *gorm.DB, submissionId int64) error {
	err := tx.Model(&models.Submission{}).Where("id = ?", submissionId).Update("status", "completed").Error
	return err
}

func (us *SubmissionRepositoryImpl) MarkSubmissionFailed(db *gorm.DB, submissionId int64, errorMsg string) error {
	err := db.Model(&models.Submission{}).Where("id = ?", submissionId).Updates(map[string]interface{}{
		"status":         "failed",
		"status_message": errorMsg,
	}).Error
	return err
}

func NewSubmissionRepository(db *gorm.DB) (SubmissionRepository, error) {
	if !db.Migrator().HasTable(&models.Submission{}) {
		err := db.Migrator().CreateTable(&models.Submission{})
		if err != nil {
			return nil, err
		}
	}
	return &SubmissionRepositoryImpl{}, nil
}
