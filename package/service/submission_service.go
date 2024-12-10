package service

import (
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"gorm.io/gorm"
)

type SubmissionService interface {
	MarkSubmissionFailed(submissionId int64, errorMsg string) error
	MarkSubmissionComplete(submissionId int64) error
	MarkSubmissionProcessing(submissionId int64) error
	CreateSubmissionResult(submissionId int64, responseMessage schemas.ResponseMessage) (int64, error)
}

type SubmissionServiceImpl struct {
	database                   database.Database
	submissionRepository       repository.SubmissionRepository
	submissionResultRepository repository.SubmissionResultRepository
	inputOutputRepository      repository.InputOutputRepository
	testResultRepository       repository.TestResultRepository
}

func (us *SubmissionServiceImpl) MarkSubmissionFailed(submissionId int64, errorMsg string) error {
	db := us.database.Connect()
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer utils.TransactionPanicRecover(tx)

	err := us.submissionRepository.MarkSubmissionFailed(tx, submissionId, errorMsg)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

func (us *SubmissionServiceImpl) MarkSubmissionComplete(submissionId int64) error {
	db := us.database.Connect()
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer utils.TransactionPanicRecover(tx)

	err := us.submissionRepository.MarkSubmissionComplete(tx, submissionId)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

func (us *SubmissionServiceImpl) MarkSubmissionProcessing(submissionId int64) error {
	db := us.database.Connect()
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer utils.TransactionPanicRecover(tx)

	err := us.submissionRepository.MarkSubmissionProcessing(tx, submissionId)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

func (us *SubmissionServiceImpl) CreateSubmissionResult(submissionId int64, responseMessage schemas.ResponseMessage) (int64, error) {
	db := us.database.Connect()
	tx := db.Begin()
	if tx.Error != nil {
		return -1, tx.Error
	}

	defer utils.TransactionPanicRecover(tx)

	submission, err := us.submissionRepository.GetSubmission(tx, submissionId)
	if err != nil {
		tx.Rollback()
		return -1, err
	}

	submissionResult := models.SubmissionResult{
		SubmissionId: submissionId,
		Code:         responseMessage.Result.Code,
		Message:      responseMessage.Result.Message,
	}
	id, err := us.submissionResultRepository.CreateSubmissionResult(tx, submissionResult)
	if err != nil {
		tx.Rollback()
		return -1, err
	}
	// Save test results
	for _, testResult := range responseMessage.Result.TestResults {
		inputOutputId, err := us.inputOutputRepository.GetInputOutputId(tx, submission.TaskId, testResult.Order)
		if err != nil {
			tx.Rollback()
			return -1, err
		}
		err = us.createTestResult(tx, id, inputOutputId, testResult)
		if err != nil {
			tx.Rollback()
			return -1, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return -1, err
	}
	return id, nil
}

func (us *SubmissionServiceImpl) createTestResult(tx *gorm.DB, submissionResultId int64, inputOutputId int64, testResult schemas.TestResult) error {
	testResultModel := models.TestResult{
		SubmissionResultId: submissionResultId,
		InputOutputId:      inputOutputId,
		Passed:             testResult.Passed,
		ErrorMessage:       testResult.ErrorMessage,
	}
	return us.testResultRepository.CreateTestResults(tx, testResultModel)
}

func NewSubmissionService(database database.Database, submissionRepository repository.SubmissionRepository, submissionResultRepository repository.SubmissionResultRepository) SubmissionService {
	return &SubmissionServiceImpl{
		database:                   database,
		submissionRepository:       submissionRepository,
		submissionResultRepository: submissionResultRepository,
	}
}
