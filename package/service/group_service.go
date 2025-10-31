package service

import (
	"errors"
	"github.com/mini-maxit/backend/internal/database"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	myerrors "github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"gorm.io/gorm"
)

type GroupService interface {
	// AddUsers adds users to a group.
	AddUsers(tx *database.DB, currentUser schemas.User, groupID int64, userIDs []int64) error
	// Create creates a new group.
	Create(tx *database.DB, currentUser schemas.User, group *schemas.Group) (int64, error)
	// Delete removes a group by its ID.
	Delete(tx *database.DB, currentUser schemas.User, groupID int64) error
	// DeleteUsers removes users from a group.
	DeleteUsers(tx *database.DB, currentUser schemas.User, groupID int64, userIDs []int64) error
	// Edit updates the information of a group.
	Edit(tx *database.DB, currentUser schemas.User, groupID int64, editInfo *schemas.EditGroup) (*schemas.Group, error)
	// Get retrieves detailed information about a group by its ID.
	Get(tx *database.DB, currentUser schemas.User, groupID int64) (*schemas.GroupDetailed, error)
	// GetAll retrieves all groups based on query parameters.
	GetAll(tx *database.DB, currentUser schemas.User, queryParams map[string]any) ([]schemas.Group, error)
	// GetTasks retrieves all tasks associated with a group.
	GetTasks(tx *database.DB, currentUser schemas.User, groupID int64) ([]schemas.Task, error)
	// GetUsers retrieves all users associated with a group.
	GetUsers(tx *database.DB, currentUser schemas.User, groupID int64) ([]schemas.User, error)
}

const defaultGroupSort = "created_at:desc"

type groupService struct {
	groupRepository repository.GroupRepository
	userRepository  repository.UserRepository
	userService     UserService
}

func (gs *groupService) Create(tx *database.DB, currentUser schemas.User, group *schemas.Group) (int64, error) {
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

	groupID, err := gs.groupRepository.Create(tx, model)
	if err != nil {
		return -1, err
	}

	return groupID, nil
}

func (gs *groupService) Delete(tx *database.DB, currentUser schemas.User, groupID int64) error {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return err
	}

	group, err := gs.groupRepository.Get(tx, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrGroupNotFound
		}
		return err
	}

	if currentUser.Role == types.UserRoleTeacher && group.CreatedBy != currentUser.ID {
		return myerrors.ErrNotAuthorized
	}

	return gs.groupRepository.Delete(tx, groupID)
}

func (gs *groupService) Edit(
	tx *database.DB,
	currentUser schemas.User,
	groupID int64,
	editInfo *schemas.EditGroup,
) (*schemas.Group, error) {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return nil, err
	}

	validate, err := utils.NewValidator()
	if err != nil {
		return nil, err
	}
	if err := validate.Struct(editInfo); err != nil {
		return nil, err
	}

	model, err := gs.groupRepository.Get(tx, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrGroupNotFound
		}
		return nil, err
	}
	if currentUser.Role == types.UserRoleTeacher && model.CreatedBy != currentUser.ID {
		return nil, myerrors.ErrNotAuthorized
	}

	gs.updateModel(model, editInfo)

	newModel, err := gs.groupRepository.Edit(tx, groupID, model)
	if err != nil {
		return nil, err
	}
	return GroupToSchema(newModel), nil
}

func (gs *groupService) GetAll(
	tx *database.DB,
	currentUser schemas.User,
	queryParams map[string]any,
) ([]schemas.Group, error) {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return nil, err
	}

	limit := queryParams["limit"].(int)
	offset := queryParams["offset"].(int)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = defaultGroupSort
	}
	var groups []models.Group

	switch currentUser.Role {
	case types.UserRoleAdmin:
		groups, err = gs.groupRepository.GetAll(tx, offset, limit, sort)
		if err != nil {
			return nil, err
		}
	case types.UserRoleTeacher:
		groups, err = gs.groupRepository.GetAllForTeacher(tx, currentUser.ID, offset, limit, sort)
		if err != nil {
			return nil, err
		}
	case types.UserRoleStudent:
		return nil, myerrors.ErrNotAuthorized
	default:
		return nil, myerrors.ErrNotAuthorized
	}
	result := make([]schemas.Group, len(groups))
	for i, group := range groups {
		result[i] = *GroupToSchema(&group)
	}

	return result, nil
}

func (gs *groupService) Get(tx *database.DB, currentUser schemas.User, groupID int64) (*schemas.GroupDetailed, error) {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return nil, err
	}

	group, err := gs.groupRepository.Get(tx, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrGroupNotFound
		}
		return nil, err
	}

	if currentUser.Role == types.UserRoleTeacher && group.CreatedBy != currentUser.ID {
		return nil, myerrors.ErrNotAuthorized
	}

	return GroupToSchemaDetailed(group), nil
}

func (gs *groupService) AddUsers(tx *database.DB, currentUser schemas.User, groupID int64, userIDs []int64) error {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return err
	}

	group, err := gs.groupRepository.Get(tx, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrGroupNotFound
		}
		return err
	}

	if currentUser.Role == types.UserRoleTeacher && group.CreatedBy != currentUser.ID {
		return myerrors.ErrNotAuthorized
	}

	for _, userID := range userIDs {
		_, err := gs.userRepository.Get(tx, userID)
		if err != nil {
			return err
		}
		exists, err := gs.groupRepository.UserBelongsTo(tx, groupID, userID)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		err = gs.groupRepository.AddUser(tx, groupID, userID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (gs *groupService) DeleteUsers(tx *database.DB, currentUser schemas.User, groupID int64, userIDs []int64) error {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return err
	}

	group, err := gs.groupRepository.Get(tx, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrGroupNotFound
		}
		return err
	}

	if currentUser.Role == types.UserRoleTeacher && group.CreatedBy != currentUser.ID {
		return myerrors.ErrNotAuthorized
	}

	for _, userID := range userIDs {
		_, err := gs.userRepository.Get(tx, userID)
		if err != nil {
			return myerrors.ErrUserNotFound
		}
		exists, err := gs.groupRepository.UserBelongsTo(tx, groupID, userID)
		if err != nil {
			return err
		}
		if !exists {
			continue
		}

		err = gs.groupRepository.DeleteUser(tx, groupID, userID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (gs *groupService) GetUsers(tx *database.DB, currentUser schemas.User, groupID int64) ([]schemas.User, error) {
	err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return nil, err
	}

	group, err := gs.groupRepository.Get(tx, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrGroupNotFound
		}
		return nil, err
	}

	if currentUser.Role == types.UserRoleTeacher && group.CreatedBy != currentUser.ID {
		return nil, myerrors.ErrNotAuthorized
	}

	users, err := gs.groupRepository.GetUsers(tx, groupID)
	if err != nil {
		return nil, err
	}

	result := make([]schemas.User, len(users))
	for i, user := range users {
		result[i] = *UserToSchema(&user)
	}

	return result, nil
}

func (gs *groupService) GetTasks(tx *database.DB, currentUser schemas.User, groupID int64) ([]schemas.Task, error) {
	group, err := gs.groupRepository.Get(tx, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrGroupNotFound
		}
		return nil, err
	}

	switch currentUser.Role {
	case types.UserRoleAdmin:
		break
	case types.UserRoleTeacher:
		if group.CreatedBy != currentUser.ID {
			return nil, myerrors.ErrNotAuthorized
		}
	case types.UserRoleStudent:
		exists, err := gs.groupRepository.UserBelongsTo(tx, groupID, currentUser.ID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, myerrors.ErrNotAuthorized
		}
	}

	tasks, err := gs.groupRepository.GetTasks(tx, groupID)
	if err != nil {
		return nil, err
	}

	result := make([]schemas.Task, len(tasks))
	for i, task := range tasks {
		result[i] = *TaskToSchema(&task)
	}

	return result, nil
}

func (gs *groupService) updateModel(model *models.Group, editInfo *schemas.EditGroup) {
	if editInfo.Name != nil {
		model.Name = *editInfo.Name
	}
}

func GroupToSchemaDetailed(model *models.Group) *schemas.GroupDetailed {
	tasks := make([]schemas.Task, len(model.Tasks))
	for i, task := range model.Tasks {
		tasks[i] = *TaskToSchema(&task)
	}
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
		Tasks:     tasks,
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
) GroupService {
	return &groupService{
		groupRepository: groupRepository,
		userRepository:  userRepository,
		userService:     userService,
	}
}
