package repository

import (
	"fmt"
	"time"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/utils"
)

type SubmissionRepository interface {
	// Create creates a new submission and returns the submission ID.
	Create(db database.Database, submission *models.Submission) (int64, error)
	// GetAll returns all submissions. The submissions are paginated.
	GetAll(db database.Database, limit, offset int, sort string) ([]models.Submission, int64, error)
	// GetAllByUser returns all submissions by a user. The submissions are paginated.
	GetAllByUser(db database.Database, userID int64, limit, offset int, sort string) ([]models.Submission, int64, error)
	// GetAllForGroup returns all submissions for a group. The submissions are paginated.
	GetAllForGroup(db database.Database, groupID int64, limit, offset int, sort string) ([]models.Submission, int64, error)
	// GetAllForTask returns all submissions for a task. The submissions are paginated.
	GetAllForTask(db database.Database, taskID int64, limit, offset int, sort string) ([]models.Submission, int64, error)
	// GetAllForTaskByUser returns all submissions for a task by a user. The submissions are paginated.
	GetAllForTaskByUser(db database.Database, taskID, userID int64, limit, offset int, sort string) ([]models.Submission, int64, error)
	// GetAllForContest returns all submissions for a contest. The submissions are paginated.
	GetAllForContest(db database.Database, contestID int64, limit, offset int, sort string) ([]models.Submission, int64, error)
	// GetAllByUserForContest returns all submissions by a user for a specific contest. The submissions are paginated.
	GetAllByUserForContest(db database.Database, userID, contestID int64, limit, offset int, sort string) ([]models.Submission, int64, error)
	// GetAllByUserForTask returns all submissions by a user for a specific task. The submissions are paginated.
	GetAllByUserForTask(db database.Database, userID, taskID int64, limit, offset int, sort string) ([]models.Submission, int64, error)
	// GetAllByUserForContestAndTask returns all submissions by a user for a specific contest and task. The submissions are paginated.
	GetAllByUserForContestAndTask(db database.Database, userID, contestID, taskID int64, limit, offset int, sort string) ([]models.Submission, int64, error)
	// GetAllForTeacher returns all submissions for a teacher, this includes submissions for tasks created by this teacher.
	// The submissions are paginated.
	GetAllForTeacher(db database.Database, currentUserID int64, limit, offset int, sort string) ([]models.Submission, int64, error)
	// GetAllByUserForTeacher returns all submissions by a specific user, filtered to only include submissions
	// for tasks created by the teacher. The submissions are paginated.
	GetAllByUserForTeacher(db database.Database, userID, teacherID int64, limit, offset int, sort string) ([]models.Submission, int64, error)
	// GetAllByUserForTaskByTeacher returns all submissions by a user for a specific task,
	// filtered to only include submissions where the teacher created the task. The submissions are paginated.
	GetAllByUserForTaskByTeacher(db database.Database, userID, taskID, teacherID int64, limit, offset int, sort string) ([]models.Submission, int64, error)
	// GetAllByUserForContestByTeacher returns all submissions by a user for a specific contest,
	// filtered to only include submissions where the teacher created the contest or the task. The submissions are paginated.
	GetAllByUserForContestByTeacher(db database.Database, userID, contestID, teacherID int64, limit, offset int, sort string) ([]models.Submission, int64, error)
	// GetAllByUserForContestAndTaskByTeacher returns all submissions by a user for a specific contest and task,
	// filtered to only include submissions where the teacher created the contest or the task. The submissions are paginated.
	GetAllByUserForContestAndTaskByTeacher(db database.Database, userID, contestID, taskID, teacherID int64, limit, offset int, sort string) ([]models.Submission, int64, error)
	// GetLatestSubmissionForTaskByUser returns the latest submission for a task by a user.
	GetLatestForTaskByUser(db database.Database, taskID, userID int64) (*models.Submission, error)
	// Get returns a submission by its ID.
	Get(db database.Database, submissionID int64) (*models.Submission, error)
	// GetBestScoreForTaskByUser returns the best score (percentage of passed tests) for a task by a user.
	GetBestScoreForTaskByUser(db database.Database, taskID, userID int64) (float64, error)
	// GetAttemptCountForTaskByUser returns the number of submission attempts for a task by a user.
	GetAttemptCountForTaskByUser(db database.Database, taskID, userID int64) (int, error)
	// MarkEvaluated marks a submission as evaluated.
	MarkEvaluated(db database.Database, submissionID int64) error
	// MarkFailed marks a submission as failed.
	MarkFailed(db database.Database, submissionID int64, errorMsg string) error
	// MarkProcessing marks a submission as processing.
	MarkProcessing(db database.Database, submissionID int64) error
	// GetPendingSubmissions returns submissions that are in "received" status (not yet sent for evaluation).
	GetPendingSubmissions(db database.Database, limit int) ([]models.Submission, error)
	// GetTaskStatsForContest returns aggregated statistics for each task in a contest
	GetTaskStatsForContest(db database.Database, contestID int64) ([]models.ContestTaskStatsModel, error)
	// GetUserStatsForContestTask returns per-user statistics for a specific task in a contest
	GetUserStatsForContestTask(db database.Database, contestID, taskID int64) ([]models.TaskUserStatsModel, error)
	// GetUserStatsForContest returns overall statistics for each user in a contest
	GetUserStatsForContest(db database.Database, contestID int64, userID *int64) ([]models.UserContestStatsFull, error)
}

type submissionRepository struct{}

func (us *submissionRepository) GetAll(db database.Database, limit, offset int, sort string) ([]models.Submission, int64, error) {
	tx := db.GetInstance()
	submissions := []models.Submission{}
	var totalCount int64

	// Get total count first
	err := tx.Model(&models.Submission{}).Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	paginatedTx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = paginatedTx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Preload("Result").
		Preload("Result.TestResults").
		Find(&submissions).Error
	if err != nil {
		return nil, 0, err
	}
	return submissions, totalCount, nil
}

func (us *submissionRepository) GetAllForTeacher(
	db database.Database,
	userID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, int64, error) {
	tx := db.GetInstance()
	submissions := []models.Submission{}
	var totalCount int64

	// Get total count first
	err := tx.Model(&models.Submission{}).
		Joins(
			fmt.Sprintf("JOIN %s ON tasks.id = submissions.task_id",
				database.ResolveTableName(tx, &models.Task{}),
			)).
		Where("tasks.created_by = ?", userID).
		Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	tx, err = utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = tx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Joins(
			fmt.Sprintf("JOIN %s ON tasks.id = submissions.task_id",
				database.ResolveTableName(tx, &models.Task{}),
			)).
		Where("tasks.created_by = ?", userID).
		Find(&submissions).Error
	if err != nil {
		return nil, 0, err
	}
	return submissions, totalCount, nil
}

func (us *submissionRepository) Get(db database.Database, submissionID int64) (*models.Submission, error) {
	tx := db.GetInstance()
	var submission models.Submission
	err := tx.Where("id = ?", submissionID).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Preload("Result").
		Preload("File").
		Preload("Result.TestResults").
		First(&submission).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

func (us *submissionRepository) GetAllByUser(
	db database.Database,
	userID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, int64, error) {
	tx := db.GetInstance()
	submissions := []models.Submission{}
	var totalCount int64

	// Get total count first
	err := tx.Model(&models.Submission{}).Where("user_id = ?", userID).Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	paginatedTx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = paginatedTx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Preload("Result").
		Preload("Result.TestResults").
		Where("user_id = ?", userID).Find(&submissions).Error
	if err != nil {
		return nil, 0, err
	}

	return submissions, totalCount, nil
}

func (us *submissionRepository) GetAllForGroup(
	db database.Database,
	groupID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, int64, error) {
	tx := db.GetInstance()
	submissions := []models.Submission{}
	var totalCount int64

	// Get total count first
	countQuery := tx.Model(&models.Submission{}).
		Joins(fmt.Sprintf("JOIN %s ON users.id = submissions.user_id", database.ResolveTableName(tx, &models.User{}))).
		Joins(fmt.Sprintf("JOIN %s ON user_group.user_id = users.id", database.ResolveTableName(tx, &models.UserGroup{}))).
		Joins(fmt.Sprintf("JOIN %s ON groups.id = user_group.group_id", database.ResolveTableName(tx, &models.Group{}))).
		Where("groups.id = ?", groupID)
	err := countQuery.Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	paginatedTx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = paginatedTx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Preload("Result").
		Joins(fmt.Sprintf("JOIN %s ON users.id = submissions.user_id", database.ResolveTableName(tx, &models.User{}))).
		Joins(fmt.Sprintf("JOIN %s ON user_group.user_id = users.id", database.ResolveTableName(tx, &models.UserGroup{}))).
		Joins(fmt.Sprintf("JOIN %s ON groups.id = user_group.group_id", database.ResolveTableName(tx, &models.Group{}))).
		Where("groups.id = ?", groupID).
		Find(&submissions).Error

	if err != nil {
		return nil, 0, err
	}
	return submissions, totalCount, nil
}

func (us *submissionRepository) GetAllForTask(
	db database.Database,
	taskID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, int64, error) {
	tx := db.GetInstance()
	submissions := []models.Submission{}
	var totalCount int64

	// Get total count first
	err := tx.Model(&models.Submission{}).Where("task_id = ?", taskID).Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	paginatedTx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = paginatedTx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Preload("Result").
		Joins(fmt.Sprintf("JOIN %s ON tasks.id = submissions.task_id", database.ResolveTableName(tx, &models.Task{}))).
		Where("tasks.id = ?", taskID).
		Find(&submissions).Error

	if err != nil {
		return nil, 0, err
	}
	return submissions, totalCount, nil
}

func (us *submissionRepository) GetAllForTaskTeacher(
	db database.Database,
	taskID, userID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, error) {
	tx := db.GetInstance()
	submissions := []models.Submission{}

	tx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, err
	}

	err = tx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Preload("Result").
		Joins(fmt.Sprintf("JOIN %s ON tasks.id = submissions.task_id", database.ResolveTableName(tx, &models.Task{}))).
		Where("tasks.id = ? AND tasks.created_by = ?", taskID, userID).
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *submissionRepository) GetAllForTaskStudent(
	db database.Database,
	taskID, studentID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, error) {
	tx := db.GetInstance()
	submissions := []models.Submission{}

	tx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, err
	}

	err = tx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Joins(fmt.Sprintf("JOIN %s ON tasks.id = submissions.task_id", database.ResolveTableName(tx, &models.Task{}))).
		Where("tasks.id = ? AND submissions.user_id = ?", taskID, studentID).
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *submissionRepository) Create(db database.Database, submission *models.Submission) (int64, error) {
	tx := db.GetInstance()
	err := tx.Create(submission).Error
	if err != nil {
		return 0, err
	}
	return submission.ID, nil
}

func (us *submissionRepository) MarkProcessing(db database.Database, submissionID int64) error {
	tx := db.GetInstance()
	err := tx.Model(&models.Submission{}).Where("id = ?", submissionID).Updates(&models.Submission{Status: types.SubmissionStatusSentForEvaluation}).Error
	return err
}

func (us *submissionRepository) MarkEvaluated(db database.Database, submissionID int64) error {
	tx := db.GetInstance()
	err := tx.Model(&models.Submission{}).Where("id = ?", submissionID).Updates(&models.Submission{Status: types.SubmissionStatusEvaluated, CheckedAt: time.Now()}).Error
	return err
}

func (us *submissionRepository) MarkFailed(db database.Database, submissionID int64, errorMsg string) error {
	tx := db.GetInstance()
	err := tx.Model(&models.Submission{}).Where("id = ?", submissionID).Updates(&models.Submission{
		Status:        types.SubmissionStatusEvaluated,
		StatusMessage: errorMsg,
		CheckedAt:     time.Now(),
	}).Error
	return err
}

func (us *submissionRepository) GetPendingSubmissions(db database.Database, limit int) ([]models.Submission, error) {
	tx := db.GetInstance()
	submissions := []models.Submission{}
	err := tx.Model(&models.Submission{}).
		Where("status = ?", types.SubmissionStatusReceived).
		Order("submitted_at ASC").
		Limit(limit).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Preload("Result").
		Preload("Result.TestResults").
		Preload("Result.TestResults.TestCase").
		Preload("Result.TestResults.TestCase.InputFile").
		Preload("Result.TestResults.TestCase.OutputFile").
		Preload("Result.TestResults.StdoutFile").
		Preload("Result.TestResults.StderrFile").
		Preload("Result.TestResults.DiffFile").
		Preload("File").
		Find(&submissions).Error
	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *submissionRepository) GetAllForTaskByUser(
	db database.Database,
	taskID, userID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, int64, error) {
	tx := db.GetInstance()
	submissions := []models.Submission{}
	var totalCount int64

	// Get total count first
	err := tx.Model(&models.Submission{}).
		Where("submissions.task_id = ? AND submissions.user_id = ?", taskID, userID).
		Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	paginatedTx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = paginatedTx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Preload("Result").
		Where("submissions.task_id = ? AND submissions.user_id = ?", taskID, userID).
		Find(&submissions).Error

	if err != nil {
		return nil, 0, err
	}
	return submissions, totalCount, nil
}

func (sr *submissionRepository) GetLatestForTaskByUser(
	db database.Database,
	taskID, userID int64,
) (*models.Submission, error) {
	tx := db.GetInstance()
	submission := models.Submission{}
	err := tx.Model(&models.Submission{}).
		Where("task_id = ? AND user_id = ?", taskID, userID).
		Order("submitted_at DESC").
		First(&submission).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

func (sr *submissionRepository) GetBestScoreForTaskByUser(db database.Database, taskID, userID int64) (float64, error) {
	tx := db.GetInstance()
	var bestScore *float64

	// Query to get the best score (highest percentage of passed tests)
	err := tx.Model(&models.Submission{}).
		Select("MAX(CASE WHEN total_tests.count > 0 THEN (passed_tests.count * 100.0 / total_tests.count) ELSE 0 END) as best_score").
		Joins(fmt.Sprintf("LEFT JOIN %s ON submissions.id = submission_results.submission_id", database.ResolveTableName(tx, &models.SubmissionResult{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT submission_result_id, COUNT(*) as count
			FROM %s
			GROUP BY submission_result_id
		) as total_tests ON submission_results.id = total_tests.submission_result_id`, database.ResolveTableName(tx, &models.TestResult{}))).
		Joins(fmt.Sprintf(`LEFT JOIN (
			SELECT submission_result_id, COUNT(*) as count
			FROM %s
			WHERE passed = true
			GROUP BY submission_result_id
		) as passed_tests ON submission_results.id = passed_tests.submission_result_id`, database.ResolveTableName(tx, &models.TestResult{}))).
		Where("submissions.task_id = ? AND submissions.user_id = ? AND submissions.status = ?", taskID, userID, types.SubmissionStatusEvaluated).
		Scan(&bestScore).Error

	if err != nil {
		return 0, err
	}

	if bestScore == nil {
		return 0, nil
	}

	return *bestScore, nil
}

func (sr *submissionRepository) GetAttemptCountForTaskByUser(db database.Database, taskID, userID int64) (int, error) {
	tx := db.GetInstance()
	var count int64

	err := tx.Model(&models.Submission{}).
		Where("task_id = ? AND user_id = ?", taskID, userID).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (us *submissionRepository) GetAllForContest(
	db database.Database,
	contestID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, int64, error) {
	tx := db.GetInstance()
	submissions := []models.Submission{}
	var totalCount int64

	// Get total count first
	err := tx.Model(&models.Submission{}).Where("contest_id = ?", contestID).Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	paginatedTx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = paginatedTx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Preload("Result").
		Preload("Result.TestResults").
		Where("contest_id = ?", contestID).
		Find(&submissions).Error

	if err != nil {
		return nil, 0, err
	}
	return submissions, totalCount, nil
}

func (us *submissionRepository) GetAllByUserForContest(
	db database.Database,
	userID, contestID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, int64, error) {
	tx := db.GetInstance()
	submissions := []models.Submission{}
	var totalCount int64

	// Get total count first
	err := tx.Model(&models.Submission{}).
		Where("user_id = ? AND contest_id = ?", userID, contestID).
		Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	tx, err = utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = tx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Preload("Result").
		Preload("Result.TestResults").
		Where("user_id = ? AND contest_id = ?", userID, contestID).
		Find(&submissions).Error

	if err != nil {
		return nil, 0, err
	}
	return submissions, totalCount, nil
}

func (us *submissionRepository) GetAllByUserForTask(
	db database.Database,
	userID, taskID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, int64, error) {
	tx := db.GetInstance()
	submissions := []models.Submission{}
	var totalCount int64

	// Get total count first
	err := tx.Model(&models.Submission{}).
		Where("user_id = ? AND task_id = ?", userID, taskID).
		Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	tx, err = utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = tx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Preload("Result").
		Preload("Result.TestResults").
		Where("user_id = ? AND task_id = ?", userID, taskID).
		Find(&submissions).Error

	if err != nil {
		return nil, 0, err
	}
	return submissions, totalCount, nil
}

func (us *submissionRepository) GetAllByUserForContestAndTask(
	db database.Database,
	userID, contestID, taskID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, int64, error) {
	tx := db.GetInstance()
	submissions := []models.Submission{}
	var totalCount int64

	// Get total count first
	err := tx.Model(&models.Submission{}).
		Where("user_id = ? AND contest_id = ? AND task_id = ?", userID, contestID, taskID).
		Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	tx, err = utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = tx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Preload("Result").
		Preload("Result.TestResults").
		Where("user_id = ? AND contest_id = ? AND task_id = ?", userID, contestID, taskID).
		Find(&submissions).Error

	if err != nil {
		return nil, 0, err
	}
	return submissions, totalCount, nil
}

func (us *submissionRepository) GetAllByUserForTeacher(
	db database.Database,
	userID, teacherID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, int64, error) {
	tx := db.GetInstance()
	submissions := []models.Submission{}
	var totalCount int64

	// Get total count first
	err := tx.Model(&models.Submission{}).
		Joins(fmt.Sprintf("JOIN %s ON tasks.id = submissions.task_id", database.ResolveTableName(tx, &models.Task{}))).
		Joins(fmt.Sprintf("LEFT JOIN %s ON contests.id = submissions.contest_id", database.ResolveTableName(tx, &models.Contest{}))).
		Where("submissions.user_id = ? AND (tasks.created_by = ? OR (submissions.contest_id IS NOT NULL AND contests.created_by = ?))", userID, teacherID, teacherID).
		Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	tx, err = utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = tx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Preload("Result").
		Preload("Result.TestResults").
		Joins(fmt.Sprintf("JOIN %s ON tasks.id = submissions.task_id", database.ResolveTableName(tx, &models.Task{}))).
		Joins(fmt.Sprintf("LEFT JOIN %s ON contests.id = submissions.contest_id", database.ResolveTableName(tx, &models.Contest{}))).
		Where("submissions.user_id = ? AND (tasks.created_by = ? OR (submissions.contest_id IS NOT NULL AND contests.created_by = ?))", userID, teacherID, teacherID).
		Find(&submissions).Error

	if err != nil {
		return nil, 0, err
	}
	return submissions, totalCount, nil
}

func (us *submissionRepository) GetAllByUserForTaskByTeacher(
	db database.Database,
	userID, taskID, teacherID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, int64, error) {
	tx := db.GetInstance()
	submissions := []models.Submission{}
	var totalCount int64

	// Get total count first
	err := tx.Model(&models.Submission{}).
		Joins(fmt.Sprintf("JOIN %s ON tasks.id = submissions.task_id", database.ResolveTableName(tx, &models.Task{}))).
		Where("submissions.user_id = ? AND submissions.task_id = ? AND tasks.created_by = ?", userID, taskID, teacherID).
		Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	tx, err = utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = tx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Preload("Result").
		Preload("Result.TestResults").
		Joins(fmt.Sprintf("JOIN %s ON tasks.id = submissions.task_id", database.ResolveTableName(tx, &models.Task{}))).
		Where("submissions.user_id = ? AND submissions.task_id = ? AND tasks.created_by = ?", userID, taskID, teacherID).
		Find(&submissions).Error

	if err != nil {
		return nil, 0, err
	}
	return submissions, totalCount, nil
}

func (us *submissionRepository) GetAllByUserForContestByTeacher(
	db database.Database,
	userID, contestID, teacherID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, int64, error) {
	tx := db.GetInstance()
	submissions := []models.Submission{}
	var totalCount int64

	// Get total count first
	err := tx.Model(&models.Submission{}).
		Joins(fmt.Sprintf("JOIN %s ON tasks.id = submissions.task_id", database.ResolveTableName(tx, &models.Task{}))).
		Joins(fmt.Sprintf("JOIN %s ON contests.id = submissions.contest_id", database.ResolveTableName(tx, &models.Contest{}))).
		Where("submissions.user_id = ? AND submissions.contest_id = ? AND (tasks.created_by = ? OR contests.created_by = ?)", userID, contestID, teacherID, teacherID).
		Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	tx, err = utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = tx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Preload("Result").
		Preload("Result.TestResults").
		Joins(fmt.Sprintf("JOIN %s ON tasks.id = submissions.task_id", database.ResolveTableName(tx, &models.Task{}))).
		Joins(fmt.Sprintf("JOIN %s ON contests.id = submissions.contest_id", database.ResolveTableName(tx, &models.Contest{}))).
		Where("submissions.user_id = ? AND submissions.contest_id = ? AND (tasks.created_by = ? OR contests.created_by = ?)", userID, contestID, teacherID, teacherID).
		Find(&submissions).Error

	if err != nil {
		return nil, 0, err
	}
	return submissions, totalCount, nil
}

func (us *submissionRepository) GetAllByUserForContestAndTaskByTeacher(
	db database.Database,
	userID, contestID, taskID, teacherID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, int64, error) {
	tx := db.GetInstance()
	submissions := []models.Submission{}
	var totalCount int64

	// Get total count first
	err := tx.Model(&models.Submission{}).
		Joins(fmt.Sprintf("JOIN %s ON tasks.id = submissions.task_id", database.ResolveTableName(tx, &models.Task{}))).
		Joins(fmt.Sprintf("JOIN %s ON contests.id = submissions.contest_id", database.ResolveTableName(tx, &models.Contest{}))).
		Where("submissions.user_id = ? AND submissions.contest_id = ? AND submissions.task_id = ? AND (tasks.created_by = ? OR contests.created_by = ?)", userID, contestID, taskID, teacherID, teacherID).
		Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	tx, err = utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	err = tx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Preload("Result").
		Preload("Result.TestResults").
		Joins(fmt.Sprintf("JOIN %s ON tasks.id = submissions.task_id", database.ResolveTableName(tx, &models.Task{}))).
		Joins(fmt.Sprintf("JOIN %s ON contests.id = submissions.contest_id", database.ResolveTableName(tx, &models.Contest{}))).
		Where("submissions.user_id = ? AND submissions.contest_id = ? AND submissions.task_id = ? AND (tasks.created_by = ? OR contests.created_by = ?)", userID, contestID, taskID, teacherID, teacherID).
		Find(&submissions).Error

	if err != nil {
		return nil, 0, err
	}
	return submissions, totalCount, nil
}

func (us *submissionRepository) GetTaskStatsForContest(db database.Database, contestID int64) ([]models.ContestTaskStatsModel, error) {
	tx := db.GetInstance()
	var results []models.ContestTaskStatsModel

	query := `
		SELECT
			t.id as task_id,
			t.title as task_title,
			COUNT(DISTINCT cp.user_id) as total_participants,
			COUNT(DISTINCT s.user_id) as submitted_count,
			COUNT(DISTINCT CASE
				WHEN sr.code = 1 THEN s.user_id
			END) as fully_solved_count,
			COUNT(DISTINCT CASE
				WHEN sr.code != 1 AND sr.code > 0 THEN s.user_id
			END) as partially_solved_count,
			COALESCE(AVG(CASE
				WHEN total_tests.count > 0
				THEN (passed_tests.count * 100.0 / total_tests.count)
				ELSE 0
			END), 0) as average_score
		FROM ` + database.ResolveTableName(tx, &models.Task{}) + ` t
		INNER JOIN ` + database.ResolveTableName(tx, &models.ContestTask{}) + ` ct ON ct.task_id = t.id
		LEFT JOIN ` + database.ResolveTableName(tx, &models.ContestParticipant{}) + ` cp ON cp.contest_id = ct.contest_id
		LEFT JOIN ` + database.ResolveTableName(tx, &models.Submission{}) + ` s ON s.task_id = t.id
			AND s.contest_id = ct.contest_id
			AND s.user_id = cp.user_id
			AND s.status = 'evaluated'
		LEFT JOIN ` + database.ResolveTableName(tx, &models.SubmissionResult{}) + ` sr ON sr.submission_id = s.id
		LEFT JOIN (
			SELECT submission_result_id, COUNT(*) as count
			FROM ` + database.ResolveTableName(tx, &models.TestResult{}) + `
			GROUP BY submission_result_id
		) as total_tests ON total_tests.submission_result_id = sr.id
		LEFT JOIN (
			SELECT submission_result_id, COUNT(*) as count
			FROM ` + database.ResolveTableName(tx, &models.TestResult{}) + `
			WHERE passed = true
			GROUP BY submission_result_id
		) as passed_tests ON passed_tests.submission_result_id = sr.id
		WHERE ct.contest_id = ?
		GROUP BY t.id, t.title
		ORDER BY t.id
	`

	err := tx.Raw(query, contestID).Scan(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (us *submissionRepository) GetUserStatsForContestTask(db database.Database, contestID, taskID int64) ([]models.TaskUserStatsModel, error) {
	tx := db.GetInstance()
	var results []models.TaskUserStatsModel

	query := `
		WITH scored_submissions AS (
			SELECT
				s.id,
				s.user_id,
				CASE
					WHEN tt.count > 0 THEN (pt.count * 100.0 / tt.count)
					ELSE 0
				END AS score
			FROM ` + database.ResolveTableName(tx, &models.Submission{}) + ` s
			LEFT JOIN ` + database.ResolveTableName(tx, &models.SubmissionResult{}) + ` sr ON sr.submission_id = s.id
			LEFT JOIN (
				SELECT submission_result_id, COUNT(*) as count
				FROM ` + database.ResolveTableName(tx, &models.TestResult{}) + `
				GROUP BY submission_result_id
			) AS tt ON tt.submission_result_id = sr.id
			LEFT JOIN (
				SELECT submission_result_id, COUNT(*) as count
				FROM ` + database.ResolveTableName(tx, &models.TestResult{}) + `
				WHERE passed = true
				GROUP BY submission_result_id
			) AS pt ON pt.submission_result_id = sr.id
			WHERE s.task_id = ?
			  AND s.contest_id = ?
			  AND s.status = 'evaluated'
		),
		best_submissions AS (
			SELECT DISTINCT ON (user_id)
				user_id,
				id AS best_submission_id,
				score AS best_score
			FROM scored_submissions
			ORDER BY user_id, score DESC, id DESC
		)
		SELECT
			u.id as user_id,
			u.username as user_username,
			u.name as user_name,
			u.surname as user_surname,
			COUNT(DISTINCT s.id) as submission_count,
			COALESCE(b.best_score, 0) as best_score,
			COALESCE(b.best_submission_id, 0) as best_submission_id
		FROM ` + database.ResolveTableName(tx, &models.User{}) + ` u
		INNER JOIN ` + database.ResolveTableName(tx, &models.ContestParticipant{}) + ` cp ON cp.user_id = u.id
		LEFT JOIN ` + database.ResolveTableName(tx, &models.Submission{}) + ` s ON s.user_id = u.id
			AND s.task_id = ?
			AND s.contest_id = ?
			AND s.status = 'evaluated'
		LEFT JOIN best_submissions b ON b.user_id = u.id
		WHERE cp.contest_id = ?
		GROUP BY u.id, u.username, u.name, u.surname, b.best_score, b.best_submission_id
		ORDER BY best_score DESC, u.username
	`
	if err := tx.Raw(query, taskID, contestID, taskID, contestID, contestID).Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

func (us *submissionRepository) GetUserStatsForContest(db database.Database, contestID int64, userID *int64) ([]models.UserContestStatsFull, error) {
	tx := db.GetInstance()
	userFilter := ""
	summaryArgs := []interface{}{contestID}
	if userID != nil {
		userFilter = " AND u.id = ?"
		summaryArgs = append(summaryArgs, *userID)
	}

	// First query: per-user aggregated counts (no JSON)
	summaryQuery := `
		SELECT
			u.id AS user_id,
			u.username AS user_username,
			u.name AS user_name,
			u.surname AS user_surname,
			COUNT(DISTINCT CASE WHEN s.id IS NOT NULL THEN t.id END) AS tasks_attempted,
			COUNT(DISTINCT CASE WHEN sr.code = 1 THEN t.id END) AS tasks_solved,
			COUNT(DISTINCT CASE WHEN sr.code != 1 AND sr.code > 0 AND NOT EXISTS (
				SELECT 1 FROM ` + database.ResolveTableName(tx, &models.SubmissionResult{}) + ` sr2
				JOIN ` + database.ResolveTableName(tx, &models.Submission{}) + ` s2 ON sr2.submission_id = s2.id
				WHERE s2.user_id = u.id AND s2.task_id = t.id AND s2.contest_id = cp.contest_id AND sr2.code = 1
			) THEN t.id END) AS tasks_partially_solved
		FROM ` + database.ResolveTableName(tx, &models.User{}) + ` u
		INNER JOIN ` + database.ResolveTableName(tx, &models.ContestParticipant{}) + ` cp ON cp.user_id = u.id
		INNER JOIN ` + database.ResolveTableName(tx, &models.ContestTask{}) + ` ct ON ct.contest_id = cp.contest_id
		INNER JOIN ` + database.ResolveTableName(tx, &models.Task{}) + ` t ON t.id = ct.task_id
		LEFT JOIN ` + database.ResolveTableName(tx, &models.Submission{}) + ` s ON s.user_id = u.id
			AND s.task_id = t.id
			AND s.contest_id = cp.contest_id
			AND s.status = 'evaluated'
		LEFT JOIN ` + database.ResolveTableName(tx, &models.SubmissionResult{}) + ` sr ON sr.submission_id = s.id
		WHERE cp.contest_id = ?` + userFilter + `
		GROUP BY u.id, u.username, u.name, u.surname
		ORDER BY tasks_solved DESC, tasks_attempted DESC, u.username
	`

	var summaryRows []models.UserContestSummaryRow
	if err := tx.Raw(summaryQuery, summaryArgs...).Scan(&summaryRows).Error; err != nil {
		return nil, err
	}
	if len(summaryRows) == 0 {
		return []models.UserContestStatsFull{}, nil
	}

	// Second query: per-user per-task performance rows
	// Placeholders: best score contest_id, solved flag contest_id, attempts contest_id
	performanceArgs := []interface{}{contestID, contestID, contestID}
	// Same user filter applied if provided
	if userID != nil {
		performanceArgs = append(performanceArgs, *userID)
	}

	performanceQuery := `
		SELECT
			u.id AS user_id,
			t.id AS task_id,
			t.title AS task_title,
			COALESCE((
				SELECT MAX(
					CASE WHEN total.count > 0
						THEN (passed.count * 100.0 / total.count)
						ELSE 0
					END
				)
				FROM ` + database.ResolveTableName(tx, &models.Submission{}) + ` s2
				LEFT JOIN ` + database.ResolveTableName(tx, &models.SubmissionResult{}) + ` sr2 ON sr2.submission_id = s2.id
				LEFT JOIN (
					SELECT submission_result_id, COUNT(*) AS count
					FROM ` + database.ResolveTableName(tx, &models.TestResult{}) + `
					GROUP BY submission_result_id
				) total ON total.submission_result_id = sr2.id
				LEFT JOIN (
					SELECT submission_result_id, COUNT(*) AS count
					FROM ` + database.ResolveTableName(tx, &models.TestResult{}) + `
					WHERE passed = true
					GROUP BY submission_result_id
				) passed ON passed.submission_result_id = sr2.id
				WHERE s2.user_id = u.id
				  AND s2.task_id = t.id
				  AND s2.contest_id = ?
				  AND s2.status = 'evaluated'
			), 0) AS best_score,
			COALESCE((
				SELECT bool_or(sr3.code = 1)
				FROM ` + database.ResolveTableName(tx, &models.Submission{}) + ` s3
				LEFT JOIN ` + database.ResolveTableName(tx, &models.SubmissionResult{}) + ` sr3 ON sr3.submission_id = s3.id
				WHERE s3.user_id = u.id
				  AND s3.task_id = t.id
				  AND s3.contest_id = ?
				  AND s3.status = 'evaluated'
			), false) AS is_solved,
			COALESCE((
				SELECT COUNT(*)
				FROM ` + database.ResolveTableName(tx, &models.Submission{}) + ` s4
				WHERE s4.user_id = u.id
				  AND s4.task_id = t.id
				  AND s4.contest_id = ?
			), 0) AS attempt_count
		FROM ` + database.ResolveTableName(tx, &models.User{}) + ` u
		INNER JOIN ` + database.ResolveTableName(tx, &models.ContestParticipant{}) + ` cp ON cp.user_id = u.id
		INNER JOIN ` + database.ResolveTableName(tx, &models.ContestTask{}) + ` ct ON ct.contest_id = cp.contest_id
		INNER JOIN ` + database.ResolveTableName(tx, &models.Task{}) + ` t ON t.id = ct.task_id
		WHERE cp.contest_id = ?` + userFilter + `
		ORDER BY u.id, t.id
	`

	// Append contestID again for the WHERE cp.contest_id = ? part
	performanceArgs = append(performanceArgs, contestID)

	var performanceRows []models.UserTaskPerformanceRow
	if err := tx.Raw(performanceQuery, performanceArgs...).Scan(&performanceRows).Error; err != nil {
		return nil, err
	}

	// Group task rows by user
	perfMap := make(map[int64][]models.UserTaskPerformanceModel)
	for _, pr := range performanceRows {
		perfMap[pr.UserID] = append(perfMap[pr.UserID], models.UserTaskPerformanceModel{
			TaskID:       pr.TaskID,
			TaskTitle:    pr.TaskTitle,
			BestScore:    pr.BestScore,
			AttemptCount: int(pr.AttemptCount),
			IsSolved:     pr.IsSolved,
		})
	}

	// Merge summaries with breakdown
	result := make([]models.UserContestStatsFull, 0, len(summaryRows))
	for _, s := range summaryRows {
		taskBreakdown := perfMap[s.UserID]
		entry := models.UserContestStatsFull{
			User: models.User{
				ID:       s.UserID,
				Username: s.UserUsername,
				Name:     s.UserName,
				Surname:  s.UserSurname,
			},
			TasksAttempted:       s.TasksAttempted,
			TasksSolved:          s.TasksSolved,
			TasksPartiallySolved: s.TasksPartiallySolved,
			TaskBreakdown:        taskBreakdown,
		}
		result = append(result, entry)
	}

	return result, nil
}

func NewSubmissionRepository() SubmissionRepository {
	return &submissionRepository{}
}
