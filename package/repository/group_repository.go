package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/utils"
	"gorm.io/gorm"
)

type GroupRepository interface {
	CreateGroup(tx *gorm.DB, group *models.Group) (int64, error)
	GetGroup(tx *gorm.DB, groupId int64) (*models.Group, error)
	DeleteGroup(tx *gorm.DB, groupId int64) error
	Edit(tx *gorm.DB, groupId int64, group *models.Group) (*models.Group, error)
	GetAllGroup(tx *gorm.DB, offset int, limit int, sort string) ([]models.Group, error)
	GetAllGroupForTeacher(tx *gorm.DB, teacherId int64, offset int, limit int, sort string) ([]models.Group, error)
	AddUserToGroup(tx *gorm.DB, groupId int64, userId int64) error
	GetGroupUsers(tx *gorm.DB, groupId int64) ([]models.User, error)
}

type groupRepository struct {
}

func (gr *groupRepository) CreateGroup(tx *gorm.DB, group *models.Group) (int64, error) {
	err := tx.Create(group).Error
	if err != nil {
		return 0, err
	}
	return group.Id, nil
}

func (gr *groupRepository) GetGroup(tx *gorm.DB, groupId int64) (*models.Group, error) {
	var group models.Group
	err := tx.Where("id = ?", groupId).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (gr *groupRepository) DeleteGroup(tx *gorm.DB, groupId int64) error {
	err := tx.Where("id = ?", groupId).Delete(&models.Group{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (gr *groupRepository) Edit(tx *gorm.DB, groupId int64, group *models.Group) (*models.Group, error) {
	err := tx.Model(&models.Group{}).Where("id = ?", groupId).Updates(group).Error
	if err != nil {
		return nil, err
	}

	return gr.GetGroup(tx, groupId)
}

func (gr *groupRepository) GetAllGroup(tx *gorm.DB, offset int, limit int, sort string) ([]models.Group, error) {
	var groups []models.Group
	tx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, err
	}
	err = tx.Model(&models.Group{}).Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (gr *groupRepository) GetAllGroupForTeacher(tx *gorm.DB, teacherId int64, offset int, limit int, sort string) ([]models.Group, error) {
	tx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, err
	}

	var groups []models.Group
	err = tx.Model(&models.Group{}).
		Where("created_by = ?", teacherId).
		Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (gr *groupRepository) GetGroupUsers(tx *gorm.DB, groupId int64) ([]models.User, error) {
	var users []models.User
	err := tx.Model(&models.User{}).
		Joins("JOIN user_groups ON user_groups.user_id = users.id").
		Where("user_groups.group_id = ?", groupId).
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (gr *groupRepository) AddUserToGroup(tx *gorm.DB, groupId int64, userId int64) error {
	userGroup := &models.UserGroup{
		GroupId: groupId,
		UserId:  userId,
	}
	err := tx.Create(userGroup).Error
	if err != nil {
		return err
	}
	return nil
}

func NewGroupRepository(db *gorm.DB) (GroupRepository, error) {
	tables := []interface{}{&models.Group{}, &models.UserGroup{}, &models.TaskGroup{}}
	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			err := db.Migrator().CreateTable(table)
			if err != nil {
				return nil, err
			}
		}
	}
	return &groupRepository{}, nil
}
