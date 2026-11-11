package repository

import (
	"fmt"
	"strings"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/utils"
)

type ContestDetailed struct {
	models.Contest
	CreatedByName    string `gorm:"column:created_by_name"`
	ParticipantCount int64  `gorm:"column:participant_count"`
	TaskCount        int64  `gorm:"column:task_count"` // number of tasks
}

type ContestRepository interface {
	// Create creates a new contest
	Create(db database.Database, contest *models.Contest) (int64, error)
	// Get retrieves a contest by ID
	Get(db database.Database, contestID int64) (*models.Contest, error)
	// GetDetailed retrieves a contest by ID with detailed information
	GetDetailed(db database.Database, contestID int64) (*ContestDetailed, error)
	// Get retrieves a contest by ID with participant and task counts
	GetWithCount(db database.Database, contestID int64) (*models.ParticipantContestStats, error)
	// GetWithCreator retrieves a contest by ID with preloaded creator and stats
	GetWithCreator(db database.Database, contestID int64) (*models.ParticipantContestStats, *models.User, error)
	// GetAll retrieves all contests with pagination and sorting
	GetAll(db database.Database, offset int, limit int, sort string) ([]models.Contest, error)
	// GetAllForCreator retrieves all contests created by a specific user with pagination and sorting
	GetAllForCreator(db database.Database, creatorID int64, offset int, limit int, sort string) ([]ContestDetailed, int64, error)
	// GetAllForCollaborator retrieves all contests where the user has any access control entry (collaborator) with pagination and sorting.
	// When all is true, every contest is returned and contests without an entry for the user get "owner" permission (admin view).
	GetAllForCollaborator(db database.Database, userID int64, all bool, offset int, limit int, sort string) ([]models.ManagedContest, int64, error)
	// Edit updates a contest
	Edit(db database.Database, contestID int64, updates map[string]any) (*models.Contest, error)
	// EditWithStats updates a contest and returns it with participant and task counts
	EditWithStats(db database.Database, contestID int64, updates map[string]any) (*models.ContestWithStats, error)
	// Delete removes a contest
	Delete(db database.Database, contestID int64) error
	// IsUserParticipant checks if user is already a participant
	IsUserParticipant(db database.Database, contestID int64, userID int64) (bool, error)
	// GetTasksForContest retrieves all tasks assigned to a contest
	GetTasksForContest(db database.Database, contestID int64) ([]models.Task, error)
	// GetContestTasksWithSettings retrieves contest-task relations with timing flags and associated task (paginated)
	GetContestTasksWithSettings(db database.Database, contestID int64, limit, offset int, sort, search string) ([]models.ContestTask, int64, error)
	// GetStartedContestTasksWithSettings retrieves already-started contest-task relations with timing flags and associated task
	GetStartedContestTasksWithSettings(db database.Database, contestID int64) ([]models.ContestTask, error)
	// GetTasksForContestWithStats retrieves all tasks assigned to a contest with submission statistics for a user
	GetTasksForContestWithStats(db database.Database, contestID, userID int64) ([]models.Task, error)
	// GetAssignableTasks retrieves all tasks NOT assigned to a contest (paginated)
	GetAssignableTasks(db database.Database, contestID int64, collaboratorID *int64, limit, offset int, sort, search string) ([]models.Task, int64, error)
	// GetContestsForUserWithStats retrieves contests with stats a user is participating in
	GetContestsForUserWithStats(db database.Database, userID int64) ([]models.ParticipantContestStats, error)
	// AddTasksToContest assigns tasks to a contest
	AddTaskToContest(db database.Database, taskContest models.ContestTask) error
	// UpdateTaskInContest updates a task's schedule in a contest
	UpdateTaskInContest(db database.Database, taskContest models.ContestTask) error
	// IsTaskInContest checks if a task is assigned to a contest
	IsTaskInContest(db database.Database, contestID, taskID int64) (bool, error)
	// RemoveTaskFromContest removes a task from a contest
	RemoveTaskFromContest(db database.Database, contestID, taskID int64) error
	// CreateContestParticipant adds a user as a participant to a contest
	CreateContestParticipant(db database.Database, contestID, userID int64) error
	// GetContestTask retrieves the ContestTask relationship for validation
	GetContestTask(db database.Database, contestID, taskID int64) (*models.ContestTask, error)
	GetTaskContests(db database.Database, taskID int64) ([]int64, error)
	// AddGroupToContest adds a group as a participant group to a contest
	AddGroupToContest(db database.Database, contestID, groupID int64) error
	// RemoveGroupFromContest removes a group from a contest's participant groups
	RemoveGroupFromContest(db database.Database, contestID, groupID int64) error
	// GetContestGroups retrieves all groups assigned to a contest
	GetContestGroups(db database.Database, contestID int64) ([]models.Group, error)
	// GetAssignableGroups retrieves all groups NOT assigned to a contest
	GetAssignableGroups(db database.Database, contestID int64) ([]models.Group, error)
	// GetGroupMembersForGroups retrieves members for a set of groups in a single query
	GetGroupMembersForGroups(db database.Database, groupIDs []int64) (map[int64][]models.User, error)
	// GetContestParticipants retrieves all effective participants of a contest (direct assignments and users in assigned groups)
	GetContestParticipants(db database.Database, contestID int64, limit, offset int, sort string) ([]models.User, int64, error)
	// GetParticipantAccessSources returns, for each requested user, how they access the contest (direct assignment and/or assigned groups)
	GetParticipantAccessSources(db database.Database, contestID int64, userIDs []int64) (map[int64]models.ParticipantAccessInfo, error)
	// GetAssignableParticipants retrieves all users NOT assigned as participants to a contest with pagination
	GetAssignableParticipants(db database.Database, contestID int64, limit, offset int, sort string) ([]models.User, int64, error)
	// AddParticipantsToContest adds multiple users as participants to a contest
	AddParticipantsToContest(db database.Database, contestID int64, userIDs []int64) error
	// RemoveParticipantsFromContest removes multiple users from contest participants
	RemoveParticipantsFromContest(db database.Database, contestID int64, userIDs []int64) error
}

type contestRepository struct{}

func (cr *contestRepository) Create(db database.Database, contest *models.Contest) (int64, error) {
	tx := db.GetInstance()
	err := tx.Create(contest).Error
	if err != nil {
		return 0, err
	}
	return contest.ID, nil
}

func (cr *contestRepository) Get(db database.Database, contestID int64) (*models.Contest, error) {
	tx := db.GetInstance()
	var contest models.Contest
	err := tx.Where("id = ?", contestID).First(&contest).Error
	if err != nil {
		return nil, err
	}
	return &contest, nil
}

func (cr *contestRepository) GetDetailed(db database.Database, contestID int64) (*ContestDetailed, error) {
	tx := db.GetInstance()
	var contest ContestDetailed
	err := tx.Model(&models.Contest{}).
		Select(`contests.*, (users.name || ' ' || users.surname) as created_by_name, pc.participant_count, tc.task_count`).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) as participant_count
			FROM %s
			WHERE contest_id = ?
			GROUP BY contest_id
		) as pc ON contests.id = pc.contest_id`, database.ResolveTableName(tx, &models.ContestParticipant{})), contestID).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT contest_id, COUNT(*) as task_count
			FROM %s
			WHERE contest_id = ? AND start_at <= NOW()
			GROUP BY contest_id
		) as tc ON contests.id = tc.contest_id`, database.ResolveTableName(tx, &models.ContestTask{})), contestID).
		Joins(fmt.Sprintf(`LEFT JOIN %s as users
			ON contests.created_by = users.id`, database.ResolveTableName(tx, &models.User{}))).
		Where("contests.id = ?", contestID).
		First(&contest).Error
	if err != nil {
		return nil, err
	}
	return &contest, nil
}

func (cr *contestRepository) GetAll(db database.Database, offset int, limit int, sort string) ([]models.Contest, error) {
	tx := db.GetInstance()
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

func (cr *contestRepository) GetAllForCreator(db database.Database, creatorID int64, offset int, limit int, sort string) ([]ContestDetailed, int64, error) {
	tx := db.GetInstance()
	var contests []ContestDetailed
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

	err = paginatedQuery.Select(`*, (SELECT COUNT(*) FROM maxit.contest_participants WHERE contest_id = contests.id) AS participant_count, (SELECT COUNT(*) FROM maxit.contest_tasks WHERE contest_id = contests.id) AS task_count`).Where("created_by = ?", creatorID).Find(&contests).Error
	if err != nil {
		return nil, 0, err
	}
	return contests, totalCount, nil
}

// GetAllForCollaborator retrieves all contests where the user has any access control entry (collaborator)
// It returns contests with pagination and total count, including participant and task counts.
// When all is true, every contest is returned; the user's own entry is used where present, otherwise "owner".
func (cr *contestRepository) GetAllForCollaborator(db database.Database, userID int64, all bool, offset int, limit int, sort string) ([]models.ManagedContest, int64, error) {
	tx := db.GetInstance()
	var contests []models.ManagedContest
	var totalCount int64

	accessControlTable := database.ResolveTableName(tx, &models.AccessControl{})
	contestParticipantTable := database.ResolveTableName(tx, &models.ContestParticipant{})
	contestTaskTable := database.ResolveTableName(tx, &models.ContestTask{})

	acJoin, permissionExpr := "JOIN", "ac.permission"
	if all {
		acJoin, permissionExpr = "LEFT JOIN", fmt.Sprintf("COALESCE(ac.permission, '%s')", types.PermissionOwner)
	}
	acJoinClause := fmt.Sprintf("%s %s ac ON ac.resource_type = ? AND ac.resource_id = contests.id AND ac.user_id = ?", acJoin, accessControlTable)

	// Base query joining access control entries for contests
	baseQuery := tx.Model(&models.Contest{}).
		Joins(acJoinClause, types.ResourceTypeContest, userID)

	err := baseQuery.Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Use JOIN-based approach for participant_count and task_count
	paginatedQuery, err := utils.ApplyPaginationAndSort(
		tx.Model(&models.Contest{}).
			Select(`contests.*,
				`+permissionExpr+` as permission_type,
				COALESCE(direct_participants.count, 0) + COALESCE(group_participants.count, 0) as participant_count,
				COALESCE(contest_task_count.count, 0) as task_count
			`).
			Joins(acJoinClause, types.ResourceTypeContest, userID).
			Joins(fmt.Sprintf(`LEFT JOIN (
				SELECT contest_id, COUNT(*) as count
				FROM %s
				GROUP BY contest_id
			) as direct_participants ON contests.id = direct_participants.contest_id`, contestParticipantTable)).
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
			) as contest_task_count ON contests.id = contest_task_count.contest_id`, contestTaskTable)),
		limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = paginatedQuery.Find(&contests).Error
	if err != nil {
		return nil, 0, err
	}

	return contests, totalCount, nil
}
func (cr *contestRepository) GetAllForCreatorWithStats(db database.Database, creatorID int64, offset int, limit int, sort string) ([]models.ContestWithStats, error) {
	tx := db.GetInstance()
	var contests []models.ContestWithStats

	// Build a query that calculates participant counts and task counts for contests created by a specific user.
	// Similar to GetAllWithStats but filtered by creator.
	query := tx.Model(&models.Contest{}).
		Select(`contests.*,
			COALESCE(direct_participants.count, 0) + COALESCE(group_participants.count, 0) as participant_count,
			COALESCE(contest_task_count.count, 0) as task_count,
			false as is_participant`).
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

func (cr *contestRepository) Edit(db database.Database, contestID int64, updates map[string]any) (*models.Contest, error) {
	tx := db.GetInstance()
	err := tx.Model(&models.Contest{}).Where("id = ?", contestID).Updates(updates).Error
	if err != nil {
		return nil, err
	}
	var updatedContest models.Contest
	err = tx.Model(&models.Contest{}).Where("id = ?", contestID).First(&updatedContest).Error
	return &updatedContest, err
}

func (cr *contestRepository) EditWithStats(db database.Database, contestID int64, updates map[string]any) (*models.ContestWithStats, error) {
	tx := db.GetInstance()
	// First update the contest
	err := tx.Model(&models.Contest{}).Where("id = ?", contestID).Updates(updates).Error
	if err != nil {
		return nil, err
	}

	// Then retrieve with stats
	var contestWithStats models.ContestWithStats
	query := tx.Model(&models.Contest{}).
		Select(`contests.*,
			COALESCE(direct_participants.count, 0) + COALESCE(group_participants.count, 0) as participant_count,
			COALESCE(contest_task_count.count, 0) as task_count,
			false as is_participant`).
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

func (cr *contestRepository) Delete(db database.Database, contestID int64) error {
	tx := db.GetInstance()
	err := tx.Where("id = ?", contestID).Delete(&models.Contest{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (cr *contestRepository) IsUserParticipant(db database.Database, contestID int64, userID int64) (bool, error) {
	tx := db.GetInstance()
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

// GetContestTasksWithSettings retrieves contest-task relations (with timing flags) and preloads the associated Task
func (cr *contestRepository) GetContestTasksWithSettings(db database.Database, contestID int64, limit, offset int, sort, search string) ([]models.ContestTask, int64, error) {
	tx := db.GetInstance()
	var relations []models.ContestTask
	var totalCount int64

	taskTable := database.ResolveTableName(tx, &models.Task{})
	contestTaskTable := database.ResolveTableName(tx, &models.ContestTask{})

	base := tx.Unscoped().Model(&models.ContestTask{}).
		Joins(fmt.Sprintf("JOIN %s ON %s.id = %s.task_id", taskTable, taskTable, contestTaskTable)).
		Where(fmt.Sprintf("%s.contest_id = ?", contestTaskTable), contestID)

	if search != "" {
		base = base.Where(fmt.Sprintf("%s.title ILIKE ?", taskTable), "%"+search+"%")
	}

	if err := base.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	query := base
	if sort != "" {
		query = query.Order(qualifyContestTaskSort(sort, contestTaskTable, taskTable))
	}

	err := query.
		Limit(limit).
		Offset(offset).
		Preload("Task").
		Preload("Task.Author").
		Find(&relations).Error
	if err != nil {
		return nil, 0, err
	}
	return relations, totalCount, nil
}

// qualifyContestTaskSort qualifies sort fields with their table prefix for a joined contest_tasks/tasks query
func qualifyContestTaskSort(sortBy, contestTaskTable, taskTable string) string {
	fieldMap := map[string]string{
		"id":       contestTaskTable + ".task_id",
		"task_id":  contestTaskTable + ".task_id",
		"start_at": contestTaskTable + ".start_at",
		"end_at":   contestTaskTable + ".end_at",
		"title":    taskTable + ".title",
	}
	parts := strings.Split(sortBy, ":")
	if len(parts) != 2 {
		return sortBy
	}
	field, ok := fieldMap[parts[0]]
	if !ok {
		return sortBy
	}
	dir := parts[1]
	if dir != "asc" && dir != "desc" {
		dir = "asc"
	}
	return field + " " + dir
}

// GetStartedContestTasksWithSettings retrieves already-started contest-task relations (with timing flags) and preloads the associated Task
func (cr *contestRepository) GetStartedContestTasksWithSettings(db database.Database, contestID int64) ([]models.ContestTask, error) {
	tx := db.GetInstance()
	var relations []models.ContestTask
	err := tx.Model(&models.ContestTask{}).
		Where("contest_id = ? AND start_at <= now()", contestID).
		Preload("Task").
		Preload("Task.Author").
		Find(&relations).Error
	if err != nil {
		return nil, err
	}
	return relations, nil
}

func (cr *contestRepository) GetTasksForContest(db database.Database, contestID int64) ([]models.Task, error) {
	tx := db.GetInstance()
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

func (cr *contestRepository) GetTasksForContestWithStats(db database.Database, contestID, userID int64) ([]models.Task, error) {
	tx := db.GetInstance()
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

func (cr *contestRepository) GetAssignableTasks(db database.Database, contestID int64, collaboratorID *int64, limit, offset int, sort, search string) ([]models.Task, int64, error) {
	tx := db.GetInstance()
	var tasks []models.Task
	var totalCount int64

	base := tx.Model(&models.Task{}).
		Where("id NOT IN (?)",
			tx.Table(database.ResolveTableName(tx, &models.ContestTask{})).
				Select("task_id").
				Where("contest_id = ?", contestID))

	if collaboratorID != nil {
		base = base.Where("id IN (?)",
			tx.Table(database.ResolveTableName(tx, &models.AccessControl{})).
				Select("resource_id").
				Where("resource_type = ? AND user_id = ?", types.ResourceTypeTask, *collaboratorID))
	}

	if search != "" {
		base = base.Where("title ILIKE ?", "%"+search+"%")
	}

	if err := base.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	query, err := utils.ApplyPaginationAndSort(base, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = query.Preload("Author").Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}
	return tasks, totalCount, nil
}

func (cr *contestRepository) GetContestsForUserWithStats(db database.Database, userID int64) ([]models.ParticipantContestStats, error) {
	tx := db.GetInstance()
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

func (cr *contestRepository) AddTaskToContest(db database.Database, taskContest models.ContestTask) error {
	tx := db.GetInstance()
	err := tx.Create(&taskContest).Error
	if err != nil {
		return err
	}
	return nil
}

func (cr *contestRepository) RemoveTaskFromContest(db database.Database, contestID, taskID int64) error {
	tx := db.GetInstance()
	err := tx.Where("contest_id = ? AND task_id = ?", contestID, taskID).Delete(&models.ContestTask{}).Error
	return err
}

func (cr *contestRepository) UpdateTaskInContest(db database.Database, taskContest models.ContestTask) error {
	tx := db.GetInstance()
	err := tx.Model(&models.ContestTask{}).
		Where("contest_id = ? AND task_id = ?", taskContest.ContestID, taskContest.TaskID).
		Updates(map[string]any{
			"start_at": taskContest.StartAt,
			"end_at":   taskContest.EndAt,
		}).Error
	return err
}

func (cr *contestRepository) IsTaskInContest(db database.Database, contestID, taskID int64) (bool, error) {
	tx := db.GetInstance()
	var count int64
	err := tx.Model(&models.ContestTask{}).
		Where("contest_id = ? AND task_id = ?", contestID, taskID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (cr *contestRepository) CreateContestParticipant(db database.Database, contestID, userID int64) error {
	tx := db.GetInstance()
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

func (cr *contestRepository) GetContestTask(db database.Database, contestID, taskID int64) (*models.ContestTask, error) {
	tx := db.GetInstance()
	var contestTask models.ContestTask
	err := tx.Where("contest_id = ? AND task_id = ?", contestID, taskID).First(&contestTask).Error
	if err != nil {
		return nil, err
	}
	return &contestTask, nil
}

func (cr *contestRepository) GetTaskContests(db database.Database, taskID int64) ([]int64, error) {
	tx := db.GetInstance()
	var contestIDs []int64
	err := tx.Model(&models.ContestTask{}).Where("task_id = ?", taskID).Pluck("contest_id", &contestIDs).Error
	if err != nil {
		return nil, err
	}
	return contestIDs, nil
}

func (cr *contestRepository) GetWithCount(db database.Database, contestID int64) (*models.ParticipantContestStats, error) {
	tx := db.GetInstance()
	var contest models.ParticipantContestStats
	err := tx.Model(&models.Contest{}).
		Select(`
		contests.*,
		COALESCE(direct_participants.count, 0) + COALESCE(group_participants.count, 0) as participant_count,
		COALESCE(contest_task_count.count, 0) as task_count
		`).
		Where("id = ?", contestID).
		Joins(fmt.Sprintf(`
		LEFT JOIN (
			SELECT contest_id, COUNT(*) AS count
			FROM %s
			GROUP BY contest_id
		) AS direct_participants ON contests.id = direct_participants.contest_id`,
			database.ResolveTableName(tx, &models.ContestParticipant{})),
		).
		// Group participants expanded to distinct users
		Joins(fmt.Sprintf(`
		LEFT JOIN (
			SELECT cpg.contest_id, COUNT(DISTINCT ug.user_id) AS count
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			GROUP BY cpg.contest_id
		) AS group_participants ON contests.id = group_participants.contest_id`,
			database.ResolveTableName(tx, &models.ContestParticipantGroup{}),
			database.ResolveTableName(tx, &models.UserGroup{})),
		).
		// Task count per contest
		Joins(fmt.Sprintf(`
		LEFT JOIN (
			SELECT contest_id, COUNT(*) AS count
			FROM %s
			GROUP BY contest_id
		) AS contest_task_count ON contests.id = contest_task_count.contest_id`,
			database.ResolveTableName(tx, &models.ContestTask{})),
		).First(&contest).Error
	if err != nil {
		return nil, err
	}
	return &contest, nil
}

func (cr *contestRepository) GetWithCreator(db database.Database, contestID int64) (*models.ParticipantContestStats, *models.User, error) {
	tx := db.GetInstance()
	var contest models.ParticipantContestStats
	err := tx.Model(&models.Contest{}).
		Select(`
		contests.*,
		COALESCE(direct_participants.count, 0) + COALESCE(group_participants.count, 0) as participant_count,
		COALESCE(contest_task_count.count, 0) as task_count
		`).
		Where("contests.id = ?", contestID).
		Joins(fmt.Sprintf(`
		LEFT JOIN (
			SELECT contest_id, COUNT(*) AS count
			FROM %s
			GROUP BY contest_id
		) AS direct_participants ON contests.id = direct_participants.contest_id`,
			database.ResolveTableName(tx, &models.ContestParticipant{})),
		).
		// Group participants expanded to distinct users
		Joins(fmt.Sprintf(`
		LEFT JOIN (
			SELECT cpg.contest_id, COUNT(DISTINCT ug.user_id) AS count
			FROM %s cpg
			JOIN %s ug ON cpg.group_id = ug.group_id
			GROUP BY cpg.contest_id
		) AS group_participants ON contests.id = group_participants.contest_id`,
			database.ResolveTableName(tx, &models.ContestParticipantGroup{}),
			database.ResolveTableName(tx, &models.UserGroup{})),
		).
		// Task count per contest
		Joins(fmt.Sprintf(`
		LEFT JOIN (
			SELECT contest_id, COUNT(*) AS count
			FROM %s
			GROUP BY contest_id
		) AS contest_task_count ON contests.id = contest_task_count.contest_id`,
			database.ResolveTableName(tx, &models.ContestTask{})),
		).First(&contest).Error
	if err != nil {
		return nil, nil, err
	}

	// Fetch the creator
	var creator models.User
	err = tx.Where("id = ?", contest.CreatedBy).First(&creator).Error
	if err != nil {
		return nil, nil, err
	}

	return &contest, &creator, nil
}

func (cr *contestRepository) AddGroupToContest(db database.Database, contestID, groupID int64) error {
	tx := db.GetInstance()
	participantGroup := &models.ContestParticipantGroup{
		ContestID: contestID,
		GroupID:   groupID,
	}
	err := tx.Create(participantGroup).Error
	if err != nil {
		return err
	}
	return nil
}

func (cr *contestRepository) RemoveGroupFromContest(db database.Database, contestID, groupID int64) error {
	tx := db.GetInstance()
	err := tx.Where("contest_id = ? AND group_id = ?", contestID, groupID).Delete(&models.ContestParticipantGroup{}).Error
	return err
}

func (cr *contestRepository) GetContestGroups(db database.Database, contestID int64) ([]models.Group, error) {
	tx := db.GetInstance()
	var groups []models.Group
	err := tx.Model(&models.Group{}).
		Joins(fmt.Sprintf("JOIN %s ON contest_participant_groups.group_id = groups.id", database.ResolveTableName(tx, &models.ContestParticipantGroup{}))).
		Where("contest_participant_groups.contest_id = ?", contestID).
		Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (cr *contestRepository) GetGroupMembersForGroups(db database.Database, groupIDs []int64) (map[int64][]models.User, error) {
	tx := db.GetInstance()
	result := make(map[int64][]models.User)

	if len(groupIDs) == 0 {
		return result, nil
	}

	ugTable := database.ResolveTableName(tx, &models.UserGroup{})

	type groupMemberRow struct {
		GroupID int64 `gorm:"column:group_id"`
		models.User
	}

	var rows []groupMemberRow
	err := tx.Model(&models.User{}).
		Select("users.*, "+ugTable+".group_id").
		Joins(fmt.Sprintf("JOIN %s ON %s.user_id = users.id", ugTable, ugTable)).
		Where(fmt.Sprintf("%s.group_id IN ?", ugTable), groupIDs).
		Order("users.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		result[row.GroupID] = append(result[row.GroupID], row.User)
	}

	return result, nil
}

func (cr *contestRepository) GetAssignableGroups(db database.Database, contestID int64) ([]models.Group, error) {
	tx := db.GetInstance()
	var groups []models.Group
	err := tx.Model(&models.Group{}).
		Where("id NOT IN (?)",
			tx.Table(database.ResolveTableName(tx, &models.ContestParticipantGroup{})).
				Select("group_id").
				Where("contest_id = ?", contestID),
		).
		Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (cr *contestRepository) GetContestParticipants(db database.Database, contestID int64, limit, offset int, sort string) ([]models.User, int64, error) {
	tx := db.GetInstance()
	var users []models.User
	var totalCount int64

	cpTable := database.ResolveTableName(tx, &models.ContestParticipant{})
	ugTable := database.ResolveTableName(tx, &models.UserGroup{})
	cpgTable := database.ResolveTableName(tx, &models.ContestParticipantGroup{})
	gTable := database.ResolveTableName(tx, &models.Group{})

	directExists := fmt.Sprintf("SELECT 1 FROM %s WHERE contest_id = ? AND user_id = users.id", cpTable)
	groupExists := fmt.Sprintf(
		"SELECT 1 FROM %s AS ug JOIN %s AS g ON g.id = ug.group_id JOIN %s AS cpg ON cpg.group_id = ug.group_id WHERE ug.user_id = users.id AND cpg.contest_id = ?",
		ugTable, gTable, cpgTable,
	)

	// Get total count of effective participants (direct assignments and users in assigned groups)
	if err := tx.Model(&models.User{}).
		Where("EXISTS ("+directExists+") OR EXISTS ("+groupExists+")", contestID, contestID).
		Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	paginatedTx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = paginatedTx.Model(&models.User{}).
		Where("EXISTS ("+directExists+") OR EXISTS ("+groupExists+")", contestID, contestID).
		Order("users.id ASC"). // unique tie-breaker so offset pages are stable on duplicate sort keys
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, totalCount, nil
}

func (cr *contestRepository) GetParticipantAccessSources(db database.Database, contestID int64, userIDs []int64) (map[int64]models.ParticipantAccessInfo, error) {
	tx := db.GetInstance()
	result := make(map[int64]models.ParticipantAccessInfo)

	if len(userIDs) == 0 {
		return result, nil
	}

	cpTable := database.ResolveTableName(tx, &models.ContestParticipant{})
	ugTable := database.ResolveTableName(tx, &models.UserGroup{})
	cpgTable := database.ResolveTableName(tx, &models.ContestParticipantGroup{})
	gTable := database.ResolveTableName(tx, &models.Group{})

	var directUserIDs []int64
	err := tx.Table(cpTable).
		Select("user_id").
		Where("contest_id = ? AND user_id IN ?", contestID, userIDs).
		Find(&directUserIDs).Error
	if err != nil {
		return nil, err
	}
	for _, userID := range directUserIDs {
		info := result[userID]
		info.Direct = true
		if info.Groups == nil {
			info.Groups = []models.Group{}
		}
		result[userID] = info
	}

	type groupAccess struct {
		UserID    int64  `gorm:"column:user_id"`
		GroupID   int64  `gorm:"column:group_id"`
		GroupName string `gorm:"column:group_name"`
	}
	var groupAccessRows []groupAccess
	err = tx.Table(ugTable+" AS ug").
		Select("ug.user_id, g.id AS group_id, g.name AS group_name").
		Joins("JOIN "+gTable+" AS g ON g.id = ug.group_id").
		Joins("JOIN "+cpgTable+" AS cpg ON cpg.group_id = ug.group_id").
		Where("ug.user_id IN ? AND cpg.contest_id = ?", userIDs, contestID).
		Order("g.id ASC").
		Find(&groupAccessRows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range groupAccessRows {
		info := result[row.UserID]
		if info.Groups == nil {
			info.Groups = []models.Group{}
		}
		info.Groups = append(info.Groups, models.Group{ID: row.GroupID, Name: row.GroupName})
		result[row.UserID] = info
	}

	return result, nil
}

func (cr *contestRepository) GetAssignableParticipants(db database.Database, contestID int64, limit, offset int, sort string) ([]models.User, int64, error) {
	tx := db.GetInstance()
	var users []models.User
	var totalCount int64

	cpTable := database.ResolveTableName(tx, &models.ContestParticipant{})
	ugTable := database.ResolveTableName(tx, &models.UserGroup{})
	cpgTable := database.ResolveTableName(tx, &models.ContestParticipantGroup{})
	assignableFilter := fmt.Sprintf(
		"id NOT IN (SELECT user_id FROM %s WHERE contest_id = ?) AND id NOT IN (SELECT ug.user_id FROM %s AS ug JOIN %s AS cpg ON cpg.group_id = ug.group_id WHERE cpg.contest_id = ?)",
		cpTable, ugTable, cpgTable,
	)

	// Get total count of assignable users (not already participants, directly or via an assigned group)
	if err := tx.Model(&models.User{}).
		Where(assignableFilter, contestID, contestID).
		Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	paginatedTx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = paginatedTx.Model(&models.User{}).
		Where(assignableFilter, contestID, contestID).
		Order("users.id ASC").
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, totalCount, nil
}

func (cr *contestRepository) AddParticipantsToContest(db database.Database, contestID int64, userIDs []int64) error {
	tx := db.GetInstance()
	participants := make([]models.ContestParticipant, len(userIDs))
	for i, userID := range userIDs {
		participants[i] = models.ContestParticipant{
			ContestID: contestID,
			UserID:    userID,
		}
	}
	return tx.Create(&participants).Error
}

func (cr *contestRepository) RemoveParticipantsFromContest(db database.Database, contestID int64, userIDs []int64) error {
	tx := db.GetInstance()
	err := tx.Where("contest_id = ? AND user_id IN ?", contestID, userIDs).Delete(&models.ContestParticipant{}).Error
	return err
}

func NewContestRepository() ContestRepository {
	return &contestRepository{}
}
