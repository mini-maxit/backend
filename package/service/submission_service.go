package service

import (
	"encoding/json"
	"errors"
	"time"

	"slices"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	myerrors "github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/filestorage"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SubmissionService interface {
	// Create creates a new submission for a given task, user, language, and order.
	Create(tx *gorm.DB, taskID int64, userID int64, languageID int64, order int, fileID int64) (int64, error)
	// CreateSubmissionResult creates a new submission result based on the response message.
	CreateSubmissionResult(tx *gorm.DB, submissionID int64, responseMessage schemas.QueueResponseMessage) (int64, error)
	// GetAll retrieves all submissions based on the user's role and query parameters.
	GetAll(tx *gorm.DB, user schemas.User, queryParams map[string]any) ([]schemas.Submission, error)
	// GetAllForGroup retrieves all submissions for a specific group based on the user's role and query parameters.
	GetAllForGroup(tx *gorm.DB, groupID int64, user schemas.User, queryParams map[string]any) ([]schemas.Submission, error)
	// GetAllForTask retrieves all submissions for a specific task based on the user's role and query parameters.
	GetAllForTask(tx *gorm.DB, taskID int64, user schemas.User, queryParams map[string]any) ([]schemas.Submission, error)
	// GetAllForUser retrieves all submissions for a specific user based on the current user's role and query parameters.
	GetAllForUser(tx *gorm.DB, userID int64, user schemas.User, queryParams map[string]any) ([]schemas.Submission, error)
	// GetAllForUserShort retrieves a short version of all submissions for a specific user
	// based on the current user's role and query parameters.
	GetAllForUserShort(
		tx *gorm.DB,
		userID int64,
		user schemas.User,
		queryParams map[string]any,
	) ([]schemas.SubmissionShort, error)
	// GetAvailableLanguages retrieves all available languages.
	GetAvailableLanguages(tx *gorm.DB) ([]schemas.LanguageConfig, error)
	// Get retrieves a specific submission based on the submission ID and user's role.
	Get(tx *gorm.DB, submissionID int64, user schemas.User) (schemas.Submission, error)
	// MarkComplete marks a submission as complete.
	MarkComplete(tx *gorm.DB, submissionID int64) error
	// MarkFailed marks a submission as failed with an error message.
	MarkFailed(tx *gorm.DB, submissionID int64, errorMsg string) error
	// MarkProcessing marks a submission as processing.
	MarkProcessing(tx *gorm.DB, submissionID int64) error
	// Submit creates new submission, publishes it to the queue, and returns the submission ID.
	Submit(tx *gorm.DB, user *schemas.User, taskID int64, languageID int64, submissionFilePath string) (int64, error)
}

const defaultSortOrder = "submitted_at:desc"

type submissionService struct {
	filestorage                filestorage.FileStorageService
	fileRepository             repository.File
	submissionRepository       repository.SubmissionRepository
	submissionResultRepository repository.SubmissionResultRepository
	inputOutputRepository      repository.TestCaseRepository
	testResultRepository       repository.TestRepository
	groupRepository            repository.GroupRepository
	taskRepository             repository.TaskRepository
	userService                UserService
	taskService                TaskService
	languageService            LanguageService
	queueService               QueueService
	logger                     *zap.SugaredLogger
}

func (ss *submissionService) GetAll(
	tx *gorm.DB,
	user schemas.User,
	queryParams map[string]any,
) ([]schemas.Submission, error) {
	submissionModels := []models.Submission{}
	var err error

	limit := queryParams["limit"].(int)
	offset := queryParams["offset"].(int)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = defaultSortOrder
	}

	switch user.Role {
	case types.UserRoleAdmin:
		submissionModels, err = ss.submissionRepository.GetAll(tx, limit, offset, sort)
	case types.UserRoleStudent:
		submissionModels, err = ss.submissionRepository.GetAllByUser(tx, user.ID, limit, offset, sort)
	case types.UserRoleTeacher:
		submissionModels, err = ss.submissionRepository.GetAllForTeacher(tx, user.ID, limit, offset, sort)
	}

	if err != nil {
		ss.logger.Errorf("Error getting all submissions: %v", err.Error())
		return nil, err
	}

	var result []schemas.Submission
	for _, submissionModel := range submissionModels {
		result = append(result, *ss.modelToSchema(&submissionModel))
	}

	return result, nil
}

func (ss *submissionService) Get(tx *gorm.DB, submissionID int64, user schemas.User) (schemas.Submission, error) {
	submissionModel, err := ss.submissionRepository.Get(tx, submissionID)
	if err != nil {
		ss.logger.Errorf("Error getting submission: %v", err.Error())
		return schemas.Submission{}, err
	}

	switch user.Role {
	case types.UserRoleAdmin:
		// Admin is allowed to view all submissions
	case types.UserRoleStudent:
		// Student is only allowed to view their own submissions
		if submissionModel.UserID != user.ID {
			ss.logger.Errorf("User %v is not allowed to view submission %v", user.ID, submissionID)
			return schemas.Submission{}, myerrors.ErrPermissionDenied
		}
	case types.UserRoleTeacher:
		// Teacher is only allowed to view submissions for tasks they created
		if submissionModel.Task.CreatedBy != user.ID {
			ss.logger.Errorf("User %v is not allowed to view submission %v", user.ID, submissionID)
			return schemas.Submission{}, myerrors.ErrPermissionDenied
		}
	}

	return *ss.modelToSchema(submissionModel), nil
}

func (ss *submissionService) GetAllForUser(
	tx *gorm.DB,
	userID int64,
	currentUser schemas.User,
	queryParams map[string]any,
) ([]schemas.Submission, error) {
	limit := queryParams["limit"].(int)
	offset := queryParams["offset"].(int)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = defaultSortOrder
	}

	submissionModels, err := ss.submissionRepository.GetAllByUser(tx, userID, limit, offset, sort)
	if err != nil {
		ss.logger.Errorf("Error getting all submissions for user: %v", err.Error())
		return nil, err
	}

	switch currentUser.Role {
	case types.UserRoleAdmin:
		// Admin is allowed to view all submissions
	case types.UserRoleStudent:
		// Student is only allowed to view their own submissions
		if userID != currentUser.ID {
			ss.logger.Errorf("User %v is not allowed to view submissions", currentUser.ID)
			return nil, myerrors.ErrPermissionDenied
		}
	case types.UserRoleTeacher:
		// Teacher is only allowed to view submissions for tasks they created
		for i, submission := range submissionModels {
			if submission.Task.CreatedBy != currentUser.ID {
				submissionModels = slices.Delete(submissionModels, i, i+1)
			}
		}
	}

	var result []schemas.Submission
	for _, submission := range submissionModels {
		result = append(result, *ss.modelToSchema(&submission))
	}

	return result, nil
}

func (ss *submissionService) GetAllForUserShort(
	tx *gorm.DB,
	userID int64,
	currentUser schemas.User,
	queryParams map[string]any,
) ([]schemas.SubmissionShort, error) {
	submissionModels, err := ss.GetAllForUser(tx, userID, currentUser, queryParams)
	if err != nil {
		return nil, err
	}
	result := []schemas.SubmissionShort{}
	for _, submission := range submissionModels {
		passed := true
		count := len(submission.Result.TestResults)
		for _, testResult := range submission.Result.TestResults {
			if !testResult.Passed {
				passed = false
				count--
			}
		}
		result = append(result, schemas.SubmissionShort{
			ID:            submission.ID,
			TaskID:        submission.TaskID,
			UserID:        submission.UserID,
			Passed:        passed,
			HowManyPassed: int64(count),
		})
	}

	return result, nil
}

func (ss *submissionService) GetAllForGroup(
	tx *gorm.DB,
	groupID int64,
	user schemas.User,
	queryParams map[string]any,
) ([]schemas.Submission, error) {
	var err error
	submissionModels := []models.Submission{}

	limit := queryParams["limit"].(int)
	offset := queryParams["offset"].(int)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = defaultSortOrder
	}

	switch user.Role {
	case types.UserRoleAdmin:
		// Admin is allowed to view all submissions
		submissionModels, err = ss.submissionRepository.GetAllForGroup(tx, groupID, limit, offset, sort)
	case types.UserRoleStudent:
		// Student is only allowed to view their own submissions
		return nil, myerrors.ErrPermissionDenied
	case types.UserRoleTeacher:
		// Teacher is only allowed to view submissions for tasks they created
		group, er := ss.groupRepository.Get(tx, groupID)
		if er != nil {
			ss.logger.Errorf("Error getting group: %v", er.Error())
			return nil, er
		}
		if group.CreatedBy != user.ID {
			return nil, myerrors.ErrPermissionDenied
		}
		submissionModels, err = ss.submissionRepository.GetAllForGroup(tx, groupID, limit, offset, sort)
	}

	if err != nil {
		ss.logger.Errorf("Error getting all submissions for group: %v", err.Error())
		return nil, err
	}

	var result []schemas.Submission
	for _, submissionModel := range submissionModels {
		result = append(result, *ss.modelToSchema(&submissionModel))
	}

	return result, nil
}

func (ss *submissionService) GetAllForTask(
	tx *gorm.DB,
	taskID int64,
	user schemas.User,
	queryParams map[string]any,
) ([]schemas.Submission, error) {
	var err error
	submissionModel := []models.Submission{}

	limit := queryParams["limit"].(int)
	offset := queryParams["offset"].(int)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = defaultSortOrder
	}

	switch user.Role {
	case types.UserRoleAdmin:
		submissionModel, err = ss.submissionRepository.GetAllForTask(tx, taskID, limit, offset, sort)
	case types.UserRoleTeacher:
		task, er := ss.taskService.Get(tx, user, taskID)
		if er != nil {
			return nil, er
		}
		if task.CreatedBy != user.ID {
			return nil, myerrors.ErrPermissionDenied
		}
		submissionModel, err = ss.submissionRepository.GetAllForTask(tx, taskID, limit, offset, sort)
	case types.UserRoleStudent:
		isAssigned, er := ss.taskRepository.IsAssignedToUser(tx, taskID, user.ID)
		if er != nil {
			return nil, er
		}
		if !isAssigned {
			return nil, myerrors.ErrPermissionDenied
		}
		submissionModel, err = ss.submissionRepository.GetAllForTaskByUser(tx, taskID, user.ID, limit, offset, sort)
	}

	if err != nil {
		return nil, err
	}

	result := []schemas.Submission{}
	for _, submission := range submissionModel {
		result = append(result, *ss.modelToSchema(&submission))
	}

	return result, nil
}

func (ss *submissionService) MarkFailed(tx *gorm.DB, submissionID int64, errorMsg string) error {
	err := ss.submissionRepository.MarkFailed(tx, submissionID, errorMsg)
	if err != nil {
		ss.logger.Errorf("Error marking submission failed: %v", err.Error())
		return err
	}

	return nil
}

func (ss *submissionService) MarkComplete(tx *gorm.DB, submissionID int64) error {
	err := ss.submissionRepository.MarkComplete(tx, submissionID)
	if err != nil {
		ss.logger.Errorf("Error marking submission complete: %v", err.Error())
		return err
	}

	return nil
}

func (ss *submissionService) MarkProcessing(tx *gorm.DB, submissionID int64) error {
	err := ss.submissionRepository.MarkProcessing(tx, submissionID)
	if err != nil {
		ss.logger.Errorf("Error marking submission processing: %v", err.Error())
		return err
	}

	return nil
}

func (ss *submissionService) Create(
	tx *gorm.DB,
	taskID int64,
	userID int64,
	languageID int64,
	order int,
	fileID int64,
) (int64, error) {
	// Create a new submission
	submission := &models.Submission{
		TaskID:     taskID,
		UserID:     userID,
		Order:      order,
		LanguageID: languageID,
		FileID:     fileID,
		Status:     types.SubmissionStatusReceived,
	}
	submissionID, err := ss.submissionRepository.Create(tx, submission)

	if err != nil {
		ss.logger.Errorf("Error creating submission: %v", err.Error())
		return 0, err
	}

	return submissionID, nil
}

func (ss *submissionService) CreateSubmissionResult(
	tx *gorm.DB,
	submissionID int64,
	responseMessage schemas.QueueResponseMessage,
) (int64, error) {
	submissionResultResponse := schemas.SubmissionResultWorkerResponse{}

	err := json.Unmarshal(responseMessage.Payload, &submissionResultResponse)
	if err != nil {
		ss.logger.Errorf("Error unmarshalling task response: %v", err.Error())
		return -1, err
	}

	submissionResult, err := ss.submissionResultRepository.GetBySubmission(tx, submissionID)
	if err != nil {
		ss.logger.Errorf("Error getting submission result: %v", err.Error())
		return -1, err
	}
	ss.logger.Info("Got submission result: ", submissionResult)
	submissionResult.Code = submissionResultResponse.Code
	if !submissionResult.Code.IsValid() {
		ss.logger.Errorf("Invalid submission result code received from wokrer: %d", submissionResult.Code)
		submissionResult.Code = types.SubmissionResultCodeInvalid
	}
	submissionResult.Message = submissionResultResponse.Message
	submissionResult.Submission.CheckedAt = time.Now()
	err = ss.submissionResultRepository.Put(tx, submissionResult)
	if err != nil {
		ss.logger.Errorf("Error creating submission result: %v", err.Error())
		return -1, err
	}
	// Save test results
	for i, responseTestResult := range submissionResultResponse.TestResults {
		ss.logger.Infof("Processing test result %d: %+v\n", i, responseTestResult)
		testResult, err := ss.testResultRepository.GetBySubmissionAndOrder(tx, submissionID, responseTestResult.Order)
		if err != nil {
			ss.logger.Errorf("Error getting test result: %v", err.Error())
			return -1, err
		}

		testResult.StatusCode = responseTestResult.StatusCode
		if !testResult.StatusCode.IsValid() {
			testResult.StatusCode = types.TestResultStatusCodeInvalid
		}
		testResult.Passed = &responseTestResult.Passed
		testResult.ExecutionTime = responseTestResult.ExecutionTime
		testResult.ErrorMessage = responseTestResult.ErrorMessage
		err = ss.testResultRepository.Put(tx, testResult)
		if err != nil {
			ss.logger.Errorf("Error storing test result: %v", err.Error())
			return -1, err
		}
	}

	err = ss.submissionRepository.MarkComplete(tx, submissionID)
	if err != nil {
		ss.logger.Errorf("Error marking submission complete: %v", err.Error())
		return -1, err
	}

	return submissionResult.ID, nil
}

func (ss *submissionService) GetAvailableLanguages(tx *gorm.DB) ([]schemas.LanguageConfig, error) {
	languages, err := ss.languageService.GetAllEnabled(tx)
	if err != nil {
		ss.logger.Errorf("Error getting all languages: %v", err.Error())
		return nil, err
	}

	return languages, nil
}

func (ss *submissionService) Submit(
	tx *gorm.DB,
	user *schemas.User,
	taskID int64,
	languageID int64,
	submissionFilePath string,
) (int64, error) {
	_, err := ss.taskService.Get(tx, *user, taskID)
	if err != nil {
		ss.logger.Errorf("Error getting task: %v", err.Error())
		return -1, err
	}
	// Upload solution file to storage
	latest, err := ss.submissionRepository.GetLatestForTaskByUser(tx, user.ID, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			latest = nil
		} else {
			return -1, err
		}
	}
	newOrder := 1
	if latest != nil {
		newOrder = latest.Order + 1
	}

	uploadedFile, err := ss.filestorage.UploadSolutionFile(taskID, user.ID, newOrder, submissionFilePath)
	if err != nil {
		ss.logger.Errorf("Error uploading solution file: %v", err.Error())
		return -1, err
	}
	file := &models.File{
		Filename:   uploadedFile.Filename,
		Path:       uploadedFile.Path,
		Bucket:     uploadedFile.Bucket,
		ServerType: uploadedFile.ServerType,
	}

	err = ss.fileRepository.Create(tx, file)
	if err != nil {
		ss.logger.Errorf("Error creating file record: %v", err.Error())
		return -1, err
	}
	submissionID, err := ss.Create(tx, taskID, user.ID, languageID, newOrder, file.ID)
	if err != nil {
		ss.logger.Errorf("Error creating submission: %v", err.Error())
		return 0, err
	}
	// Create submissionResult if the submission is created successfully
	submissionResultID, err := ss.createSubmissionResult(tx, submissionID)
	if err != nil {
		ss.logger.Errorf("Error creating submission result: %v", err.Error())
		return -1, err
	}

	err = ss.queueService.PublishSubmission(tx, submissionID, submissionResultID)
	if err != nil {
		ss.logger.Errorf("Error publishing submission to queue: %v", err.Error())
		return -1, err
	}
	return submissionID, nil
}

// creates blank submission result for the submission together with testResults
func (ss *submissionService) createSubmissionResult(tx *gorm.DB, submissionID int64) (int64, error) {
	submissionResult := models.SubmissionResult{
		SubmissionID: submissionID,
		Code:         types.SubmissionResultCodeUnknown,
		Message:      "Awaiting processing",
	}

	submission, err := ss.submissionRepository.Get(tx, submissionID)
	if err != nil {
		ss.logger.Errorf("Error getting submission: %v", err.Error())
		return -1, err
	}

	submissionResultID, err := ss.submissionResultRepository.Create(tx, submissionResult)
	if err != nil {
		ss.logger.Errorf("Error creating submission result: %v", err.Error())
		return -1, err
	}

	inputOutputs, err := ss.inputOutputRepository.GetByTask(tx, submission.TaskID)
	if err != nil {
		ss.logger.Errorf("Error getting input outputs: %v", err.Error())
		return -1, err
	}
	ss.logger.Info("Got input outputs: ", inputOutputs)

	for _, inputOutput := range inputOutputs {
		stdoutFile := ss.filestorage.GetTestResultStdoutPath(submission.TaskID, submission.UserID, submission.Order, inputOutput.Order)
		stderrFile := ss.filestorage.GetTestResultStderrPath(submission.TaskID, submission.UserID, submission.Order, inputOutput.Order)
		diffFile := ss.filestorage.GetTestResultDiffPath(submission.TaskID, submission.UserID, submission.Order, inputOutput.Order)
		stdoutFileModel := &models.File{
			Filename:   stdoutFile.Filename,
			Path:       stdoutFile.Path,
			Bucket:     stdoutFile.Bucket,
			ServerType: stdoutFile.ServerType,
		}
		err = ss.fileRepository.Create(tx, stdoutFileModel)
		if err != nil {
			ss.logger.Errorf("Error creating stdout file record: %v", err.Error())
			return -1, err
		}
		stderrFileMode := &models.File{
			Filename:   stderrFile.Filename,
			Path:       stderrFile.Path,
			Bucket:     stderrFile.Bucket,
			ServerType: stderrFile.ServerType,
		}
		err = ss.fileRepository.Create(tx, stderrFileMode)
		if err != nil {
			ss.logger.Errorf("Error creating stdout file record: %v", err.Error())
			return -1, err
		}
		diffFileModel := &models.File{
			Filename:   diffFile.Filename,
			Path:       diffFile.Path,
			Bucket:     diffFile.Bucket,
			ServerType: diffFile.ServerType,
		}
		err = ss.fileRepository.Create(tx, diffFileModel)
		if err != nil {
			ss.logger.Errorf("Error creating diff file record: %v", err.Error())
			return -1, err
		}

		falseVal := false
		testResult := models.TestResult{
			SubmissionResultID: submissionResultID,
			TestCaseID:         inputOutput.ID,
			Passed:             &falseVal,
			ExecutionTime:      float64(-1),
			StatusCode:         types.TestResultStatusCodeNotExecuted,
			ErrorMessage:       "Not executed",
			StderrFileID:       stderrFileMode.ID,
			StdoutFileID:       stdoutFileModel.ID,
			DiffFileID:         diffFileModel.ID,
		}
		err = ss.testResultRepository.Create(tx, &testResult)
		if err != nil {
			ss.logger.Errorf("Error creating test result: %v", err.Error())
			return -1, err
		}
		ss.logger.Info("Created testResult", testResult)
	}
	return submissionResultID, nil
}

func NewSubmissionService(
	filestorage filestorage.FileStorageService,
	fileRepository repository.File,
	submissionRepository repository.SubmissionRepository,
	submissionResultRepository repository.SubmissionResultRepository,
	inputOutputRepository repository.TestCaseRepository,
	testResultRepository repository.TestRepository,
	groupRepository repository.GroupRepository,
	taskRepository repository.TaskRepository,
	languageService LanguageService,
	taskService TaskService,
	userService UserService,
	queueService QueueService,
) SubmissionService {
	log := utils.NewNamedLogger("submission_service")
	return &submissionService{
		filestorage:                filestorage,
		fileRepository:             fileRepository,
		submissionRepository:       submissionRepository,
		submissionResultRepository: submissionResultRepository,
		inputOutputRepository:      inputOutputRepository,
		testResultRepository:       testResultRepository,
		groupRepository:            groupRepository,
		taskRepository:             taskRepository,
		languageService:            languageService,
		taskService:                taskService,
		userService:                userService,
		queueService:               queueService,
		logger:                     log,
	}
}

func (ss *submissionService) testResultsModelToSchema(testResults []models.TestResult) []schemas.TestResult {
	var result []schemas.TestResult
	for _, testResult := range testResults {
		result = append(result, schemas.TestResult{
			ID:                 testResult.ID,
			SubmissionResultID: testResult.SubmissionResultID,
			TestCaseID:         testResult.TestCaseID,
			Passed:             *testResult.Passed,
			ErrorMessage:       testResult.ErrorMessage,
		})
	}
	return result
}

func (ss *submissionService) resultModelToSchema(result *models.SubmissionResult) *schemas.SubmissionResult {
	if result == nil {
		return nil
	}
	return &schemas.SubmissionResult{
		ID:           result.ID,
		SubmissionID: result.SubmissionID,
		Code:         result.Code.String(),
		Message:      result.Message,
		CreatedAt:    result.CreatedAt,
		TestResults:  ss.testResultsModelToSchema(result.TestResults),
	}
}

func (ss *submissionService) modelToSchema(submission *models.Submission) *schemas.Submission {
	return &schemas.Submission{
		ID:          submission.ID,
		TaskID:      submission.TaskID,
		UserID:      submission.UserID,
		Order:       submission.Order,
		LanguageID:  submission.LanguageID,
		Status:      submission.Status,
		SubmittedAt: submission.SubmittedAt,
		CheckedAt:   submission.CheckedAt,
		Language:    *LanguageToSchema(&submission.Language),
		Task:        *TaskToSchema(&submission.Task),
		User:        *UserToSchema(&submission.User),
		Result:      ss.resultModelToSchema(submission.Result),
	}
}
