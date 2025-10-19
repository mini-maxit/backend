package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type SubmissionResultRepository interface {
	// Create creates a new submission result in the database
	Create(tx *gorm.DB, solutionResult models.SubmissionResult) (int64, error)
	//
	Get(tx *gorm.DB, submissionResultID int64) (*models.SubmissionResult, error)

	GetBySubmission(tx *gorm.DB, submissionID int64) (*models.SubmissionResult, error)

	Put(tx *gorm.DB, submissionResult *models.SubmissionResult) error
}

type submissionResultRepository struct{}

func (usr *submissionResultRepository) Create(tx *gorm.DB, submissionResult models.SubmissionResult) (int64, error) {
	if err := tx.Create(&submissionResult).Error; err != nil {
		return 0, err
	}
	return submissionResult.ID, nil
}

func (usr *submissionResultRepository) Get(tx *gorm.DB, submissionResultID int64) (*models.SubmissionResult, error) {
	submissionResult := &models.SubmissionResult{}
	if err := tx.
		Preload("TestResult").
		Preload("TestResult.TestCase").
		Preload("TestResult.TestCase.InputFile").
		Preload("TestResult.TestCase.OutputFile").
		Preload("TestResult.StdoutFile").
		Preload("TestResult.StderrFile").
		Preload("TestResult.DiffFile").
		Where("id = ?", submissionResultID).First(submissionResult).Error; err != nil {
		return nil, err
	}
	return submissionResult, nil
}

func (usr *submissionResultRepository) GetBySubmission(tx *gorm.DB, submissionID int64) (*models.SubmissionResult, error) {
	submissionResult := &models.SubmissionResult{}
	if err := tx.Model(submissionResult).
		Preload("TestResult").
		Preload("TestResult.TestCase").
		Preload("TestResult.StdoutFile").
		Preload("TestResult.StderrFile").
		Preload("TestResult.DiffFile").
		Where("submission_id = ?", submissionID).First(submissionResult).Error; err != nil {
		return nil, err
	}
	return submissionResult, nil
}

func (usr *submissionResultRepository) Put(tx *gorm.DB, submissionResult *models.SubmissionResult) error {
	err := tx.Model(&models.SubmissionResult{}).Where("id = ?", submissionResult.ID).Save(submissionResult).Error
	return err
}

func NewSubmissionResultRepository(db *gorm.DB) (SubmissionResultRepository, error) {
	if !db.Migrator().HasTable(&models.SubmissionResult{}) {
		if err := db.Migrator().CreateTable(&models.SubmissionResult{}); err != nil {
			return nil, err
		}
	}
	return &submissionResultRepository{}, nil
}
