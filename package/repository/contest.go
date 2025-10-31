package repository

import (
	"fmt"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/types"
	"gorm.io/gorm"
)

type ContestRepository interface {
	// Create creates a new contest
	Create(tx *database.DB, contest *models.Contest) (int64, error)
	// Get retrieves a contest by ID
	Get(tx *database.DB, contestID int64) (*models.Contest, error)
	// GetAll retrieves all contests with pagination and sorting
	GetAll(tx *database.DB, offset int, limit int, sort string) ([]models.Contest, error)
	// GetAllWithStats retrieves all contests with participant counts and user registration status.
	// This method efficiently calculates participant counts (both direct participants and those via groups)
	// and determines the user's registration status for each contest in a single SQL query.
	// Returns ContestWithStats which includes ParticipantCount, IsParticipant, and HasPendingReg fields.
	GetAllWithStats(tx *database.DB, userID int64, offset int, limit int, sort string) ([]models.ContestWithStats, error)
	// GetOngoingContestsWithStats retrieves contests that are currently running with stats
	GetOngoingContestsWithStats(tx *database.DB, userID int64, offset int, limit int, sort string) ([]models.ContestWithStats, error)
	// GetPastContestsWithStats retrieves contests that have ended with stats
	GetPastContestsWithStats(tx *database.DB, userID int64, offset int, limit int, sort string) ([]models.ContestWithStats, error)
	// GetUpcomingContestsWithStats retrieves contests that haven't started yet with stats
	GetUpcomingContestsWithStats(tx *database.DB, userID int64, offset int, limit int, sort string) ([]models.ContestWithStats, error)
	// GetAllForCreator retrieves all contests created by a specific user with pagination and sorting
	GetAllForCreator(tx *database.DB, creatorID int64, offset int, limit int, sort string) ([]models.Contest, error)
	// Edit updates a contest
	Edit(tx *database.DB, contestID int64, contest *models.Contest) (*models.Contest, error)
	// Delete removes a contest
	Delete(tx *database.DB, contestID int64) error
	// CreatePendingRegistration creates a pending registration request
	CreatePendingRegistration(tx *database.DB, registration *models.ContestRegistrationRequests) (int64, error)
	// IsPendingRegistrationExists checks if pending registration already exists
	IsPendingRegistrationExists(tx *database.DB, contestID int64, userID int64) (bool, error)
	// IsUserParticipant checks if user is already a participant
	IsUserParticipant(tx *database.DB, contestID int64, userID int64) (bool, error)
	// GetTasksForContest retrieves all tasks assigned to a contest
	GetTasksForContest(tx *database.DB, contestID int64) ([]models.Task, error)
	// GetTasksForContestWithStats retrieves all tasks assigned to a contest with submission statistics for a user
	GetTasksForContestWithStats(tx *database.DB, contestID, userID int64) ([]models.Task, error)
	// GetContestsForUserWithStats retrieves contests with stats a user is participating in
	GetContestsForUserWithStats(tx *database.DB, userID int64) ([]models.ParticipantContestStats, error)
	// AddTasksToContest assigns tasks to a contest
	AddTaskToContest(tx *database.DB, taskContest models.ContestTask) error
	// GetRegistrationRequests retrieves 'status' registration requests for a contest
	GetRegistrationRequests(tx *database.DB, contestID int64, status types.RegistrationRequestStatus) ([]models.ContestRegistrationRequests, error)
	// DeleteRegistrationRequest deletes a pending registration request
	DeleteRegistrationRequest(tx *database.DB, requestID int64) error
	// CreateContestParticipant adds a user as a participant to a contest
	CreateContestParticipant(tx *database.DB, contestID, userID int64) error
	// RejectRegistrationRequest rejects a pending registration request
	UpdateRegistrationRequestStatus(tx *database.DB, requestID int64, status types.RegistrationRequestStatus) error
	// GetPendingRegistrationRequest retrieves a pending registration request for a user in a contest
	GetPendingRegistrationRequest(tx *database.DB, contestID, userID int64) (*models.ContestRegistrationRequests, error)
	// GetContestTask retrieves the ContestTask relationship for validation
	GetContestTask(tx *database.DB, contestID, taskID int64) (*models.ContestTask, error)
}

type contestRepository struct{}

func (cr *contestRepository) Create(tx *database.DB, contest *models.Contest) (int64, error) {
	err := tx.Create(contest).Error()
	if err != nil {
		return 0, err
	}
	return contest.ID, nil
}

func (cr *contestRepository) Get(tx *database.DB, contestID int64) (*models.Contest, error) {
	var contest models.Contest
	err := tx.Where("id = ?", contestID).First(&contest).Error()
	if err != nil {
		return nil, err
	}
	return &contest, nil
}

func (cr *contestRepository) GetAll(tx *database.DB, offset int, limit int, sort string) ([]models.Contest, error) {
	var contests []models.Contest
	tx = tx.ApplyPaginationAndSort(limit, offset, sort)
	err := tx.Model(&models.Contest{}).Find(&contests).Error()
	if err != nil {
		return nil, err
	}
	return contests, nil
}

func (cr *contestRepository) GetAllForCreator(tx *database.DB, creatorID int64, offset int, limit int, sort string) ([]models.Contest, error) {
	var contests []models.Contest
	tx = tx.ApplyPaginationAndSort(limit, offset, sort)
	err := tx.Model(&models.Contest{}).Where("created_by = ?", creatorID).Find(&contests).Error()
	if err != nil {
		return nil, err
	}
	return contests, nil
}

func (cr *contestRepository) Edit(tx *database.DB, contestID int64, contest *models.Contest) (*models.Contest, error) {
	err := tx.Model(&models.Contest{}).Where("id = ?", contestID).Updates(contest).Error()
	if err != nil {
		return nil, err
	}
	return cr.Get(tx, contestID)
}

func (cr *contestRepository) Delete(tx *database.DB, contestID int64) error {
	err := tx.Where("id = ?", contestID).Delete(&models.Contest{}).Error()
	if err != nil {
		return err
	}
	return nil
}

func (cr *contestRepository) CreatePendingRegistration(tx *database.DB, registration *models.ContestRegistrationRequests) (int64, error) {
	err := tx.Create(registration).Error()
	if err != nil {
		return 0, err
	}
	return registration.ID, nil
}

func (cr *contestRepository) IsPendingRegistrationExists(tx *database.DB, contestID int64, userID int64) (bool, error) {
	var count int64
	err := tx.Model(&models.ContestRegistrationRequests{}).
		Where("contest_id = ? AND user_id = ? AND status = ?", contestID, userID, types.RegistrationRequestStatusPending).
		Count(&count).Error()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (cr *contestRepository) IsUserParticipant(tx *database.DB, contestID int64, userID int64) (bool, error) {
	var userCount int64
	err := tx.Model(&models.ContestParticipant{}).
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		Count(&userCount).Error()
	if err != nil {
		return false, err
	}
	var groupCount int64
	err = tx.Model(&models.ContestParticipantGroup{}).Where("contest_id = ?", contestID).
		Joins(fmt.Sprintf("JOIN %s ON contest_participant_groups.group_id = user_groups.group_id", database.ResolveTableName(tx.GormDB(), &models.UserGroup{}))).
		Where("user_groups.user_id = ?", userID).
		Count(&groupCount).Error()
	if err != nil {
		return false, err
	}
	return userCount > 0 || groupCount > 0, nil
}

func (cr *contestRepository) GetAllWithStats(tx *database.DB, userID int64, offset int, limit int, sort string) ([]models.ContestWithStats, error) {
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
		) as direct_participants ON contests.id = direct_participants.contest_id`, database.ResolveTableName(tx.GormDB(), &models.ContestParticipant{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT cpg.contest_id, COUNT(DISTINCT ug.user_id) as count
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			GROUP BY cpg.contest_id
		) as group_participants ON contests.id = group_participants.contest_id`, database.ResolveTableName(tx.GormDB(), &models.ContestParticipantGroup{}), database.ResolveTableName(tx.GormDB(), &models.UserGroup{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) as count
			FROM %s
			GROUP BY contest_id
		) as contest_task_count ON contests.id = contest_task_count.contest_id`, database.ResolveTableName(tx.GormDB(), &models.ContestTask{}))).
		Joins(fmt.Sprintf(`LEFT JOIN %s user_participants ON contests.id = user_participants.contest_id AND user_participants.user_id = ?`, database.ResolveTableName(tx.GormDB(), &models.ContestParticipant{})), userID).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT DISTINCT cpg.contest_id, ug.user_id
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			WHERE ug.user_id = ?
		) as user_group_participants ON contests.id = user_group_participants.contest_id`, database.ResolveTableName(tx.GormDB(), &models.ContestParticipantGroup{}), database.ResolveTableName(tx.GormDB(), &models.UserGroup{})), userID).
		Joins(fmt.Sprintf(`LEFT JOIN %s pending_regs ON contests.id = pending_regs.contest_id AND pending_regs.user_id = ?`, database.ResolveTableName(tx.GormDB(), &models.ContestRegistrationRequests{})), userID)

	// Apply pagination and sorting
	query = query.ApplyPaginationAndSort(limit, offset, sort)

	err := query.Find(&contests).Error()
	if err != nil {
		return nil, err
	}

	return contests, nil
}

func (cr *contestRepository) GetOngoingContestsWithStats(tx *database.DB, userID int64, offset int, limit int, sort string) ([]models.ContestWithStats, error) {
	var contests []models.ContestWithStats

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
		) as direct_participants ON contests.id = direct_participants.contest_id`, database.ResolveTableName(tx.GormDB(), &models.ContestParticipant{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT cpg.contest_id, COUNT(DISTINCT ug.user_id) as count
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			GROUP BY cpg.contest_id
		) as group_participants ON contests.id = group_participants.contest_id`, database.ResolveTableName(tx.GormDB(), &models.ContestParticipantGroup{}), database.ResolveTableName(tx.GormDB(), &models.UserGroup{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) as count
			FROM %s
			GROUP BY contest_id
		) as contest_task_count ON contests.id = contest_task_count.contest_id`, database.ResolveTableName(tx.GormDB(), &models.ContestTask{}))).
		Joins(fmt.Sprintf(`LEFT JOIN %s user_participants ON contests.id = user_participants.contest_id AND user_participants.user_id = ?`, database.ResolveTableName(tx.GormDB(), &models.ContestParticipant{})), userID).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT DISTINCT cpg.contest_id, ug.user_id
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			WHERE ug.user_id = ?
		) as user_group_participants ON contests.id = user_group_participants.contest_id`, database.ResolveTableName(tx.GormDB(), &models.ContestParticipantGroup{}), database.ResolveTableName(tx.GormDB(), &models.UserGroup{})), userID).
		Joins(fmt.Sprintf(`LEFT JOIN %s pending_regs ON contests.id = pending_regs.contest_id AND pending_regs.user_id = ?`, database.ResolveTableName(tx.GormDB(), &models.ContestRegistrationRequests{})), userID).
		Where(`(
			(start_at IS NOT NULL AND start_at <= NOW() AND (end_at IS NULL OR end_at > NOW()))
		)`)

	// Apply pagination and sorting
	query = query.ApplyPaginationAndSort(limit, offset, sort)

	err := query.Find(&contests).Error()
	if err != nil {
		return nil, err
	}

	return contests, nil
}

func (cr *contestRepository) GetPastContestsWithStats(tx *database.DB, userID int64, offset int, limit int, sort string) ([]models.ContestWithStats, error) {
	var contests []models.ContestWithStats

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
		) as direct_participants ON contests.id = direct_participants.contest_id`, database.ResolveTableName(tx.GormDB(), &models.ContestParticipant{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT cpg.contest_id, COUNT(DISTINCT ug.user_id) as count
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			GROUP BY cpg.contest_id
		) as group_participants ON contests.id = group_participants.contest_id`, database.ResolveTableName(tx.GormDB(), &models.ContestParticipantGroup{}), database.ResolveTableName(tx.GormDB(), &models.UserGroup{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) as count
			FROM %s
			GROUP BY contest_id
		) as contest_task_count ON contests.id = contest_task_count.contest_id`, database.ResolveTableName(tx.GormDB(), &models.ContestTask{}))).
		Joins(fmt.Sprintf(`LEFT JOIN %s user_participants ON contests.id = user_participants.contest_id AND user_participants.user_id = ?`, database.ResolveTableName(tx.GormDB(), &models.ContestParticipant{})), userID).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT DISTINCT cpg.contest_id, ug.user_id
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			WHERE ug.user_id = ?
		) as user_group_participants ON contests.id = user_group_participants.contest_id`, database.ResolveTableName(tx.GormDB(), &models.ContestParticipantGroup{}), database.ResolveTableName(tx.GormDB(), &models.UserGroup{})), userID).
		Joins(fmt.Sprintf(`LEFT JOIN %s pending_regs ON contests.id = pending_regs.contest_id AND pending_regs.user_id = ?`, database.ResolveTableName(tx.GormDB(), &models.ContestRegistrationRequests{})), userID).
		Where("end_at IS NOT NULL AND end_at <= NOW()")

	// Apply pagination and sorting
	query = query.ApplyPaginationAndSort(limit, offset, sort)

	err := query.Find(&contests).Error()
	if err != nil {
		return nil, err
	}

	return contests, nil
}

func (cr *contestRepository) GetUpcomingContestsWithStats(tx *database.DB, userID int64, offset int, limit int, sort string) ([]models.ContestWithStats, error) {
	var contests []models.ContestWithStats

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
		) as direct_participants ON contests.id = direct_participants.contest_id`, database.ResolveTableName(tx.GormDB(), &models.ContestParticipant{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT cpg.contest_id, COUNT(DISTINCT ug.user_id) as count
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			GROUP BY cpg.contest_id
		) as group_participants ON contests.id = group_participants.contest_id`, database.ResolveTableName(tx.GormDB(), &models.ContestParticipantGroup{}), database.ResolveTableName(tx.GormDB(), &models.UserGroup{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) as count
			FROM %s
			GROUP BY contest_id
		) as contest_task_count ON contests.id = contest_task_count.contest_id`, database.ResolveTableName(tx.GormDB(), &models.ContestTask{}))).
		Joins(fmt.Sprintf(`LEFT JOIN %s user_participants ON contests.id = user_participants.contest_id AND user_participants.user_id = ?`, database.ResolveTableName(tx.GormDB(), &models.ContestParticipant{})), userID).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT DISTINCT cpg.contest_id, ug.user_id
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			WHERE ug.user_id = ?
		) as user_group_participants ON contests.id = user_group_participants.contest_id`, database.ResolveTableName(tx.GormDB(), &models.ContestParticipantGroup{}), database.ResolveTableName(tx.GormDB(), &models.UserGroup{})), userID).
		Joins(fmt.Sprintf(`LEFT JOIN %s pending_regs ON contests.id = pending_regs.contest_id AND pending_regs.user_id = ?`, database.ResolveTableName(tx.GormDB(), &models.ContestRegistrationRequests{})), userID).
		Where("start_at IS NOT NULL AND start_at > NOW()")

	// Apply pagination and sorting
	query = query.ApplyPaginationAndSort(limit, offset, sort)

	err := query.Find(&contests).Error()
	if err != nil {
		return nil, err
	}

	return contests, nil
}

func (cr *contestRepository) GetTasksForContest(tx *database.DB, contestID int64) ([]models.Task, error) {
	var tasks []models.Task
	err := tx.Model(&models.Task{}).
		Joins(fmt.Sprintf("JOIN %s ON contest_tasks.task_id = tasks.id", database.ResolveTableName(tx.GormDB(), &models.ContestTask{}))).
		Where("contest_tasks.contest_id = ?", contestID).
		Find(&tasks).Error()
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (cr *contestRepository) GetTasksForContestWithStats(tx *database.DB, contestID, userID int64) ([]models.Task, error) {
	var tasks []models.Task
	err := tx.Model(&models.Task{}).
		Joins(fmt.Sprintf("JOIN %s ON contest_tasks.task_id = tasks.id", database.ResolveTableName(tx.GormDB(), &models.ContestTask{}))).
		Where("contest_tasks.contest_id = ?", contestID).
		Find(&tasks).Error()
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (cr *contestRepository) GetContestsForUserWithStats(tx *database.DB, userID int64) ([]models.ParticipantContestStats, error) {
	var contests []models.ParticipantContestStats

	// Build query to get contests where user is a participant with statistics
	// including how many tasks the user has solved (achieved 100% score)
	query := tx.Model(&models.Contest{}).
		Select(`contests.*,
			COALESCE(direct_participants.count, 0) + COALESCE(group_participants.count, 0) as participant_count,
			COALESCE(contest_task_count.count, 0) as task_count,
			COALESCE(user_solved_tasks.count, 0) as solved_count`).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) as count
			FROM %s
			GROUP BY contest_id
		) as direct_participants ON contests.id = direct_participants.contest_id`, database.ResolveTableName(tx.GormDB(), &models.ContestParticipant{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT cpg.contest_id, COUNT(DISTINCT ug.user_id) as count
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			GROUP BY cpg.contest_id
		) as group_participants ON contests.id = group_participants.contest_id`, database.ResolveTableName(tx.GormDB(), &models.ContestParticipantGroup{}), database.ResolveTableName(tx.GormDB(), &models.UserGroup{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) as count
			FROM %s
			GROUP BY contest_id
		) as contest_task_count ON contests.id = contest_task_count.contest_id`, database.ResolveTableName(tx.GormDB(), &models.ContestTask{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT ct.contest_id, COUNT(*) as count
			FROM %s ct
			INNER JOIN %s s ON s.task_id = ct.task_id AND s.user_id = ?
			INNER JOIN %s sr ON sr.submission_id = s.id
			INNER JOIN (
				SELECT submission_result_id,
					   COUNT(*) as total_tests,
					   COUNT(CASE WHEN passed = true THEN 1 END) as passed_tests
				FROM %s
				GROUP BY submission_result_id
				HAVING COUNT(*) = COUNT(CASE WHEN passed = true THEN 1 END)
			) perfect_results ON perfect_results.submission_result_id = sr.id
			GROUP BY ct.contest_id
		) as user_solved_tasks ON contests.id = user_solved_tasks.contest_id`, database.ResolveTableName(tx.GormDB(), &models.ContestTask{}), database.ResolveTableName(tx.GormDB(), &models.Submission{}), database.ResolveTableName(tx.GormDB(), &models.SubmissionResult{}), database.ResolveTableName(tx.GormDB(), &models.TestResult{})), userID).
		Where(fmt.Sprintf(`EXISTS (
			SELECT 1 FROM %s cp
			WHERE cp.contest_id = contests.id AND cp.user_id = ?
		) OR EXISTS (
			SELECT 1 FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			WHERE cpg.contest_id = contests.id AND ug.user_id = ?
		)`, database.ResolveTableName(tx.GormDB(), &models.ContestParticipant{}), database.ResolveTableName(tx.GormDB(), &models.ContestParticipantGroup{}), database.ResolveTableName(tx.GormDB(), &models.UserGroup{})), userID, userID)

	err := query.Find(&contests).Error()
	if err != nil {
		return nil, err
	}

	return contests, nil
}

func (cr *contestRepository) AddTaskToContest(tx *database.DB, taskContest models.ContestTask) error {
	err := tx.Create(&taskContest).Error()
	if err != nil {
		return err
	}
	return nil
}

func (cr *contestRepository) GetRegistrationRequests(tx *database.DB, contestID int64, status types.RegistrationRequestStatus) ([]models.ContestRegistrationRequests, error) {
	var requests []models.ContestRegistrationRequests
	err := tx.Preload("User").Where("contest_id = ? AND status = ?", contestID, status).Find(&requests).Error()
	if err != nil {
		return nil, err
	}
	return requests, nil
}

func (cr *contestRepository) UpdateRegistrationRequestStatus(tx *database.DB, requestID int64, status types.RegistrationRequestStatus) error {
	return tx.Model(models.ContestRegistrationRequests{}).Where("id = ?", requestID).Updates(models.ContestRegistrationRequests{Status: status}).Error()
}

func (cr *contestRepository) DeleteRegistrationRequest(tx *database.DB, requestID int64) error {
	return tx.Model(models.ContestRegistrationRequests{}).Delete(&models.ContestRegistrationRequests{ID: requestID}).Error()
}

func (cr *contestRepository) CreateContestParticipant(tx *database.DB, contestID, userID int64) error {
	participant := &models.ContestParticipant{
		ContestID: contestID,
		UserID:    userID,
	}
	err := tx.Create(participant).Error()
	if err != nil {
		return err
	}
	return nil
}

func (cr *contestRepository) GetPendingRegistrationRequest(tx *database.DB, contestID, userID int64) (*models.ContestRegistrationRequests, error) {
	var request []*models.ContestRegistrationRequests
	err := tx.Where("contest_id = ? AND user_id = ? AND status = ?", contestID, userID, types.RegistrationRequestStatusPending).Limit(1).Find(&request).Error()
	if err != nil {
		return nil, err
	}
	if len(request) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return request[0], nil
}

func (cr *contestRepository) GetContestTask(tx *database.DB, contestID, taskID int64) (*models.ContestTask, error) {
	var contestTask models.ContestTask
	err := tx.Where("contest_id = ? AND task_id = ?", contestID, taskID).First(&contestTask).Error()
	if err != nil {
		return nil, err
	}
	return &contestTask, nil
}

func NewContestRepository() ContestRepository {
	return &contestRepository{}
}
