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
	// CanUserAccess checks if a user can access a resource with the required permission.
	// Checks in order: admin role → creator → required minimal permission
	CanUserAccess(tx *gorm.DB, resourceType models.ResourceType, resourceID int64, user *schemas.User, requiredPermission types.Permission) error

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

// CanUserAccess checks if a user can access a resource with the required permission.
// Checks in order: admin role → creator → collaborator permissions.
func (acs *accessControlService) CanUserAccess(tx *gorm.DB, resourceType models.ResourceType, resourceID int64, user *schemas.User, requiredPermission types.Permission) error {
	if user.Role == types.UserRoleAdmin {
		return nil
	}
	permission, err := acs.accessControlRepository.GetUserPermission(tx, resourceType, resourceID, user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrForbidden
		}
		acs.logger.Errorw("Error checking permission", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", user.ID)
		return err
	}
	if permission.HasPermission(requiredPermission) {
		return nil
	}
	return myerrors.ErrForbidden
}

// AddCollaborator adds a collaborator with specified permission to a resource.
func (acs *accessControlService) AddCollaborator(tx *gorm.DB, currentUser *schemas.User, resourceType models.ResourceType, resourceID int64, userID int64, permission types.Permission) error {
	// Owner permission cannot be manually assigned
	if permission == types.PermissionOwner {
		return myerrors.ErrForbidden
	}

	err := acs.CanUserAccess(tx, resourceType, resourceID, currentUser, types.PermissionManage)
	if err != nil {
		if !errors.Is(err, myerrors.ErrForbidden) {
			acs.logger.Errorw("Error checking permission", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", currentUser.ID)
		}
		return err
	}
	// Check if resource Exists
	exists, err := acs.checkResourceExists(tx, resourceType, resourceID)
	if err != nil {
		acs.logger.Errorw("Error getting resource", "error", err, "resourceType", resourceType, "resourceID", resourceID)
		return err
	}
	if !exists {
		return myerrors.ErrNotFound
	}

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
	err := acs.CanUserAccess(tx, resourceType, resourceID, currentUser, types.PermissionEdit)
	if err != nil {
		if !errors.Is(err, myerrors.ErrForbidden) {
			acs.logger.Errorw("Error checking permission", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", currentUser.ID)
		}
		return nil, err
	}

	if _, err = acs.checkResourceExists(tx, resourceType, resourceID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrNotFound
		}
		acs.logger.Errorw("Error getting resource", "error", err, "resourceType", resourceType, "resourceID", resourceID)
		return nil, err
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
		return myerrors.ErrCannotAssignOwner
	}

	err := acs.CanUserAccess(tx, resourceType, resourceID, currentUser, types.PermissionManage)
	if err != nil {
		if !errors.Is(err, myerrors.ErrForbidden) {
			acs.logger.Errorw("Error checking permission", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", currentUser.ID)
		}
		return err
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
	err := acs.CanUserAccess(tx, resourceType, resourceID, currentUser, types.PermissionManage)
	if err != nil {
		if !errors.Is(err, myerrors.ErrForbidden) {
			acs.logger.Errorw("Error checking permission", "error", err, "resourceType", resourceType, "resourceID", resourceID, "actorUserID", currentUser.ID)
		}
		return err
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

func (acs *accessControlService) checkResourceExists(tx *gorm.DB, resourceType models.ResourceType, resourceID int64) (bool, error) {
	switch resourceType {
	case models.ResourceTypeContest:
		_, err := acs.contestRepository.Get(tx, resourceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	case models.ResourceTypeTask:
		_, err := acs.taskRepository.Get(tx, resourceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	default:
		return false, myerrors.ErrInvalidData
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
