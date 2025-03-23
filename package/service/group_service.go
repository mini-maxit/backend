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
	// AddUsers adds users to a group.
	AddUsers(tx *gorm.DB, current_user schemas.User, groupId int64, userIds []int64) error
	// Create creates a new group.
	Create(tx *gorm.DB, current_user schemas.User, group *schemas.Group) (int64, error)
	// Delete removes a group by its Id.
	Delete(tx *gorm.DB, current_user schemas.User, groupId int64) error
	// DeleteUsers removes users from a group.
	DeleteUsers(tx *gorm.DB, current_user schemas.User, groupId int64, userIds []int64) error
	// Edit updates the information of a group.
	Edit(tx *gorm.DB, current_user schemas.User, groupId int64, editInfo *schemas.EditGroup) (*schemas.Group, error)
	// Get retrieves detailed information about a group by its ID.
	Get(tx *gorm.DB, current_user schemas.User, groupId int64) (*schemas.GroupDetailed, error)
	// GetAll retrieves all groups based on query parameters.
	GetAll(tx *gorm.DB, current_user schemas.User, queryParams map[string]any) ([]schemas.Group, error)
	// GetTasks retrieves all tasks associated with a group.
	GetTasks(tx *gorm.DB, current_user schemas.User, groupId int64) ([]schemas.Task, error)
	// GetUsers retrieves all users associated with a group.
	GetUsers(tx *gorm.DB, current_user schemas.User, groupId int64) ([]schemas.User, error)
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

func (gs *groupService) Create(tx *gorm.DB, current_user schemas.User, group *schemas.Group) (int64, error) {
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

	groupId, err := gs.groupRepository.Create(tx, model)
	if err != nil {
		return -1, err
	}

	return groupId, nil
}

func (gs *groupService) Delete(tx *gorm.DB, current_user schemas.User, groupId int64) error {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return err
	}

	group, err := gs.groupRepository.Get(tx, groupId)
	if err != nil {
		return err
	}

	if current_user.Role == types.UserRoleTeacher && group.CreatedBy != current_user.Id {
		return errors.ErrNotAuthorized
	}

	return gs.groupRepository.Delete(tx, groupId)
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

	model, err := gs.groupRepository.Get(tx, groupId)
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

func (gs *groupService) GetAll(tx *gorm.DB, current_user schemas.User, queryParams map[string]any) ([]schemas.Group, error) {
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
		groups, err = gs.groupRepository.GetAll(tx, int(offset), int(limit), sort)
		if err != nil {
			return nil, err
		}
	case types.UserRoleTeacher:
		groups, err = gs.groupRepository.GetAllForTeacher(tx, current_user.Id, int(offset), int(limit), sort)
		if err != nil {
			return nil, err
		}
	}
	result := make([]schemas.Group, len(groups))
	for i, group := range groups {
		result[i] = *GroupToSchema(&group)
	}

	return result, nil
}

func (gs *groupService) Get(tx *gorm.DB, current_user schemas.User, groupId int64) (*schemas.GroupDetailed, error) {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return nil, err
	}

	group, err := gs.groupRepository.Get(tx, groupId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}

	if current_user.Role == types.UserRoleTeacher && group.CreatedBy != current_user.Id {
		return nil, errors.ErrNotAuthorized
	}

	return GroupToSchemaDetailed(group), nil
}

func (gs *groupService) AddUsers(tx *gorm.DB, current_user schemas.User, groupId int64, userIds []int64) error {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return err
	}

	group, err := gs.groupRepository.Get(tx, groupId)
	if err != nil {
		return err
	}

	if current_user.Role == types.UserRoleTeacher && group.CreatedBy != current_user.Id {
		return errors.ErrNotAuthorized
	}

	for _, userId := range userIds {
		_, err := gs.userRepository.Get(tx, userId)
		if err != nil {
			return err
		}
		exists, err := gs.groupRepository.UserBelongsTo(tx, groupId, userId)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		err = gs.groupRepository.AddUser(tx, groupId, userId)
		if err != nil {
			return err
		}
	}

	return nil
}

func (gs *groupService) DeleteUsers(tx *gorm.DB, current_user schemas.User, groupId int64, userIds []int64) error {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return err
	}

	group, err := gs.groupRepository.Get(tx, groupId)
	if err != nil {
		return errors.ErrNotFound
	}

	if current_user.Role == types.UserRoleTeacher && group.CreatedBy != current_user.Id {
		return errors.ErrNotAuthorized
	}

	for _, userId := range userIds {
		_, err := gs.userRepository.Get(tx, userId)
		if err != nil {
			return errors.ErrUserNotFound
		}
		exists, err := gs.groupRepository.UserBelongsTo(tx, groupId, userId)
		if err != nil {
			return err
		}
		if !exists {
			continue
		}

		err = gs.groupRepository.DeleteUser(tx, groupId, userId)
		if err != nil {
			return err
		}
	}

	return nil
}

func (gs *groupService) GetUsers(tx *gorm.DB, current_user schemas.User, groupId int64) ([]schemas.User, error) {
	err := utils.ValidateRoleAccess(current_user.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher})
	if err != nil {
		return nil, err
	}

	group, err := gs.groupRepository.Get(tx, groupId)
	if err != nil {
		return nil, err
	}

	if current_user.Role == types.UserRoleTeacher && group.CreatedBy != current_user.Id {
		return nil, errors.ErrNotAuthorized
	}

	users, err := gs.groupRepository.GetUsers(tx, groupId)
	if err != nil {
		return nil, err
	}

	result := make([]schemas.User, len(users))
	for i, user := range users {
		result[i] = *UserToSchema(&user)
	}

	return result, nil
}

func (gs *groupService) GetTasks(tx *gorm.DB, current_user schemas.User, groupId int64) ([]schemas.Task, error) {
	group, err := gs.groupRepository.Get(tx, groupId)
	if err != nil {
		return nil, err
	}

	if current_user.Role == types.UserRoleTeacher && group.CreatedBy != current_user.Id {
		return nil, errors.ErrNotAuthorized
	} else {
		exists, err := gs.groupRepository.UserBelongsTo(tx, groupId, current_user.Id)
		if err != nil {
			return nil, err
		}
		if current_user.Role == types.UserRoleStudent && !exists {
			return nil, errors.ErrNotAuthorized
		}
	}

	tasks, err := gs.groupRepository.GetTasks(tx, groupId)
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
		Id:        model.Id,
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
		Id:        model.Id,
		Name:      model.Name,
		CreatedBy: model.CreatedBy,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}
func NewGroupService(groupRepository repository.GroupRepository, userRepository repository.UserRepository, userService UserService) GroupService {
	return &groupService{
		groupRepository: groupRepository,
		userRepository:  userRepository,
		userService:     userService,
	}
}
