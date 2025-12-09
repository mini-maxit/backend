package service

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/mini-maxit/backend/internal/database"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/filestorage"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SubmissionService interface {
	// Create creates a new submission for a given task, user, language, and order.
	Create(db database.Database, taskID, userID, languageID int64, contestID *int64, order int, fileID int64) (int64, error)
	// CreateSubmissionResult creates a new submission result based on the response message.
	CreateSubmissionResult(db database.Database, submissionID int64, responseMessage schemas.QueueResponseMessage) (int64, error)
	// GetAll retrieves all submissions based on the user's role and query parameters.
	GetAll(db database.Database, user *schemas.User, userID, taskID, contestID *int64, paginationParams schemas.PaginationParams) (*schemas.PaginatedResult[[]schemas.Submission], error)
	// GetAllForTask retrieves all submissions for a specific task based on the user's role and query parameters.
	GetAllForTask(db database.Database, taskID int64, user *schemas.User, paginationParams schemas.PaginationParams) (*schemas.PaginatedResult[[]schemas.Submission], error)
	// GetAllForContest retrieves all submissions for a specific contest based on the user's role and query parameters.
	GetAllForContest(db database.Database, contestID int64, user *schemas.User, paginationParams schemas.PaginationParams) (*schemas.PaginatedResult[[]schemas.Submission], error)
	// GetAllForUser retrieves all submissions for a specific user based on the current user's role and query parameters.
	GetAllForUser(db database.Database, userID int64, user *schemas.User, paginationParams schemas.PaginationParams) (*schemas.PaginatedResult[[]schemas.Submission], error)
	// GetAvailableLanguages retrieves all available languages.
	GetAvailableLanguages(db database.Database) ([]schemas.LanguageConfig, error)
	// Get retrieves a specific submission based on the submission ID and user's role.
	Get(db database.Database, submissionID int64, user *schemas.User) (schemas.Submission, error)
	// MarkComplete marks a submission as complete.
	MarkComplete(db database.Database, submissionID int64) error
	// MarkFailed marks a submission as failed with an error message.
	MarkFailed(db database.Database, submissionID int64, errorMsg string) error
	// MarkProcessing marks a submission as processing.
	MarkProcessing(db database.Database, submissionID int64) error
	// Submit creates new submission, publishes it to the queue, and returns the submission ID.
	Submit(db database.Database, user *schemas.User, taskID, languageID int64, contestID *int64, submissionFilePath string) (int64, error)
	// GetTaskStatsForContest retrieves aggregated statistics for each task in a contest.
	GetTaskStatsForContest(db database.Database, user *schemas.User, contestID int64) ([]schemas.ContestTaskStats, error)
	// GetUserStatsForContestTask retrieves per-user statistics for a specific task in a contest.
	GetUserStatsForContestTask(db database.Database, user *schemas.User, contestID, taskID int64) ([]schemas.TaskUserStats, error)
	// GetUserStatsForContest retrieves overall statistics for users in a contest.
	GetUserStatsForContest(db database.Database, user *schemas.User, contestID int64, userID *int64) ([]schemas.UserContestStats, error)
}

const defaultSortOrder = "submitted_at:desc"

type submissionService struct {
	accessControlService       AccessControlService
	contestService             ContestService
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
	db database.Database,
	user *schemas.User,
	userID, taskID, contestID *int64,
	paginationParams schemas.PaginationParams,
) (*schemas.PaginatedResult[[]schemas.Submission], error) {
	if paginationParams.Sort == "" {
		paginationParams.Sort = defaultSortOrder
	}

	// Get submissions based on filters
	var submissionModels []models.Submission
	var err error
	var totalCount int64

	if userID != nil || contestID != nil || taskID != nil {
		submissionModels, totalCount, err = ss.getFilteredSubmissions(db, user, userID, contestID, taskID, paginationParams)
	} else {
		submissionModels, totalCount, err = ss.getUnfilteredSubmissions(db, user, paginationParams)
	}

	if err != nil {
		return nil, err
	}

	result := ss.modelsToSchemas(submissionModels)
	paginatedResult := schemas.NewPaginatedResult(result, paginationParams.Offset, paginationParams.Limit, totalCount)
	return &paginatedResult, nil
}

func (ss *submissionService) getFilteredSubmissions(
	db database.Database,
	user *schemas.User,
	userID, contestID, taskID *int64,
	paginationParams schemas.PaginationParams,
) ([]models.Submission, int64, error) {
	// Determine target user ID
	targetUserID := user.ID
	if userID != nil {
		targetUserID = *userID
	}

	// Authorization check for students
	if user.Role == types.UserRoleStudent && targetUserID != user.ID {
		ss.logger.Errorf("Student %v is not allowed to view submissions for user %v", user.ID, targetUserID)
		return nil, 0, errors.ErrPermissionDenied
	}

	// Fetch submissions based on filters and user role
	limit := paginationParams.Limit
	offset := paginationParams.Offset
	sort := paginationParams.Sort

	var submissionModels []models.Submission
	var err error
	var totalCount int64

	// For teachers viewing other users' submissions, use teacher-specific repository methods
	// that filter at the database level using JOINs
	if user.Role == types.UserRoleTeacher && targetUserID != user.ID {
		submissionModels, totalCount, err = ss.fetchSubmissionsByFiltersForTeacher(db, targetUserID, user.ID, contestID, taskID, limit, offset, sort)
	} else {
		// For admins and users viewing their own submissions, use standard repository methods
		submissionModels, totalCount, err = ss.fetchSubmissionsByFilters(db, targetUserID, contestID, taskID, limit, offset, sort)
	}

	if err != nil {
		ss.logger.Errorf("Error getting filtered submissions: %v", err.Error())
		return nil, 0, err
	}

	return submissionModels, totalCount, nil
}

func (ss *submissionService) fetchSubmissionsByFiltersForTeacher(
	db database.Database,
	userID, teacherID int64,
	contestID, taskID *int64,
	limit, offset int,
	sort string,
) ([]models.Submission, int64, error) {
	if contestID != nil && taskID != nil {
		return ss.submissionRepository.GetAllByUserForContestAndTaskByTeacher(db, userID, *contestID, *taskID, teacherID, limit, offset, sort)
	} else if contestID != nil {
		return ss.submissionRepository.GetAllByUserForContestByTeacher(db, userID, *contestID, teacherID, limit, offset, sort)
	} else if taskID != nil {
		return ss.submissionRepository.GetAllByUserForTaskByTeacher(db, userID, *taskID, teacherID, limit, offset, sort)
	}
	return ss.submissionRepository.GetAllByUserForTeacher(db, userID, teacherID, limit, offset, sort)
}

func (ss *submissionService) fetchSubmissionsByFilters(
	db database.Database,
	userID int64,
	contestID, taskID *int64,
	limit, offset int,
	sort string,
) ([]models.Submission, int64, error) {
	if contestID != nil && taskID != nil {
		return ss.submissionRepository.GetAllByUserForContestAndTask(db, userID, *contestID, *taskID, limit, offset, sort)
	} else if contestID != nil {
		return ss.submissionRepository.GetAllByUserForContest(db, userID, *contestID, limit, offset, sort)
	} else if taskID != nil {
		return ss.submissionRepository.GetAllByUserForTask(db, userID, *taskID, limit, offset, sort)
	}
	models, totalCount, err := ss.submissionRepository.GetAllByUser(db, userID, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}
	return models, totalCount, nil
}

func (ss *submissionService) getUnfilteredSubmissions(
	db database.Database,
	user *schemas.User,
	paginationParams schemas.PaginationParams,
) ([]models.Submission, int64, error) {
	var submissionModels []models.Submission
	var totalCount int64
	var err error

	limit := paginationParams.Limit
	offset := paginationParams.Offset
	sort := paginationParams.Sort

	switch user.Role {
	case types.UserRoleAdmin:
		submissionModels, totalCount, err = ss.submissionRepository.GetAll(db, limit, offset, sort)
	case types.UserRoleStudent:
		submissionModels, totalCount, err = ss.submissionRepository.GetAllByUser(db, user.ID, limit, offset, sort)
	case types.UserRoleTeacher:
		submissionModels, totalCount, err = ss.submissionRepository.GetAllForTeacher(db, user.ID, limit, offset, sort)
	}

	if err != nil {
		ss.logger.Errorf("Error getting all submissions: %v", err.Error())
		return nil, 0, err
	}

	return submissionModels, totalCount, nil
}

func (ss *submissionService) modelsToSchemas(submissionModels []models.Submission) []schemas.Submission {
	result := make([]schemas.Submission, len(submissionModels))
	for i, submissionModel := range submissionModels {
		result[i] = *ss.modelToSchema(&submissionModel)
	}
	return result
}

func (ss *submissionService) Get(db database.Database, submissionID int64, user *schemas.User) (schemas.Submission, error) {
	submissionModel, err := ss.submissionRepository.Get(db, submissionID)
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
			return schemas.Submission{}, errors.ErrPermissionDenied
		}
	case types.UserRoleTeacher:
		// Teacher is only allowed to view submissions for tasks they created
		if submissionModel.Task.CreatedBy != user.ID {
			ss.logger.Errorf("User %v is not allowed to view submission %v", user.ID, submissionID)
			return schemas.Submission{}, errors.ErrPermissionDenied
		}
	}

	return *ss.modelToSchema(submissionModel), nil
}

func (ss *submissionService) GetAllForUser(
	db database.Database,
	userID int64,
	currentUser *schemas.User,
	paginationParams schemas.PaginationParams,
) (*schemas.PaginatedResult[[]schemas.Submission], error) {
	if paginationParams.Sort == "" {
		paginationParams.Sort = defaultSortOrder
	}

	submissionModels, totalCount, err := ss.submissionRepository.GetAllByUser(db, userID, paginationParams.Limit, paginationParams.Offset, paginationParams.Sort)
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
			return nil, errors.ErrPermissionDenied
		}
	case types.UserRoleTeacher:
		// Teacher is only allowed to view submissions for tasks they created
		for i, submission := range submissionModels {
			if submission.Task.CreatedBy != currentUser.ID {
				submissionModels = slices.Delete(submissionModels, i, i+1)
			}
		}
	}

	result := make([]schemas.Submission, len(submissionModels))
	for i, submission := range submissionModels {
		result[i] = *ss.modelToSchema(&submission)
	}
	paginatedResult := schemas.NewPaginatedResult(result, paginationParams.Offset, paginationParams.Limit, totalCount)

	return &paginatedResult, nil
}

func (ss *submissionService) GetAllForGroup(
	db database.Database,
	groupID int64,
	user *schemas.User,
	paginationParams schemas.PaginationParams,
) ([]schemas.Submission, int64, error) {
	var err error
	var totalCount int64
	submissionModels := []models.Submission{}

	if paginationParams.Sort == "" {
		paginationParams.Sort = defaultSortOrder
	}

	switch user.Role {
	case types.UserRoleAdmin:
		// Admin is allowed to view all submissions
		submissionModels, totalCount, err = ss.submissionRepository.GetAllForGroup(db, groupID, paginationParams.Limit, paginationParams.Offset, paginationParams.Sort)
	case types.UserRoleStudent:
		// Student is only allowed to view their own submissions
		return nil, 0, errors.ErrPermissionDenied
	case types.UserRoleTeacher:
		// Teacher is only allowed to view submissions for tasks they created
		group, er := ss.groupRepository.Get(db, groupID)
		if er != nil {
			ss.logger.Errorf("Error getting group: %v", er.Error())
			return nil, 0, er
		}
		if group.CreatedBy != user.ID {
			return nil, 0, errors.ErrPermissionDenied
		}
		submissionModels, totalCount, err = ss.submissionRepository.GetAllForGroup(db, groupID, paginationParams.Limit, paginationParams.Offset, paginationParams.Sort)
	}

	if err != nil {
		ss.logger.Errorf("Error getting all submissions for group: %v", err.Error())
		return nil, 0, err
	}

	result := make([]schemas.Submission, len(submissionModels))
	for i, submissionModel := range submissionModels {
		result[i] = *ss.modelToSchema(&submissionModel)
	}

	return result, totalCount, nil
}

func (ss *submissionService) GetAllForTask(
	db database.Database,
	taskID int64,
	user *schemas.User,
	paginationParams schemas.PaginationParams,
) (*schemas.PaginatedResult[[]schemas.Submission], error) {
	var err error
	submissionModel := []models.Submission{}

	if paginationParams.Sort == "" {
		paginationParams.Sort = defaultSortOrder
	}

	var totalCount int64
	switch user.Role {
	case types.UserRoleAdmin:
		submissionModel, totalCount, err = ss.submissionRepository.GetAllForTask(db, taskID, paginationParams.Limit, paginationParams.Offset, paginationParams.Sort)
	case types.UserRoleTeacher:
		task, er := ss.taskService.Get(db, user, taskID)
		if er != nil {
			return nil, er
		}
		if task.CreatedBy != user.ID {
			return nil, errors.ErrPermissionDenied
		}
		submissionModel, totalCount, err = ss.submissionRepository.GetAllForTask(db, taskID, paginationParams.Limit, paginationParams.Offset, paginationParams.Sort)
	case types.UserRoleStudent:
		isAssigned, er := ss.userService.IsTaskAssignedToUser(db, user.ID, taskID)
		if er != nil {
			return nil, er
		}
		if !isAssigned {
			return nil, errors.ErrPermissionDenied
		}
		submissionModel, totalCount, err = ss.submissionRepository.GetAllForTaskByUser(db, taskID, user.ID, paginationParams.Limit, paginationParams.Offset, paginationParams.Sort)
	}

	if err != nil {
		return nil, err
	}

	result := make([]schemas.Submission, len(submissionModel))
	for i, submission := range submissionModel {
		result[i] = *ss.modelToSchema(&submission)
	}
	paginatedResult := schemas.NewPaginatedResult(result, paginationParams.Offset, paginationParams.Limit, totalCount)

	return &paginatedResult, nil
}

func (ss *submissionService) GetAllForContest(
	db database.Database,
	contestID int64,
	user *schemas.User,
	paginationParams schemas.PaginationParams,
) (*schemas.PaginatedResult[[]schemas.Submission], error) {
	var err error
	submissionModels := []models.Submission{}

	if paginationParams.Sort == "" {
		paginationParams.Sort = defaultSortOrder
	}

	var totalCount int64
	switch user.Role {
	case types.UserRoleAdmin:
		// Admin is allowed to view all submissions for the contest
		submissionModels, totalCount, err = ss.submissionRepository.GetAllForContest(db, contestID, paginationParams.Limit, paginationParams.Offset, paginationParams.Sort)
	case types.UserRoleTeacher:
		// Teacher is allowed to view all submissions for contests they created
		contest, er := ss.contestService.GetDetailed(db, user, contestID)
		if er != nil {
			ss.logger.Errorf("Error getting contest: %v", er.Error())
			return nil, er
		}
		if contest.CreatedBy != user.ID {
			return nil, errors.ErrPermissionDenied
		}
		submissionModels, totalCount, err = ss.submissionRepository.GetAllForContest(db, contestID, paginationParams.Limit, paginationParams.Offset, paginationParams.Sort)
	case types.UserRoleStudent:
		// Students are not allowed to view all submissions for a contest
		return nil, errors.ErrPermissionDenied
	}

	if err != nil {
		ss.logger.Errorf("Error getting submissions for contest: %v", err.Error())
		return nil, err
	}

	result := []schemas.Submission{}
	for _, submission := range submissionModels {
		result = append(result, *ss.modelToSchema(&submission))
	}
	paginatedResult := schemas.NewPaginatedResult(result, paginationParams.Offset, paginationParams.Limit, totalCount)
	return &paginatedResult, nil
}

func (ss *submissionService) MarkFailed(db database.Database, submissionID int64, errorMsg string) error {
	err := ss.submissionRepository.MarkFailed(db, submissionID, errorMsg)
	if err != nil {
		ss.logger.Errorf("Error marking submission failed: %v", err.Error())
		return err
	}

	return nil
}

func (ss *submissionService) MarkComplete(db database.Database, submissionID int64) error {
	err := ss.submissionRepository.MarkEvaluated(db, submissionID)
	if err != nil {
		ss.logger.Errorf("Error marking submission complete: %v", err.Error())
		return err
	}

	return nil
}

func (ss *submissionService) MarkProcessing(db database.Database, submissionID int64) error {
	err := ss.submissionRepository.MarkProcessing(db, submissionID)
	if err != nil {
		ss.logger.Errorf("Error marking submission processing: %v", err.Error())
		return err
	}

	return nil
}

func (ss *submissionService) Create(
	db database.Database,
	taskID int64,
	userID int64,
	languageID int64,
	contestID *int64,
	order int,
	fileID int64,
) (int64, error) {
	// Create a new submission
	submission := &models.Submission{
		TaskID:     taskID,
		UserID:     userID,
		Order:      order,
		LanguageID: languageID,
		ContestID:  contestID,
		FileID:     fileID,
		Status:     types.SubmissionStatusReceived,
	}
	submissionID, err := ss.submissionRepository.Create(db, submission)

	if err != nil {
		ss.logger.Errorf("Error creating submission: %v", err.Error())
		return 0, err
	}

	return submissionID, nil
}

func (ss *submissionService) CreateSubmissionResult(
	db database.Database,
	submissionID int64,
	responseMessage schemas.QueueResponseMessage,
) (int64, error) {
	submissionResultResponse := schemas.SubmissionResultWorkerResponse{}

	err := json.Unmarshal(responseMessage.Payload, &submissionResultResponse)
	if err != nil {
		ss.logger.Errorf("Error unmarshalling task response: %v", err.Error())
		return -1, err
	}

	submissionResult, err := ss.submissionResultRepository.GetBySubmission(db, submissionID)
	if err != nil {
		ss.logger.Errorf("Error getting submission result: %v", err.Error())
		return -1, err
	}
	submissionResult.Code = submissionResultResponse.Code
	if !submissionResult.Code.IsValid() {
		ss.logger.Errorf("Invalid submission result code received from wokrer: %d", submissionResult.Code)
		submissionResult.Code = types.SubmissionResultCodeInvalid
	}
	submissionResult.Message = submissionResultResponse.Message
	submissionResult.Submission.CheckedAt = time.Now()
	err = ss.submissionResultRepository.Put(db, submissionResult)
	if err != nil {
		ss.logger.Errorf("Error creating submission result: %v", err.Error())
		return -1, err
	}
	// Save test results
	for _, responseTestResult := range submissionResultResponse.TestResults {
		testResult, err := ss.testResultRepository.GetBySubmissionAndOrder(db, submissionID, responseTestResult.Order)
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
		err = ss.testResultRepository.Put(db, testResult)
		if err != nil {
			ss.logger.Errorf("Error storing test result: %v", err.Error())
			return -1, err
		}
	}

	err = ss.submissionRepository.MarkEvaluated(db, submissionID)
	if err != nil {
		ss.logger.Errorf("Error marking submission complete: %v", err.Error())
		return -1, err
	}

	return submissionResult.ID, nil
}

func (ss *submissionService) GetAvailableLanguages(db database.Database) ([]schemas.LanguageConfig, error) {
	languages, err := ss.languageService.GetAllEnabled(db)
	if err != nil {
		ss.logger.Errorf("Error getting all languages: %v", err.Error())
		return nil, err
	}
	if languages == nil {
		languages = []schemas.LanguageConfig{}
	}
	return languages, nil
}

func (ss *submissionService) Submit(
	db database.Database,
	user *schemas.User,
	taskID, languageID int64,
	contestID *int64, // null means no contest
	submissionFilePath string,
) (int64, error) {
	_, err := ss.taskService.Get(db, user, taskID)
	if err != nil {
		ss.logger.Errorf("Error getting task: %v", err.Error())
		return -1, err
	}
	if contestID != nil {
		// Validate contest submission (checks contest status, task in contest, user participation, submission windows, etc.)
		err = ss.contestService.ValidateContestSubmission(db, *contestID, taskID, user.ID)
		if err != nil {
			ss.logger.Errorf("Contest submission validation failed: %v", err.Error())
			return -1, err
		}
	}
	// Upload solution file to storage
	latest, err := ss.submissionRepository.GetLatestForTaskByUser(db, user.ID, taskID)
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

	err = ss.fileRepository.Create(db, file)
	if err != nil {
		ss.logger.Errorf("Error creating file record: %v", err.Error())
		return -1, err
	}
	submissionID, err := ss.Create(db, taskID, user.ID, languageID, contestID, newOrder, file.ID)
	if err != nil {
		ss.logger.Errorf("Error creating submission: %v", err.Error())
		return 0, err
	}
	// Create submissionResult if the submission is created successfully
	submissionResultID, err := ss.createSubmissionResult(db, submissionID)
	if err != nil {
		ss.logger.Errorf("Error creating submission result: %v", err.Error())
		return -1, err
	}

	err = ss.queueService.PublishSubmission(db, submissionID, submissionResultID)
	if err != nil {
		ss.logger.Errorf("Error publishing submission to queue: %v", err.Error())
		return -1, err
	}
	return submissionID, nil
}

// creates blank submission result for the submission together with testResults
func (ss *submissionService) createSubmissionResult(db database.Database, submissionID int64) (int64, error) {
	submissionResult := models.SubmissionResult{
		SubmissionID: submissionID,
		Code:         types.SubmissionResultCodeUnknown,
		Message:      "Awaiting processing",
	}

	submission, err := ss.submissionRepository.Get(db, submissionID)
	if err != nil {
		ss.logger.Errorf("Error getting submission: %v", err.Error())
		return -1, err
	}

	submissionResultID, err := ss.submissionResultRepository.Create(db, submissionResult)
	if err != nil {
		ss.logger.Errorf("Error creating submission result: %v", err.Error())
		return -1, err
	}

	inputOutputs, err := ss.inputOutputRepository.GetByTask(db, submission.TaskID)
	if err != nil {
		ss.logger.Errorf("Error getting input outputs: %v", err.Error())
		return -1, err
	}

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
		err = ss.fileRepository.Create(db, stdoutFileModel)
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
		err = ss.fileRepository.Create(db, stderrFileMode)
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
		err = ss.fileRepository.Create(db, diffFileModel)
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
		err = ss.testResultRepository.Create(db, &testResult)
		if err != nil {
			ss.logger.Errorf("Error creating test result: %v", err.Error())
			return -1, err
		}
	}
	return submissionResultID, nil
}

func NewSubmissionService(
	accessControlService AccessControlService,
	contestService ContestService,
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
	service := &submissionService{
		accessControlService:       accessControlService,
		contestService:             contestService,
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
	if err := utils.ValidateStruct(*service); err != nil {
		log.Fatalf("Invalid submission service: %v", err)
	}
	return service
}

func (ss *submissionService) testResultsModelToSchema(testResults []models.TestResult) []schemas.TestResult {
	var result []schemas.TestResult
	for _, testResult := range testResults {
		result = append(result, schemas.TestResult{
			ID:                 testResult.ID,
			SubmissionResultID: testResult.SubmissionResultID,
			TestCaseID:         testResult.TestCaseID,
			ExecutionTimeMs:    testResult.ExecutionTime * 1000,
			Passed:             *testResult.Passed,
			Code:               testResult.StatusCode.String(),
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
		ContestID:   submission.ContestID,
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

func (ss *submissionService) GetTaskStatsForContest(db database.Database, user *schemas.User, contestID int64) ([]schemas.ContestTaskStats, error) {
	err := ss.accessControlService.CanUserAccess(db, types.ResourceTypeContest, contestID, user, types.PermissionEdit)
	if err != nil {
		return nil, err
	}

	_, err = ss.contestService.GetDetailed(db, user, contestID)
	if err != nil {
		ss.logger.Errorw("Error getting contest for contest task stats", "error", err)
		return nil, err
	}

	// Get raw stats from repository
	rawStats, err := ss.submissionRepository.GetTaskStatsForContest(db, contestID)
	if err != nil {
		ss.logger.Errorw("Error getting task stats for contest", "error", err)
		return nil, err
	}

	// Convert typed model slice to schema slice
	stats := make([]schemas.ContestTaskStats, 0, len(rawStats))
	for _, raw := range rawStats {
		stat := schemas.ContestTaskStats{
			Task:                 *TaskToInfoSchema(&raw.Task),
			TotalParticipants:    raw.TotalParticipants,
			SubmittedCount:       raw.SubmittedCount,
			FullySolvedCount:     raw.FullySolvedCount,
			PartiallySolvedCount: raw.PartiallySolvedCount,
			AverageScore:         raw.AverageScore,
		}
		if stat.SubmittedCount > 0 {
			stat.SuccessRate = float64(stat.FullySolvedCount) / float64(stat.SubmittedCount) * 100
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (ss *submissionService) GetUserStatsForContestTask(db database.Database, user *schemas.User, contestID, taskID int64) ([]schemas.TaskUserStats, error) {
	err := ss.accessControlService.CanUserAccess(db, types.ResourceTypeContest, contestID, user, types.PermissionEdit)
	if err != nil {
		return nil, err
	}

	isInContest, err := ss.contestService.IsTaskInContest(db, contestID, taskID)
	if err != nil {
		ss.logger.Errorw("Error checking if task is in contest for contest task stats", "error", err)
		return nil, err
	}
	if !isInContest {
		ss.logger.Errorw("Task is not in contest for contest task stats", "contestID", contestID, "taskID", taskID)
		return nil, errors.ErrNotFound
	}

	// Get raw stats from repository
	rawStats, err := ss.submissionRepository.GetUserStatsForContestTask(db, contestID, taskID)
	if err != nil {
		ss.logger.Errorw("Error getting user stats for contest task", "error", err)
		return nil, err
	}

	// Convert typed model slice to schema slice
	stats := make([]schemas.TaskUserStats, 0, len(rawStats))
	for _, raw := range rawStats {
		var bestSubmissionID *int64
		if raw.BestSubmissionID != 0 {
			bestSubmissionID = &raw.BestSubmissionID
		}
		stat := schemas.TaskUserStats{
			User:             *UserToInfoSchema(&raw.User),
			SubmissionCount:  int(raw.SubmissionCount),
			BestScore:        raw.BestScore,
			BestSubmissionID: bestSubmissionID,
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (ss *submissionService) GetUserStatsForContest(db database.Database, user *schemas.User, contestID int64, userID *int64) ([]schemas.UserContestStats, error) {
	err := ss.accessControlService.CanUserAccess(db, types.ResourceTypeContest, contestID, user, types.PermissionEdit)
	if err != nil {
		return nil, err
	}

	rawStats, err := ss.submissionRepository.GetUserStatsForContest(db, contestID, userID)
	if err != nil {
		ss.logger.Errorw("Error getting user stats for contest", "error", err)
		return nil, err
	}

	stats := make([]schemas.UserContestStats, 0, len(rawStats))
	for _, raw := range rawStats {
		stat := schemas.UserContestStats{
			User:                 *UserToInfoSchema(&raw.User),
			TasksAttempted:       int(raw.TasksAttempted),
			TasksSolved:          int(raw.TasksSolved),
			TasksPartiallySolved: int(raw.TasksPartiallySolved),
			TaskBreakdown:        make([]schemas.UserTaskPerformance, 0, len(raw.TaskBreakdown)),
		}

		for _, bm := range raw.TaskBreakdown {
			stat.TaskBreakdown = append(stat.TaskBreakdown, schemas.UserTaskPerformance{
				TaskID:       bm.TaskID,
				TaskTitle:    bm.TaskTitle,
				BestScore:    bm.BestScore,
				AttemptCount: bm.AttemptCount,
				IsSolved:     bm.IsSolved,
			})
		}

		stats = append(stats, stat)
	}

	return stats, nil
}
