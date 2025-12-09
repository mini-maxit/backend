package service

import (
	"time"

	"github.com/mini-maxit/backend/internal/database"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AccessControlService handles all access control and permission management for resources.
type AccessControlService interface {
	// CanUserAccess checks if a user can access a resource with the required permission.
	// Checks in order: admin role → creator → required minimal permission
	CanUserAccess(db database.Database, resourceType types.ResourceType, resourceID int64, user *schemas.User, requiredPermission types.Permission) error

	// AddCollaborator adds a collaborator with specified permission to a resource.
	AddCollaborator(db database.Database, currentUser *schemas.User, resourceType types.ResourceType, resourceID int64, userID int64, permission types.Permission) error

	// GetCollaborators retrieves all collaborators for a resource.
	GetCollaborators(db database.Database, currentUser *schemas.User, resourceType types.ResourceType, resourceID int64) ([]schemas.Collaborator, error)

	// GetAssignableUsers returns users (teachers) that currently have no access to the resource and can be granted access. Supports pagination.
	GetAssignableUsers(db database.Database, currentUser *schemas.User, resourceType types.ResourceType, resourceID int64, params schemas.PaginationParams) (*schemas.PaginatedResult[[]schemas.User], error)

	// UpdateCollaborator updates a collaborator's permission.
	UpdateCollaborator(db database.Database, currentUser *schemas.User, resourceType types.ResourceType, resourceID int64, userID int64, permission types.Permission) error

	// RemoveCollaborator removes a collaborator from a resource.
	RemoveCollaborator(db database.Database, currentUser *schemas.User, resourceType types.ResourceType, resourceID int64, userID int64) error

	// GetUserPermission gets a user's permission level for a resource.
	GetUserPermission(db database.Database, resourceType types.ResourceType, resourceID int64, userID int64) (types.Permission, error)

	// GrantOwnerAccess auto-grants owner permission to the creator/owner of a resource.
	GrantOwnerAccess(db database.Database, resourceType types.ResourceType, resourceID int64, ownerID int64) error
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
func (acs *accessControlService) CanUserAccess(db database.Database, resourceType types.ResourceType, resourceID int64, user *schemas.User, requiredPermission types.Permission) error {
	if user.Role == types.UserRoleAdmin {
		return nil
	}
	permission, err := acs.accessControlRepository.GetUserPermission(db, resourceType, resourceID, user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrForbidden
		}
		acs.logger.Errorw("Error checking permission", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", user.ID)
		return err
	}
	if permission.HasPermission(requiredPermission) {
		return nil
	}
	return errors.ErrForbidden
}

// AddCollaborator adds a collaborator with specified permission to a resource.
func (acs *accessControlService) AddCollaborator(db database.Database, currentUser *schemas.User, resourceType types.ResourceType, resourceID int64, userID int64, permission types.Permission) error {
	// Owner permission cannot be manually assigned
	if permission == types.PermissionOwner {
		return errors.ErrForbidden
	}

	err := acs.CanUserAccess(db, resourceType, resourceID, currentUser, types.PermissionManage)
	if err != nil {
		if !errors.Is(err, errors.ErrForbidden) {
			acs.logger.Errorw("Error checking permission", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", currentUser.ID)
		}
		return err
	}
	// Check if resource Exists
	exists, err := acs.checkResourceExists(db, resourceType, resourceID)
	if err != nil {
		acs.logger.Errorw("Error getting resource", "error", err, "resourceType", resourceType, "resourceID", resourceID)
		return err
	}
	if !exists {
		return errors.ErrNotFound
	}

	existing, err := acs.accessControlRepository.GetAccess(db, resourceType, resourceID, userID)
	if err == nil && existing != nil {
		return errors.ErrAccessAlreadyExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		acs.logger.Errorw("Error checking existing access", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", userID)
		return err
	}

	// Check if target user exists
	_, err = acs.userRepository.Get(db, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrNotFound
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

	err = acs.accessControlRepository.AddAccess(db, access)
	if err != nil {
		acs.logger.Errorw("Error adding collaborator", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", userID)
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errors.ErrAccessAlreadyExists
		}
		return err
	}

	return nil
}

// GetCollaborators retrieves all collaborators for a resource and converts them to schema objects. Requires edit permission.
func (acs *accessControlService) GetCollaborators(db database.Database, currentUser *schemas.User, resourceType types.ResourceType, resourceID int64) ([]schemas.Collaborator, error) {
	// Check if resource exists first before checking permissions
	exists, err := acs.checkResourceExists(db, resourceType, resourceID)
	if err != nil {
		acs.logger.Errorw("Error checking resource existence", "error", err, "resourceType", resourceType, "resourceID", resourceID)
		return nil, err
	}
	if !exists {
		return nil, errors.ErrNotFound
	}

	err = acs.CanUserAccess(db, resourceType, resourceID, currentUser, types.PermissionEdit)
	if err != nil {
		if !errors.Is(err, errors.ErrForbidden) {
			acs.logger.Errorw("Error checking permission", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", currentUser.ID)
		}
		return nil, err
	}

	accesses, err := acs.accessControlRepository.GetResourceAccess(db, resourceType, resourceID)
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

// GetAssignableUsers returns users (teachers) that currently have no access to the resource and can be granted access.
// Requires manage permission. Performs existence check first. Supports pagination via params.
func (acs *accessControlService) GetAssignableUsers(db database.Database, currentUser *schemas.User, resourceType types.ResourceType, resourceID int64, params schemas.PaginationParams) (*schemas.PaginatedResult[[]schemas.User], error) {
	// Check if resource exists
	exists, err := acs.checkResourceExists(db, resourceType, resourceID)
	if err != nil {
		acs.logger.Errorw("Error checking resource existence", "error", err, "resourceType", resourceType, "resourceID", resourceID)
		return nil, err
	}
	if !exists {
		return nil, errors.ErrNotFound
	}

	// Require manage permission
	if err := acs.CanUserAccess(db, resourceType, resourceID, currentUser, types.PermissionManage); err != nil {
		if !errors.Is(err, errors.ErrForbidden) {
			acs.logger.Errorw("Error checking permission", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", currentUser.ID)
		}
		return nil, err
	}

	// Use repository-level query to fetch assignable teachers without access, with pagination and sorting
	usersModels, total, err := acs.accessControlRepository.GetAssignableUsers(db, resourceType, resourceID, params.Limit, params.Offset, params.Sort)
	if err != nil {
		acs.logger.Errorw("Error fetching assignable users", "error", err, "resourceType", resourceType, "resourceID", resourceID, "limit", params.Limit, "offset", params.Offset, "sort", params.Sort)
		return nil, err
	}

	assignable := make([]schemas.User, 0, len(usersModels))
	for i := range usersModels {
		u := usersModels[i]
		assignable = append(assignable, schemas.User{
			ID:        u.ID,
			Name:      u.Name,
			Surname:   u.Surname,
			Email:     u.Email,
			Username:  u.Username,
			Role:      u.Role,
			CreatedAt: u.CreatedAt,
		})
	}
	resp := schemas.NewPaginatedResult(assignable, params.Offset, params.Limit, total)
	return &resp, nil
}

// UpdateCollaborator updates a collaborator's permission.
func (acs *accessControlService) UpdateCollaborator(db database.Database, currentUser *schemas.User, resourceType types.ResourceType, resourceID int64, userID int64, permission types.Permission) error {
	// Cannot assign owner via update
	if permission == types.PermissionOwner {
		return errors.ErrCannotAssignOwner
	}

	err := acs.CanUserAccess(db, resourceType, resourceID, currentUser, types.PermissionManage)
	if err != nil {
		if !errors.Is(err, errors.ErrForbidden) {
			acs.logger.Errorw("Error checking permission", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", currentUser.ID)
		}
		return err
	}

	// Check if collaborator exists
	existing, err := acs.accessControlRepository.GetAccess(db, resourceType, resourceID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrNotFound
		}
		acs.logger.Errorw("Error getting access", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", userID)
		return err
	}

	// Owner entry is immutable
	if existing.Permission == types.PermissionOwner {
		return errors.ErrForbidden
	}

	// Update permission
	err = acs.accessControlRepository.UpdatePermission(db, resourceType, resourceID, userID, permission)
	if err != nil {
		acs.logger.Errorw("Error updating collaborator permission", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", userID)
		return err
	}

	return nil
}

// RemoveCollaborator removes a collaborator from a resource.
func (acs *accessControlService) RemoveCollaborator(db database.Database, currentUser *schemas.User, resourceType types.ResourceType, resourceID int64, userID int64) error {
	// Actor must have manage permission on the resource
	err := acs.CanUserAccess(db, resourceType, resourceID, currentUser, types.PermissionManage)
	if err != nil {
		if !errors.Is(err, errors.ErrForbidden) {
			acs.logger.Errorw("Error checking permission", "error", err, "resourceType", resourceType, "resourceID", resourceID, "actorUserID", currentUser.ID)
		}
		return err
	}

	// Fetch target access entry
	access, err := acs.accessControlRepository.GetAccess(db, resourceType, resourceID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrNotFound
		}
		acs.logger.Errorw("Error getting access for removal", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", userID)
		return err
	}

	// Owner permission is immutable and cannot be removed
	if access.Permission == types.PermissionOwner {
		return errors.ErrForbidden
	}

	// Remove collaborator
	if err = acs.accessControlRepository.RemoveAccess(db, resourceType, resourceID, userID); err != nil {
		acs.logger.Errorw("Error removing collaborator", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", userID)
		return err
	}
	return nil
}

// GetUserPermission gets a user's permission level for a resource.
func (acs *accessControlService) GetUserPermission(db database.Database, resourceType types.ResourceType, resourceID int64, userID int64) (types.Permission, error) {
	permission, err := acs.accessControlRepository.GetUserPermission(db, resourceType, resourceID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.ErrNotFound
		}
		acs.logger.Errorw("Error getting user permission", "error", err, "resourceType", resourceType, "resourceID", resourceID, "userID", userID)
		return "", err
	}
	return permission, nil
}

// GrantOwnerAccess auto-grants owner permission to the creator/owner of a resource.
func (acs *accessControlService) GrantOwnerAccess(db database.Database, resourceType types.ResourceType, resourceID int64, ownerID int64) error {
	access := &models.AccessControl{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		UserID:       ownerID,
		Permission:   types.PermissionOwner,
	}

	err := acs.accessControlRepository.AddAccess(db, access)
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
	var userName, userEmail, firstName, lastName string
	if (access.User != models.User{}) {
		userName = access.User.Name
		userEmail = access.User.Email
		firstName = access.User.Name
		lastName = access.User.Surname
	}
	return &schemas.Collaborator{
		UserID:     access.UserID,
		UserName:   userName,
		FirstName:  firstName,
		LastName:   lastName,
		UserEmail:  userEmail,
		Permission: access.Permission,
		AddedAt:    access.CreatedAt.Format(time.RFC3339),
	}
}

func (acs *accessControlService) checkResourceExists(db database.Database, resourceType types.ResourceType, resourceID int64) (bool, error) {
	switch resourceType {
	case types.ResourceTypeContest:
		_, err := acs.contestRepository.Get(db, resourceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	case types.ResourceTypeTask:
		_, err := acs.taskRepository.Get(db, resourceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	default:
		return false, errors.ErrInvalidData
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
