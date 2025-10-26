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
	GetOngoingContests(tx *gorm.DB, currentUser schemas.User, queryParams map[string]any) ([]schemas.AvailableContest, error)
	// GetPastContests retrieves contests that have ended
	GetPastContests(tx *gorm.DB, currentUser schemas.User, queryParams map[string]any) ([]schemas.AvailableContest, error)
	// GetUpcomingContests retrieves contests that haven't started yet
	GetUpcomingContests(tx *gorm.DB, currentUser schemas.User, queryParams map[string]any) ([]schemas.AvailableContest, error)
	// Edit updates a contest
	Edit(tx *gorm.DB, currentUser schemas.User, contestID int64, editInfo *schemas.EditContest) (*schemas.Contest, error)
	// Delete removes a contest
	Delete(tx *gorm.DB, currentUser schemas.User, contestID int64) error
	// RegisterForContest creates a pending registration for a contest
	RegisterForContest(tx *gorm.DB, currentUser schemas.User, contestID int64) error
	// GetTasksForContest retrieves all tasks associated with a contest with submission stats and optional name filter (for authorized users)
	GetTasksForContest(tx *gorm.DB, currentUser schemas.User, contestID int64, nameFilter string) ([]schemas.TaskWithContestStats, error)
	// GetUserContests retrieves all contests a user is participating in
	GetUserContests(tx *gorm.DB, userID int64) (schemas.UserContestsWithStats, error)
	// AddTaskToContest adds a task to a contest
	AddTaskToContest(tx *gorm.DB, currentUser *schemas.User, contestID int64, request *schemas.AddTaskToContest) error
}

const defaultContestSort = "created_at:desc"

type contestService struct {
	contestRepository    repository.ContestRepository
	taskRepository       repository.TaskRepository
	userRepository       repository.UserRepository
	submissionRepository repository.SubmissionRepository

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
	}

	if contest.StartAt != nil {
		model.StartAt = contest.StartAt
	}
	if contest.EndAt != nil {
		model.EndAt = contest.EndAt
	}

	contestID, err := cs.contestRepository.Create(tx, model)
	if err != nil {
		return -1, err
	}

	return contestID, nil
}

func (cs *contestService) Get(tx *gorm.DB, currentUser schemas.User, contestID int64) (*schemas.Contest, error) {
	contest, err := cs.contestRepository.Get(tx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrNotFound
		}
		return nil, err
	}

	// Check visibility and permissions
	if !*contest.IsVisible {
		if currentUser.Role == types.UserRoleStudent {
			return nil, myerrors.ErrNotAuthorized
		}
		if currentUser.Role == types.UserRoleTeacher && contest.CreatedBy != currentUser.ID {
			return nil, myerrors.ErrNotAuthorized
		}
	}

	return ContestToSchema(contest), nil
}

func (cs *contestService) GetOngoingContests(tx *gorm.DB, currentUser schemas.User, queryParams map[string]any) ([]schemas.AvailableContest, error) {
	limit := queryParams["limit"].(int)
	offset := queryParams["offset"].(int)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = defaultContestSort
	}

	contestsWithStats, err := cs.contestRepository.GetOngoingContestsWithStats(tx, currentUser.ID, offset, limit, sort)
	if err != nil {
		return nil, err
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

	return result, nil
}

func (cs *contestService) GetPastContests(tx *gorm.DB, currentUser schemas.User, queryParams map[string]any) ([]schemas.AvailableContest, error) {
	limit := queryParams["limit"].(int)
	offset := queryParams["offset"].(int)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = defaultContestSort
	}

	contestsWithStats, err := cs.contestRepository.GetPastContestsWithStats(tx, currentUser.ID, offset, limit, sort)
	if err != nil {
		return nil, err
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

	return result, nil
}

func (cs *contestService) GetUpcomingContests(tx *gorm.DB, currentUser schemas.User, queryParams map[string]any) ([]schemas.AvailableContest, error) {
	limit := queryParams["limit"].(int)
	offset := queryParams["offset"].(int)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = defaultContestSort
	}

	contestsWithStats, err := cs.contestRepository.GetUpcomingContestsWithStats(tx, currentUser.ID, offset, limit, sort)
	if err != nil {
		return nil, err
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

	return result, nil
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

func (cs *contestService) Edit(tx *gorm.DB, currentUser schemas.User, contestID int64, editInfo *schemas.EditContest) (*schemas.Contest, error) {
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

	// Check permissions
	if currentUser.Role == types.UserRoleTeacher && model.CreatedBy != currentUser.ID {
		return nil, myerrors.ErrNotAuthorized
	}

	cs.updateModel(model, editInfo)

	newModel, err := cs.contestRepository.Edit(tx, contestID, model)
	if err != nil {
		return nil, err
	}

	return ContestToSchema(newModel), nil
}

func (cs *contestService) Delete(tx *gorm.DB, currentUser schemas.User, contestID int64) error {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return err
	}

	contest, err := cs.contestRepository.Get(tx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrNotFound
		}
		return err
	}

	if currentUser.Role == types.UserRoleTeacher && contest.CreatedBy != currentUser.ID {
		return myerrors.ErrNotAuthorized
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
			return myerrors.ErrNotAuthorized
		}
		if currentUser.Role == types.UserRoleTeacher && contest.CreatedBy != currentUser.ID {
			return myerrors.ErrNotAuthorized
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
	registration := &models.ContestPendingRegistration{
		ContestID: contestID,
		UserID:    currentUser.ID,
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
	// TODO: handle when setting to nil is intended
	if editInfo.StartAt != nil {
		model.StartAt = editInfo.StartAt
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

func (cs *contestService) GetTasksForContest(tx *gorm.DB, currentUser schemas.User, contestID int64, nameFilter string) ([]schemas.TaskWithContestStats, error) {
	contest, err := cs.contestRepository.Get(tx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrNotFound
		}
		return nil, err
	}

	if !cs.isContestVisibleToUser(tx, contest, &currentUser) {
		return nil, myerrors.ErrNotAuthorized
	}

	tasks, err := cs.contestRepository.GetTasksForContest(tx, contestID, nameFilter)
	if err != nil {
		return nil, err
	}

	result := make([]schemas.TaskWithContestStats, len(tasks))
	for i, task := range tasks {
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
			ID:           task.ID,
			Title:        task.Title,
			CreatedBy:    task.CreatedBy,
			CreatedAt:    task.CreatedAt,
			UpdatedAt:    task.UpdatedAt,
			BestScore:    bestScore,
			AttemptCount: attemptCount,
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
	past := make([]schemas.ContestWithStats, 0)
	upcoming := make([]schemas.ContestWithStats, 0)
	now := time.Now()

	for _, contest := range contestsWithStats {
		contestSchema := *ParticipantContestStatsToSchema(&contest)

		// Determine contest status
		if contest.StartAt != nil && !contest.StartAt.After(now) {
			// Contest has started
			if contest.EndAt == nil || contest.EndAt.After(now) {
				// Contest is ongoing
				ongoing = append(ongoing, contestSchema)
			} else {
				// Contest has ended
				past = append(past, contestSchema)
			}
		} else if contest.EndAt != nil && !contest.EndAt.After(now) {
			// Contest has ended (edge case)
			past = append(past, contestSchema)
		} else {
			// Contest is upcoming
			upcoming = append(upcoming, contestSchema)
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
	if currentUser.Role == types.UserRoleTeacher && contest.CreatedBy != currentUser.ID {
		return myerrors.ErrNotAuthorized
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

func ContestToSchema(model *models.Contest) *schemas.Contest {
	return &schemas.Contest{
		ID:               model.ID,
		Name:             model.Name,
		Description:      model.Description,
		CreatedBy:        model.CreatedBy,
		StartAt:          model.StartAt,
		EndAt:            model.EndAt,
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
		ParticipantCount: 0,
		TaskCount:        0,
		Status:           getContestStatus(model.StartAt, model.EndAt),
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
			ID:               model.ID,
			Name:             model.Name,
			Description:      model.Description,
			CreatedBy:        model.CreatedBy,
			StartAt:          model.StartAt,
			EndAt:            model.EndAt,
			CreatedAt:        model.CreatedAt,
			UpdatedAt:        model.UpdatedAt,
			ParticipantCount: model.ParticipantCount,
			TaskCount:        model.TaskCount,
			Status:           status,
		},
		RegistrationStatus: registrationStatus,
	}
}

func getContestStatus(startAt, endAt *time.Time) string {
	status := "upcoming"
	now := time.Now()

	if startAt != nil && !startAt.After(now) {
		// Started
		if endAt == nil || endAt.After(now) {
			status = "ongoing"
		} else {
			status = "past"
		}
	} else if endAt != nil && !endAt.After(now) {
		status = "past"
	}
	return status
}

func ParticipantContestStatsToSchema(model *models.ParticipantContestStats) *schemas.ContestWithStats {
	status := getContestStatus(model.StartAt, model.EndAt)
	return &schemas.ContestWithStats{
		Contest: schemas.Contest{
			ID:               model.ID,
			Name:             model.Name,
			Description:      model.Description,
			CreatedBy:        model.CreatedBy,
			StartAt:          model.StartAt,
			EndAt:            model.EndAt,
			CreatedAt:        model.CreatedAt,
			UpdatedAt:        model.UpdatedAt,
			ParticipantCount: model.ParticipantCount,
			TaskCount:        model.TaskCount,
			Status:           status,
		},
		SolvedTaskCount: model.SolvedCount,
	}
}

func NewContestService(contestRepository repository.ContestRepository, userRepository repository.UserRepository, submissionRepository repository.SubmissionRepository, taskRepository repository.TaskRepository) ContestService {
	return &contestService{
		contestRepository:    contestRepository,
		taskRepository:       taskRepository,
		userRepository:       userRepository,
		submissionRepository: submissionRepository,
		logger:               utils.NewNamedLogger("contest_service"),
	}
}
