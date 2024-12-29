package service

import (
	"errors"

	"github.com/mini-maxit/backend/internal/logger"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SubmissionService interface {
	MarkSubmissionFailed(tx *gorm.DB, submissionId int64, errorMsg string) error
	MarkSubmissionComplete(tx *gorm.DB, submissionId int64) error
	MarkSubmissionProcessing(tx *gorm.DB, submissionId int64) error
	CreateSubmissionResult(tx *gorm.DB, submissionId int64, responseMessage schemas.ResponseMessage) (int64, error)
	GetAll(tx *gorm.DB, limit, offset int64, user schemas.UserSession) ([]schemas.Submission, error)
	GetById(tx *gorm.DB, submissionId int64, user schemas.UserSession) (schemas.Submission, error)
	GetAllForUser(tx *gorm.DB, userId int64, limit, offset int64 ,user schemas.UserSession) ([]schemas.Submission, error)
	GetAllForGroup(tx *gorm.DB, groupId, limit, offset int64, user schemas.UserSession) ([]schemas.Submission, error)
	GetAllForTask(tx *gorm.DB, taskId, limit, offset int64, user schemas.UserSession) ([]schemas.Submission, error)
}

type SubmissionServiceImpl struct {
	submissionRepository       repository.SubmissionRepository
	submissionResultRepository repository.SubmissionResultRepository
	inputOutputRepository      repository.InputOutputRepository
	testResultRepository       repository.TestResultRepository
	userService 			  UserService
	taskService 			  TaskService
	languageService 		  LanguageService
	logger                     *zap.SugaredLogger
}

var ErrPermissionDenied = errors.New("User is not allowed to view this submission")

func (us *SubmissionServiceImpl) GetAll(tx *gorm.DB, limit int64, offset int64, user schemas.UserSession) ([]schemas.Submission, error) {
	submission_models := []models.Submission{}
	var err error

	switch user.Role {
	case "admin":
		submission_models, err = us.submissionRepository.GetAll(tx)
		break
	case "student":
		submission_models, err = us.submissionRepository.GetAllForStudent(tx, user.Id)
		break
	case "teacher":
		submission_models, err = us.submissionRepository.GetAllForTeacher(tx, user.Id)
		break
	}

	if err != nil {
		us.logger.Errorf("Error getting all submissions: %v", err.Error())
		return nil, err
	}

	var result []schemas.Submission
	for _, submission_model := range submission_models {
		result = append(result, *us.modelToSchema(&submission_model))
	}

	// Handle pagination
	if offset >= int64(len(result)) {
		return []schemas.Submission{}, nil
	}

	end := offset + limit
	if end > int64(len(result)) {
		end = int64(len(result))
	}

	return result[offset:end], nil
}

func (us *SubmissionServiceImpl) GetById(tx *gorm.DB, submissionId int64, user schemas.UserSession) (schemas.Submission, error) {
	submission_model, err := us.submissionRepository.GetSubmission(tx, submissionId)
	if err != nil {
		us.logger.Errorf("Error getting submission: %v", err.Error())
		return schemas.Submission{}, err
	}

	switch user.Role {
	case "admin":
		// Admin is allowed to view all submissions
		break
	case "student":
		// Student is only allowed to view their own submissions
		if submission_model.UserId != user.Id {
			us.logger.Errorf("User %v is not allowed to view submission %v", user.Id, submissionId)
			return schemas.Submission{}, ErrPermissionDenied
		}
		break
	case "teacher":
		// Teacher is only allowed to view submissions for tasks they created
		if submission_model.Task.CreatedBy != user.Id {
			us.logger.Errorf("User %v is not allowed to view submission %v", user.Id, submissionId)
			return schemas.Submission{}, ErrPermissionDenied
		}
		break
	}

	return *us.modelToSchema(submission_model), nil
}

func (us *SubmissionServiceImpl) GetAllForUser(tx *gorm.DB, userId int64, limit, offset int64 ,user schemas.UserSession) ([]schemas.Submission, error) {
	submission_models, err := us.submissionRepository.GetAllByUserId(tx, userId)
	if err != nil {
		us.logger.Errorf("Error getting all submissions for user: %v", err.Error())
		return nil, err
	}

	switch user.Role {
	case "admin":
		// Admin is allowed to view all submissions
		break
	case "student":
		// Student is only allowed to view their own submissions
		if userId != user.Id {
			us.logger.Errorf("User %v is not allowed to view submissions", user.Id)
			return []schemas.Submission{}, ErrPermissionDenied
		}
		break
	case "teacher":
		// Teacher is only allowed to view submissions for tasks they created
		for i, submission := range submission_models {
			if submission.Task.CreatedBy != user.Id {
				submission_models = append(submission_models[:i], submission_models[i+1:]...)
			}
		}
		break
	}

	var result []schemas.Submission
	for _, submission_model := range submission_models {
		result = append(result, *us.modelToSchema(&submission_model))
	}

	// Handle pagination
	if offset >= int64(len(result)) {
		return []schemas.Submission{}, nil
	}

	end := offset + limit
	if end > int64(len(result)) {
		end = int64(len(result))
	}

	return result[offset:end], nil
}

func (us *SubmissionServiceImpl) GetAllForGroup(tx *gorm.DB, groupId int64, limit, offset int64 ,user schemas.UserSession) ([]schemas.Submission, error) {
	var err error
	submission_models := []models.Submission{}

	switch user.Role {
	case "admin":
		// Admin is allowed to view all submissions
		submission_models, err = us.submissionRepository.GetAllForGroup(tx, groupId)
		break
	case "student":
		// Student is only allowed to view their own submissions
		return []schemas.Submission{}, ErrPermissionDenied
	case "teacher":
		// Teacher is only allowed to view submissions for tasks they created
		submission_models, err = us.submissionRepository.GetAllForGroupTeacher(tx, groupId, user.Id)
	}

	if err != nil {
		us.logger.Errorf("Error getting all submissions for group: %v", err.Error())
		return nil, err
	}

	var result []schemas.Submission
	for _, submission_model := range submission_models {
		result = append(result, *us.modelToSchema(&submission_model))
	}

	// Handle pagination
	if offset >= int64(len(result)) {
		return []schemas.Submission{}, nil
	}

	end := offset + limit
	if end > int64(len(result)) {
		end = int64(len(result))
	}

	return result[offset:end], nil
}

func (us *SubmissionServiceImpl) GetAllForTask(tx *gorm.DB, taskId, limit, offset int64, user schemas.UserSession) ([]schemas.Submission, error) {
	var err error
	submissions_model := []models.Submission{}

	switch user.Role {
	case "admin":
		submissions_model, err = us.submissionRepository.GetAllForTask(tx, taskId)
		break

	case "teacher":
		submissions_model, err = us.submissionRepository.GetAllForTaskTeacher(tx, taskId, user.Id)
		break

	case "student":
		submissions_model, err = us.submissionRepository.GetAllForTaskStudent(tx, taskId, user.Id)
		break
	}

	if err != nil {
		return nil, err
	}

	response := []schemas.Submission{}
	for _, submission := range submissions_model {
		response = append(response, *us.modelToSchema(&submission))
	}

	// Handle pagination
	if offset >= int64(len(response)) {
		return []schemas.Submission{}, nil
	}

	end := offset + limit
	if end > int64(len(response)) {
		end = int64(len(response))
	}

	return response[offset:end], nil
}


func (us *SubmissionServiceImpl) MarkSubmissionFailed(tx *gorm.DB, submissionId int64, errorMsg string) error {
	err := us.submissionRepository.MarkSubmissionFailed(tx, submissionId, errorMsg)
	if err != nil {
		us.logger.Errorf("Error marking submission failed: %v", err.Error())
		return err
	}

	return nil
}

func (us *SubmissionServiceImpl) MarkSubmissionComplete(tx *gorm.DB, submissionId int64) error {
	err := us.submissionRepository.MarkSubmissionComplete(tx, submissionId)
	if err != nil {
		us.logger.Errorf("Error marking submission complete: %v", err.Error())
		return err
	}

	return nil
}

func (us *SubmissionServiceImpl) MarkSubmissionProcessing(tx *gorm.DB, submissionId int64) error {
	err := us.submissionRepository.MarkSubmissionProcessing(tx, submissionId)
	if err != nil {
		us.logger.Errorf("Error marking submission processing: %v", err.Error())
		return err
	}

	return nil
}

func (us *SubmissionServiceImpl) CreateSubmissionResult(tx *gorm.DB, submissionId int64, responseMessage schemas.ResponseMessage) (int64, error) {
	submission, err := us.submissionRepository.GetSubmission(tx, submissionId)
	if err != nil {
		us.logger.Errorf("Error getting submission: %v", err.Error())
		return -1, err
	}

	submissionResult := models.SubmissionResult{
		SubmissionId: submissionId,
		Code:         responseMessage.Result.Code,
		Message:      responseMessage.Result.Message,
	}
	id, err := us.submissionResultRepository.CreateSubmissionResult(tx, submissionResult)
	if err != nil {
		us.logger.Errorf("Error creating submission result: %v", err.Error())
		return -1, err
	}
	// Save test results
	for _, testResult := range responseMessage.Result.TestResults {
		inputOutputId, err := us.inputOutputRepository.GetInputOutputId(tx, submission.TaskId, testResult.Order)
		if err != nil {
			us.logger.Errorf("Error getting input output id: %v", err.Error())
			return -1, err
		}
		err = us.createTestResult(tx, id, inputOutputId, testResult)
		if err != nil {
			us.logger.Errorf("Error creating test result: %v", err.Error())
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

func NewSubmissionService(submissionRepository repository.SubmissionRepository, submissionResultRepository repository.SubmissionResultRepository, languageService LanguageService, taskService TaskService, userService UserService) SubmissionService {
	log := logger.NewNamedLogger("submission_service")
	return &SubmissionServiceImpl{
		submissionRepository:       submissionRepository,
		submissionResultRepository: submissionResultRepository,
		languageService: languageService,
		taskService: taskService,
		userService: userService,
		logger:                     log,
	}
}

func (us *SubmissionServiceImpl) modelToSchema(submission *models.Submission) *schemas.Submission {
	return &schemas.Submission{
		Id:            submission.Id,
		TaskId:        submission.TaskId,
		UserId:        submission.UserId,
		Order:         submission.Order,
		LanguageId:    submission.LanguageId,
		Status:        submission.Status,
		StatusMessage: submission.StatusMessage,
		SubmittedAt:   submission.SubmittedAt,
		CheckedAt:     submission.CheckedAt,
		Language: 	   *us.languageService.modelToSchema(&submission.Language),
		Task: 		   *us.taskService.modelToSchema(&submission.Task),
		User: 		   *us.userService.modelToSchema(&submission.User),
	}
}
