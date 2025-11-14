package repository

import (
	"fmt"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/utils"
	"gorm.io/gorm"
)

type ContestRepository interface {
	// Create creates a new contest
	Create(tx *gorm.DB, contest *models.Contest) (int64, error)
	// Get retrieves a contest by ID
	Get(tx *gorm.DB, contestID int64) (*models.Contest, error)
	// GetAll retrieves all contests with pagination and sorting
	GetAll(tx *gorm.DB, offset int, limit int, sort string) ([]models.Contest, error)
	// GetAllWithStats retrieves all contests with participant counts and user registration status.
	// This method efficiently calculates participant counts (both direct participants and those via groups)
	// and determines the user's registration status for each contest in a single SQL query.
	// Returns ContestWithStats which includes ParticipantCount, IsParticipant, and HasPendingReg fields.
	GetAllWithStats(tx *gorm.DB, userID int64, offset int, limit int, sort string) ([]models.ContestWithStats, error)
	// GetOngoingContestsWithStats retrieves contests that are currently running with stats
	GetOngoingContestsWithStats(tx *gorm.DB, userID int64, offset int, limit int, sort string) ([]models.ContestWithStats, int64, error)
	// GetPastContestsWithStats retrieves contests that have ended with stats
	GetPastContestsWithStats(tx *gorm.DB, userID int64, offset int, limit int, sort string) ([]models.ContestWithStats, int64, error)
	// GetUpcomingContestsWithStats retrieves contests that haven't started yet with stats
	GetUpcomingContestsWithStats(tx *gorm.DB, userID int64, offset int, limit int, sort string) ([]models.ContestWithStats, int64, error)
	// GetAllForCreator retrieves all contests created by a specific user with pagination and sorting
	GetAllForCreatorWithStats(tx *gorm.DB, creatorID int64, offset int, limit int, sort string) ([]models.ContestWithStats, error)
	GetAllForCreator(tx *gorm.DB, creatorID int64, offset int, limit int, sort string) ([]models.Contest, int64, error)
	// Edit updates a contest
	Edit(tx *gorm.DB, contestID int64, contest *models.Contest) (*models.Contest, error)
	// EditWithStats updates a contest and returns it with participant and task counts
	EditWithStats(tx *gorm.DB, contestID int64, contest *models.Contest) (*models.ContestWithStats, error)
	// Delete removes a contest
	Delete(tx *gorm.DB, contestID int64) error
	// CreatePendingRegistration creates a pending registration request
	CreatePendingRegistration(tx *gorm.DB, registration *models.ContestRegistrationRequests) (int64, error)
	// IsPendingRegistrationExists checks if pending registration already exists
	IsPendingRegistrationExists(tx *gorm.DB, contestID int64, userID int64) (bool, error)
	// IsUserParticipant checks if user is already a participant
	IsUserParticipant(tx *gorm.DB, contestID int64, userID int64) (bool, error)
	// GetTasksForContest retrieves all tasks assigned to a contest
	GetTasksForContest(tx *gorm.DB, contestID int64) ([]models.Task, error)
	// GetContestTasksWithSettings retrieves contest-task relations with timing flags and associated task
	GetContestTasksWithSettings(tx *gorm.DB, contestID int64) ([]models.ContestTask, error)
	// GetTasksForContestWithStats retrieves all tasks assigned to a contest with submission statistics for a user
	GetTasksForContestWithStats(tx *gorm.DB, contestID, userID int64) ([]models.Task, error)
	// GetAssignableTasks retrieves all tasks NOT assigned to a contest
	GetAssignableTasks(tx *gorm.DB, contestID int64) ([]models.Task, error)
	// GetContestsForUserWithStats retrieves contests with stats a user is participating in
	GetContestsForUserWithStats(tx *gorm.DB, userID int64) ([]models.ParticipantContestStats, error)
	// AddTasksToContest assigns tasks to a contest
	AddTaskToContest(tx *gorm.DB, taskContest models.ContestTask) error
	// GetRegistrationRequests retrieves 'status' registration requests for a contest
	GetRegistrationRequests(tx *gorm.DB, contestID int64, status types.RegistrationRequestStatus) ([]models.ContestRegistrationRequests, error)
	// DeleteRegistrationRequest deletes a pending registration request
	DeleteRegistrationRequest(tx *gorm.DB, requestID int64) error
	// CreateContestParticipant adds a user as a participant to a contest
	CreateContestParticipant(tx *gorm.DB, contestID, userID int64) error
	// RejectRegistrationRequest rejects a pending registration request
	UpdateRegistrationRequestStatus(tx *gorm.DB, requestID int64, status types.RegistrationRequestStatus) error
	// GetPendingRegistrationRequest retrieves a pending registration request for a user in a contest
	GetPendingRegistrationRequest(tx *gorm.DB, contestID, userID int64) (*models.ContestRegistrationRequests, error)
	// GetContestTask retrieves the ContestTask relationship for validation
	GetContestTask(tx *gorm.DB, contestID, taskID int64) (*models.ContestTask, error)
	GetTaskContests(tx *gorm.DB, taskID int64) ([]int64, error)
}

type contestRepository struct{}

func (cr *contestRepository) Create(tx *gorm.DB, contest *models.Contest) (int64, error) {
	err := tx.Create(contest).Error
	if err != nil {
		return 0, err
	}
	return contest.ID, nil
}

func (cr *contestRepository) Get(tx *gorm.DB, contestID int64) (*models.Contest, error) {
	var contest models.Contest
	err := tx.Where("id = ?", contestID).First(&contest).Error
	if err != nil {
		return nil, err
	}
	return &contest, nil
}

func (cr *contestRepository) GetAll(tx *gorm.DB, offset int, limit int, sort string) ([]models.Contest, error) {
	var contests []models.Contest
	tx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, err
	}
	err = tx.Model(&models.Contest{}).Find(&contests).Error
	if err != nil {
		return nil, err
	}
	return contests, nil
}

func (cr *contestRepository) GetAllForCreator(tx *gorm.DB, creatorID int64, offset int, limit int, sort string) ([]models.Contest, int64, error) {
	var contests []models.Contest
	var totalCount int64

	// Get total count first
	baseQuery := tx.Model(&models.Contest{}).Where("created_by = ?", creatorID)
	err := baseQuery.Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting to a new query
	paginatedQuery, err := utils.ApplyPaginationAndSort(tx.Model(&models.Contest{}), limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}
	err = paginatedQuery.Where("created_by = ?", creatorID).Find(&contests).Error
	if err != nil {
		return nil, 0, err
	}
	return contests, totalCount, nil
}

func (cr *contestRepository) GetAllForCreatorWithStats(tx *gorm.DB, creatorID int64, offset int, limit int, sort string) ([]models.ContestWithStats, error) {
	var contests []models.ContestWithStats

	// Build a query that calculates participant counts and task counts for contests created by a specific user.
	// Similar to GetAllWithStats but filtered by creator.
	query := tx.Model(&models.Contest{}).
		Select(`contests.*,
			COALESCE(direct_participants.count, 0) + COALESCE(group_participants.count, 0) as participant_count,
			COALESCE(contest_task_count.count, 0) as task_count,
			false as is_participant,
			false as has_pending_reg`).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) as count
			FROM %s
			GROUP BY contest_id
		) as direct_participants ON contests.id = direct_participants.contest_id`, database.ResolveTableName(tx, &models.ContestParticipant{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT cpg.contest_id, COUNT(DISTINCT ug.user_id) as count
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			GROUP BY cpg.contest_id
		) as group_participants ON contests.id = group_participants.contest_id`, database.ResolveTableName(tx, &models.ContestParticipantGroup{}), database.ResolveTableName(tx, &models.UserGroup{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) as count
			FROM %s
			GROUP BY contest_id
		) as contest_task_count ON contests.id = contest_task_count.contest_id`, database.ResolveTableName(tx, &models.ContestTask{}))).
		Where("contests.created_by = ?", creatorID)

	// Apply pagination and sorting
	query, err := utils.ApplyPaginationAndSort(query, limit, offset, sort)
	if err != nil {
		return nil, err
	}

	err = query.Find(&contests).Error
	if err != nil {
		return nil, err
	}

	return contests, nil
}

func (cr *contestRepository) Edit(tx *gorm.DB, contestID int64, contest *models.Contest) (*models.Contest, error) {
	err := tx.Model(&models.Contest{}).Where("id = ?", contestID).Updates(contest).Error
	if err != nil {
		return nil, err
	}
	return cr.Get(tx, contestID)
}

func (cr *contestRepository) EditWithStats(tx *gorm.DB, contestID int64, contest *models.Contest) (*models.ContestWithStats, error) {
	// First update the contest
	err := tx.Model(&models.Contest{}).Where("id = ?", contestID).Updates(contest).Error
	if err != nil {
		return nil, err
	}

	// Then retrieve with stats
	var contestWithStats models.ContestWithStats
	query := tx.Model(&models.Contest{}).
		Select(`contests.*,
			COALESCE(direct_participants.count, 0) + COALESCE(group_participants.count, 0) as participant_count,
			COALESCE(contest_task_count.count, 0) as task_count,
			false as is_participant,
			false as has_pending_reg`).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) as count
			FROM %s
			GROUP BY contest_id
		) as direct_participants ON contests.id = direct_participants.contest_id`, database.ResolveTableName(tx, &models.ContestParticipant{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT cpg.contest_id, COUNT(DISTINCT ug.user_id) as count
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			GROUP BY cpg.contest_id
		) as group_participants ON contests.id = group_participants.contest_id`, database.ResolveTableName(tx, &models.ContestParticipantGroup{}), database.ResolveTableName(tx, &models.UserGroup{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) as count
			FROM %s
			GROUP BY contest_id
		) as contest_task_count ON contests.id = contest_task_count.contest_id`, database.ResolveTableName(tx, &models.ContestTask{}))).
		Where("contests.id = ?", contestID)

	err = query.First(&contestWithStats).Error
	if err != nil {
		return nil, err
	}

	return &contestWithStats, nil
}

func (cr *contestRepository) Delete(tx *gorm.DB, contestID int64) error {
	err := tx.Where("id = ?", contestID).Delete(&models.Contest{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (cr *contestRepository) CreatePendingRegistration(tx *gorm.DB, registration *models.ContestRegistrationRequests) (int64, error) {
	err := tx.Create(registration).Error
	if err != nil {
		return 0, err
	}
	return registration.ID, nil
}

func (cr *contestRepository) IsPendingRegistrationExists(tx *gorm.DB, contestID int64, userID int64) (bool, error) {
	var count int64
	err := tx.Model(&models.ContestRegistrationRequests{}).
		Where("contest_id = ? AND user_id = ? AND status = ?", contestID, userID, types.RegistrationRequestStatusPending).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (cr *contestRepository) IsUserParticipant(tx *gorm.DB, contestID int64, userID int64) (bool, error) {
	var userCount int64
	err := tx.Model(&models.ContestParticipant{}).
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		Count(&userCount).Error
	if err != nil {
		return false, err
	}
	var groupCount int64
	err = tx.Model(&models.ContestParticipantGroup{}).Where("contest_id = ?", contestID).
		Joins(fmt.Sprintf("JOIN %s ON contest_participant_groups.group_id = user_groups.group_id", database.ResolveTableName(tx, &models.UserGroup{}))).
		Where("user_groups.user_id = ?", userID).
		Count(&groupCount).Error
	if err != nil {
		return false, err
	}
	return userCount > 0 || groupCount > 0, nil
}

func (cr *contestRepository) GetAllWithStats(tx *gorm.DB, userID int64, offset int, limit int, sort string) ([]models.ContestWithStats, error) {
	var contests []models.ContestWithStats

	// Build an efficient query that calculates participant counts and user registration status
	// in a single SQL query to avoid N+1 problems when fetching contests with their statistics.
	//
	// The query:
	// 1. Counts direct participants via contest_participants table
	// 2. Counts group participants via contest_participant_groups + user_groups tables
	// 3. Determines if the current user is a participant (directly or via group)
	// 4. Checks if the current user has a pending registration
	query := tx.Model(&models.Contest{}).
		Select(`contests.*,
			COALESCE(direct_participants.count, 0) + COALESCE(group_participants.count, 0) as participant_count,
			COALESCE(contest_task_count.count, 0) as task_count,
			CASE WHEN user_participants.user_id IS NOT NULL OR user_group_participants.user_id IS NOT NULL THEN true ELSE false END as is_participant,
			CASE WHEN pending_regs.user_id IS NOT NULL THEN true ELSE false END as has_pending_reg`).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) as count
			FROM %s
			GROUP BY contest_id
		) as direct_participants ON contests.id = direct_participants.contest_id`, database.ResolveTableName(tx, &models.ContestParticipant{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT cpg.contest_id, COUNT(DISTINCT ug.user_id) as count
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			GROUP BY cpg.contest_id
		) as group_participants ON contests.id = group_participants.contest_id`, database.ResolveTableName(tx, &models.ContestParticipantGroup{}), database.ResolveTableName(tx, &models.UserGroup{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) as count
			FROM %s
			GROUP BY contest_id
		) as contest_task_count ON contests.id = contest_task_count.contest_id`, database.ResolveTableName(tx, &models.ContestTask{}))).
		Joins(fmt.Sprintf(`LEFT JOIN %s user_participants ON contests.id = user_participants.contest_id AND user_participants.user_id = ?`, database.ResolveTableName(tx, &models.ContestParticipant{})), userID).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT DISTINCT cpg.contest_id, ug.user_id
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			WHERE ug.user_id = ?
		) as user_group_participants ON contests.id = user_group_participants.contest_id`, database.ResolveTableName(tx, &models.ContestParticipantGroup{}), database.ResolveTableName(tx, &models.UserGroup{})), userID).
		Joins(fmt.Sprintf(`LEFT JOIN %s pending_regs ON contests.id = pending_regs.contest_id AND pending_regs.user_id = ?`, database.ResolveTableName(tx, &models.ContestRegistrationRequests{})), userID)

	// Apply pagination and sorting
	query, err := utils.ApplyPaginationAndSort(query, limit, offset, sort)
	if err != nil {
		return nil, err
	}

	err = query.Find(&contests).Error
	if err != nil {
		return nil, err
	}

	return contests, nil
}

func (cr *contestRepository) GetOngoingContestsWithStats(tx *gorm.DB, userID int64, offset int, limit int, sort string) ([]models.ContestWithStats, int64, error) {
	var contests []models.ContestWithStats
	var totalCount int64

	// Get total count first
	countQuery := tx.Model(&models.Contest{}).
		Where(`(
			(start_at IS NULL OR start_at <= NOW()) AND (end_at IS NULL OR end_at > NOW())
		)`)
	err := countQuery.Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Build query for ongoing contests (started but not ended, or no end date but started)
	query := tx.Model(&models.Contest{}).
		Select(`contests.*,
			COALESCE(direct_participants.count, 0) + COALESCE(group_participants.count, 0) as participant_count,
			COALESCE(contest_task_count.count, 0) as task_count,
			CASE WHEN user_participants.user_id IS NOT NULL OR user_group_participants.user_id IS NOT NULL THEN true ELSE false END as is_participant,
			CASE WHEN pending_regs.user_id IS NOT NULL THEN true ELSE false END as has_pending_reg`).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) as count
			FROM %s
			GROUP BY contest_id
		) as direct_participants ON contests.id = direct_participants.contest_id`, database.ResolveTableName(tx, &models.ContestParticipant{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT cpg.contest_id, COUNT(DISTINCT ug.user_id) as count
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			GROUP BY cpg.contest_id
		) as group_participants ON contests.id = group_participants.contest_id`, database.ResolveTableName(tx, &models.ContestParticipantGroup{}), database.ResolveTableName(tx, &models.UserGroup{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) as count
			FROM %s
			GROUP BY contest_id
		) as contest_task_count ON contests.id = contest_task_count.contest_id`, database.ResolveTableName(tx, &models.ContestTask{}))).
		Joins(fmt.Sprintf(`LEFT JOIN %s user_participants ON contests.id = user_participants.contest_id AND user_participants.user_id = ?`, database.ResolveTableName(tx, &models.ContestParticipant{})), userID).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT DISTINCT cpg.contest_id, ug.user_id
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			WHERE ug.user_id = ?
		) as user_group_participants ON contests.id = user_group_participants.contest_id`, database.ResolveTableName(tx, &models.ContestParticipantGroup{}), database.ResolveTableName(tx, &models.UserGroup{})), userID).
		Joins(fmt.Sprintf(`LEFT JOIN %s pending_regs ON contests.id = pending_regs.contest_id AND pending_regs.user_id = ?`, database.ResolveTableName(tx, &models.ContestRegistrationRequests{})), userID).
		Where(`(
			(start_at IS NULL OR start_at <= NOW()) AND (end_at IS NULL OR end_at > NOW())
		)`)

	// Apply pagination and sorting
	query, err = utils.ApplyPaginationAndSort(query, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = query.Find(&contests).Error
	if err != nil {
		return nil, 0, err
	}

	return contests, totalCount, nil
}

func (cr *contestRepository) GetPastContestsWithStats(tx *gorm.DB, userID int64, offset int, limit int, sort string) ([]models.ContestWithStats, int64, error) {
	var contests []models.ContestWithStats
	var totalCount int64

	// Get total count first
	countQuery := tx.Model(&models.Contest{}).
		Where("end_at IS NOT NULL AND end_at <= NOW()")
	err := countQuery.Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Build query for past contests (ended)
	query := tx.Model(&models.Contest{}).
		Select(`contests.*,
			COALESCE(direct_participants.count, 0) + COALESCE(group_participants.count, 0) as participant_count,
			COALESCE(contest_task_count.count, 0) as task_count,
			CASE WHEN user_participants.user_id IS NOT NULL OR user_group_participants.user_id IS NOT NULL THEN true ELSE false END as is_participant,
			CASE WHEN pending_regs.user_id IS NOT NULL THEN true ELSE false END as has_pending_reg`).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) as count
			FROM %s
			GROUP BY contest_id
		) as direct_participants ON contests.id = direct_participants.contest_id`, database.ResolveTableName(tx, &models.ContestParticipant{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT cpg.contest_id, COUNT(DISTINCT ug.user_id) as count
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			GROUP BY cpg.contest_id
		) as group_participants ON contests.id = group_participants.contest_id`, database.ResolveTableName(tx, &models.ContestParticipantGroup{}), database.ResolveTableName(tx, &models.UserGroup{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) as count
			FROM %s
			GROUP BY contest_id
		) as contest_task_count ON contests.id = contest_task_count.contest_id`, database.ResolveTableName(tx, &models.ContestTask{}))).
		Joins(fmt.Sprintf(`LEFT JOIN %s user_participants ON contests.id = user_participants.contest_id AND user_participants.user_id = ?`, database.ResolveTableName(tx, &models.ContestParticipant{})), userID).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT DISTINCT cpg.contest_id, ug.user_id
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			WHERE ug.user_id = ?
		) as user_group_participants ON contests.id = user_group_participants.contest_id`, database.ResolveTableName(tx, &models.ContestParticipantGroup{}), database.ResolveTableName(tx, &models.UserGroup{})), userID).
		Joins(fmt.Sprintf(`LEFT JOIN %s pending_regs ON contests.id = pending_regs.contest_id AND pending_regs.user_id = ?`, database.ResolveTableName(tx, &models.ContestRegistrationRequests{})), userID).
		Where("end_at IS NOT NULL AND end_at <= NOW()")

	// Apply pagination and sorting
	query, err = utils.ApplyPaginationAndSort(query, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = query.Find(&contests).Error
	if err != nil {
		return nil, 0, err
	}

	return contests, totalCount, nil
}

func (cr *contestRepository) GetUpcomingContestsWithStats(tx *gorm.DB, userID int64, offset int, limit int, sort string) ([]models.ContestWithStats, int64, error) {
	var contests []models.ContestWithStats
	var totalCount int64

	// Get total count first
	countQuery := tx.Model(&models.Contest{}).
		Where("start_at IS NOT NULL AND start_at > NOW()")
	err := countQuery.Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Build query for upcoming contests (not started yet)
	query := tx.Model(&models.Contest{}).
		Select(`contests.*,
			COALESCE(direct_participants.count, 0) + COALESCE(group_participants.count, 0) as participant_count,
			COALESCE(contest_task_count.count, 0) as task_count,
			CASE WHEN user_participants.user_id IS NOT NULL OR user_group_participants.user_id IS NOT NULL THEN true ELSE false END as is_participant,
			CASE WHEN pending_regs.user_id IS NOT NULL THEN true ELSE false END as has_pending_reg`).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) as count
			FROM %s
			GROUP BY contest_id
		) as direct_participants ON contests.id = direct_participants.contest_id`, database.ResolveTableName(tx, &models.ContestParticipant{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT cpg.contest_id, COUNT(DISTINCT ug.user_id) as count
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			GROUP BY cpg.contest_id
		) as group_participants ON contests.id = group_participants.contest_id`, database.ResolveTableName(tx, &models.ContestParticipantGroup{}), database.ResolveTableName(tx, &models.UserGroup{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) as count
			FROM %s
			GROUP BY contest_id
		) as contest_task_count ON contests.id = contest_task_count.contest_id`, database.ResolveTableName(tx, &models.ContestTask{}))).
		Joins(fmt.Sprintf(`LEFT JOIN %s user_participants ON contests.id = user_participants.contest_id AND user_participants.user_id = ?`, database.ResolveTableName(tx, &models.ContestParticipant{})), userID).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT DISTINCT cpg.contest_id, ug.user_id
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			WHERE ug.user_id = ?
		) as user_group_participants ON contests.id = user_group_participants.contest_id`, database.ResolveTableName(tx, &models.ContestParticipantGroup{}), database.ResolveTableName(tx, &models.UserGroup{})), userID).
		Joins(fmt.Sprintf(`LEFT JOIN %s pending_regs ON contests.id = pending_regs.contest_id AND pending_regs.user_id = ?`, database.ResolveTableName(tx, &models.ContestRegistrationRequests{})), userID).
		Where("start_at IS NOT NULL AND start_at > NOW()")

	// Apply pagination and sorting
	query, err = utils.ApplyPaginationAndSort(query, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = query.Find(&contests).Error
	if err != nil {
		return nil, 0, err
	}

	return contests, totalCount, nil
}

// GetContestTasksWithSettings retrieves contest-task relations (with timing flags) and preloads the associated Task
func (cr *contestRepository) GetContestTasksWithSettings(tx *gorm.DB, contestID int64) ([]models.ContestTask, error) {
	var relations []models.ContestTask
	err := tx.Model(&models.ContestTask{}).
		Where("contest_id = ?", contestID).
		Preload("Task").
		Preload("Task.Author").
		Find(&relations).Error
	if err != nil {
		return nil, err
	}
	return relations, nil
}

func (cr *contestRepository) GetTasksForContest(tx *gorm.DB, contestID int64) ([]models.Task, error) {
	var tasks []models.Task
	err := tx.Model(&models.Task{}).
		Joins(fmt.Sprintf("JOIN %s ON contest_tasks.task_id = tasks.id", database.ResolveTableName(tx, &models.ContestTask{}))).
		Where("contest_tasks.contest_id = ?", contestID).
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (cr *contestRepository) GetTasksForContestWithStats(tx *gorm.DB, contestID, userID int64) ([]models.Task, error) {
	var tasks []models.Task
	err := tx.Model(&models.Task{}).
		Joins(fmt.Sprintf("JOIN %s ON contest_tasks.task_id = tasks.id", database.ResolveTableName(tx, &models.ContestTask{}))).
		Where("contest_tasks.contest_id = ?", contestID).
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (cr *contestRepository) GetAssignableTasks(tx *gorm.DB, contestID int64) ([]models.Task, error) {
	var tasks []models.Task
	err := tx.Model(&models.Task{}).
		Where("id NOT IN (?)",
			tx.Table(database.ResolveTableName(tx, &models.ContestTask{})).
				Select("task_id").
				Where("contest_id = ?", contestID),
		).
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (cr *contestRepository) GetContestsForUserWithStats(tx *gorm.DB, userID int64) ([]models.ParticipantContestStats, error) {
	var contests []models.ParticipantContestStats

	// Build query to get contests where user is a participant with statistics.
	// solved_task_count: number of tasks for which the user has at least one perfect (all tests passed) submission.
	// test_count: total number of test cases across all tasks in the contest.
	// solved_test_count: sum of passed tests from the BEST evaluated submission per task (max passed tests; tie -> latest submitted_at).
	query := tx.Model(&models.Contest{}).
		Select(`contests.*,
			COALESCE(direct_participants.count, 0) + COALESCE(group_participants.count, 0) AS participant_count,
			COALESCE(contest_task_count.count, 0) AS task_count,
			COALESCE(user_solved_tasks.count, 0) AS solved_task_count,
			COALESCE(contest_test_count.test_count, 0) AS test_count,
			COALESCE(user_best_solved_tests.solved_test_count, 0) AS solved_test_count
		`).
		// Direct participants
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) AS count
			FROM %s
			GROUP BY contest_id
		) AS direct_participants ON contests.id = direct_participants.contest_id`,
			database.ResolveTableName(tx, &models.ContestParticipant{}))).
		// Group participants expanded to distinct users
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT cpg.contest_id, COUNT(DISTINCT ug.user_id) AS count
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			GROUP BY cpg.contest_id
		) AS group_participants ON contests.id = group_participants.contest_id`,
			database.ResolveTableName(tx, &models.ContestParticipantGroup{}),
			database.ResolveTableName(tx, &models.UserGroup{}))).
		// Task count per contest
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) AS count
			FROM %s
			GROUP BY contest_id
		) AS contest_task_count ON contests.id = contest_task_count.contest_id`,
			database.ResolveTableName(tx, &models.ContestTask{}))).
		// Total test cases per contest
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT ct.contest_id, COUNT(tc.id) AS test_count
			FROM %s ct
			JOIN %s tc ON tc.task_id = ct.task_id
			GROUP BY ct.contest_id
		) AS contest_test_count ON contests.id = contest_test_count.contest_id`,
			database.ResolveTableName(tx, &models.ContestTask{}),
			database.ResolveTableName(tx, &models.TestCase{}))).
		// Perfectly solved tasks (all tests passed in at least one submission result)
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT ct.contest_id, COUNT(DISTINCT ct.task_id) AS count
			FROM %s ct
			INNER JOIN %s s ON s.task_id = ct.task_id AND s.user_id = ? AND s.status = '%s'
			INNER JOIN %s sr ON sr.submission_id = s.id
			INNER JOIN (
				SELECT submission_result_id,
					   COUNT(*) AS total_tests,
					   COUNT(CASE WHEN passed = true THEN 1 END) AS passed_tests
				FROM %s
				GROUP BY submission_result_id
				HAVING COUNT(*) = COUNT(CASE WHEN passed = true THEN 1 END)
			) perfect_results ON perfect_results.submission_result_id = sr.id
			GROUP BY ct.contest_id
		) AS user_solved_tasks ON contests.id = user_solved_tasks.contest_id`,
			database.ResolveTableName(tx, &models.ContestTask{}),
			database.ResolveTableName(tx, &models.Submission{}),
			types.SubmissionStatusEvaluated,
			database.ResolveTableName(tx, &models.SubmissionResult{}),
			database.ResolveTableName(tx, &models.TestResult{})), userID).
		// Best evaluated submission per task (highest passed tests; tie -> latest submitted_at)
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT ct.contest_id,
				   SUM(best.passed_tests) AS solved_test_count
			FROM %s ct
			JOIN (
				SELECT s.task_id,
					   sr.id AS submission_result_id,
					   COUNT(tr.id) AS total_tests,
					   COUNT(CASE WHEN tr.passed = true THEN 1 END) AS passed_tests,
					   s.submitted_at,
					   ROW_NUMBER() OVER (
						   PARTITION BY s.task_id
						   ORDER BY COUNT(CASE WHEN tr.passed = true THEN 1 END) DESC,
									s.submitted_at DESC
					   ) AS rn
				FROM %s s
				JOIN %s sr ON sr.submission_id = s.id
				LEFT JOIN %s tr ON tr.submission_result_id = sr.id
				WHERE s.user_id = ? AND s.status = '%s'
				GROUP BY s.task_id, sr.id, s.submitted_at
			) best ON best.task_id = ct.task_id AND best.rn = 1
			GROUP BY ct.contest_id
		) AS user_best_solved_tests ON contests.id = user_best_solved_tests.contest_id`,
			database.ResolveTableName(tx, &models.ContestTask{}),
			database.ResolveTableName(tx, &models.Submission{}),
			database.ResolveTableName(tx, &models.SubmissionResult{}),
			database.ResolveTableName(tx, &models.TestResult{}),
			types.SubmissionStatusEvaluated), userID).
		// Participation filter (direct or via group)
		Where(fmt.Sprintf(`EXISTS (
			SELECT 1 FROM %s cp
			WHERE cp.contest_id = contests.id AND cp.user_id = ?
		) OR EXISTS (
			SELECT 1 FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			WHERE cpg.contest_id = contests.id AND ug.user_id = ?
		)`,
			database.ResolveTableName(tx, &models.ContestParticipant{}),
			database.ResolveTableName(tx, &models.ContestParticipantGroup{}),
			database.ResolveTableName(tx, &models.UserGroup{})),
			userID, userID)

	err := query.Find(&contests).Error
	if err != nil {
		return nil, err
	}

	return contests, nil
}

func (cr *contestRepository) AddTaskToContest(tx *gorm.DB, taskContest models.ContestTask) error {
	err := tx.Create(&taskContest).Error
	if err != nil {
		return err
	}
	return nil
}

func (cr *contestRepository) GetRegistrationRequests(tx *gorm.DB, contestID int64, status types.RegistrationRequestStatus) ([]models.ContestRegistrationRequests, error) {
	var requests []models.ContestRegistrationRequests
	err := tx.Preload("User").Where("contest_id = ? AND status = ?", contestID, status).Find(&requests).Error
	if err != nil {
		return nil, err
	}
	return requests, nil
}

func (cr *contestRepository) UpdateRegistrationRequestStatus(tx *gorm.DB, requestID int64, status types.RegistrationRequestStatus) error {
	return tx.Model(models.ContestRegistrationRequests{}).Where("id = ?", requestID).Updates(models.ContestRegistrationRequests{Status: status}).Error
}

func (cr *contestRepository) DeleteRegistrationRequest(tx *gorm.DB, requestID int64) error {
	return tx.Model(models.ContestRegistrationRequests{}).Delete(&models.ContestRegistrationRequests{ID: requestID}).Error
}

func (cr *contestRepository) CreateContestParticipant(tx *gorm.DB, contestID, userID int64) error {
	participant := &models.ContestParticipant{
		ContestID: contestID,
		UserID:    userID,
	}
	err := tx.Create(participant).Error
	if err != nil {
		return err
	}
	return nil
}

func (cr *contestRepository) GetPendingRegistrationRequest(tx *gorm.DB, contestID, userID int64) (*models.ContestRegistrationRequests, error) {
	var request []*models.ContestRegistrationRequests
	err := tx.Where("contest_id = ? AND user_id = ? AND status = ?", contestID, userID, types.RegistrationRequestStatusPending).Limit(1).Find(&request).Error
	if err != nil {
		return nil, err
	}
	if len(request) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return request[0], nil
}

func (cr *contestRepository) GetContestTask(tx *gorm.DB, contestID, taskID int64) (*models.ContestTask, error) {
	var contestTask models.ContestTask
	err := tx.Where("contest_id = ? AND task_id = ?", contestID, taskID).First(&contestTask).Error
	if err != nil {
		return nil, err
	}
	return &contestTask, nil
}

func (cr *contestRepository) GetTaskContests(tx *gorm.DB, taskID int64) ([]int64, error) {
	var contestIDs []int64
	err := tx.Model(&models.ContestTask{}).Where("task_id = ?", taskID).Pluck("contest_id", &contestIDs).Error
	if err != nil {
		return nil, err
	}
	return contestIDs, nil
}

func NewContestRepository() ContestRepository {
	return &contestRepository{}
}
