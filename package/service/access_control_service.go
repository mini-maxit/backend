package service

import (
	"errors"
	"time"

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
	AddCollaborator(tx *gorm.DB, currentUser *schemas.User, resourceType models.ResourceType, resourceID int64, userID int64, permission types.Permission) error

	// GetCollaborators retrieves all collaborators for a resource.
	GetCollaborators(tx *gorm.DB, currentUser *schemas.User, resourceType models.ResourceType, resourceID int64) ([]schemas.Collaborator, error)

	// UpdateCollaborator updates a collaborator's permission.
	UpdateCollaborator(tx *gorm.DB, currentUser *schemas.User, resourceType models.ResourceType, resourceID int64, userID int64, permission types.Permission) error

	// RemoveCollaborator removes a collaborator from a resource.
	RemoveCollaborator(tx *gorm.DB, currentUser *schemas.User, resourceType models.ResourceType, resourceID int64, userID int64) error

	// GetUserPermission gets a user's permission level for a resource.
	GetUserPermission(tx *gorm.DB, resourceType models.ResourceType, resourceID int64, userID int64) (types.Permission, error)

	// GrantOwnerAccess auto-grants owner permission to the creator/owner of a resource.
	GrantOwnerAccess(tx *gorm.DB, resourceType models.ResourceType, resourceID int64, ownerID int64) error
}

type accessControlService struct {
	accessControlRepository repository.AccessControlRepository
	userRepository          repository.UserRepository
	taskRepository          repository.TaskRepository
	contestRepository       repository.ContestRepository
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
func (acs *accessControlService) AddCollaborator(tx *gorm.DB, currentUser *schemas.User, resourceType models.ResourceType, resourceID int64, userID int64, permission types.Permission) error {
	// Owner permission cannot be manually assigned
	if permission == types.PermissionOwner {
		return myerrors.ErrForbidden
	}

	hasPermission, err := acs.HasPermission(tx, resourceType, resourceID, currentUser.ID, types.PermissionManage)
	if err != nil {
		acs.logger.Errorw("Error checking permission", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", currentUser.ID)
		return err
	}
	if !hasPermission {
		return myerrors.ErrForbidden
	}

	// Check if collaborator already exists to return a user-friendly error
	existing, err := acs.accessControlRepository.GetAccess(tx, resourceType, resourceID, userID)
	if err == nil && existing != nil {
		return myerrors.ErrAccessAlreadyExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		acs.logger.Errorw("Error checking existing access", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", userID)
		return err
	}

	// Check if target user exists
	_, err = acs.userRepository.Get(tx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrNotFound
		}
		acs.logger.Errorw("Error getting user", "error", err, "userID", userID)
		return err
	}

	access := &models.AccessControl{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		UserID:       userID,
		Permission:   permission,
	}

	err = acs.accessControlRepository.AddAccess(tx, access)
	if err != nil {
		acs.logger.Errorw("Error adding collaborator", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", userID)
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return myerrors.ErrAccessAlreadyExists
		}
		return err
	}

	return nil
}

// GetCollaborators retrieves all collaborators for a resource and converts them to schema objects. Requires edit permission.
func (acs *accessControlService) GetCollaborators(tx *gorm.DB, currentUser *schemas.User, resourceType models.ResourceType, resourceID int64) ([]schemas.Collaborator, error) {
	hasPermission, err := acs.HasPermission(tx, resourceType, resourceID, currentUser.ID, types.PermissionEdit)
	if err != nil {
		acs.logger.Errorw("Error checking permission", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", currentUser.ID)
		return nil, err
	}
	if !hasPermission {
		return nil, myerrors.ErrForbidden
	}
	accesses, err := acs.accessControlRepository.GetResourceAccess(tx, resourceType, resourceID)
	if err != nil {
		acs.logger.Errorw("Error getting collaborators", "error", err, "resourceType", resourceType, "resourceID", resourceID)
		return nil, err
	}
	result := make([]schemas.Collaborator, len(accesses))
	for i := range accesses {
		result[i] = *accessControlToCollaborator(&accesses[i])
	}
	return result, nil
}

// UpdateCollaborator updates a collaborator's permission.
func (acs *accessControlService) UpdateCollaborator(tx *gorm.DB, currentUser *schemas.User, resourceType models.ResourceType, resourceID int64, userID int64, permission types.Permission) error {
	// Cannot assign owner via update
	if permission == types.PermissionOwner {
		return myerrors.ErrForbidden
	}

	hasPermission, err := acs.HasPermission(tx, resourceType, resourceID, currentUser.ID, types.PermissionManage)
	if err != nil {
		acs.logger.Errorw("Error checking permission", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", currentUser.ID)
		return err
	}
	if !hasPermission {
		return myerrors.ErrForbidden
	}

	// Check if collaborator exists
	existing, err := acs.accessControlRepository.GetAccess(tx, resourceType, resourceID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrNotFound
		}
		acs.logger.Errorw("Error getting access", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", userID)
		return err
	}

	// Owner entry is immutable
	if existing.Permission == types.PermissionOwner {
		return myerrors.ErrForbidden
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
func (acs *accessControlService) RemoveCollaborator(tx *gorm.DB, currentUser *schemas.User, resourceType models.ResourceType, resourceID int64, userID int64) error {
	// Actor must have manage permission on the resource
	hasPermission, err := acs.HasPermission(tx, resourceType, resourceID, currentUser.ID, types.PermissionManage)
	if err != nil {
		acs.logger.Errorw("Error checking permission", "error", err, "resourceType", resourceType, "resourceID", resourceID, "actorUserID", currentUser.ID)
		return err
	}
	if !hasPermission {
		return myerrors.ErrForbidden
	}

	// Fetch target access entry
	access, err := acs.accessControlRepository.GetAccess(tx, resourceType, resourceID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrNotFound
		}
		acs.logger.Errorw("Error getting access for removal", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", userID)
		return err
	}

	// Owner permission is immutable and cannot be removed
	if access.Permission == types.PermissionOwner {
		return myerrors.ErrForbidden
	}

	// Remove collaborator
	if err = acs.accessControlRepository.RemoveAccess(tx, resourceType, resourceID, userID); err != nil {
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

// GrantOwnerAccess auto-grants owner permission to the creator/owner of a resource.
func (acs *accessControlService) GrantOwnerAccess(tx *gorm.DB, resourceType models.ResourceType, resourceID int64, ownerID int64) error {
	access := &models.AccessControl{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		UserID:       ownerID,
		Permission:   types.PermissionOwner,
	}

	err := acs.accessControlRepository.AddAccess(tx, access)
	if err != nil {
		acs.logger.Warnw("Failed to grant owner access", "error", err, "resourceType", resourceType, "resourceID", resourceID, "ownerID", ownerID)
		// Don't fail the operation if we can't add owner as collaborator
	}
	return nil
}

// accessControlToCollaborator converts a models.AccessControl to a schemas.Collaborator.
func accessControlToCollaborator(access *models.AccessControl) *schemas.Collaborator {
	if access == nil {
		return nil
	}
	var userName, userEmail string
	if (access.User != models.User{}) {
		userName = access.User.Name
		userEmail = access.User.Email
	}
	return &schemas.Collaborator{
		UserID:     access.UserID,
		UserName:   userName,
		UserEmail:  userEmail,
		Permission: access.Permission,
		AddedAt:    access.CreatedAt.Format(time.RFC3339),
	}
}

// accessControlsToCollaborators maps a slice of AccessControl models to Collaborator schemas.
func accessControlsToCollaborators(accesses []models.AccessControl) []schemas.Collaborator {
	result := make([]schemas.Collaborator, len(accesses))
	for i := range accesses {
		result[i] = *accessControlToCollaborator(&accesses[i])
	}
	return result
}

func (acs *accessControlService) getResource(tx *gorm.DB, resourceType models.ResourceType, resourceID int64) (interface{}, error) {
	switch resourceType {
	case models.ResourceTypeContest:
		contest, err := acs.contestRepository.Get(tx, resourceID)
		if err != nil {
			return nil, err
		}
		return contest, nil
	case models.ResourceTypeTask:
		task, err := acs.taskRepository.Get(tx, resourceID)
		if err != nil {
			return nil, err
		}
		return task, nil
	default:
		return nil, myerrors.ErrInvalidData
	}
}

// NewAccessControlService creates a new AccessControlService.
func NewAccessControlService(
	accessControlRepository repository.AccessControlRepository,
	userRepository repository.UserRepository,
	taskRepository repository.TaskRepository,
	contestRepository repository.ContestRepository,
) AccessControlService {
	log := utils.NewNamedLogger("access_control_service")
	return &accessControlService{
		accessControlRepository: accessControlRepository,
		userRepository:          userRepository,
		taskRepository:          taskRepository,
		contestRepository:       contestRepository,
		logger:                  log,
	}
}
