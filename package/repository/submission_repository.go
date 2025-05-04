package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
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
	// Get returns a submission by its ID.
	Get(tx *gorm.DB, submissionID int64) (*models.Submission, error)
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
		Preload("Result.TestResult").
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
		Preload("Result.TestResult").
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
	err := tx.Model(&models.Submission{}).Where("id = ?", submissionID).Update("status", "completed").Error
	return err
}

func (us *submissionRepository) MarkFailed(tx *gorm.DB, submissionID int64, errorMsg string) error {
	err := tx.Model(&models.Submission{}).Where("id = ?", submissionID).Updates(map[string]any{
		"status":         "failed",
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

func NewSubmissionRepository(db *gorm.DB) (SubmissionRepository, error) {
	if !db.Migrator().HasTable(&models.Submission{}) {
		err := db.Migrator().CreateTable(&models.Submission{})
		if err != nil {
			return nil, err
		}
	}
	return &submissionRepository{}, nil
}
