package repository

import (
	"fmt"
	"time"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/utils"
	"gorm.io/gorm"
)

type SubmissionRepository interface {
	// Create creates a new submission and returns the submission ID.
	Create(tx *gorm.DB, submission *models.Submission) (int64, error)
	// GetAll returns all submissions. The submissions are paginated.
	GetAll(tx *gorm.DB, limit, offset int, sort string) ([]models.Submission, error)
	// GetAllByUser returns all submissions by a user. The submissions are paginated.
	GetAllByUser(tx *gorm.DB, userID int64, limit, offset int, sort string) ([]models.Submission, error)
	// GetAllForGroup returns all submissions for a group. The submissions are paginated.
	GetAllForGroup(tx *gorm.DB, groupID int64, limit, offset int, sort string) ([]models.Submission, error)
	// GetAllForTask returns all submissions for a task. The submissions are paginated.
	GetAllForTask(tx *gorm.DB, taskID int64, limit, offset int, sort string) ([]models.Submission, error)
	// GetAllForTaskByUser returns all submissions for a task by a user. The submissions are paginated.
	GetAllForTaskByUser(tx *gorm.DB, taskID, userID int64, limit, offset int, sort string) ([]models.Submission, error)
	// GetAllForContest returns all submissions for a contest. The submissions are paginated.
	GetAllForContest(tx *gorm.DB, contestID int64, limit, offset int, sort string) ([]models.Submission, error)
	// GetAllByUserForContest returns all submissions by a user for a specific contest. The submissions are paginated.
	GetAllByUserForContest(tx *gorm.DB, userID, contestID int64, limit, offset int, sort string) ([]models.Submission, error)
	// GetAllByUserForTask returns all submissions by a user for a specific task. The submissions are paginated.
	GetAllByUserForTask(tx *gorm.DB, userID, taskID int64, limit, offset int, sort string) ([]models.Submission, error)
	// GetAllByUserForContestAndTask returns all submissions by a user for a specific contest and task. The submissions are paginated.
	GetAllByUserForContestAndTask(tx *gorm.DB, userID, contestID, taskID int64, limit, offset int, sort string) ([]models.Submission, error)
	// GetAllForTeacher returns all submissions for a teacher, this includes submissions for tasks created by this teacher.
	// The submissions are paginated.
	GetAllForTeacher(tx *gorm.DB, currentUserID int64, limit, offset int, sort string) ([]models.Submission, error)
	// GetAllByUserForTeacher returns all submissions by a specific user, filtered to only include submissions
	// for tasks created by the teacher. The submissions are paginated.
	GetAllByUserForTeacher(tx *gorm.DB, userID, teacherID int64, limit, offset int, sort string) ([]models.Submission, error)
	// GetAllByUserForTaskByTeacher returns all submissions by a user for a specific task,
	// filtered to only include submissions where the teacher created the task. The submissions are paginated.
	GetAllByUserForTaskByTeacher(tx *gorm.DB, userID, taskID, teacherID int64, limit, offset int, sort string) ([]models.Submission, error)
	// GetAllByUserForContestByTeacher returns all submissions by a user for a specific contest,
	// filtered to only include submissions where the teacher created the contest or the task. The submissions are paginated.
	GetAllByUserForContestByTeacher(tx *gorm.DB, userID, contestID, teacherID int64, limit, offset int, sort string) ([]models.Submission, error)
	// GetAllByUserForContestAndTaskByTeacher returns all submissions by a user for a specific contest and task,
	// filtered to only include submissions where the teacher created the contest or the task. The submissions are paginated.
	GetAllByUserForContestAndTaskByTeacher(tx *gorm.DB, userID, contestID, taskID, teacherID int64, limit, offset int, sort string) ([]models.Submission, error)
	// GetLatestSubmissionForTaskByUser returns the latest submission for a task by a user.
	GetLatestForTaskByUser(tx *gorm.DB, taskID, userID int64) (*models.Submission, error)
	// Get returns a submission by its ID.
	Get(tx *gorm.DB, submissionID int64) (*models.Submission, error)
	// GetBestScoreForTaskByUser returns the best score (percentage of passed tests) for a task by a user.
	GetBestScoreForTaskByUser(tx *gorm.DB, taskID, userID int64) (*float64, error)
	// GetAttemptCountForTaskByUser returns the number of submission attempts for a task by a user.
	GetAttemptCountForTaskByUser(tx *gorm.DB, taskID, userID int64) (int, error)
	// MarkEvaluated marks a submission as evaluated.
	MarkEvaluated(tx *gorm.DB, submissionID int64) error
	// MarkFailed marks a submission as failed.
	MarkFailed(db *gorm.DB, submissionID int64, errorMsg string) error
	// MarkProcessing marks a submission as processing.
	MarkProcessing(tx *gorm.DB, submissionID int64) error
	// GetPendingSubmissions returns submissions that are in "received" status (not yet sent for evaluation).
	GetPendingSubmissions(tx *gorm.DB, limit int) ([]models.Submission, error)
}

type submissionRepository struct{}

func (us *submissionRepository) GetAll(tx *gorm.DB, limit, offset int, sort string) ([]models.Submission, error) {
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
		Preload("Result.TestResults").
		Find(&submissions).Error
	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *submissionRepository) GetAllForTeacher(
	tx *gorm.DB,
	userID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, error) {
	submissions := []models.Submission{}

	tx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return submissions, nil
}

func (us *submissionRepository) Get(tx *gorm.DB, submissionID int64) (*models.Submission, error) {
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
	tx *gorm.DB,
	userID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, error) {
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
		Preload("Result.TestResults").
		Where("user_id = ?", userID).Find(&submissions).Error
	if err != nil {
		return nil, err
	}

	return submissions, nil
}

func (us *submissionRepository) GetAllForGroup(
	tx *gorm.DB,
	groupID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, error) {
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
		Joins(fmt.Sprintf("JOIN %s ON users.id = submissions.user_id", database.ResolveTableName(tx, &models.User{}))).
		Joins(fmt.Sprintf("JOIN %s ON user_group.user_id = users.id", database.ResolveTableName(tx, &models.UserGroup{}))).
		Joins(fmt.Sprintf("JOIN %s ON groups.id = user_group.group_id", database.ResolveTableName(tx, &models.Group{}))).
		Where("groups.id = ?", groupID).
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *submissionRepository) GetAllForGroupTeacher(
	tx *gorm.DB,
	groupID, userID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, error) {
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
		Joins(fmt.Sprintf("JOIN %s ON task_group.task_id = tasks.id", database.ResolveTableName(tx, &models.TaskGroup{}))).
		Joins(fmt.Sprintf("JOIN %s ON groups.id = task_group.group_id", database.ResolveTableName(tx, &models.Group{}))).
		Where("groups.id = ? AND tasks.created_by_id = ?", groupID, userID).
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *submissionRepository) GetAllForTask(
	tx *gorm.DB,
	taskID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, error) {
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
		Where("tasks.id = ?", taskID).
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *submissionRepository) GetAllForTaskTeacher(
	tx *gorm.DB,
	taskID, userID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, error) {
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
	tx *gorm.DB,
	taskID, studentID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, error) {
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

func (us *submissionRepository) Create(tx *gorm.DB, submission *models.Submission) (int64, error) {
	err := tx.Create(submission).Error
	if err != nil {
		return 0, err
	}
	return submission.ID, nil
}

func (us *submissionRepository) MarkProcessing(tx *gorm.DB, submissionID int64) error {
	err := tx.Model(&models.Submission{}).Where("id = ?", submissionID).Updates(&models.Submission{Status: types.SubmissionStatusSentForEvaluation}).Error
	return err
}

func (us *submissionRepository) MarkEvaluated(tx *gorm.DB, submissionID int64) error {
	err := tx.Model(&models.Submission{}).Where("id = ?", submissionID).Updates(&models.Submission{Status: types.SubmissionStatusEvaluated, CheckedAt: time.Now()}).Error
	return err
}

func (us *submissionRepository) MarkFailed(tx *gorm.DB, submissionID int64, errorMsg string) error {
	err := tx.Model(&models.Submission{}).Where("id = ?", submissionID).Updates(&models.Submission{
		Status:        types.SubmissionStatusEvaluated,
		StatusMessage: errorMsg,
		CheckedAt:     time.Now(),
	}).Error
	return err
}

func (us *submissionRepository) GetPendingSubmissions(tx *gorm.DB, limit int) ([]models.Submission, error) {
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
	tx *gorm.DB,
	taskID, userID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, error) {
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
		Where("submissions.task_id = ? AND submissions.user_id = ?", taskID, userID).
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (sr *submissionRepository) GetLatestForTaskByUser(
	tx *gorm.DB,
	taskID, userID int64,
) (*models.Submission, error) {
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

func (sr *submissionRepository) GetBestScoreForTaskByUser(tx *gorm.DB, taskID, userID int64) (*float64, error) {
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
		return nil, err
	}

	return bestScore, nil
}

func (sr *submissionRepository) GetAttemptCountForTaskByUser(tx *gorm.DB, taskID, userID int64) (int, error) {
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
	tx *gorm.DB,
	contestID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, error) {
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
		Preload("Result.TestResults").
		Where("contest_id = ?", contestID).
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *submissionRepository) GetAllByUserForContest(
	tx *gorm.DB,
	userID, contestID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, error) {
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
		Preload("Result.TestResults").
		Where("user_id = ? AND contest_id = ?", userID, contestID).
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *submissionRepository) GetAllByUserForTask(
	tx *gorm.DB,
	userID, taskID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, error) {
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
		Preload("Result.TestResults").
		Where("user_id = ? AND task_id = ?", userID, taskID).
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *submissionRepository) GetAllByUserForContestAndTask(
	tx *gorm.DB,
	userID, contestID, taskID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, error) {
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
		Preload("Result.TestResults").
		Where("user_id = ? AND contest_id = ? AND task_id = ?", userID, contestID, taskID).
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *submissionRepository) GetAllByUserForTeacher(
	tx *gorm.DB,
	userID, teacherID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, error) {
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
		Preload("Result.TestResults").
		Joins(fmt.Sprintf("JOIN %s ON tasks.id = submissions.task_id", database.ResolveTableName(tx, &models.Task{}))).
		Joins(fmt.Sprintf("LEFT JOIN %s ON contests.id = submissions.contest_id", database.ResolveTableName(tx, &models.Contest{}))).
		Where("submissions.user_id = ? AND (tasks.created_by = ? OR (submissions.contest_id IS NOT NULL AND contests.created_by = ?))", userID, teacherID, teacherID).
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *submissionRepository) GetAllByUserForTaskByTeacher(
	tx *gorm.DB,
	userID, taskID, teacherID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, error) {
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
		Preload("Result.TestResults").
		Joins(fmt.Sprintf("JOIN %s ON tasks.id = submissions.task_id", database.ResolveTableName(tx, &models.Task{}))).
		Where("submissions.user_id = ? AND submissions.task_id = ? AND tasks.created_by = ?", userID, taskID, teacherID).
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *submissionRepository) GetAllByUserForContestByTeacher(
	tx *gorm.DB,
	userID, contestID, teacherID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, error) {
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
		Preload("Result.TestResults").
		Joins(fmt.Sprintf("JOIN %s ON tasks.id = submissions.task_id", database.ResolveTableName(tx, &models.Task{}))).
		Joins(fmt.Sprintf("JOIN %s ON contests.id = submissions.contest_id", database.ResolveTableName(tx, &models.Contest{}))).
		Where("submissions.user_id = ? AND submissions.contest_id = ? AND (tasks.created_by = ? OR contests.created_by = ?)", userID, contestID, teacherID, teacherID).
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *submissionRepository) GetAllByUserForContestAndTaskByTeacher(
	tx *gorm.DB,
	userID, contestID, taskID, teacherID int64,
	limit, offset int,
	sort string,
) ([]models.Submission, error) {
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
		Preload("Result.TestResults").
		Joins(fmt.Sprintf("JOIN %s ON tasks.id = submissions.task_id", database.ResolveTableName(tx, &models.Task{}))).
		Joins(fmt.Sprintf("JOIN %s ON contests.id = submissions.contest_id", database.ResolveTableName(tx, &models.Contest{}))).
		Where("submissions.user_id = ? AND submissions.contest_id = ? AND submissions.task_id = ? AND (tasks.created_by = ? OR contests.created_by = ?)", userID, contestID, taskID, teacherID, teacherID).
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func NewSubmissionRepository() SubmissionRepository {
	return &submissionRepository{}
}
