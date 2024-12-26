package service

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"gorm.io/gorm"
)

type SubmissionService interface {
	MarkSubmissionFailed(tx *gorm.DB, submissionId int64, errorMsg string) error
	MarkSubmissionComplete(tx *gorm.DB, submissionId int64) error
	MarkSubmissionProcessing(tx *gorm.DB, submissionId int64) error
	CreateSubmissionResult(tx *gorm.DB, submissionId int64, responseMessage schemas.ResponseMessage) (int64, error)
}

type SubmissionServiceImpl struct {
	submissionRepository       repository.SubmissionRepository
	submissionResultRepository repository.SubmissionResultRepository
	inputOutputRepository      repository.InputOutputRepository
	testResultRepository       repository.TestResultRepository
}

func (us *SubmissionServiceImpl) MarkSubmissionFailed(tx *gorm.DB, submissionId int64, errorMsg string) error {
	err := us.submissionRepository.MarkSubmissionFailed(tx, submissionId, errorMsg)
	if err != nil {
		return err
	}
	return nil
}

func (us *SubmissionServiceImpl) MarkSubmissionComplete(tx *gorm.DB, submissionId int64) error {
	err := us.submissionRepository.MarkSubmissionComplete(tx, submissionId)
	if err != nil {
		return err
	}
	return nil
}

func (us *SubmissionServiceImpl) MarkSubmissionProcessing(tx *gorm.DB, submissionId int64) error {
	err := us.submissionRepository.MarkSubmissionProcessing(tx, submissionId)
	if err != nil {
		return err
	}
	return nil
}

func (us *SubmissionServiceImpl) CreateSubmissionResult(tx *gorm.DB, submissionId int64, responseMessage schemas.ResponseMessage) (int64, error) {
	submission, err := us.submissionRepository.GetSubmission(tx, submissionId)
	if err != nil {
		return -1, err
	}

	submissionResult := models.SubmissionResult{
		SubmissionId: submissionId,
		Code:         responseMessage.Result.Code,
		Message:      responseMessage.Result.Message,
	}
	id, err := us.submissionResultRepository.CreateSubmissionResult(tx, submissionResult)
	if err != nil {
		return -1, err
	}
	// Save test results
	for _, testResult := range responseMessage.Result.TestResults {
		inputOutputId, err := us.inputOutputRepository.GetInputOutputId(tx, submission.TaskId, testResult.Order)
		if err != nil {
			return -1, err
		}
		err = us.createTestResult(tx, id, inputOutputId, testResult)
		if err != nil {
			return -1, err
		}
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

func NewSubmissionService(submissionRepository repository.SubmissionRepository, submissionResultRepository repository.SubmissionResultRepository) SubmissionService {
	return &SubmissionServiceImpl{
		submissionRepository:       submissionRepository,
		submissionResultRepository: submissionResultRepository,
	}
}
