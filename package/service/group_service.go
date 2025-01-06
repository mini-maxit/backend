package service

import (
	"fmt"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"gorm.io/gorm"
)

type GroupService interface {
	CreateGroup(tx *gorm.DB, group *schemas.Group) (int64, error)
	DeleteGroup(tx *gorm.DB, groupId int64) error
	Edit(tx *gorm.DB, groupId int64, editInfo *schemas.EditGroup) (*schemas.Group, error)
	GetAllGroup(tx *gorm.DB, queryParams map[string]string) ([]schemas.Group, error)
	GetGroup(tx *gorm.DB, groupId int64) (*schemas.Group, error)
}

var (
	ErrGroupNotFound      = fmt.Errorf("group not found")
	ErrInvalidLimitParam  = fmt.Errorf("invalid limit parameter")
	ErrInvalidOffsetParam = fmt.Errorf("invalid offset parameter")
)

type GroupServiceImpl struct {
	groupRepository repository.GroupRepository
}

func (gs *GroupServiceImpl) CreateGroup(tx *gorm.DB, group *schemas.Group) (int64, error) {
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

func (gs *GroupServiceImpl) DeleteGroup(tx *gorm.DB, groupId int64) error {
	return gs.groupRepository.DeleteGroup(tx, groupId)
}

func (gs *GroupServiceImpl) Edit(tx *gorm.DB, groupId int64, editInfo *schemas.EditGroup) (*schemas.Group, error) {
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
	gs.updateModel(model, editInfo)

	newModel, err := gs.groupRepository.Edit(tx, groupId, model)
	if err != nil {
		return nil, err
	}
	return gs.modelToSchema(newModel), nil

}

func (gs *GroupServiceImpl) GetAllGroup(tx *gorm.DB, queryParams map[string]string) ([]schemas.Group, error) {
	limit, err := utils.GetLimit(queryParams["limit"])
	if err != nil {
		return nil, ErrInvalidLimitParam
	}
	offset, err := utils.GetOffset(queryParams["offset"])
	if err != nil {
		return nil, ErrInvalidOffsetParam
	}
	sort := utils.GetSort(queryParams["sort"])
	groups, err := gs.groupRepository.GetAllGroup(tx, offset, limit, sort)
	if err != nil {
		return nil, err
	}

	var result []schemas.Group
	for _, group := range groups {
		result = append(result, *gs.modelToSchema(&group))
	}

	return result, nil
}

func (gs *GroupServiceImpl) GetGroup(tx *gorm.DB, groupId int64) (*schemas.Group, error) {
	group, err := gs.groupRepository.GetGroup(tx, groupId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}

	return gs.modelToSchema(group), nil
}

func (gs *GroupServiceImpl) updateModel(model *models.Group, editInfo *schemas.EditGroup) {
	if editInfo.Name != nil {
		model.Name = *editInfo.Name
	}
}

func (gs *GroupServiceImpl) modelToSchema(model *models.Group) *schemas.Group {
	return &schemas.Group{
		Id:        model.Id,
		Name:      model.Name,
		CreatedBy: model.CreatedBy,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func NewGroupService(groupRepository repository.GroupRepository) GroupService {
	return &GroupServiceImpl{
		groupRepository: groupRepository,
	}
}
