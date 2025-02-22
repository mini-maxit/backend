package service

import (
	"fmt"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"gorm.io/gorm"
)

type GroupService interface {
	CreateGroup(tx *gorm.DB, current_user schemas.User, group *schemas.Group) (int64, error)
	DeleteGroup(tx *gorm.DB, current_user schemas.User, groupId int64) error
	Edit(tx *gorm.DB, current_user schemas.User, groupId int64, editInfo *schemas.EditGroup) (*schemas.Group, error)
	GetAllGroup(tx *gorm.DB, current_user schemas.User, queryParams map[string]interface{}) ([]schemas.Group, error)
	GetGroup(tx *gorm.DB, current_user schemas.User, groupId int64) (*schemas.Group, error)
	AddUsersToGroup(tx *gorm.DB, current_user schemas.User, groupId int64, userIds []int64) error
	GetGroupUsers(tx *gorm.DB, current_user schemas.User, groupId int64) ([]schemas.User, error)
}

var (
	ErrGroupNotFound      = fmt.Errorf("group not found")
	ErrInvalidLimitParam  = fmt.Errorf("invalid limit parameter")
	ErrInvalidOffsetParam = fmt.Errorf("invalid offset parameter")
)

type groupService struct {
	groupRepository repository.GroupRepository
	userRepository  repository.UserRepository
	userService     UserService
}

func (gs *groupService) CreateGroup(tx *gorm.DB, current_user schemas.User, group *schemas.Group) (int64, error) {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
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

	groupId, err := gs.groupRepository.CreateGroup(tx, model)
	if err != nil {
		return -1, err
	}

	return groupId, nil
}

func (gs *groupService) DeleteGroup(tx *gorm.DB, current_user schemas.User, groupId int64) error {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return err
	}

	group, err := gs.groupRepository.GetGroup(tx, groupId)
	if err != nil {
		return err
	}

	if current_user.Role == types.UserRoleTeacher && group.CreatedBy != current_user.Id {
		return errors.ErrNotAuthorized
	}

	return gs.groupRepository.DeleteGroup(tx, groupId)
}

func (gs *groupService) Edit(tx *gorm.DB, current_user schemas.User, groupId int64, editInfo *schemas.EditGroup) (*schemas.Group, error) {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
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

	model, err := gs.groupRepository.GetGroup(tx, groupId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}
	if current_user.Role == types.UserRoleTeacher && model.CreatedBy != current_user.Id {
		return nil, errors.ErrNotAuthorized
	}

	gs.updateModel(model, editInfo)

	newModel, err := gs.groupRepository.Edit(tx, groupId, model)
	if err != nil {
		return nil, err
	}
	return GroupToSchema(newModel), nil
}

func (gs *groupService) GetAllGroup(tx *gorm.DB, current_user schemas.User, queryParams map[string]interface{}) ([]schemas.Group, error) {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return nil, err
	}

	limit := queryParams["limit"].(uint64)
	offset := queryParams["offset"].(uint64)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = "created_at:desc"
	}
	var groups []models.Group

	switch current_user.Role {
	case types.UserRoleAdmin:
		groups, err = gs.groupRepository.GetAllGroup(tx, int(offset), int(limit), sort)
		if err != nil {
			return nil, err
		}
	case types.UserRoleTeacher:
		groups, err = gs.groupRepository.GetAllGroupForTeacher(tx, current_user.Id, int(offset), int(limit), sort)
		if err != nil {
			return nil, err
		}
	}
	var result []schemas.Group
	for _, group := range groups {
		result = append(result, *GroupToSchema(&group))
	}

	return result, nil
}

func (gs *groupService) GetGroup(tx *gorm.DB, current_user schemas.User, groupId int64) (*schemas.Group, error) {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return nil, err
	}

	group, err := gs.groupRepository.GetGroup(tx, groupId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}

	if current_user.Role == types.UserRoleTeacher && group.CreatedBy != current_user.Id {
		return nil, errors.ErrNotAuthorized
	}

	return GroupToSchema(group), nil
}

func (gs *groupService) AddUsersToGroup(tx *gorm.DB, current_user schemas.User, groupId int64, userIds []int64) error {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return err
	}

	group, err := gs.groupRepository.GetGroup(tx, groupId)
	if err != nil {
		return err
	}

	if current_user.Role == types.UserRoleTeacher && group.CreatedBy != current_user.Id {
		return errors.ErrNotAuthorized
	}

	for _, userId := range userIds {
		_, err := gs.userRepository.GetUser(tx, userId)
		if err != nil {
			return err
		}

		err = gs.groupRepository.AddUserToGroup(tx, groupId, userId)
		if err != nil {
			return err
		}
	}

	return nil
}

func (gs *groupService) GetGroupUsers(tx *gorm.DB, current_user schemas.User, groupId int64) ([]schemas.User, error) {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return nil, err
	}

	group, err := gs.groupRepository.GetGroup(tx, groupId)
	if err != nil {
		return nil, err
	}

	if current_user.Role == types.UserRoleTeacher && group.CreatedBy != current_user.Id {
		return nil, errors.ErrNotAuthorized
	}

	users, err := gs.groupRepository.GetGroupUsers(tx, groupId)
	if err != nil {
		return nil, err
	}

	var result []schemas.User
	for _, user := range users {
		result = append(result, *UserToSchema(&user))
	}

	return result, nil
}

func (gs *groupService) updateModel(model *models.Group, editInfo *schemas.EditGroup) {
	if editInfo.Name != nil {
		model.Name = *editInfo.Name
	}
}

func GroupToSchema(model *models.Group) *schemas.Group {
	tasks := make([]schemas.Task, len(model.Tasks))
	for i, task := range model.Tasks {
		tasks[i] = *TaskToSchema(&task)
	}
	users := make([]schemas.User, len(model.Users))
	for i, user := range model.Users {
		users[i] = *UserToSchema(&user)
	}

	return &schemas.Group{
		Id:        model.Id,
		Name:      model.Name,
		CreatedBy: model.CreatedBy,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		Tasks:     tasks,
		Users:     users,
	}
}

func NewGroupService(groupRepository repository.GroupRepository, userRepository repository.UserRepository, userService UserService) GroupService {
	return &groupService{
		groupRepository: groupRepository,
		userRepository:  userRepository,
		userService:     userService,
	}
}
