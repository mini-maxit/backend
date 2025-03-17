package service

import (
	"encoding/json"
	"strconv"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SubmissionService interface {
	MarkSubmissionFailed(tx *gorm.DB, submissionId int64, errorMsg string) error
	MarkSubmissionComplete(tx *gorm.DB, submissionId int64) error
	MarkSubmissionProcessing(tx *gorm.DB, submissionId int64) error
	CreateSubmission(tx *gorm.DB, taskId int64, userId int64, languageId int64, order int64) (int64, error)
	CreateSubmissionResult(tx *gorm.DB, submissionId int64, responseMessage schemas.QueueResponseMessage) (int64, error)
	GetAll(tx *gorm.DB, user schemas.User, queryParams map[string]interface{}) ([]schemas.Submission, error)
	GetById(tx *gorm.DB, submissionId int64, user schemas.User) (schemas.Submission, error)
	GetAllForUser(tx *gorm.DB, userId int64, user schemas.User, queryParams map[string]interface{}) ([]schemas.Submission, error)
	GetAllForUserShort(tx *gorm.DB, userId int64, user schemas.User, queryParams map[string]interface{}) ([]schemas.SubmissionShort, error)
	GetAllForGroup(tx *gorm.DB, groupId int64, user schemas.User, queryParams map[string]interface{}) ([]schemas.Submission, error)
	GetAllForTask(tx *gorm.DB, taskId int64, user schemas.User, queryParams map[string]interface{}) ([]schemas.Submission, error)
	GetAvailableLanguages(tx *gorm.DB) ([]schemas.LanguageConfig, error)
}

type submissionService struct {
	submissionRepository       repository.SubmissionRepository
	submissionResultRepository repository.SubmissionResultRepository
	inputOutputRepository      repository.InputOutputRepository
	testResultRepository       repository.TestRepository
	userService                UserService
	taskService                TaskService
	languageService            LanguageService
	logger                     *zap.SugaredLogger
}

func (us *submissionService) GetAll(tx *gorm.DB, user schemas.User, queryParams map[string]interface{}) ([]schemas.Submission, error) {
	submission_models := []models.Submission{}
	var err error

	limit := queryParams["limit"].(uint64)
	offset := queryParams["offset"].(uint64)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = "submitted_at:desc"
	}

	switch user.Role {
	case "admin":
		submission_models, err = us.submissionRepository.GetAll(tx, int(limit), int(offset), sort)
	case "student":
		submission_models, err = us.submissionRepository.GetAllForStudent(tx, user.Id, int(limit), int(offset), sort)
	case "teacher":
		submission_models, err = us.submissionRepository.GetAllForTeacher(tx, user.Id, int(limit), int(offset), sort)
	}

	if err != nil {
		us.logger.Errorf("Error getting all submissions: %v", err.Error())
		return nil, err
	}

	var result []schemas.Submission
	for _, submission_model := range submission_models {
		result = append(result, *us.modelToSchema(&submission_model))
	}

	return result, nil
}

func (us *submissionService) GetById(tx *gorm.DB, submissionId int64, user schemas.User) (schemas.Submission, error) {
	submission_model, err := us.submissionRepository.GetSubmission(tx, submissionId)
	if err != nil {
		us.logger.Errorf("Error getting submission: %v", err.Error())
		return schemas.Submission{}, err
	}

	switch user.Role {
	case "admin":
		// Admin is allowed to view all submissions
	case "student":
		// Student is only allowed to view their own submissions
		if submission_model.UserId != user.Id {
			us.logger.Errorf("User %v is not allowed to view submission %v", user.Id, submissionId)
			return schemas.Submission{}, errors.ErrPermissionDenied
		}
	case "teacher":
		// Teacher is only allowed to view submissions for tasks they created
		if submission_model.Task.CreatedBy != user.Id {
			us.logger.Errorf("User %v is not allowed to view submission %v", user.Id, submissionId)
			return schemas.Submission{}, errors.ErrPermissionDenied
		}
	}

	return *us.modelToSchema(submission_model), nil
}

func (us *submissionService) GetAllForUser(tx *gorm.DB, userId int64, currentUser schemas.User, queryParams map[string]interface{}) ([]schemas.Submission, error) {
	limit := queryParams["limit"].(uint64)
	offset := queryParams["offset"].(uint64)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = "submitted_at:desc"
	}

	submission_models, err := us.submissionRepository.GetAllByUserId(tx, userId, int(limit), int(offset), sort)
	if err != nil {
		us.logger.Errorf("Error getting all submissions for user: %v", err.Error())
		return nil, err
	}

	switch currentUser.Role {
	case "admin":
		// Admin is allowed to view all submissions
	case "student":
		// Student is only allowed to view their own submissions
		if userId != currentUser.Id {
			us.logger.Errorf("User %v is not allowed to view submissions", currentUser.Id)
			return nil, errors.ErrPermissionDenied
		}
	case "teacher":
		// Teacher is only allowed to view submissions for tasks they created
		for i, submission := range submission_models {
			if submission.Task.CreatedBy != currentUser.Id {
				submission_models = append(submission_models[:i], submission_models[i+1:]...)
			}
		}
	}

	var result []schemas.Submission
	for _, submission_model := range submission_models {
		result = append(result, *us.modelToSchema(&submission_model))
	}

	return result, nil
}

func (us *submissionService) GetAllForUserShort(tx *gorm.DB, userId int64, currentUser schemas.User, queryParams map[string]interface{}) ([]schemas.SubmissionShort, error) {
	submission_models, err := us.GetAllForUser(tx, userId, currentUser, queryParams)
	if err != nil {
		return nil, err
	}
	result := []schemas.SubmissionShort{}
	for _, submission := range submission_models {
		passed := true
		how_many := len(submission.Result.TestResults)
		for _, test_result := range submission.Result.TestResults {
			if !test_result.Passed {
				passed = false
				how_many--
			}
		}
		result = append(result, schemas.SubmissionShort{
			Id:            submission.Id,
			TaskId:        submission.TaskId,
			UserId:        submission.UserId,
			Passed:        passed,
			HowManyPassed: int64(how_many),
		})
	}

	return result, nil
}

func (us *submissionService) GetAllForGroup(tx *gorm.DB, groupId int64, user schemas.User, queryParams map[string]interface{}) ([]schemas.Submission, error) {
	var err error
	submission_models := []models.Submission{}

	limit := queryParams["limit"].(uint64)
	offset := queryParams["offset"].(uint64)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = "submitted_at:desc"
	}

	switch user.Role {
	case "admin":
		// Admin is allowed to view all submissions
		submission_models, err = us.submissionRepository.GetAllForGroup(tx, groupId, int(limit), int(offset), sort)
	case "student":
		// Student is only allowed to view their own submissions
		return nil, errors.ErrPermissionDenied
	case "teacher":
		// Teacher is only allowed to view submissions for tasks they created
		submission_models, err = us.submissionRepository.GetAllForGroupTeacher(tx, groupId, user.Id, int(limit), int(offset), sort)
	}

	if err != nil {
		us.logger.Errorf("Error getting all submissions for group: %v", err.Error())
		return nil, err
	}

	var result []schemas.Submission
	for _, submission_model := range submission_models {
		result = append(result, *us.modelToSchema(&submission_model))
	}

	return result, nil
}

func (us *submissionService) GetAllForTask(tx *gorm.DB, taskId int64, user schemas.User, queryParams map[string]interface{}) ([]schemas.Submission, error) {
	var err error
	submissions_model := []models.Submission{}

	limit := queryParams["limit"].(uint64)
	offset := queryParams["offset"].(uint64)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = "submitted_at:desc"
	}

	switch user.Role {
	case "admin":
		submissions_model, err = us.submissionRepository.GetAllForTask(tx, taskId, int(limit), int(offset), sort)

	case "teacher":
		submissions_model, err = us.submissionRepository.GetAllForTaskTeacher(tx, taskId, user.Id, int(limit), int(offset), sort)

	case "student":
		submissions_model, err = us.submissionRepository.GetAllForTaskStudent(tx, taskId, user.Id, int(limit), int(offset), sort)
	}

	if err != nil {
		return nil, err
	}

	result := []schemas.Submission{}
	for _, submission := range submissions_model {
		result = append(result, *us.modelToSchema(&submission))
	}

	return result, nil
}

func (us *submissionService) MarkSubmissionFailed(tx *gorm.DB, submissionId int64, errorMsg string) error {
	err := us.submissionRepository.MarkSubmissionFailed(tx, submissionId, errorMsg)
	if err != nil {
		us.logger.Errorf("Error marking submission failed: %v", err.Error())
		return err
	}

	return nil
}

func (us *submissionService) MarkSubmissionComplete(tx *gorm.DB, submissionId int64) error {
	err := us.submissionRepository.MarkSubmissionComplete(tx, submissionId)
	if err != nil {
		us.logger.Errorf("Error marking submission complete: %v", err.Error())
		return err
	}

	return nil
}

func (us *submissionService) MarkSubmissionProcessing(tx *gorm.DB, submissionId int64) error {
	err := us.submissionRepository.MarkSubmissionProcessing(tx, submissionId)
	if err != nil {
		us.logger.Errorf("Error marking submission processing: %v", err.Error())
		return err
	}

	return nil
}

func (us *submissionService) CreateSubmission(tx *gorm.DB, taskId int64, userId int64, languageId int64, order int64) (int64, error) {
	// Create a new submission
	submission := &models.Submission{
		TaskId:     taskId,
		UserId:     userId,
		Order:      order,
		LanguageId: languageId,
		Status:     models.StatusReceived,
	}
	submissionId, err := us.submissionRepository.CreateSubmission(tx, submission)

	if err != nil {
		us.logger.Errorf("Error creating submission: %v", err.Error())
		return 0, err
	}

	return submissionId, nil
}

func (us *submissionService) CreateSubmissionResult(tx *gorm.DB, submissionId int64, responseMessage schemas.QueueResponseMessage) (int64, error) {
	submission, err := us.submissionRepository.GetSubmission(tx, submissionId)
	if err != nil {
		us.logger.Errorf("Error getting submission: %v", err.Error())
		return -1, err
	}

	taskResponse := schemas.TaskResponsePayload{}

	err = json.Unmarshal(responseMessage.Payload, &taskResponse)
	if err != nil {
		us.logger.Errorf("Error unmarshalling task response: %v", err.Error())
		return -1, err
	}

	submissionResult := models.SubmissionResult{
		SubmissionId: submissionId,
		Code:         strconv.FormatInt(taskResponse.StatusCode, 10),
		Message:      taskResponse.Message,
	}
	id, err := us.submissionResultRepository.CreateSubmissionResult(tx, submissionResult)
	if err != nil {
		us.logger.Errorf("Error creating submission result: %v", err.Error())
		return -1, err
	}
	// Save test results
	for _, testResult := range taskResponse.TestResults {
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

func (ss *submissionService) GetAvailableLanguages(tx *gorm.DB) ([]schemas.LanguageConfig, error) {
	languages, err := ss.languageService.GetAll(tx)
	if err != nil {
		ss.logger.Errorf("Error getting all languages: %v", err.Error())
		return nil, err
	}

	return languages, nil
}

func (us *submissionService) createTestResult(tx *gorm.DB, submissionResultId int64, inputOutputId int64, testResult schemas.QueueTestResult) error {
	testResultModel := models.TestResult{
		SubmissionResultId: submissionResultId,
		InputOutputId:      inputOutputId,
		Passed:             testResult.Passed,
		ErrorMessage:       testResult.ErrorMessage,
	}
	return us.testResultRepository.CreateTestResults(tx, testResultModel)
}

func NewSubmissionService(submissionRepository repository.SubmissionRepository, submissionResultRepository repository.SubmissionResultRepository, inputOutputRepository repository.InputOutputRepository, testResultRepository repository.TestRepository, languageService LanguageService, taskService TaskService, userService UserService) SubmissionService {
	log := utils.NewNamedLogger("submission_service")
	return &submissionService{
		submissionRepository:       submissionRepository,
		submissionResultRepository: submissionResultRepository,
		inputOutputRepository:      inputOutputRepository,
		testResultRepository:       testResultRepository,
		languageService:            languageService,
		taskService:                taskService,
		userService:                userService,
		logger:                     log,
	}
}

func (us *submissionService) testResultsModelToSchema(testResults []models.TestResult) []schemas.TestResult {
	var result []schemas.TestResult
	for _, testResult := range testResults {
		result = append(result, schemas.TestResult{
			ID:                 testResult.ID,
			SubmissionResultId: testResult.SubmissionResultId,
			InputOutputId:      testResult.InputOutputId,
			Passed:             testResult.Passed,
			ErrorMessage:       testResult.ErrorMessage,
		})
	}
	return result
}

func (us *submissionService) resultModelToSchema(result *models.SubmissionResult) *schemas.SubmissionResult {
	if result == nil {
		return nil
	}
	return &schemas.SubmissionResult{
		Id:           result.Id,
		SubmissionId: result.SubmissionId,
		Code:         result.Code,
		Message:      result.Message,
		CreatedAt:    result.CreatedAt,
		TestResults:  us.testResultsModelToSchema(result.TestResult),
	}
}

func (us *submissionService) modelToSchema(submission *models.Submission) *schemas.Submission {
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
		Language:      *LanguageToSchema(&submission.Language),
		Task:          *TaskToSchema(&submission.Task),
		User:          *UserToSchema(&submission.User),
		Result:        us.resultModelToSchema(submission.Result),
	}
}
