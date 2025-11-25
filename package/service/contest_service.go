package service

import (
	"fmt"
	"time"

	"github.com/mini-maxit/backend/internal/database"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ContestService interface {
	// Create creates a new contest
	Create(db database.Database, currentUser *schemas.User, contest *schemas.CreateContest) (int64, error)
	// Get retrieves a contest by ID
	Get(db database.Database, currentUser *schemas.User, contestID int64) (*schemas.Contest, error)
	// GetOngoingContests retrieves contests that are currently running
	GetOngoingContests(db database.Database, currentUser *schemas.User, paginationParams schemas.PaginationParams) (schemas.PaginatedResult[[]schemas.AvailableContest], error)
	// GetPastContests retrieves contests that have ended
	GetPastContests(db database.Database, currentUser *schemas.User, paginationParams schemas.PaginationParams) (schemas.PaginatedResult[[]schemas.AvailableContest], error)
	// GetUpcomingContests retrieves contests that haven't started yet
	GetUpcomingContests(db database.Database, currentUser *schemas.User, paginationParams schemas.PaginationParams) (schemas.PaginatedResult[[]schemas.AvailableContest], error)
	// Edit updates a contest
	Edit(db database.Database, currentUser *schemas.User, contestID int64, editInfo *schemas.EditContest) (*schemas.CreatedContest, error)
	// Delete removes a contest
	Delete(db database.Database, currentUser *schemas.User, contestID int64) error
	// RegisterForContest creates a pending registration for a contest
	RegisterForContest(db database.Database, currentUser *schemas.User, contestID int64) error
	// GetTasksForContest retrieves all contest task relations (with timing/submission flags) for a contest (for authorized users)
	GetTasksForContest(db database.Database, currentUser *schemas.User, contestID int64) ([]schemas.ContestTask, error)
	// GetTaskProgressForContest retrieves all tasks associated with a contest with submission stats for the requesting user
	GetTaskProgressForContest(db database.Database, currentUser *schemas.User, contestID int64) ([]schemas.TaskWithContestStats, error)
	// GetAssignableTasks retrieves all tasks NOT assigned to a contest (for authorized users)
	GetAssignableTasks(db database.Database, currentUser *schemas.User, contestID int64) ([]schemas.Task, error)
	// GetUserContests retrieves all contests a user is participating in
	GetUserContests(db database.Database, userID int64) (*schemas.UserContestsWithStats, error)
	// AddTaskToContest adds a task to a contest
	AddTaskToContest(db database.Database, currentUser *schemas.User, contestID int64, request *schemas.AddTaskToContest) error
	// GetRegistrationRequests retrieves pending registration requests for a contest
	GetRegistrationRequests(db database.Database, currentUser *schemas.User, contestID int64, statusFilter types.RegistrationRequestStatus) ([]schemas.RegistrationRequest, error)
	// ApproveRegistrationRequest approves a pending registration request for a contest
	ApproveRegistrationRequest(db database.Database, currentUser *schemas.User, contestID, userID int64) error
	// RejectRegistrationRequest rejects a pending registration request for a contest
	RejectRegistrationRequest(db database.Database, currentUser *schemas.User, contestID, userID int64) error
	// IsTaskInContest checks if a task is part of a contest
	IsTaskInContest(db database.Database, contestID, taskID int64) (bool, error)
	// IsUserParticipant checks if a user is a participant in a contest
	IsUserParticipant(db database.Database, contestID, userID int64) (bool, error)
	// GetTaskContests retrieves all contest IDs that a task is assigned to
	GetTaskContests(db database.Database, taskID int64) ([]int64, error)
	// ValidateContestSubmission validates if a user can submit a solution for a task in a contest
	// Returns an error if submission is not allowed (contest/task not open, user not participant, etc.)
	ValidateContestSubmission(db database.Database, contestID, taskID, userID int64) error

	GetContestsCreatedByUser(db database.Database, userID int64, paginationParams schemas.PaginationParams) (schemas.PaginatedResult[[]schemas.CreatedContest], error)
	// GetManagedContests retrieves contests where the user is listed in access_control (any permission)
	GetManagedContests(db database.Database, userID int64, paginationParams schemas.PaginationParams) (schemas.PaginatedResult[[]schemas.ManagedContest], error)
	GetContestTask(db database.Database, currentUser *schemas.User, contestID, taskID int64) (*schemas.TaskDetailed, error)
	// GetMyContestResults returns the results of the current user for a given contest
	GetMyContestResults(db database.Database, currentUser *schemas.User, contestID int64) (*schemas.ContestResults, error)
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

func (cs *contestService) Create(db database.Database, currentUser *schemas.User, contest *schemas.CreateContest) (int64, error) {
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

	contestID, err := cs.contestRepository.Create(db, model)
	if err != nil {
		return -1, err
	}

	// Automatically grant owner permission to the creator (immutable highest level)
	if err := cs.accessControlService.GrantOwnerAccess(db, models.ResourceTypeContest, contestID, currentUser.ID); err != nil {
		cs.logger.Warnf("Failed to grant owner permission: %v", err)
		// Don't fail the creation if we can't add owner permission entry
	}

	return contestID, nil
}

func (cs *contestService) Get(db database.Database, currentUser *schemas.User, contestID int64) (*schemas.Contest, error) {
	contest, err := cs.contestRepository.GetWithCount(db, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	if !cs.isContestVisibleToUser(db, &contest.Contest, currentUser) {
		return nil, errors.ErrForbidden
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

func (cs *contestService) GetOngoingContests(db database.Database, currentUser *schemas.User, paginationParams schemas.PaginationParams) (schemas.PaginatedResult[[]schemas.AvailableContest], error) {
	if paginationParams.Sort == "" {
		paginationParams.Sort = defaultContestSort
	}

	contestsWithStats, totalCount, err := cs.contestRepository.GetOngoingContestsWithStats(db, currentUser.ID, paginationParams.Offset, paginationParams.Limit, paginationParams.Sort)
	if err != nil {
		return schemas.PaginatedResult[[]schemas.AvailableContest]{}, err
	}

	visibleContests := []models.ContestWithStats{}
	for _, contest := range contestsWithStats {
		if cs.isContestVisibleToUser(db, &contest.Contest, currentUser) {
			visibleContests = append(visibleContests, contest)
		}
	}

	result := make([]schemas.AvailableContest, len(visibleContests))
	for i, contest := range visibleContests {
		result[i] = *ContestWithStatsToAvailableContest(&contest)
	}

	return schemas.NewPaginatedResult(result, paginationParams.Offset, paginationParams.Limit, totalCount), nil
}

func (cs *contestService) GetPastContests(db database.Database, currentUser *schemas.User, paginationParams schemas.PaginationParams) (schemas.PaginatedResult[[]schemas.AvailableContest], error) {
	if paginationParams.Sort == "" {
		paginationParams.Sort = defaultContestSort
	}

	contestsWithStats, totalCount, err := cs.contestRepository.GetPastContestsWithStats(db, currentUser.ID, paginationParams.Offset, paginationParams.Limit, paginationParams.Sort)
	if err != nil {
		return schemas.PaginatedResult[[]schemas.AvailableContest]{}, err
	}

	visibleContests := []models.ContestWithStats{}
	for _, contest := range contestsWithStats {
		if cs.isContestVisibleToUser(db, &contest.Contest, currentUser) {
			visibleContests = append(visibleContests, contest)
		}
	}

	result := make([]schemas.AvailableContest, len(visibleContests))
	for i, contest := range visibleContests {
		result[i] = *ContestWithStatsToAvailableContest(&contest)
	}

	return schemas.NewPaginatedResult(result, paginationParams.Offset, paginationParams.Limit, totalCount), nil
}

func (cs *contestService) GetUpcomingContests(db database.Database, currentUser *schemas.User, paginationParams schemas.PaginationParams) (schemas.PaginatedResult[[]schemas.AvailableContest], error) {
	if paginationParams.Sort == "" {
		paginationParams.Sort = defaultContestSort
	}

	contestsWithStats, totalCount, err := cs.contestRepository.GetUpcomingContestsWithStats(db, currentUser.ID, paginationParams.Offset, paginationParams.Limit, paginationParams.Sort)
	if err != nil {
		return schemas.PaginatedResult[[]schemas.AvailableContest]{}, err
	}

	visibleContests := []models.ContestWithStats{}
	for _, contest := range contestsWithStats {
		if cs.isContestVisibleToUser(db, &contest.Contest, currentUser) {
			visibleContests = append(visibleContests, contest)
		}
	}

	result := make([]schemas.AvailableContest, len(visibleContests))
	for i, contest := range visibleContests {
		result[i] = *ContestWithStatsToAvailableContest(&contest)
	}

	return schemas.NewPaginatedResult(result, paginationParams.Offset, paginationParams.Limit, totalCount), nil
}

func (cs *contestService) isContestVisibleToUser(db database.Database, contest *models.Contest, user *schemas.User) bool {
	if *contest.IsVisible {
		return true
	}

	if user.Role == types.UserRoleAdmin {
		return true
	}
	err := cs.hasContestPermission(db, contest.ID, user, types.PermissionEdit)
	if err != nil {
		if !errors.Is(err, errors.ErrForbidden) {
			return false
		}
	} else {
		return true
	}

	isParticipant, err := cs.contestRepository.IsUserParticipant(db, contest.ID, user.ID)
	if err != nil {
		cs.logger.Errorf("Error checking if user is participant: %v", err)
		return false
	}
	return isParticipant
}

func (cs *contestService) Edit(db database.Database, currentUser *schemas.User, contestID int64, editInfo *schemas.EditContest) (*schemas.CreatedContest, error) {
	err := cs.hasContestPermission(db, contestID, currentUser, types.PermissionEdit)
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

	model, err := cs.contestRepository.Get(db, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}

	// Check permissions using collaborator system

	cs.updateModel(model, editInfo)

	newModel, err := cs.contestRepository.EditWithStats(db, contestID, model)
	if err != nil {
		return nil, err
	}

	return ContestWithStatsToCreatedContest(newModel), nil
}

func (cs *contestService) Delete(db database.Database, currentUser *schemas.User, contestID int64) error {
	err := cs.hasContestPermission(db, contestID, currentUser, types.PermissionOwner)
	if err != nil {
		return err
	}

	return cs.contestRepository.Delete(db, contestID)
}

func (cs *contestService) RegisterForContest(db database.Database, currentUser *schemas.User, contestID int64) error {
	// Get the contest
	contest, err := cs.contestRepository.Get(db, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrNotFound
		}
		return err
	}

	if !cs.isContestVisibleToUser(db, contest, currentUser) {
		return errors.ErrNotFound
	}

	// Check if registration is open
	if contest.IsRegistrationOpen == nil || !*contest.IsRegistrationOpen {
		return errors.ErrContestRegistrationClosed
	}

	// Check if contest has ended
	if contest.EndAt != nil && contest.EndAt.Before(time.Now()) {
		return errors.ErrContestEnded
	}

	// Check if user is already a participant
	isParticipant, err := cs.contestRepository.IsUserParticipant(db, contestID, currentUser.ID)
	if err != nil {
		return err
	}
	if isParticipant {
		return errors.ErrAlreadyParticipant
	}

	// Check if user already has a pending registration
	hasPending, err := cs.contestRepository.IsPendingRegistrationExists(db, contestID, currentUser.ID)
	if err != nil {
		return err
	}
	if hasPending {
		return errors.ErrAlreadyRegistered
	}

	// Create pending registration
	registration := &models.ContestRegistrationRequests{
		ContestID: contestID,
		UserID:    currentUser.ID,
		Status:    types.RegistrationRequestStatusPending,
	}

	_, err = cs.contestRepository.CreatePendingRegistration(db, registration)
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

func (cs *contestService) GetTasksForContest(db database.Database, currentUser *schemas.User, contestID int64) ([]schemas.ContestTask, error) {
	contest, err := cs.contestRepository.Get(db, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}

	if !cs.isContestVisibleToUser(db, contest, currentUser) {
		return nil, errors.ErrForbidden
	}

	relations, err := cs.contestRepository.GetContestTasksWithSettings(db, contestID)
	if err != nil {
		return nil, err
	}

	result := make([]schemas.ContestTask, len(relations))
	for i, rel := range relations {
		result[i] = schemas.ContestTask{
			Task:             *TaskToInfoSchema(&rel.Task),
			CreatorName:      rel.Task.Author.Name,
			StartAt:          rel.StartAt,
			EndAt:            rel.EndAt,
			IsSubmissionOpen: rel.IsSubmissionOpen,
		}
	}
	return result, nil
}

func (cs *contestService) GetTaskProgressForContest(db database.Database, currentUser *schemas.User, contestID int64) ([]schemas.TaskWithContestStats, error) {
	contest, err := cs.contestRepository.Get(db, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}

	if !cs.isContestVisibleToUser(db, contest, currentUser) {
		return nil, errors.ErrForbidden
	}

	// Fetch raw task models (repository method unchanged - does not include per-contest timing fields)
	taskModels, err := cs.contestRepository.GetTasksForContest(db, contestID)
	if err != nil {
		return nil, err
	}

	result := make([]schemas.TaskWithContestStats, len(taskModels))
	for i, task := range taskModels {
		// Get submission statistics for this user and task
		bestScore, err := cs.submissionRepository.GetBestScoreForTaskByUser(db, task.ID, currentUser.ID)
		if err != nil {
			cs.logger.Warnf("Error getting best score for task %d, user %d: %v", task.ID, currentUser.ID, err)
		}

		attemptCount, err := cs.submissionRepository.GetAttemptCountForTaskByUser(db, task.ID, currentUser.ID)
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

func (cs *contestService) GetAssignableTasks(db database.Database, currentUser *schemas.User, contestID int64) ([]schemas.Task, error) {
	err := cs.hasContestPermission(db, contestID, currentUser, types.PermissionEdit)
	if err != nil {
		return nil, err
	}

	tasks, err := cs.contestRepository.GetAssignableTasks(db, contestID)
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

func (cs *contestService) GetUserContests(db database.Database, userID int64) (*schemas.UserContestsWithStats, error) {
	user, err := cs.userRepository.Get(db, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}

	contestsWithStats, err := cs.contestRepository.GetContestsForUserWithStats(db, user.ID)
	if err != nil {
		return nil, err
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

	return &result, nil
}

func (cs *contestService) AddTaskToContest(db database.Database, currentUser *schemas.User, contestID int64, request *schemas.AddTaskToContest) error {
	err := cs.hasContestPermission(db, contestID, currentUser, types.PermissionEdit)
	if err != nil {
		return err
	}
	contest, err := cs.contestRepository.Get(db, contestID)
	if err != nil {
		return err
	}

	_, err = cs.taskRepository.Get(db, request.TaskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrNotFound
		}
		return err
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
		return errors.ErrEndBeforeStart
	}

	taskContest := models.ContestTask{
		ContestID:        contestID,
		TaskID:           request.TaskID,
		StartAt:          startAt,
		EndAt:            endAt,
		IsSubmissionOpen: true,
	}

	return cs.contestRepository.AddTaskToContest(db, taskContest)
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

func (cs *contestService) GetRegistrationRequests(db database.Database, currentUser *schemas.User, contestID int64, status types.RegistrationRequestStatus) ([]schemas.RegistrationRequest, error) {
	err := cs.hasContestPermission(db, contestID, currentUser, types.PermissionManage)
	if err != nil {
		return nil, err
	}

	// Get registration requests from repository
	requests, err := cs.contestRepository.GetRegistrationRequests(db, contestID, status)
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

func (cs *contestService) ApproveRegistrationRequest(db database.Database, currentUser *schemas.User, contestID, userID int64) error {
	err := cs.hasContestPermission(db, contestID, currentUser, types.PermissionEdit)
	if err != nil {
		return err
	}

	// Check if user exists
	_, err = cs.userRepository.Get(db, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrNotFound
		}
		return err
	}

	// Check if already a participant
	isParticipant, err := cs.contestRepository.IsUserParticipant(db, contestID, userID)
	if err != nil {
		return err
	}

	// Check if pending registration exists
	request, err := cs.contestRepository.GetPendingRegistrationRequest(db, contestID, userID)
	if err != nil {
		return err
	}
	hasPending := request != nil
	if !hasPending {
		return errors.ErrNoPendingRegistration
	}
	if isParticipant {
		err = cs.contestRepository.DeleteRegistrationRequest(db, request.ID)
		if err != nil {
			return err
		}
		return errors.ErrAlreadyParticipant
	}
	err = cs.contestRepository.UpdateRegistrationRequestStatus(db, request.ID, types.RegistrationRequestStatusApproved)
	if err != nil {
		return err
	}
	return cs.contestRepository.CreateContestParticipant(db, contestID, userID)
}

func (cs *contestService) RejectRegistrationRequest(db database.Database, currentUser *schemas.User, contestID, userID int64) error {
	err := cs.hasContestPermission(db, contestID, currentUser, types.PermissionEdit)
	if err != nil {
		return err
	}

	// Check if user exists
	_, err = cs.userRepository.Get(db, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrNotFound
		}
		return err
	}

	// Check if already a participant
	isParticipant, err := cs.contestRepository.IsUserParticipant(db, contestID, userID)
	if err != nil {
		return err
	}

	// Check if pending registration exists
	request, err := cs.contestRepository.GetPendingRegistrationRequest(db, contestID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrNoPendingRegistration
		}
		return err
	}
	if request == nil {
		return errors.ErrNoPendingRegistration
	}
	if isParticipant {
		err = cs.contestRepository.DeleteRegistrationRequest(db, request.ID)
		if err != nil {
			return err
		}
		return errors.ErrAlreadyParticipant
	}

	return cs.contestRepository.UpdateRegistrationRequestStatus(db, request.ID, types.RegistrationRequestStatusRejected)
}

func (cs *contestService) IsTaskInContest(db database.Database, contestID, taskID int64) (bool, error) {
	contestTask, err := cs.contestRepository.GetContestTask(db, contestID, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return contestTask != nil, nil
}

func (cs *contestService) IsUserParticipant(db database.Database, contestID, userID int64) (bool, error) {
	return cs.contestRepository.IsUserParticipant(db, contestID, userID)
}

func (cs *contestService) GetTaskContests(db database.Database, taskID int64) ([]int64, error) {
	return cs.contestRepository.GetTaskContests(db, taskID)
}

func (cs *contestService) ValidateContestSubmission(db database.Database, contestID, taskID, userID int64) error {
	// Get contest
	contest, err := cs.contestRepository.Get(db, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrNotFound
		}
		return err
	}

	// Check if contest submissions are open
	if contest.IsSubmissionOpen == nil || !*contest.IsSubmissionOpen {
		return errors.ErrContestSubmissionClosed
	}

	// Check if contest has started
	if time.Now().Before(contest.StartAt) {
		return errors.ErrContestNotStarted
	}

	// Check if contest has ended
	if contest.EndAt != nil && time.Now().After(*contest.EndAt) {
		return errors.ErrContestEnded
	}

	// Check if user is a participant
	isParticipant, err := cs.contestRepository.IsUserParticipant(db, contestID, userID)
	if err != nil {
		return err
	}
	if !isParticipant {
		return errors.ErrNotContestParticipant
	}

	// Check if task is in contest
	contestTask, err := cs.contestRepository.GetContestTask(db, contestID, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrTaskNotInContest
		}
		return err
	}

	// Check if task submissions are open for this contest
	if !contestTask.IsSubmissionOpen {
		return errors.ErrTaskSubmissionClosed
	}

	// Check if task submission period has started
	if time.Now().Before(contestTask.StartAt) {
		return errors.ErrTaskNotStarted
	}

	// Check if task submission period has ended
	if contestTask.EndAt != nil && time.Now().After(*contestTask.EndAt) {
		return errors.ErrTaskSubmissionEnded
	}

	return nil
}

func (cs *contestService) GetContestsCreatedByUser(db database.Database, userID int64, paginationParams schemas.PaginationParams) (schemas.PaginatedResult[[]schemas.CreatedContest], error) {
	if paginationParams.Sort == "" {
		paginationParams.Sort = defaultContestSort
	}

	contests, totalCount, err := cs.contestRepository.GetAllForCreator(db, userID, paginationParams.Offset, paginationParams.Limit, paginationParams.Sort)
	if err != nil {
		return schemas.PaginatedResult[[]schemas.CreatedContest]{}, err
	}

	result := make([]schemas.CreatedContest, len(contests))
	for i, contest := range contests {
		result[i] = *ContestToCreatedSchema(&contest)
	}
	return schemas.NewPaginatedResult(result, paginationParams.Offset, paginationParams.Limit, totalCount), nil
}

// GetManagedContests retrieves contests where the user has any collaborator access (view/edit/manage)
func (cs *contestService) GetManagedContests(db database.Database, userID int64, paginationParams schemas.PaginationParams) (schemas.PaginatedResult[[]schemas.ManagedContest], error) {
	if paginationParams.Sort == "" {
		paginationParams.Sort = defaultContestSort
	}

	contests, totalCount, err := cs.contestRepository.GetAllForCollaborator(db, userID, paginationParams.Offset, paginationParams.Limit, paginationParams.Sort)
	if err != nil {
		return schemas.PaginatedResult[[]schemas.ManagedContest]{}, err
	}

	result := make([]schemas.ManagedContest, len(contests))
	for i, contest := range contests {
		result[i] = *ManagedContestToSchema(&contest)
	}
	return schemas.NewPaginatedResult(result, paginationParams.Offset, paginationParams.Limit, totalCount), nil
}

func (cs *contestService) GetContestTask(db database.Database, currentUser *schemas.User, contestID, taskID int64) (*schemas.TaskDetailed, error) {
	contest, err := cs.contestRepository.Get(db, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	if !cs.isContestVisibleToUser(db, contest, currentUser) {
		return nil, errors.ErrForbidden
	}

	_, err = cs.contestRepository.GetContestTask(db, contestID, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}

	task, err := cs.taskService.Get(db, currentUser, taskID)
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
		CreatedAt:          model.CreatedAt,
	}
}

func (cs *contestService) GetMyContestResults(db database.Database, currentUser *schemas.User, contestID int64) (*schemas.ContestResults, error) {
	// Get contest to verify it exists and user has access
	contest, err := cs.contestRepository.Get(db, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}

	// Check if user has access to the contest
	if !cs.isContestVisibleToUser(db, contest, currentUser) {
		return nil, errors.ErrForbidden
	}

	// Get all tasks for this contest
	tasks, err := cs.contestRepository.GetTasksForContest(db, contestID)
	if err != nil {
		return nil, err
	}

	// For each task, get the user's statistics
	taskResults := make([]schemas.TaskResult, 0, len(tasks))
	for _, task := range tasks {
		// Get attempt count
		attemptCount, err := cs.submissionRepository.GetAttemptCountForTaskByUser(db, task.ID, currentUser.ID)
		if err != nil {
			cs.logger.Warnf("Error getting attempt count for task %d, user %d: %v", task.ID, currentUser.ID, err)
			// Continue with default values on error
			attemptCount = 0
		}

		// Get best score
		bestScore, err := cs.submissionRepository.GetBestScoreForTaskByUser(db, task.ID, currentUser.ID)
		if err != nil {
			cs.logger.Warnf("Error getting best score for task %d, user %d: %v", task.ID, currentUser.ID, err)
			// Continue with default values on error
			bestScore = 0
		}

		// Get best submission ID
		bestSubmissionID, err := cs.submissionRepository.GetBestSubmissionIDForTaskByUser(db, task.ID, currentUser.ID)
		if err != nil {
			cs.logger.Warnf("Error getting best submission ID for task %d, user %d: %v", task.ID, currentUser.ID, err)
			// Continue with nil on error
			bestSubmissionID = nil
		}

		taskResults = append(taskResults, schemas.TaskResult{
			Task:             *TaskToInfoSchema(&task),
			AttemptCount:     attemptCount,
			BestResult:       bestScore,
			BestSubmissionID: bestSubmissionID,
		})
	}

	return &schemas.ContestResults{
		Contest: schemas.BaseContest{
			ID:            contest.ID,
			Name:          contest.Name,
			Description:   contest.Description,
			CreatedBy:     contest.CreatedBy,
			CreatedByName: fmt.Sprintf("%s %s", contest.Creator.Name, contest.Creator.Surname),
			StartAt:       contest.StartAt,
			EndAt:         contest.EndAt,
		},
		TaskResults: taskResults,
	}, nil
}

// hasContestPermission checks if the user has the required permission for the contest.
// Returns true if:
// 1. User is admin
// 2. User is the creator (which should also have manage permission via collaborator)
// 3. User has the required permission via collaborator
func (cs *contestService) hasContestPermission(db database.Database, contestID int64, user *schemas.User, requiredPermission types.Permission) error {
	_, err := cs.contestRepository.Get(db, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrNotFound
		}
		return err
	}
	return cs.accessControlService.CanUserAccess(db, models.ResourceTypeContest, contestID, user, requiredPermission)
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

func ManagedContestToSchema(model *models.ManagedContest) *schemas.ManagedContest {
	return &schemas.ManagedContest{
		CreatedContest: schemas.CreatedContest{
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
			CreatedAt:          model.CreatedAt,
		},
		PermissionType: model.PermissionType,
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
