package service

import (
	"errors"
	"time"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	myerrors "github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ContestService interface {
	// Create creates a new contest
	Create(tx *gorm.DB, currentUser schemas.User, contest *schemas.CreateContest) (int64, error)
	// Get retrieves a contest by ID
	Get(tx *gorm.DB, currentUser schemas.User, contestID int64) (*schemas.Contest, error)
	// GetOngoingContests retrieves contests that are currently running
	GetOngoingContests(tx *gorm.DB, currentUser schemas.User, paginationParams schemas.PaginationParams) (schemas.PaginatedResult[[]schemas.AvailableContest], error)
	// GetPastContests retrieves contests that have ended
	GetPastContests(tx *gorm.DB, currentUser schemas.User, paginationParams schemas.PaginationParams) (schemas.PaginatedResult[[]schemas.AvailableContest], error)
	// GetUpcomingContests retrieves contests that haven't started yet
	GetUpcomingContests(tx *gorm.DB, currentUser schemas.User, paginationParams schemas.PaginationParams) (schemas.PaginatedResult[[]schemas.AvailableContest], error)
	// Edit updates a contest
	Edit(tx *gorm.DB, currentUser schemas.User, contestID int64, editInfo *schemas.EditContest) (*schemas.CreatedContest, error)
	// Delete removes a contest
	Delete(tx *gorm.DB, currentUser schemas.User, contestID int64) error
	// RegisterForContest creates a pending registration for a contest
	RegisterForContest(tx *gorm.DB, currentUser schemas.User, contestID int64) error
	// GetTasksForContest retrieves all contest task relations (with timing/submission flags) for a contest (for authorized users)
	GetTasksForContest(tx *gorm.DB, currentUser schemas.User, contestID int64) ([]schemas.ContestTask, error)
	// GetTaskProgressForContest retrieves all tasks associated with a contest with submission stats for the requesting user
	GetTaskProgressForContest(tx *gorm.DB, currentUser schemas.User, contestID int64) ([]schemas.TaskWithContestStats, error)
	// GetAssignableTasks retrieves all tasks NOT assigned to a contest (for authorized users)
	GetAssignableTasks(tx *gorm.DB, currentUser schemas.User, contestID int64) ([]schemas.Task, error)
	// GetUserContests retrieves all contests a user is participating in
	GetUserContests(tx *gorm.DB, userID int64) (schemas.UserContestsWithStats, error)
	// AddTaskToContest adds a task to a contest
	AddTaskToContest(tx *gorm.DB, currentUser *schemas.User, contestID int64, request *schemas.AddTaskToContest) error
	// GetRegistrationRequests retrieves pending registration requests for a contest
	GetRegistrationRequests(tx *gorm.DB, currentUser schemas.User, contestID int64, statusFilter types.RegistrationRequestStatus) ([]schemas.RegistrationRequest, error)
	// ApproveRegistrationRequest approves a pending registration request for a contest
	ApproveRegistrationRequest(tx *gorm.DB, currentUser schemas.User, contestID, userID int64) error
	// RejectRegistrationRequest rejects a pending registration request for a contest
	RejectRegistrationRequest(tx *gorm.DB, currentUser schemas.User, contestID, userID int64) error
	// IsTaskInContest checks if a task is part of a contest
	IsTaskInContest(tx *gorm.DB, contestID, taskID int64) (bool, error)
	// IsUserParticipant checks if a user is a participant in a contest
	IsUserParticipant(tx *gorm.DB, contestID, userID int64) (bool, error)
	// GetTaskContests retrieves all contest IDs that a task is assigned to
	GetTaskContests(tx *gorm.DB, taskID int64) ([]int64, error)
	// ValidateContestSubmission validates if a user can submit a solution for a task in a contest
	// Returns an error if submission is not allowed (contest/task not open, user not participant, etc.)
	ValidateContestSubmission(tx *gorm.DB, contestID, taskID, userID int64) error

	GetContestsCreatedByUser(tx *gorm.DB, userID int64, paginationParams schemas.PaginationParams) (schemas.PaginatedResult[[]schemas.CreatedContest], error)
	GetContestTask(tx *gorm.DB, currentUser *schemas.User, contestID, taskID int64) (*schemas.TaskDetailed, error)
}

const defaultContestSort = "created_at:desc"

type contestService struct {
	contestRepository    repository.ContestRepository
	taskRepository       repository.TaskRepository
	userRepository       repository.UserRepository
	submissionRepository repository.SubmissionRepository
	accessControlService AccessControlService

	taskService TaskService

	logger *zap.SugaredLogger
}

func (cs *contestService) Create(tx *gorm.DB, currentUser schemas.User, contest *schemas.CreateContest) (int64, error) {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return 0, err
	}

	validate, err := utils.NewValidator()
	if err != nil {
		return -1, err
	}
	if err := validate.Struct(contest); err != nil {
		return 0, err
	}

	model := &models.Contest{
		Name:               contest.Name,
		Description:        contest.Description,
		CreatedBy:          currentUser.ID,
		IsRegistrationOpen: contest.IsRegistrationOpen,
		IsSubmissionOpen:   contest.IsSubmissionOpen,
		IsVisible:          contest.IsVisible,
		StartAt:            contest.StartAt,
	}

	if contest.EndAt != nil {
		model.EndAt = contest.EndAt
	}

	contestID, err := cs.contestRepository.Create(tx, model)
	if err != nil {
		return -1, err
	}

	// Automatically grant manage permission to the creator
	if err := cs.accessControlService.GrantCreatorAccess(tx, models.ResourceTypeContest, contestID, currentUser.ID); err != nil {
		cs.logger.Warnf("Failed to add creator as collaborator: %v", err)
		// Don't fail the creation if we can't add creator as collaborator
	}

	return contestID, nil
}

func (cs *contestService) Get(tx *gorm.DB, currentUser schemas.User, contestID int64) (*schemas.Contest, error) {
	contest, err := cs.contestRepository.GetWithCount(tx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrNotFound
		}
		return nil, err
	}

	// Check visibility and permissions
	if !*contest.IsVisible {
		if currentUser.Role == types.UserRoleStudent {
			return nil, myerrors.ErrForbidden
		}
		if currentUser.Role == types.UserRoleTeacher && contest.CreatedBy != currentUser.ID {
			return nil, myerrors.ErrForbidden
		}
	}

	return &schemas.Contest{
		BaseContest: schemas.BaseContest{
			ID:          contest.ID,
			Name:        contest.Name,
			Description: contest.Description,
			CreatedBy:   contest.CreatedBy,
			StartAt:     contest.StartAt,
			EndAt:       contest.EndAt,
		},
		ParticipantCount: contest.ParticipantCount,
		TaskCount:        contest.TaskCount,
		Status:           getContestStatus(contest.StartAt, contest.EndAt),
	}, nil
}

func (cs *contestService) GetOngoingContests(tx *gorm.DB, currentUser schemas.User, paginationParams schemas.PaginationParams) (schemas.PaginatedResult[[]schemas.AvailableContest], error) {
	if paginationParams.Sort == "" {
		paginationParams.Sort = defaultContestSort
	}

	contestsWithStats, totalCount, err := cs.contestRepository.GetOngoingContestsWithStats(tx, currentUser.ID, paginationParams.Offset, paginationParams.Limit, paginationParams.Sort)
	if err != nil {
		return schemas.PaginatedResult[[]schemas.AvailableContest]{}, err
	}

	visibleContests := []models.ContestWithStats{}
	for _, contest := range contestsWithStats {
		if cs.isContestVisibleToUser(tx, &contest.Contest, &currentUser) {
			visibleContests = append(visibleContests, contest)
		}
	}

	result := make([]schemas.AvailableContest, len(visibleContests))
	for i, contest := range visibleContests {
		result[i] = *ContestWithStatsToAvailableContest(&contest)
	}

	return schemas.NewPaginatedResult(result, paginationParams.Offset, paginationParams.Limit, totalCount), nil
}

func (cs *contestService) GetPastContests(tx *gorm.DB, currentUser schemas.User, paginationParams schemas.PaginationParams) (schemas.PaginatedResult[[]schemas.AvailableContest], error) {
	if paginationParams.Sort == "" {
		paginationParams.Sort = defaultContestSort
	}

	contestsWithStats, totalCount, err := cs.contestRepository.GetPastContestsWithStats(tx, currentUser.ID, paginationParams.Offset, paginationParams.Limit, paginationParams.Sort)
	if err != nil {
		return schemas.PaginatedResult[[]schemas.AvailableContest]{}, err
	}

	visibleContests := []models.ContestWithStats{}
	for _, contest := range contestsWithStats {
		if cs.isContestVisibleToUser(tx, &contest.Contest, &currentUser) {
			visibleContests = append(visibleContests, contest)
		}
	}

	result := make([]schemas.AvailableContest, len(visibleContests))
	for i, contest := range visibleContests {
		result[i] = *ContestWithStatsToAvailableContest(&contest)
	}

	return schemas.NewPaginatedResult(result, paginationParams.Offset, paginationParams.Limit, totalCount), nil
}

func (cs *contestService) GetUpcomingContests(tx *gorm.DB, currentUser schemas.User, paginationParams schemas.PaginationParams) (schemas.PaginatedResult[[]schemas.AvailableContest], error) {
	if paginationParams.Sort == "" {
		paginationParams.Sort = defaultContestSort
	}

	contestsWithStats, totalCount, err := cs.contestRepository.GetUpcomingContestsWithStats(tx, currentUser.ID, paginationParams.Offset, paginationParams.Limit, paginationParams.Sort)
	if err != nil {
		return schemas.PaginatedResult[[]schemas.AvailableContest]{}, err
	}

	visibleContests := []models.ContestWithStats{}
	for _, contest := range contestsWithStats {
		if cs.isContestVisibleToUser(tx, &contest.Contest, &currentUser) {
			visibleContests = append(visibleContests, contest)
		}
	}

	result := make([]schemas.AvailableContest, len(visibleContests))
	for i, contest := range visibleContests {
		result[i] = *ContestWithStatsToAvailableContest(&contest)
	}

	return schemas.NewPaginatedResult(result, paginationParams.Offset, paginationParams.Limit, totalCount), nil
}

func (cs *contestService) isContestVisibleToUser(tx *gorm.DB, contest *models.Contest, user *schemas.User) bool {
	if *contest.IsVisible {
		return true
	}

	if user.Role == types.UserRoleAdmin {
		return true
	}
	if user.Role == types.UserRoleTeacher && contest.CreatedBy == user.ID {
		return true
	}
	isParticipant, err := cs.contestRepository.IsUserParticipant(tx, contest.ID, user.ID)
	if err != nil {
		cs.logger.Errorf("Error checking if user is participant: %v", err)
		return false
	}
	return isParticipant
}

func (cs *contestService) Edit(tx *gorm.DB, currentUser schemas.User, contestID int64, editInfo *schemas.EditContest) (*schemas.CreatedContest, error) {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return nil, err
	}

	validate, err := utils.NewValidator()
	if err != nil {
		return nil, err
	}
	if err := validate.Struct(editInfo); err != nil {
		return nil, err
	}

	model, err := cs.contestRepository.Get(tx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrNotFound
		}
		return nil, err
	}

	// Check permissions using collaborator system
	hasPermission, err := cs.hasContestPermission(tx, contestID, currentUser, types.PermissionEdit)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, myerrors.ErrForbidden
	}

	cs.updateModel(model, editInfo)

	newModel, err := cs.contestRepository.EditWithStats(tx, contestID, model)
	if err != nil {
		return nil, err
	}

	return ContestWithStatsToCreatedContest(newModel), nil
}

func (cs *contestService) Delete(tx *gorm.DB, currentUser schemas.User, contestID int64) error {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return err
	}

	// Check permissions using collaborator system - only manage permission can delete
	// This also verifies the contest exists
	hasPermission, err := cs.hasContestPermission(tx, contestID, currentUser, types.PermissionManage)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrNotFound
		}
		return err
	}
	if !hasPermission {
		return myerrors.ErrForbidden
	}

	return cs.contestRepository.Delete(tx, contestID)
}

func (cs *contestService) RegisterForContest(tx *gorm.DB, currentUser schemas.User, contestID int64) error {
	// Get the contest
	contest, err := cs.contestRepository.Get(tx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrNotFound
		}
		return err
	}

	// Check if contest is visible to user
	if !*contest.IsVisible {
		if currentUser.Role == types.UserRoleStudent {
			return myerrors.ErrForbidden
		}
		if currentUser.Role == types.UserRoleTeacher && contest.CreatedBy != currentUser.ID {
			return myerrors.ErrForbidden
		}
	}

	// Check if registration is open
	if contest.IsRegistrationOpen == nil || !*contest.IsRegistrationOpen {
		return myerrors.ErrContestRegistrationClosed
	}

	// Check if contest has ended
	if contest.EndAt != nil && contest.EndAt.Before(time.Now()) {
		return myerrors.ErrContestEnded
	}

	// Check if user is already a participant
	isParticipant, err := cs.contestRepository.IsUserParticipant(tx, contestID, currentUser.ID)
	if err != nil {
		return err
	}
	if isParticipant {
		return myerrors.ErrAlreadyParticipant
	}

	// Check if user already has a pending registration
	hasPending, err := cs.contestRepository.IsPendingRegistrationExists(tx, contestID, currentUser.ID)
	if err != nil {
		return err
	}
	if hasPending {
		return myerrors.ErrAlreadyRegistered
	}

	// Create pending registration
	registration := &models.ContestRegistrationRequests{
		ContestID: contestID,
		UserID:    currentUser.ID,
		Status:    types.RegistrationRequestStatusPending,
	}

	_, err = cs.contestRepository.CreatePendingRegistration(tx, registration)
	if err != nil {
		return err
	}

	return nil
}

func (cs *contestService) updateModel(model *models.Contest, editInfo *schemas.EditContest) {
	if editInfo.Name != nil {
		model.Name = *editInfo.Name
	}
	if editInfo.Description != nil {
		model.Description = *editInfo.Description
	}
	if editInfo.StartAt != nil {
		model.StartAt = *editInfo.StartAt
	}
	// TODO: handle when setting to nil is intended
	if editInfo.EndAt != nil {
		model.EndAt = editInfo.EndAt
	}
	if editInfo.IsRegistrationOpen != nil {
		model.IsRegistrationOpen = editInfo.IsRegistrationOpen
	}
	if editInfo.IsSubmissionOpen != nil {
		model.IsSubmissionOpen = editInfo.IsSubmissionOpen
	}
	if editInfo.IsVisible != nil {
		model.IsVisible = editInfo.IsVisible
	}
}

func (cs *contestService) GetTasksForContest(tx *gorm.DB, currentUser schemas.User, contestID int64) ([]schemas.ContestTask, error) {
	contest, err := cs.contestRepository.Get(tx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrNotFound
		}
		return nil, err
	}

	if !cs.isContestVisibleToUser(tx, contest, &currentUser) {
		return nil, myerrors.ErrForbidden
	}

	relations, err := cs.contestRepository.GetContestTasksWithSettings(tx, contestID)
	if err != nil {
		return nil, err
	}

	result := make([]schemas.ContestTask, len(relations))
	for i, rel := range relations {
		result[i] = schemas.ContestTask{
			Task:             *TaskToSchema(&rel.Task),
			CreatorName:      rel.Task.Author.Name,
			StartAt:          rel.StartAt,
			EndAt:            rel.EndAt,
			IsSubmissionOpen: rel.IsSubmissionOpen,
		}
	}
	return result, nil
}

func (cs *contestService) GetTaskProgressForContest(tx *gorm.DB, currentUser schemas.User, contestID int64) ([]schemas.TaskWithContestStats, error) {
	contest, err := cs.contestRepository.Get(tx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrNotFound
		}
		return nil, err
	}

	if !cs.isContestVisibleToUser(tx, contest, &currentUser) {
		return nil, myerrors.ErrForbidden
	}

	// Fetch raw task models (repository method unchanged - does not include per-contest timing fields)
	taskModels, err := cs.contestRepository.GetTasksForContest(tx, contestID)
	if err != nil {
		return nil, err
	}

	result := make([]schemas.TaskWithContestStats, len(taskModels))
	for i, task := range taskModels {
		// Get submission statistics for this user and task
		bestScore, err := cs.submissionRepository.GetBestScoreForTaskByUser(tx, task.ID, currentUser.ID)
		if err != nil {
			cs.logger.Warnf("Error getting best score for task %d, user %d: %v", task.ID, currentUser.ID, err)
		}

		attemptCount, err := cs.submissionRepository.GetAttemptCountForTaskByUser(tx, task.ID, currentUser.ID)
		if err != nil {
			cs.logger.Warnf("Error getting attempt count for task %d, user %d: %v", task.ID, currentUser.ID, err)
		}

		result[i] = schemas.TaskWithContestStats{
			ID:        task.ID,
			Title:     task.Title,
			CreatedBy: task.CreatedBy,
			CreatedAt: task.CreatedAt,
			UpdatedAt: task.UpdatedAt,
			AttemptsSummary: schemas.AttemptsSummary{
				BestScore:    bestScore,
				AttemptCount: attemptCount,
			},
		}
	}
	return result, nil
}

func (cs *contestService) GetAssignableTasks(tx *gorm.DB, currentUser schemas.User, contestID int64) ([]schemas.Task, error) {
	// Only admins and teachers can see available tasks for adding to contest
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return nil, myerrors.ErrForbidden
	}

	// Check permissions using collaborator system - need edit permission
	// This also verifies the contest exists
	hasPermission, err := cs.hasContestPermission(tx, contestID, currentUser, types.PermissionEdit)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrNotFound
		}
		return nil, err
	}
	if !hasPermission {
		return nil, myerrors.ErrForbidden
	}

	tasks, err := cs.contestRepository.GetAssignableTasks(tx, contestID)
	if err != nil {
		return nil, err
	}

	result := make([]schemas.Task, len(tasks))
	for i, task := range tasks {
		result[i] = schemas.Task{
			ID:        task.ID,
			Title:     task.Title,
			CreatedBy: task.CreatedBy,
			CreatedAt: task.CreatedAt,
			UpdatedAt: task.UpdatedAt,
		}
	}
	return result, nil
}

func (cs *contestService) GetUserContests(tx *gorm.DB, userID int64) (schemas.UserContestsWithStats, error) {
	user, err := cs.userRepository.Get(tx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return schemas.UserContestsWithStats{}, myerrors.ErrNotFound
		}
		return schemas.UserContestsWithStats{}, err
	}

	contestsWithStats, err := cs.contestRepository.GetContestsForUserWithStats(tx, user.ID)
	if err != nil {
		return schemas.UserContestsWithStats{}, err
	}

	ongoing := make([]schemas.ContestWithStats, 0)
	past := make([]schemas.PastContestWithStats, 0)
	upcoming := make([]schemas.ContestWithStats, 0)
	now := time.Now()

	for _, contest := range contestsWithStats {
		contestSchema := *ParticipantContestStatsToSchema(&contest)

		// Determine contest status
		if contest.StartAt.After(now) {
			// Contest is upcoming
			upcoming = append(upcoming, contestSchema)
		} else {
			// Contest has started
			if contest.EndAt == nil || contest.EndAt.After(now) {
				// Contest is ongoing
				ongoing = append(ongoing, contestSchema)
			} else {
				// Contest has ended
				pastContest := *ParticipantContestStatsToPastSchema(&contest)
				past = append(past, pastContest)
			}
		}
	}

	result := schemas.UserContestsWithStats{
		Ongoing:  ongoing,
		Past:     past,
		Upcoming: upcoming,
	}

	return result, nil
}

func (cs *contestService) AddTaskToContest(tx *gorm.DB, currentUser *schemas.User, contestID int64, request *schemas.AddTaskToContest) error {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return err
	}
	// Check if contest exists
	contest, err := cs.contestRepository.Get(tx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrNotFound
		}
		return err
	}
	_, err = cs.taskRepository.Get(tx, request.TaskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrNotFound
		}
		return err
	}

	// Check permissions using collaborator system - need edit permission
	hasPermission, err := cs.hasContestPermission(tx, contestID, *currentUser, types.PermissionEdit)
	if err != nil {
		return err
	}
	if !hasPermission {
		return myerrors.ErrForbidden
	}

	startAt := time.Now()
	if request.StartAt != nil {
		startAt = *request.StartAt
	}
	endAt := contest.EndAt
	if request.EndAt != nil {
		endAt = request.EndAt
	}
	if endAt != nil && startAt.After(*endAt) {
		return errors.New("task end time cannot be before start time")
	}

	taskContest := models.ContestTask{
		ContestID:        contestID,
		TaskID:           request.TaskID,
		StartAt:          startAt,
		EndAt:            endAt,
		IsSubmissionOpen: true,
	}

	return cs.contestRepository.AddTaskToContest(tx, taskContest)
}

func ContestWithStatsToCreatedContest(model *models.ContestWithStats) *schemas.CreatedContest {
	return &schemas.CreatedContest{
		BaseContest: schemas.BaseContest{
			ID:          model.ID,
			Name:        model.Name,
			Description: model.Description,
			CreatedBy:   model.CreatedBy,
			StartAt:     model.StartAt,
			EndAt:       model.EndAt,
		},
		IsRegistrationOpen: model.IsRegistrationOpen,
		IsSubmissionOpen:   model.IsSubmissionOpen,
		IsVisible:          model.IsVisible,
	}
}

func ContestWithStatsToAvailableContest(model *models.ContestWithStats) *schemas.AvailableContest {
	registrationStatus := "registrationClosed"

	// Check if contest has ended
	contestEnded := model.EndAt != nil && model.EndAt.Before(time.Now())

	// If user is already a participant
	if model.IsParticipant {
		registrationStatus = "registered"
	} else if model.HasPendingReg {
		// If user has pending registration
		registrationStatus = "awaitingApproval"
	} else if !contestEnded && model.IsRegistrationOpen != nil && *model.IsRegistrationOpen {
		// If registration is open, contest hasn't ended, and user can register
		registrationStatus = "canRegister"
	}
	// Default is "registrationClosed" for all other cases (registration closed or contest ended)

	status := getContestStatus(model.StartAt, model.EndAt)
	return &schemas.AvailableContest{
		Contest: schemas.Contest{
			BaseContest: schemas.BaseContest{
				ID:          model.ID,
				Name:        model.Name,
				Description: model.Description,
				CreatedBy:   model.CreatedBy,
				StartAt:     model.StartAt,
				EndAt:       model.EndAt,
			},
			ParticipantCount: model.ParticipantCount,
			TaskCount:        model.TaskCount,
			Status:           status,
		},
		RegistrationStatus: registrationStatus,
	}
}

func getContestStatus(startAt time.Time, endAt *time.Time) types.ContestStatus {
	status := types.ContestStatusUpcoming
	now := time.Now()

	if !startAt.After(now) {
		// Started
		if endAt == nil || endAt.After(now) {
			status = types.ContestStatusOngoing
		} else {
			status = types.ContestStatusPast
		}
	} else if endAt != nil && !endAt.After(now) {
		status = types.ContestStatusPast
	}
	return status
}

func ParticipantContestStatsToSchema(model *models.ParticipantContestStats) *schemas.ContestWithStats {
	status := getContestStatus(model.StartAt, model.EndAt)
	return &schemas.ContestWithStats{
		Contest: schemas.Contest{
			BaseContest: schemas.BaseContest{
				ID:          model.ID,
				Name:        model.Name,
				Description: model.Description,
				CreatedBy:   model.CreatedBy,
				StartAt:     model.StartAt,
				EndAt:       model.EndAt,
			},
			ParticipantCount: model.ParticipantCount,
			TaskCount:        model.TaskCount,
			Status:           status,
		},
		SolvedTaskCount: model.SolvedTaskCount,
	}
}

func (cs *contestService) GetRegistrationRequests(tx *gorm.DB, currentUser schemas.User, contestID int64, status types.RegistrationRequestStatus) ([]schemas.RegistrationRequest, error) {
	// Check authorization - need manage permission to view registration requests
	// This also verifies the contest exists
	hasPermission, err := cs.hasContestPermission(tx, contestID, currentUser, types.PermissionManage)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrNotFound
		}
		return nil, err
	}
	if !hasPermission {
		return nil, myerrors.ErrForbidden
	}

	// Get registration requests from repository
	requests, err := cs.contestRepository.GetRegistrationRequests(tx, contestID, status)
	if err != nil {
		return nil, err
	}

	// Convert to schema
	result := make([]schemas.RegistrationRequest, len(requests))
	for i, req := range requests {
		result[i] = schemas.RegistrationRequest{
			ID:        req.ID,
			ContestID: req.ContestID,
			UserID:    req.UserID,
			User:      *UserToSchema(&req.User),
			CreatedAt: req.CreatedAt,
			Status:    req.Status,
		}
	}

	return result, nil
}

func (cs *contestService) ApproveRegistrationRequest(tx *gorm.DB, currentUser schemas.User, contestID, userID int64) error {
	// Check permissions using collaborator system - need manage permission
	// This also verifies the contest exists
	hasPermission, err := cs.hasContestPermission(tx, contestID, currentUser, types.PermissionManage)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrNotFound
		}
		return err
	}
	if !hasPermission {
		return myerrors.ErrForbidden
	}

	// Check if user exists
	_, err = cs.userRepository.Get(tx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrNotFound
		}
		return err
	}

	// Check if already a participant
	isParticipant, err := cs.contestRepository.IsUserParticipant(tx, contestID, userID)
	if err != nil {
		return err
	}

	// Check if pending registration exists
	request, err := cs.contestRepository.GetPendingRegistrationRequest(tx, contestID, userID)
	if err != nil {
		return err
	}
	hasPending := request != nil
	if !hasPending {
		return myerrors.ErrNoPendingRegistration
	}
	if isParticipant {
		err = cs.contestRepository.DeleteRegistrationRequest(tx, request.ID)
		if err != nil {
			return err
		}
		return myerrors.ErrAlreadyParticipant
	}
	err = cs.contestRepository.UpdateRegistrationRequestStatus(tx, request.ID, types.RegistrationRequestStatusApproved)
	if err != nil {
		return err
	}
	return cs.contestRepository.CreateContestParticipant(tx, contestID, userID)
}

func (cs *contestService) RejectRegistrationRequest(tx *gorm.DB, currentUser schemas.User, contestID, userID int64) error {
	// Check permissions using collaborator system - need manage permission
	// This also verifies the contest exists
	hasPermission, err := cs.hasContestPermission(tx, contestID, currentUser, types.PermissionManage)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrNotFound
		}
		return err
	}
	if !hasPermission {
		return myerrors.ErrForbidden
	}

	// Check if user exists
	_, err = cs.userRepository.Get(tx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrNotFound
		}
		return err
	}

	// Check if already a participant
	isParticipant, err := cs.contestRepository.IsUserParticipant(tx, contestID, userID)
	if err != nil {
		return err
	}

	// Check if pending registration exists
	request, err := cs.contestRepository.GetPendingRegistrationRequest(tx, contestID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrNoPendingRegistration
		}
		return err
	}
	if request == nil {
		return myerrors.ErrNoPendingRegistration
	}
	if isParticipant {
		err = cs.contestRepository.DeleteRegistrationRequest(tx, request.ID)
		if err != nil {
			return err
		}
		return myerrors.ErrAlreadyParticipant
	}

	return cs.contestRepository.UpdateRegistrationRequestStatus(tx, request.ID, types.RegistrationRequestStatusRejected)
}

func (cs *contestService) IsTaskInContest(tx *gorm.DB, contestID, taskID int64) (bool, error) {
	contestTask, err := cs.contestRepository.GetContestTask(tx, contestID, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return contestTask != nil, nil
}

func (cs *contestService) IsUserParticipant(tx *gorm.DB, contestID, userID int64) (bool, error) {
	return cs.contestRepository.IsUserParticipant(tx, contestID, userID)
}

func (cs *contestService) GetTaskContests(tx *gorm.DB, taskID int64) ([]int64, error) {
	return cs.contestRepository.GetTaskContests(tx, taskID)
}

func (cs *contestService) ValidateContestSubmission(tx *gorm.DB, contestID, taskID, userID int64) error {
	// Get contest
	contest, err := cs.contestRepository.Get(tx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrNotFound
		}
		return err
	}

	// Check if contest submissions are open
	if contest.IsSubmissionOpen == nil || !*contest.IsSubmissionOpen {
		return myerrors.ErrContestSubmissionClosed
	}

	// Check if contest has started
	if time.Now().Before(contest.StartAt) {
		return myerrors.ErrContestNotStarted
	}

	// Check if contest has ended
	if contest.EndAt != nil && time.Now().After(*contest.EndAt) {
		return myerrors.ErrContestEnded
	}

	// Check if user is a participant
	isParticipant, err := cs.contestRepository.IsUserParticipant(tx, contestID, userID)
	if err != nil {
		return err
	}
	if !isParticipant {
		return myerrors.ErrNotContestParticipant
	}

	// Check if task is in contest
	contestTask, err := cs.contestRepository.GetContestTask(tx, contestID, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrTaskNotInContest
		}
		return err
	}

	// Check if task submissions are open for this contest
	if !contestTask.IsSubmissionOpen {
		return myerrors.ErrTaskSubmissionClosed
	}

	// Check if task submission period has started
	if time.Now().Before(contestTask.StartAt) {
		return myerrors.ErrTaskNotStarted
	}

	// Check if task submission period has ended
	if contestTask.EndAt != nil && time.Now().After(*contestTask.EndAt) {
		return myerrors.ErrTaskSubmissionEnded
	}

	return nil
}

func (cs *contestService) GetContestsCreatedByUser(tx *gorm.DB, userID int64, paginationParams schemas.PaginationParams) (schemas.PaginatedResult[[]schemas.CreatedContest], error) {
	if paginationParams.Sort == "" {
		paginationParams.Sort = defaultContestSort
	}

	contests, totalCount, err := cs.contestRepository.GetAllForCreator(tx, userID, paginationParams.Offset, paginationParams.Limit, paginationParams.Sort)
	if err != nil {
		return schemas.PaginatedResult[[]schemas.CreatedContest]{}, err
	}

	result := make([]schemas.CreatedContest, len(contests))
	for i, contest := range contests {
		result[i] = *ContestToCreatedSchema(&contest)
	}
	return schemas.NewPaginatedResult(result, paginationParams.Offset, paginationParams.Limit, totalCount), nil
}

func (cs *contestService) GetContestTask(tx *gorm.DB, currentUser *schemas.User, contestID, taskID int64) (*schemas.TaskDetailed, error) {
	contest, err := cs.contestRepository.Get(tx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrNotFound
		}
		return nil, err
	}
	if !cs.isContestVisibleToUser(tx, contest, currentUser) {
		return nil, myerrors.ErrForbidden
	}

	_, err = cs.contestRepository.GetContestTask(tx, contestID, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrNotFound
		}
		return nil, err
	}

	task, err := cs.taskService.Get(tx, *currentUser, taskID)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func ContestToCreatedSchema(model *models.Contest) *schemas.CreatedContest {
	return &schemas.CreatedContest{
		BaseContest: schemas.BaseContest{
			ID:          model.ID,
			Name:        model.Name,
			Description: model.Description,
			CreatedBy:   model.CreatedBy,
			StartAt:     model.StartAt,
			EndAt:       model.EndAt,
		},
		IsVisible:          model.IsVisible,
		IsRegistrationOpen: model.IsRegistrationOpen,
		IsSubmissionOpen:   model.IsSubmissionOpen,
	}
}

// hasContestPermission checks if the user has the required permission for the contest.
// Returns true if:
// 1. User is admin
// 2. User is the creator (which should also have manage permission via collaborator)
// 3. User has the required permission via collaborator
func (cs *contestService) hasContestPermission(tx *gorm.DB, contestID int64, user schemas.User, requiredPermission types.Permission) (bool, error) {
	// Fetch contest (verifies existence)
	contest, err := cs.contestRepository.Get(tx, contestID)
	if err != nil {
		return false, err
	}

	// Delegate permission logic to AccessControlService
	return cs.accessControlService.CanUserAccess(tx, models.ResourceTypeContest, contestID, user, contest.CreatedBy, requiredPermission)
}

func ParticipantContestStatsToPastSchema(contest *models.ParticipantContestStats) *schemas.PastContestWithStats {
	solvedPercentage := 0.0
	if contest.TaskCount > 0 {
		solvedPercentage = float64(contest.SolvedTaskCount) / float64(contest.TaskCount) * 100
	}
	return &schemas.PastContestWithStats{
		Contest: schemas.Contest{
			BaseContest: schemas.BaseContest{
				ID:          contest.ID,
				Name:        contest.Name,
				Description: contest.Description,
				CreatedBy:   contest.CreatedBy,
				StartAt:     contest.StartAt,
				EndAt:       contest.EndAt,
			},
			TaskCount:        contest.TaskCount,
			ParticipantCount: contest.ParticipantCount,
			Status:           getContestStatus(contest.StartAt, contest.EndAt),
		},
		SolvedTaskPercentage: solvedPercentage,
		Score:                contest.SolvedTestCount,
		MaximumScore:         contest.TestCount,
		Rank:                 0, // TODO: implement
	}
}

func NewContestService(contestRepository repository.ContestRepository, userRepository repository.UserRepository, submissionRepository repository.SubmissionRepository, taskRepository repository.TaskRepository, accessControlService AccessControlService, taskService TaskService) ContestService {
	return &contestService{
		contestRepository:    contestRepository,
		taskRepository:       taskRepository,
		userRepository:       userRepository,
		submissionRepository: submissionRepository,
		accessControlService: accessControlService,
		taskService:          taskService,
		logger:               utils.NewNamedLogger("contest_service"),
	}
}
