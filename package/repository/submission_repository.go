package repository

import (
	"time"

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
	// GetAllForTeacher returns all submissions for a teacher, this includes submissions for tasks created by this teacher.
	// The submissions are paginated.
	GetAllForTeacher(tx *gorm.DB, currentUserID int64, limit, offset int, sort string) ([]models.Submission, error)
	// GetLatestSubmissionForTaskByUser returns the latest submission for a task by a user.
	GetLatestForTaskByUser(tx *gorm.DB, taskID, userID int64) (*models.Submission, error) // Get returns a submission by its ID.
	// Get returns a submission by its ID.
	Get(tx *gorm.DB, submissionID int64) (*models.Submission, error)
	// GetBestScoreForTaskByUser returns the best score (percentage of passed tests) for a task by a user.
	GetBestScoreForTaskByUser(tx *gorm.DB, taskID, userID int64) (*float64, error)
	// GetAttemptCountForTaskByUser returns the number of submission attempts for a task by a user.
	GetAttemptCountForTaskByUser(tx *gorm.DB, taskID, userID int64) (int, error)
	// MarkComplete marks a submission as completed.
	MarkComplete(tx *gorm.DB, submissionID int64) error
	// MarkFailed marks a submission as failed.
	MarkFailed(db *gorm.DB, submissionID int64, errorMsg string) error
	// MarkProcessing marks a submission as processing.
	MarkProcessing(tx *gorm.DB, submissionID int64) error
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
		Joins("JOIN tasks ON tasks.id = submissions.task_id").Where(
		"tasks.created_by = ?",
		userID,
	).Find(&submissions).Error
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
		Joins("JOIN users ON users.id = submissions.user_id").
		Joins("JOIN user_group ON user_group.user_id = users.id").
		Joins("JOIN groups ON groups.id = user_group.group_id").
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
		Joins("JOIN tasks ON tasks.id = submissions.task_id").
		Joins("JOIN task_group ON task_group.task_id = tasks.id").
		Joins("JOIN groups ON groups.id = task_group.group_id").
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
		Joins("JOIN tasks ON tasks.id = submissions.task_id").
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
		Joins("JOIN tasks ON tasks.id = submissions.task_id").
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
		Joins("JOIN tasks ON tasks.id = submissions.task_id").
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
	err := tx.Model(&models.Submission{}).Where("id = ?", submissionID).Update("status", "processing").Error
	return err
}

func (us *submissionRepository) MarkComplete(tx *gorm.DB, submissionID int64) error {
	err := tx.Model(&models.Submission{}).Where("id = ?", submissionID).Updates(map[string]any{
		"status":     "completed",
		"checked_at": time.Now(),
	}).Error
	return err
}

func (us *submissionRepository) MarkFailed(tx *gorm.DB, submissionID int64, errorMsg string) error {
	err := tx.Model(&models.Submission{}).Where("id = ?", submissionID).Updates(map[string]any{
		"status":         types.SubmissionStatusEvaluated,
		"status_message": errorMsg,
	}).Error
	return err
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
		Joins("LEFT JOIN submission_results ON submissions.id = submission_results.submission_id").
		Joins(`LEFT JOIN (
			SELECT submission_result_id, COUNT(*) as count
			FROM test_results
			GROUP BY submission_result_id
		) as total_tests ON submission_results.id = total_tests.submission_result_id`).
		Joins(`LEFT JOIN (
			SELECT submission_result_id, COUNT(*) as count
			FROM test_results
			WHERE passed = true
			GROUP BY submission_result_id
		) as passed_tests ON submission_results.id = passed_tests.submission_result_id`).
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

func NewSubmissionRepository(db *gorm.DB) (SubmissionRepository, error) {
	if !db.Migrator().HasTable(&models.Submission{}) {
		err := db.Migrator().CreateTable(&models.Submission{})
		if err != nil {
			return nil, err
		}
	}
	return &submissionRepository{}, nil
}
