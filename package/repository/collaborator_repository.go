package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/types"
	"gorm.io/gorm"
)

type CollaboratorRepository interface {
	// Contest Collaborators
	AddContestCollaborator(tx *gorm.DB, collaborator *models.ContestCollaborator) error
	GetContestCollaborators(tx *gorm.DB, contestID int64) ([]models.ContestCollaborator, error)
	GetContestCollaborator(tx *gorm.DB, contestID, userID int64) (*models.ContestCollaborator, error)
	UpdateContestCollaboratorPermission(tx *gorm.DB, contestID, userID int64, permission types.Permission) error
	RemoveContestCollaborator(tx *gorm.DB, contestID, userID int64) error
	HasContestPermission(tx *gorm.DB, contestID, userID int64, requiredPermission types.Permission) (bool, error)
	GetUserContestPermission(tx *gorm.DB, contestID, userID int64) (types.Permission, error)

	// Task Collaborators
	AddTaskCollaborator(tx *gorm.DB, collaborator *models.TaskCollaborator) error
	GetTaskCollaborators(tx *gorm.DB, taskID int64) ([]models.TaskCollaborator, error)
	GetTaskCollaborator(tx *gorm.DB, taskID, userID int64) (*models.TaskCollaborator, error)
	UpdateTaskCollaboratorPermission(tx *gorm.DB, taskID, userID int64, permission types.Permission) error
	RemoveTaskCollaborator(tx *gorm.DB, taskID, userID int64) error
	HasTaskPermission(tx *gorm.DB, taskID, userID int64, requiredPermission types.Permission) (bool, error)
	GetUserTaskPermission(tx *gorm.DB, taskID, userID int64) (types.Permission, error)
}

type collaboratorRepository struct{}

// Contest Collaborators

func (r *collaboratorRepository) AddContestCollaborator(tx *gorm.DB, collaborator *models.ContestCollaborator) error {
	return tx.Create(collaborator).Error
}

func (r *collaboratorRepository) GetContestCollaborators(tx *gorm.DB, contestID int64) ([]models.ContestCollaborator, error) {
	var collaborators []models.ContestCollaborator
	err := tx.Where("contest_id = ?", contestID).Preload("User").Find(&collaborators).Error
	return collaborators, err
}

func (r *collaboratorRepository) GetContestCollaborator(tx *gorm.DB, contestID, userID int64) (*models.ContestCollaborator, error) {
	var collaborator models.ContestCollaborator
	err := tx.Where("contest_id = ? AND user_id = ?", contestID, userID).Preload("User").First(&collaborator).Error
	if err != nil {
		return nil, err
	}
	return &collaborator, nil
}

func (r *collaboratorRepository) UpdateContestCollaboratorPermission(tx *gorm.DB, contestID, userID int64, permission types.Permission) error {
	return tx.Model(&models.ContestCollaborator{}).
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		Update("permission", permission).Error
}

func (r *collaboratorRepository) RemoveContestCollaborator(tx *gorm.DB, contestID, userID int64) error {
	return tx.Where("contest_id = ? AND user_id = ?", contestID, userID).
		Delete(&models.ContestCollaborator{}).Error
}

func (r *collaboratorRepository) HasContestPermission(tx *gorm.DB, contestID, userID int64, requiredPermission types.Permission) (bool, error) {
	permission, err := r.GetUserContestPermission(tx, contestID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return permission.HasPermission(requiredPermission), nil
}

func (r *collaboratorRepository) GetUserContestPermission(tx *gorm.DB, contestID, userID int64) (types.Permission, error) {
	var collaborator models.ContestCollaborator
	err := tx.Where("contest_id = ? AND user_id = ?", contestID, userID).
		Select("permission").
		First(&collaborator).Error
	if err != nil {
		return "", err
	}
	return collaborator.Permission, nil
}

// Task Collaborators

func (r *collaboratorRepository) AddTaskCollaborator(tx *gorm.DB, collaborator *models.TaskCollaborator) error {
	return tx.Create(collaborator).Error
}

func (r *collaboratorRepository) GetTaskCollaborators(tx *gorm.DB, taskID int64) ([]models.TaskCollaborator, error) {
	var collaborators []models.TaskCollaborator
	err := tx.Where("task_id = ?", taskID).Preload("User").Find(&collaborators).Error
	return collaborators, err
}

func (r *collaboratorRepository) GetTaskCollaborator(tx *gorm.DB, taskID, userID int64) (*models.TaskCollaborator, error) {
	var collaborator models.TaskCollaborator
	err := tx.Where("task_id = ? AND user_id = ?", taskID, userID).Preload("User").First(&collaborator).Error
	if err != nil {
		return nil, err
	}
	return &collaborator, nil
}

func (r *collaboratorRepository) UpdateTaskCollaboratorPermission(tx *gorm.DB, taskID, userID int64, permission types.Permission) error {
	return tx.Model(&models.TaskCollaborator{}).
		Where("task_id = ? AND user_id = ?", taskID, userID).
		Update("permission", permission).Error
}

func (r *collaboratorRepository) RemoveTaskCollaborator(tx *gorm.DB, taskID, userID int64) error {
	return tx.Where("task_id = ? AND user_id = ?", taskID, userID).
		Delete(&models.TaskCollaborator{}).Error
}

func (r *collaboratorRepository) HasTaskPermission(tx *gorm.DB, taskID, userID int64, requiredPermission types.Permission) (bool, error) {
	permission, err := r.GetUserTaskPermission(tx, taskID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return permission.HasPermission(requiredPermission), nil
}

func (r *collaboratorRepository) GetUserTaskPermission(tx *gorm.DB, taskID, userID int64) (types.Permission, error) {
	var collaborator models.TaskCollaborator
	err := tx.Where("task_id = ? AND user_id = ?", taskID, userID).
		Select("permission").
		First(&collaborator).Error
	if err != nil {
		return "", err
	}
	return collaborator.Permission, nil
}

func NewCollaboratorRepository() CollaboratorRepository {
	return &collaboratorRepository{}
}
