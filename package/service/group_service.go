package service

import (
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

type GroupService interface {
	// AddUsers adds users to a group.
	AddUsers(db database.Database, currentUser schemas.User, groupID int64, userIDs []int64) error
	// Create creates a new group.
	Create(db database.Database, currentUser schemas.User, group *schemas.Group) (int64, error)
	// Delete removes a group by its ID.
	Delete(db database.Database, currentUser schemas.User, groupID int64) error
	// DeleteUsers removes users from a group.
	DeleteUsers(db database.Database, currentUser schemas.User, groupID int64, userIDs []int64) error
	// Edit updates the information of a group.
	Edit(db database.Database, currentUser schemas.User, groupID int64, editInfo *schemas.EditGroup) (*schemas.Group, error)
	// Get retrieves detailed information about a group by its ID.
	Get(db database.Database, currentUser schemas.User, groupID int64) (*schemas.GroupDetailed, error)
	// GetAll retrieves all groups based on query parameters.
	GetAll(db database.Database, currentUser schemas.User, paginationParams schemas.PaginationParams) ([]schemas.Group, error)
	// GetUsers retrieves all users associated with a group.
	GetUsers(db database.Database, currentUser schemas.User, groupID int64) ([]schemas.User, error)
}

const defaultGroupSort = "created_at:desc"

type groupService struct {
	groupRepository      repository.GroupRepository
	userRepository       repository.UserRepository
	userService          UserService
	accessControlService AccessControlService
	logger               *zap.SugaredLogger
}

func (gs *groupService) Create(db database.Database, currentUser schemas.User, group *schemas.Group) (int64, error) {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return 0, err
	}

	validate, err := utils.NewValidator()
	if err != nil {
		return -1, err
	}
	if err := validate.Struct(group); err != nil {
		return 0, err
	}

	model := &models.Group{
		Name:      group.Name,
		CreatedBy: group.CreatedBy,
	}

	groupID, err := gs.groupRepository.Create(db, model)
	if err != nil {
		return -1, err
	}

	// Grant owner access to the creator
	err = gs.accessControlService.GrantOwnerAccess(db, types.ResourceTypeGroup, groupID, currentUser.ID)
	if err != nil {
		return -1, err
	}

	return groupID, nil
}

func (gs *groupService) Delete(db database.Database, currentUser schemas.User, groupID int64) error {
	// Check if group exists
	_, err := gs.groupRepository.Get(db, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrGroupNotFound
		}
		return err
	}

	// Check if user has manage permission (owner can delete)
	if currentUser.Role != types.UserRoleAdmin {
		err = gs.accessControlService.CanUserAccess(db, types.ResourceTypeGroup, groupID, &currentUser, types.PermissionOwner)
		if err != nil {
			return err
		}
	}

	return gs.groupRepository.Delete(db, groupID)
}

func (gs *groupService) Edit(
	db database.Database,
	currentUser schemas.User,
	groupID int64,
	editInfo *schemas.EditGroup,
) (*schemas.Group, error) {
	validate, err := utils.NewValidator()
	if err != nil {
		return nil, err
	}
	if err := validate.Struct(editInfo); err != nil {
		return nil, err
	}

	model, err := gs.groupRepository.Get(db, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrGroupNotFound
		}
		return nil, err
	}

	// Check if user has edit permission
	if currentUser.Role != types.UserRoleAdmin {
		err = gs.accessControlService.CanUserAccess(db, types.ResourceTypeGroup, groupID, &currentUser, types.PermissionEdit)
		if err != nil {
			return nil, err
		}
	}

	gs.updateModel(model, editInfo)

	newModel, err := gs.groupRepository.Edit(db, groupID, model)
	if err != nil {
		return nil, err
	}
	return GroupToSchema(newModel), nil
}

func (gs *groupService) GetAll(
	db database.Database,
	currentUser schemas.User,
	paginationParams schemas.PaginationParams,
) ([]schemas.Group, error) {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return nil, err
	}

	if paginationParams.Sort == "" {
		paginationParams.Sort = defaultGroupSort
	}
	var groups []models.Group

	switch currentUser.Role {
	case types.UserRoleAdmin:
		groups, err = gs.groupRepository.GetAll(db, paginationParams.Offset, paginationParams.Limit, paginationParams.Sort)
		if err != nil {
			return nil, err
		}
	case types.UserRoleTeacher:
		groups, err = gs.groupRepository.GetAllForTeacher(db, currentUser.ID, paginationParams.Offset, paginationParams.Limit, paginationParams.Sort)
		if err != nil {
			return nil, err
		}
	case types.UserRoleStudent:
		return nil, errors.ErrForbidden
	default:
		return nil, errors.ErrForbidden
	}
	result := make([]schemas.Group, len(groups))
	for i, group := range groups {
		result[i] = *GroupToSchema(&group)
	}

	return result, nil
}

func (gs *groupService) Get(db database.Database, currentUser schemas.User, groupID int64) (*schemas.GroupDetailed, error) {
	group, err := gs.groupRepository.Get(db, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrGroupNotFound
		}
		return nil, err
	}

	// Check if user has edit permission (to view group details)
	if currentUser.Role != types.UserRoleAdmin {
		err = gs.accessControlService.CanUserAccess(db, types.ResourceTypeGroup, groupID, &currentUser, types.PermissionEdit)
		if err != nil {
			return nil, err
		}
	}

	return GroupToSchemaDetailed(group), nil
}

func (gs *groupService) AddUsers(db database.Database, currentUser schemas.User, groupID int64, userIDs []int64) error {
	_, err := gs.groupRepository.Get(db, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrGroupNotFound
		}
		return err
	}

	// Check if user has edit permission
	if currentUser.Role != types.UserRoleAdmin {
		err = gs.accessControlService.CanUserAccess(db, types.ResourceTypeGroup, groupID, &currentUser, types.PermissionEdit)
		if err != nil {
			return err
		}
	}

	for _, userID := range userIDs {
		_, err := gs.userRepository.Get(db, userID)
		if err != nil {
			return err
		}
		exists, err := gs.groupRepository.UserBelongsTo(db, groupID, userID)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		err = gs.groupRepository.AddUser(db, groupID, userID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (gs *groupService) DeleteUsers(db database.Database, currentUser schemas.User, groupID int64, userIDs []int64) error {
	_, err := gs.groupRepository.Get(db, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrGroupNotFound
		}
		return err
	}

	// Check if user has edit permission
	if currentUser.Role != types.UserRoleAdmin {
		err = gs.accessControlService.CanUserAccess(db, types.ResourceTypeGroup, groupID, &currentUser, types.PermissionEdit)
		if err != nil {
			return err
		}
	}

	for _, userID := range userIDs {
		_, err := gs.userRepository.Get(db, userID)
		if err != nil {
			return errors.ErrUserNotFound
		}
		exists, err := gs.groupRepository.UserBelongsTo(db, groupID, userID)
		if err != nil {
			return err
		}
		if !exists {
			continue
		}

		err = gs.groupRepository.DeleteUser(db, groupID, userID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (gs *groupService) GetUsers(db database.Database, currentUser schemas.User, groupID int64) ([]schemas.User, error) {
	_, err := gs.groupRepository.Get(db, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrGroupNotFound
		}
		return nil, err
	}

	// Check if user has edit permission
	if currentUser.Role != types.UserRoleAdmin {
		err = gs.accessControlService.CanUserAccess(db, types.ResourceTypeGroup, groupID, &currentUser, types.PermissionEdit)
		if err != nil {
			return nil, err
		}
	}

	users, err := gs.groupRepository.GetUsers(db, groupID)
	if err != nil {
		return nil, err
	}

	result := make([]schemas.User, len(users))
	for i, user := range users {
		result[i] = *UserToSchema(&user)
	}

	return result, nil
}
func (gs *groupService) updateModel(model *models.Group, editInfo *schemas.EditGroup) {
	if editInfo.Name != nil {
		model.Name = *editInfo.Name
	}
}

func GroupToSchemaDetailed(model *models.Group) *schemas.GroupDetailed {
	users := make([]schemas.User, len(model.Users))
	for i, user := range model.Users {
		users[i] = *UserToSchema(&user)
	}

	return &schemas.GroupDetailed{
		ID:        model.ID,
		Name:      model.Name,
		CreatedBy: model.CreatedBy,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		Users:     users,
	}
}

func GroupToSchema(model *models.Group) *schemas.Group {
	return &schemas.Group{
		ID:        model.ID,
		Name:      model.Name,
		CreatedBy: model.CreatedBy,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}
func NewGroupService(
	groupRepository repository.GroupRepository,
	userRepository repository.UserRepository,
	userService UserService,
	accessControlService AccessControlService,
) GroupService {
	return &groupService{
		groupRepository:      groupRepository,
		userRepository:       userRepository,
		userService:          userService,
		accessControlService: accessControlService,
		logger:               utils.NewNamedLogger("group_service"),
	}
}
