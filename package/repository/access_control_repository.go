package repository

import (
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/types"
)

// AccessControlRepository manages access control for resources (contests and tasks).
type AccessControlRepository interface {
	// Generic access control methods
	AddAccess(db database.Database, access *models.AccessControl) error
	GetAccess(db database.Database, resourceType models.ResourceType, resourceID, userID int64) (*models.AccessControl, error)
	GetResourceAccess(db database.Database, resourceType models.ResourceType, resourceID int64) ([]models.AccessControl, error)
	UpdatePermission(db database.Database, resourceType models.ResourceType, resourceID, userID int64, permission types.Permission) error
	RemoveAccess(db database.Database, resourceType models.ResourceType, resourceID, userID int64) error
	GetUserPermission(db database.Database, resourceType models.ResourceType, resourceID, userID int64) (types.Permission, error)

	// Convenience methods for contests
	AddContestCollaborator(db database.Database, contestID, userID int64, permission types.Permission) error
	GetContestCollaborators(db database.Database, contestID int64) ([]models.AccessControl, error)
	GetUserContestPermission(db database.Database, contestID, userID int64) (types.Permission, error)
	UpdateContestCollaboratorPermission(db database.Database, contestID, userID int64, permission types.Permission) error
	RemoveContestCollaborator(db database.Database, contestID, userID int64) error

	// Convenience methods for tasks
	AddTaskCollaborator(db database.Database, taskID, userID int64, permission types.Permission) error
	GetTaskCollaborators(db database.Database, taskID int64) ([]models.AccessControl, error)
	GetUserTaskPermission(db database.Database, taskID, userID int64) (types.Permission, error)
	UpdateTaskCollaboratorPermission(db database.Database, taskID, userID int64, permission types.Permission) error
	RemoveTaskCollaborator(db database.Database, taskID, userID int64) error
}

type accessControlRepository struct{}

// Generic Methods

func (r *accessControlRepository) AddAccess(db database.Database, access *models.AccessControl) error {
	tx := db.GetInstance()
	return tx.Create(access).Error
}

func (r *accessControlRepository) GetAccess(db database.Database, resourceType models.ResourceType, resourceID, userID int64) (*models.AccessControl, error) {
	tx := db.GetInstance()
	var access models.AccessControl
	err := tx.Where("resource_type = ? AND resource_id = ? AND user_id = ?", resourceType, resourceID, userID).
		Preload("User").
		First(&access).Error
	if err != nil {
		return nil, err
	}
	return &access, nil
}

func (r *accessControlRepository) GetResourceAccess(db database.Database, resourceType models.ResourceType, resourceID int64) ([]models.AccessControl, error) {
	tx := db.GetInstance()
	var accesses []models.AccessControl
	err := tx.Model(&models.AccessControl{}).Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).
		Preload("User").
		Find(&accesses).Error
	return accesses, err
}

func (r *accessControlRepository) UpdatePermission(db database.Database, resourceType models.ResourceType, resourceID, userID int64, permission types.Permission) error {
	tx := db.GetInstance()
	return tx.Model(&models.AccessControl{}).
		Where("resource_type = ? AND resource_id = ? AND user_id = ?", resourceType, resourceID, userID).
		Update("permission", permission).Error
}

func (r *accessControlRepository) RemoveAccess(db database.Database, resourceType models.ResourceType, resourceID, userID int64) error {
	tx := db.GetInstance()
	return tx.Where("resource_type = ? AND resource_id = ? AND user_id = ?", resourceType, resourceID, userID).
		Delete(&models.AccessControl{}).Error
}

func (r *accessControlRepository) GetUserPermission(db database.Database, resourceType models.ResourceType, resourceID, userID int64) (types.Permission, error) {
	tx := db.GetInstance()
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

func (r *accessControlRepository) AddContestCollaborator(db database.Database, contestID, userID int64, permission types.Permission) error {
	access := &models.AccessControl{
		ResourceType: models.ResourceTypeContest,
		ResourceID:   contestID,
		UserID:       userID,
		Permission:   permission,
	}
	return r.AddAccess(db, access)
}

func (r *accessControlRepository) GetContestCollaborators(db database.Database, contestID int64) ([]models.AccessControl, error) {
	return r.GetResourceAccess(db, models.ResourceTypeContest, contestID)
}

func (r *accessControlRepository) GetUserContestPermission(db database.Database, contestID, userID int64) (types.Permission, error) {
	return r.GetUserPermission(db, models.ResourceTypeContest, contestID, userID)
}

func (r *accessControlRepository) UpdateContestCollaboratorPermission(db database.Database, contestID, userID int64, permission types.Permission) error {
	return r.UpdatePermission(db, models.ResourceTypeContest, contestID, userID, permission)
}

func (r *accessControlRepository) RemoveContestCollaborator(db database.Database, contestID, userID int64) error {
	return r.RemoveAccess(db, models.ResourceTypeContest, contestID, userID)
}

// Task Convenience Methods

func (r *accessControlRepository) AddTaskCollaborator(db database.Database, taskID, userID int64, permission types.Permission) error {
	access := &models.AccessControl{
		ResourceType: models.ResourceTypeTask,
		ResourceID:   taskID,
		UserID:       userID,
		Permission:   permission,
	}
	return r.AddAccess(db, access)
}

func (r *accessControlRepository) GetTaskCollaborators(db database.Database, taskID int64) ([]models.AccessControl, error) {
	return r.GetResourceAccess(db, models.ResourceTypeTask, taskID)
}

func (r *accessControlRepository) GetUserTaskPermission(db database.Database, taskID, userID int64) (types.Permission, error) {
	return r.GetUserPermission(db, models.ResourceTypeTask, taskID, userID)
}

func (r *accessControlRepository) UpdateTaskCollaboratorPermission(db database.Database, taskID, userID int64, permission types.Permission) error {
	return r.UpdatePermission(db, models.ResourceTypeTask, taskID, userID, permission)
}

func (r *accessControlRepository) RemoveTaskCollaborator(db database.Database, taskID, userID int64) error {
	return r.RemoveAccess(db, models.ResourceTypeTask, taskID, userID)
}

func NewAccessControlRepository() AccessControlRepository {
	return &accessControlRepository{}
}
