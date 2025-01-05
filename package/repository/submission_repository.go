package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/utils"
	"gorm.io/gorm"
)

type SubmissionRepository interface {
	GetSubmission(tx *gorm.DB, submissionId int64) (*models.Submission, error)
	CreateSubmission(tx *gorm.DB, submission models.Submission) (int64, error)
	MarkSubmissionProcessing(tx *gorm.DB, submissionId int64) error
	MarkSubmissionComplete(tx *gorm.DB, submissionId int64) error
	MarkSubmissionFailed(db *gorm.DB, submissionId int64, errorMsg string) error
	GetAll(tx *gorm.DB, queryParams map[string][]string) ([]models.Submission, error)
	GetAllForStudent(tx *gorm.DB, currentUserId int64, queryParams map[string][]string) ([]models.Submission, error)
	GetAllForTeacher(tx *gorm.DB, currentUserId int64, queryParams map[string][]string) ([]models.Submission, error)
	GetAllByUserId(tx *gorm.DB, userId int64, queryParams map[string][]string) ([]models.Submission, error)
	GetAllForGroup(tx *gorm.DB, groupId int64, queryParams map[string][]string) ([]models.Submission, error)
	GetAllForGroupTeacher(tx *gorm.DB, groupId, teacherId int64, queryParams map[string][]string) ([]models.Submission, error)
	GetAllForTask(tx *gorm.DB, taskId int64, queryParams map[string][]string) ([]models.Submission, error)
	GetAllForTaskTeacher(tx *gorm.DB, taskId, teacherId int64, queryParams map[string][]string) ([]models.Submission, error)
	GetAllForTaskStudent(tx *gorm.DB, taskId, studentId int64, queryParams map[string][]string) ([]models.Submission, error)
}

type SubmissionRepositoryImpl struct{}

func (us *SubmissionRepositoryImpl) GetAll(tx *gorm.DB, queryParams map[string][]string) ([]models.Submission, error) {
	submissions := []models.Submission{}

	tx = utils.ApplyQueryParams(tx, queryParams)

	err := tx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Find(&submissions).Error
	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *SubmissionRepositoryImpl) GetAllForStudent(tx *gorm.DB, currentUserId int64, queryParams map[string][]string) ([]models.Submission, error) {
	submissions := []models.Submission{}

	tx = utils.ApplyQueryParams(tx, queryParams)

	err := tx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Where("user_id = ?", currentUserId).Find(&submissions).Error
	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *SubmissionRepositoryImpl) GetAllForTeacher(tx *gorm.DB, currentUserId int64, queryParams map[string][]string) ([]models.Submission, error) {
	submissions := []models.Submission{}

	tx = utils.ApplyQueryParams(tx, queryParams)

	err := tx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Joins("JOIN tasks ON tasks.id = submissions.task_id").Where("tasks.created_by = ?", currentUserId).Find(&submissions).Error
	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *SubmissionRepositoryImpl) GetSubmission(tx *gorm.DB, submissionId int64) (*models.Submission, error) {
	var submission models.Submission
	err := tx.Where("id = ?", submissionId).
		Preload("Language").
		Preload("Task").
		Preload("User").
		First(&submission).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

func (us *SubmissionRepositoryImpl) GetAllByUserId(tx *gorm.DB, userId int64, queryParams map[string][]string) ([]models.Submission, error) {
	submissions := []models.Submission{}

	tx = utils.ApplyQueryParams(tx, queryParams)

	err := tx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Where("user_id = ?", userId).Find(&submissions).Error
	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *SubmissionRepositoryImpl) GetAllForGroup(tx *gorm.DB, groupId int64, queryParams map[string][]string) ([]models.Submission, error) {
	submissions := []models.Submission{}

	tx = utils.ApplyQueryParams(tx, queryParams)

	err := tx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Joins("JOIN users ON users.id = submissions.user_id").
		Joins("JOIN user_group ON user_group.user_id = users.id").
		Joins("JOIN groups ON groups.id = user_group.group_id").
		Where("groups.id = ?", groupId).
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *SubmissionRepositoryImpl) GetAllForGroupTeacher(tx *gorm.DB, groupId, teacherId int64, queryParams map[string][]string) ([]models.Submission, error) {
	submissions := []models.Submission{}

	tx = utils.ApplyQueryParams(tx, queryParams)

	err := tx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Joins("JOIN tasks ON tasks.id = submissions.task_id").
		Joins("JOIN task_group ON task_group.task_id = tasks.id").
		Joins("JOIN groups ON groups.id = task_group.group_id").
		Where("groups.id = ? AND tasks.created_by_id = ?", groupId, teacherId).
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *SubmissionRepositoryImpl) GetAllForTask(tx *gorm.DB, taskId int64, queryParams map[string][]string) ([]models.Submission, error) {
	submissions := []models.Submission{}

	tx = utils.ApplyQueryParams(tx, queryParams)

	err := tx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Joins("JOIN tasks ON tasks.id = submissions.task_id").
		Where("tasks.id = ?", taskId).
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}
	return submissions, nil
}


func (us *SubmissionRepositoryImpl) GetAllForTaskTeacher(tx *gorm.DB, taskId, teacherId int64, queryParams map[string][]string) ([]models.Submission, error) {
	submissions := []models.Submission{}

	tx = utils.ApplyQueryParams(tx, queryParams)

	err := tx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Joins("JOIN tasks ON tasks.id = submissions.task_id").
		Where("tasks.id = ? AND tasks.created_by_id = ?", taskId, teacherId).
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *SubmissionRepositoryImpl) GetAllForTaskStudent(tx *gorm.DB, taskId, studentId int64, queryParams map[string][]string) ([]models.Submission, error) {
	submissions := []models.Submission{}

	tx = utils.ApplyQueryParams(tx, queryParams)

	err := tx.Model(&models.Submission{}).
		Preload("Language").
		Preload("Task").
		Preload("User").
		Joins("JOIN tasks ON tasks.id = submissions.task_id").
		Where("tasks.id = ? AND submissions.user_id = ?", taskId, studentId).
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (us *SubmissionRepositoryImpl) CreateSubmission(tx *gorm.DB, submission models.Submission) (int64, error) {
	err := tx.Create(&submission).Error
	if err != nil {
		return 0, err
	}
	return submission.Id, nil
}

func (us *SubmissionRepositoryImpl) MarkSubmissionProcessing(tx *gorm.DB, submissionId int64) error {
	err := tx.Model(&models.Submission{}).Where("id = ?", submissionId).Update("status", "processing").Error
	return err
}

func (us *SubmissionRepositoryImpl) MarkSubmissionComplete(tx *gorm.DB, submissionId int64) error {
	err := tx.Model(&models.Submission{}).Where("id = ?", submissionId).Update("status", "completed").Error
	return err
}

func (us *SubmissionRepositoryImpl) MarkSubmissionFailed(db *gorm.DB, submissionId int64, errorMsg string) error {
	err := db.Model(&models.Submission{}).Where("id = ?", submissionId).Updates(map[string]interface{}{
		"status":         "failed",
		"status_message": errorMsg,
	}).Error
	return err
}

func NewSubmissionRepository(db *gorm.DB) (SubmissionRepository, error) {
	if !db.Migrator().HasTable(&models.Submission{}) {
		err := db.Migrator().CreateTable(&models.Submission{})
		if err != nil {
			return nil, err
		}
	}
	return &SubmissionRepositoryImpl{}, nil
}
