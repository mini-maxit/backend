package repository

import (
	"errors"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/types"
	"gorm.io/gorm"
)

// AccessControlRepository manages access control for resources (contests and tasks).
type AccessControlRepository interface {
	// Generic access control methods
	AddAccess(tx *gorm.DB, access *models.AccessControl) error
	GetAccess(tx *gorm.DB, resourceType models.ResourceType, resourceID, userID int64) (*models.AccessControl, error)
	GetResourceAccess(tx *gorm.DB, resourceType models.ResourceType, resourceID int64) ([]models.AccessControl, error)
	UpdatePermission(tx *gorm.DB, resourceType models.ResourceType, resourceID, userID int64, permission types.Permission) error
	RemoveAccess(tx *gorm.DB, resourceType models.ResourceType, resourceID, userID int64) error
	HasPermission(tx *gorm.DB, resourceType models.ResourceType, resourceID, userID int64, requiredPermission types.Permission) (bool, error)
	GetUserPermission(tx *gorm.DB, resourceType models.ResourceType, resourceID, userID int64) (types.Permission, error)

	// Convenience methods for contests
	AddContestCollaborator(tx *gorm.DB, contestID, userID int64, permission types.Permission) error
	GetContestCollaborators(tx *gorm.DB, contestID int64) ([]models.AccessControl, error)
	HasContestPermission(tx *gorm.DB, contestID, userID int64, requiredPermission types.Permission) (bool, error)
	GetUserContestPermission(tx *gorm.DB, contestID, userID int64) (types.Permission, error)
	UpdateContestCollaboratorPermission(tx *gorm.DB, contestID, userID int64, permission types.Permission) error
	RemoveContestCollaborator(tx *gorm.DB, contestID, userID int64) error

	// Convenience methods for tasks
	AddTaskCollaborator(tx *gorm.DB, taskID, userID int64, permission types.Permission) error
	GetTaskCollaborators(tx *gorm.DB, taskID int64) ([]models.AccessControl, error)
	HasTaskPermission(tx *gorm.DB, taskID, userID int64, requiredPermission types.Permission) (bool, error)
	GetUserTaskPermission(tx *gorm.DB, taskID, userID int64) (types.Permission, error)
	UpdateTaskCollaboratorPermission(tx *gorm.DB, taskID, userID int64, permission types.Permission) error
	RemoveTaskCollaborator(tx *gorm.DB, taskID, userID int64) error
}

type accessControlRepository struct{}

// Generic Methods

func (r *accessControlRepository) AddAccess(tx *gorm.DB, access *models.AccessControl) error {
	return tx.Create(access).Error
}

func (r *accessControlRepository) GetAccess(tx *gorm.DB, resourceType models.ResourceType, resourceID, userID int64) (*models.AccessControl, error) {
	var access models.AccessControl
	err := tx.Where("resource_type = ? AND resource_id = ? AND user_id = ?", resourceType, resourceID, userID).
		Preload("User").
		First(&access).Error
	if err != nil {
		return nil, err
	}
	return &access, nil
}

func (r *accessControlRepository) GetResourceAccess(tx *gorm.DB, resourceType models.ResourceType, resourceID int64) ([]models.AccessControl, error) {
	var accesses []models.AccessControl
	err := tx.Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).
		Preload("User").
		Find(&accesses).Error
	return accesses, err
}

func (r *accessControlRepository) UpdatePermission(tx *gorm.DB, resourceType models.ResourceType, resourceID, userID int64, permission types.Permission) error {
	return tx.Model(&models.AccessControl{}).
		Where("resource_type = ? AND resource_id = ? AND user_id = ?", resourceType, resourceID, userID).
		Update("permission", permission).Error
}

func (r *accessControlRepository) RemoveAccess(tx *gorm.DB, resourceType models.ResourceType, resourceID, userID int64) error {
	return tx.Where("resource_type = ? AND resource_id = ? AND user_id = ?", resourceType, resourceID, userID).
		Delete(&models.AccessControl{}).Error
}

func (r *accessControlRepository) HasPermission(tx *gorm.DB, resourceType models.ResourceType, resourceID, userID int64, requiredPermission types.Permission) (bool, error) {
	permission, err := r.GetUserPermission(tx, resourceType, resourceID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return permission.HasPermission(requiredPermission), nil
}

func (r *accessControlRepository) GetUserPermission(tx *gorm.DB, resourceType models.ResourceType, resourceID, userID int64) (types.Permission, error) {
	var access models.AccessControl
	err := tx.Where("resource_type = ? AND resource_id = ? AND user_id = ?", resourceType, resourceID, userID).
		Select("permission").
		First(&access).Error
	if err != nil {
		return "", err
	}
	return access.Permission, nil
}

// Contest Convenience Methods

func (r *accessControlRepository) AddContestCollaborator(tx *gorm.DB, contestID, userID int64, permission types.Permission) error {
	access := &models.AccessControl{
		ResourceType: models.ResourceTypeContest,
		ResourceID:   contestID,
		UserID:       userID,
		Permission:   permission,
	}
	return r.AddAccess(tx, access)
}

func (r *accessControlRepository) GetContestCollaborators(tx *gorm.DB, contestID int64) ([]models.AccessControl, error) {
	return r.GetResourceAccess(tx, models.ResourceTypeContest, contestID)
}

func (r *accessControlRepository) HasContestPermission(tx *gorm.DB, contestID, userID int64, requiredPermission types.Permission) (bool, error) {
	return r.HasPermission(tx, models.ResourceTypeContest, contestID, userID, requiredPermission)
}

func (r *accessControlRepository) GetUserContestPermission(tx *gorm.DB, contestID, userID int64) (types.Permission, error) {
	return r.GetUserPermission(tx, models.ResourceTypeContest, contestID, userID)
}

func (r *accessControlRepository) UpdateContestCollaboratorPermission(tx *gorm.DB, contestID, userID int64, permission types.Permission) error {
	return r.UpdatePermission(tx, models.ResourceTypeContest, contestID, userID, permission)
}

func (r *accessControlRepository) RemoveContestCollaborator(tx *gorm.DB, contestID, userID int64) error {
	return r.RemoveAccess(tx, models.ResourceTypeContest, contestID, userID)
}

// Task Convenience Methods

func (r *accessControlRepository) AddTaskCollaborator(tx *gorm.DB, taskID, userID int64, permission types.Permission) error {
	access := &models.AccessControl{
		ResourceType: models.ResourceTypeTask,
		ResourceID:   taskID,
		UserID:       userID,
		Permission:   permission,
	}
	return r.AddAccess(tx, access)
}

func (r *accessControlRepository) GetTaskCollaborators(tx *gorm.DB, taskID int64) ([]models.AccessControl, error) {
	return r.GetResourceAccess(tx, models.ResourceTypeTask, taskID)
}

func (r *accessControlRepository) HasTaskPermission(tx *gorm.DB, taskID, userID int64, requiredPermission types.Permission) (bool, error) {
	return r.HasPermission(tx, models.ResourceTypeTask, taskID, userID, requiredPermission)
}

func (r *accessControlRepository) GetUserTaskPermission(tx *gorm.DB, taskID, userID int64) (types.Permission, error) {
	return r.GetUserPermission(tx, models.ResourceTypeTask, taskID, userID)
}

func (r *accessControlRepository) UpdateTaskCollaboratorPermission(tx *gorm.DB, taskID, userID int64, permission types.Permission) error {
	return r.UpdatePermission(tx, models.ResourceTypeTask, taskID, userID, permission)
}

func (r *accessControlRepository) RemoveTaskCollaborator(tx *gorm.DB, taskID, userID int64) error {
	return r.RemoveAccess(tx, models.ResourceTypeTask, taskID, userID)
}

func NewAccessControlRepository() AccessControlRepository {
	return &accessControlRepository{}
}
