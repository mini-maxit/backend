package repository

import (
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
)

type SubmissionResultRepository interface {
	// Create creates a new submission result in the database
	Create(db database.Database, solutionResult models.SubmissionResult) (int64, error)
	//
	Get(db database.Database, submissionResultID int64) (*models.SubmissionResult, error)

	GetBySubmission(db database.Database, submissionID int64) (*models.SubmissionResult, error)

	Put(db database.Database, submissionResult *models.SubmissionResult) error
}

type submissionResultRepository struct{}

func (usr *submissionResultRepository) Create(db database.Database, submissionResult models.SubmissionResult) (int64, error) {
	tx := db.GetInstance()
	if err := tx.Create(&submissionResult).Error; err != nil {
		return 0, err
	}
	return submissionResult.ID, nil
}

func (usr *submissionResultRepository) Get(db database.Database, submissionResultID int64) (*models.SubmissionResult, error) {
	tx := db.GetInstance()
	submissionResult := &models.SubmissionResult{}
	if err := tx.
		Preload("TestResults").
		Preload("TestResults.TestCase").
		Preload("TestResults.TestCase.InputFile").
		Preload("TestResults.TestCase.OutputFile").
		Preload("TestResults.StdoutFile").
		Preload("TestResults.StderrFile").
		Preload("TestResults.DiffFile").
		Where("id = ?", submissionResultID).First(submissionResult).Error; err != nil {
		return nil, err
	}
	return submissionResult, nil
}

func (usr *submissionResultRepository) GetBySubmission(db database.Database, submissionID int64) (*models.SubmissionResult, error) {
	tx := db.GetInstance()
	submissionResult := &models.SubmissionResult{}
	if err := tx.Model(submissionResult).
		Preload("TestResults").
		Preload("TestResults.TestCase").
		Preload("TestResults.StdoutFile").
		Preload("TestResults.StderrFile").
		Preload("TestResults.DiffFile").
		Where("submission_id = ?", submissionID).First(submissionResult).Error; err != nil {
		return nil, err
	}
	return submissionResult, nil
}

func (usr *submissionResultRepository) Put(db database.Database, submissionResult *models.SubmissionResult) error {
	tx := db.GetInstance()
	err := tx.Model(&models.SubmissionResult{}).Where("id = ?", submissionResult.ID).Save(submissionResult).Error
	return err
}

func NewSubmissionResultRepository() SubmissionResultRepository {
	return &submissionResultRepository{}
}
