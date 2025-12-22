package repository

import (
	"fmt"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/utils"
)

// AccessControlRepository manages access control for resources (contests and tasks).
type AccessControlRepository interface {
	// Generic access control methods
	AddAccess(db database.Database, access *models.AccessControl) error
	GetAccess(db database.Database, resourceType types.ResourceType, resourceID, userID int64) (*models.AccessControl, error)
	GetResourceAccess(db database.Database, resourceType types.ResourceType, resourceID int64) ([]models.AccessControl, error)
	UpdatePermission(db database.Database, resourceType types.ResourceType, resourceID, userID int64, permission types.Permission) error
	RemoveAccess(db database.Database, resourceType types.ResourceType, resourceID, userID int64) error
	GetUserPermission(db database.Database, resourceType types.ResourceType, resourceID, userID int64) (types.Permission, error)
	// GetAssignableUsers returns teachers without any access entry for the given resource. Supports pagination and sorting.
	GetAssignableUsers(db database.Database, resourceType types.ResourceType, resourceID int64, limit, offset int, sort string) ([]models.User, int64, error)
}

type accessControlRepository struct{}

// Generic Methods

func (r *accessControlRepository) AddAccess(db database.Database, access *models.AccessControl) error {
	tx := db.GetInstance()

	// Check if a soft-deleted record exists
	var existing models.AccessControl
	err := tx.Unscoped().
		Where("resource_type = ? AND resource_id = ? AND user_id = ?",
			access.ResourceType, access.ResourceID, access.UserID).
		First(&existing).Error

	if err == nil && existing.DeletedAt.Valid {
		// Restore the soft-deleted record with the new permission
		return tx.Unscoped().Model(&existing).Updates(map[string]interface{}{
			"deleted_at": nil,
			"permission": access.Permission,
		}).Error
	}

	return tx.Create(access).Error
}

func (r *accessControlRepository) GetAccess(db database.Database, resourceType types.ResourceType, resourceID, userID int64) (*models.AccessControl, error) {
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

func (r *accessControlRepository) GetResourceAccess(db database.Database, resourceType types.ResourceType, resourceID int64) ([]models.AccessControl, error) {
	tx := db.GetInstance()
	var accesses []models.AccessControl
	err := tx.Model(&models.AccessControl{}).Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).
		Preload("User").
		Find(&accesses).Error
	return accesses, err
}

func (r *accessControlRepository) UpdatePermission(db database.Database, resourceType types.ResourceType, resourceID, userID int64, permission types.Permission) error {
	tx := db.GetInstance()
	return tx.Model(&models.AccessControl{}).
		Where("resource_type = ? AND resource_id = ? AND user_id = ?", resourceType, resourceID, userID).
		Update("permission", permission).Error
}

func (r *accessControlRepository) RemoveAccess(db database.Database, resourceType types.ResourceType, resourceID, userID int64) error {
	tx := db.GetInstance()
	return tx.Where("resource_type = ? AND resource_id = ? AND user_id = ?", resourceType, resourceID, userID).
		Delete(&models.AccessControl{}).Error
}

func (r *accessControlRepository) GetUserPermission(db database.Database, resourceType types.ResourceType, resourceID, userID int64) (types.Permission, error) {
	tx := db.GetInstance()
	var access models.AccessControl
	err := tx.Model(&models.AccessControl{}).
		Where("resource_type = ? AND resource_id = ? AND user_id = ?", resourceType, resourceID, userID).
		Select("permission").
		First(&access).Error
	if err != nil {
		return "", err
	}
	return access.Permission, nil
}

func NewAccessControlRepository() AccessControlRepository {
	return &accessControlRepository{}
}

// GetAssignableUsers returns teachers who currently do not have any access entry for the given resource.
func (r *accessControlRepository) GetAssignableUsers(db database.Database, resourceType types.ResourceType, resourceID int64, limit, offset int, sort string) ([]models.User, int64, error) {
	tx := db.GetInstance()

	// Count total assignable teachers (without existing access for the resource)
	accessControlsTable := database.ResolveTableName(tx, &models.AccessControl{})
	var total int64
	if err := tx.Model(&models.User{}).
		Where("role = ? OR role = ?", types.UserRoleTeacher, types.UserRoleAdmin).
		Where(fmt.Sprintf("NOT EXISTS (SELECT 1 FROM %s ac WHERE ac.user_id = users.id AND ac.resource_type = ? AND ac.resource_id = ?)", accessControlsTable), resourceType, resourceID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	paginatedTx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	// Fetch page of assignable teachers
	var users []models.User
	err = paginatedTx.Model(&models.User{}).
		Where("role = ?", types.UserRoleTeacher).
		Where(fmt.Sprintf("NOT EXISTS (SELECT 1 FROM %s ac WHERE ac.user_id = users.id AND ac.resource_type = ? AND ac.resource_id = ?)", accessControlsTable), resourceType, resourceID).
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}
