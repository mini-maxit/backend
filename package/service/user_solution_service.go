package service

import (
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"gorm.io/gorm"
)

type UserSolutionService interface {
	MarkSolutionFailed(submissionId int64, errorMsg string) error
	MarkSolutionCompleted(submissionId int64) error
	MarkUserSolutionProcessing(submissionId int64) error
	CreateUserSolutionResult(userSolutionResult schemas.ResponseMessage) (int64, error)
}

type UserSolutionServiceImpl struct {
	database               database.Database
	userSolutionRepository repository.UserSolutionRepository
	inputOutputRepository  repository.InputOutputRepository
	testResultRepository   repository.TestResultRepository
}

func (us *UserSolutionServiceImpl) MarkSolutionFailed(submissionId int64, errorMsg string) error {
	db := us.database.Connect()
	tx := db.Begin()

	err := us.userSolutionRepository.MarkUserSolutionFailed(tx, submissionId, errorMsg)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (us *UserSolutionServiceImpl) MarkSolutionCompleted(submissionId int64) error {
	db := us.database.Connect()
	tx := db.Begin()

	err := us.userSolutionRepository.MarkUserSolutionComplete(tx, submissionId)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (us *UserSolutionServiceImpl) MarkUserSolutionProcessing(submissionId int64) error {
	db := us.database.Connect()
	tx := db.Begin()

	err := us.userSolutionRepository.MarkUserSolutionProcessing(tx, submissionId)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (us *UserSolutionServiceImpl) CreateUserSolutionResult(userSolutionResult schemas.ResponseMessage) (int64, error) {
	db := us.database.Connect()
	tx := db.Begin()

	userSolutionResultModel := models.UserSolutionResult{
		UserSolutionId: userSolutionResult.UserSolutionId,
		Code:           userSolutionResult.Result.Code,
		Message:        userSolutionResult.Result.Message,
	}
	id, err := us.userSolutionRepository.CreateUserSolutionResult(tx, userSolutionResultModel)
	if err != nil {
		tx.Rollback()
		return -1, err
	}
	// Save test results
	for _, testResult := range userSolutionResult.Result.TestResults {
		inputOutputId, err := us.inputOutputRepository.GetInputOutputId(tx, userSolutionResult.TaskId, testResult.Order)
		err = us.createTestResult(tx, id, inputOutputId, testResult)
		if err != nil {
			tx.Rollback()
			return -1, err
		}
	}

	tx.Commit()
	return id, nil
}

func (us *UserSolutionServiceImpl) createTestResult(tx *gorm.DB, userSolutionResultId int64, inputOutputId int64, testResult schemas.TestResult) error {
	testResultModel := models.TestResult{
		UserSolutionResultId: userSolutionResultId,
		InputOutputId:        inputOutputId,
		Passed:               testResult.Passed,
		ErrorMessage:         testResult.ErrorMessage,
	}
	return us.testResultRepository.CreateTestResults(tx, testResultModel)
}

func NewUserSolutionService(database database.Database, userSolutionRepository repository.UserSolutionRepository) UserSolutionService {
	return &UserSolutionServiceImpl{
		database:               database,
		userSolutionRepository: userSolutionRepository,
	}
}
