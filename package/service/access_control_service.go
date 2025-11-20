package service

import (
	"errors"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	myerrors "github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AccessControlService handles all access control and permission management for resources.
type AccessControlService interface {
	// HasPermission checks if a user has the required permission for a resource.
	// Returns true if the user has the permission or higher.
	HasPermission(tx *gorm.DB, resourceType models.ResourceType, resourceID int64, userID int64, requiredPermission types.Permission) (bool, error)

	// CanUserAccess checks if a user can access a resource with the required permission.
	// Checks in order: admin role → creator → collaborator permissions.
	CanUserAccess(tx *gorm.DB, resourceType models.ResourceType, resourceID int64, user schemas.User, creatorID int64, requiredPermission types.Permission) (bool, error)

	// AddCollaborator adds a collaborator with specified permission to a resource.
	AddCollaborator(tx *gorm.DB, resourceType models.ResourceType, resourceID int64, userID int64, permission types.Permission) error

	// GetCollaborators retrieves all collaborators for a resource.
	GetCollaborators(tx *gorm.DB, resourceType models.ResourceType, resourceID int64) ([]models.AccessControl, error)

	// UpdateCollaborator updates a collaborator's permission.
	UpdateCollaborator(tx *gorm.DB, resourceType models.ResourceType, resourceID int64, userID int64, permission types.Permission) error

	// RemoveCollaborator removes a collaborator from a resource.
	RemoveCollaborator(tx *gorm.DB, resourceType models.ResourceType, resourceID int64, userID int64) error

	// GetUserPermission gets a user's permission level for a resource.
	GetUserPermission(tx *gorm.DB, resourceType models.ResourceType, resourceID int64, userID int64) (types.Permission, error)

	// GrantCreatorAccess auto-grants manage permission to the creator of a resource.
	GrantCreatorAccess(tx *gorm.DB, resourceType models.ResourceType, resourceID int64, creatorID int64) error
}

type accessControlService struct {
	accessControlRepository repository.AccessControlRepository
	userRepository          repository.UserRepository
	logger                  *zap.SugaredLogger
}

// HasPermission checks if a user has the required permission for a resource.
func (acs *accessControlService) HasPermission(tx *gorm.DB, resourceType models.ResourceType, resourceID int64, userID int64, requiredPermission types.Permission) (bool, error) {
	hasPermission, err := acs.accessControlRepository.HasPermission(tx, resourceType, resourceID, userID, requiredPermission)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		acs.logger.Errorw("Error checking permission", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", userID)
		return false, err
	}
	return hasPermission, nil
}

// CanUserAccess checks if a user can access a resource with the required permission.
// Checks in order: admin role → creator → collaborator permissions.
func (acs *accessControlService) CanUserAccess(tx *gorm.DB, resourceType models.ResourceType, resourceID int64, user schemas.User, creatorID int64, requiredPermission types.Permission) (bool, error) {
	// Admins have all permissions
	if user.Role == types.UserRoleAdmin {
		return true, nil
	}

	// Creators have full access
	if creatorID == user.ID {
		return true, nil
	}

	// Check collaborator permissions
	hasPermission, err := acs.HasPermission(tx, resourceType, resourceID, user.ID, requiredPermission)
	if err != nil {
		return false, err
	}

	return hasPermission, nil
}

// AddCollaborator adds a collaborator with specified permission to a resource.
func (acs *accessControlService) AddCollaborator(tx *gorm.DB, resourceType models.ResourceType, resourceID int64, userID int64, permission types.Permission) error {
	// Check if user exists
	_, err := acs.userRepository.Get(tx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrNotFound
		}
		acs.logger.Errorw("Error getting user", "error", err, "userID", userID)
		return err
	}

	// Create access control model
	access := &models.AccessControl{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		UserID:       userID,
		Permission:   permission,
	}

	// Add access
	err = acs.accessControlRepository.AddAccess(tx, access)
	if err != nil {
		acs.logger.Errorw("Error adding collaborator", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", userID)
		return err
	}

	return nil
}

// GetCollaborators retrieves all collaborators for a resource.
func (acs *accessControlService) GetCollaborators(tx *gorm.DB, resourceType models.ResourceType, resourceID int64) ([]models.AccessControl, error) {
	collaborators, err := acs.accessControlRepository.GetResourceAccess(tx, resourceType, resourceID)
	if err != nil {
		acs.logger.Errorw("Error getting collaborators", "error", err, "resourceType", resourceType, "resourceID", resourceID)
		return nil, err
	}
	return collaborators, nil
}

// UpdateCollaborator updates a collaborator's permission.
func (acs *accessControlService) UpdateCollaborator(tx *gorm.DB, resourceType models.ResourceType, resourceID int64, userID int64, permission types.Permission) error {
	// Check if collaborator exists
	_, err := acs.accessControlRepository.GetAccess(tx, resourceType, resourceID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrNotFound
		}
		acs.logger.Errorw("Error getting access", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", userID)
		return err
	}

	// Update permission
	err = acs.accessControlRepository.UpdatePermission(tx, resourceType, resourceID, userID, permission)
	if err != nil {
		acs.logger.Errorw("Error updating collaborator permission", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", userID)
		return err
	}

	return nil
}

// RemoveCollaborator removes a collaborator from a resource.
func (acs *accessControlService) RemoveCollaborator(tx *gorm.DB, resourceType models.ResourceType, resourceID int64, userID int64) error {
	err := acs.accessControlRepository.RemoveAccess(tx, resourceType, resourceID, userID)
	if err != nil {
		acs.logger.Errorw("Error removing collaborator", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", userID)
		return err
	}
	return nil
}

// GetUserPermission gets a user's permission level for a resource.
func (acs *accessControlService) GetUserPermission(tx *gorm.DB, resourceType models.ResourceType, resourceID int64, userID int64) (types.Permission, error) {
	permission, err := acs.accessControlRepository.GetUserPermission(tx, resourceType, resourceID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", myerrors.ErrNotFound
		}
		acs.logger.Errorw("Error getting user permission", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", userID)
		return "", err
	}
	return permission, nil
}

// GrantCreatorAccess auto-grants manage permission to the creator of a resource.
func (acs *accessControlService) GrantCreatorAccess(tx *gorm.DB, resourceType models.ResourceType, resourceID int64, creatorID int64) error {
	access := &models.AccessControl{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		UserID:       creatorID,
		Permission:   types.PermissionManage,
	}

	err := acs.accessControlRepository.AddAccess(tx, access)
	if err != nil {
		acs.logger.Warnw("Failed to grant creator access", "error", err, "resourceType", resourceType, "resourceID", resourceID, "creatorID", creatorID)
		// Don't fail the operation if we can't add creator as collaborator
	}
	return nil
}

// NewAccessControlService creates a new AccessControlService.
func NewAccessControlService(
	accessControlRepository repository.AccessControlRepository,
	userRepository repository.UserRepository,
) AccessControlService {
	log := utils.NewNamedLogger("access_control_service")
	return &accessControlService{
		accessControlRepository: accessControlRepository,
		userRepository:          userRepository,
		logger:                  log,
	}
}
