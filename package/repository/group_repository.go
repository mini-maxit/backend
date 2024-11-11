package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type GroupRepository interface {
	CreateGroup(tx *gorm.DB, group models.Group) (int64, error)
	GetGroup(tx *gorm.DB, groupId int64) (*models.Group, error)
	DeleteGroup(tx *gorm.DB, groupId int64) error
}

type GroupRepositoryImpl struct {
}

func (gr *GroupRepositoryImpl) CreateGroup(tx *gorm.DB, group models.Group) (int64, error) {
	err := tx.Create(&group).Error
	if err != nil {
		return 0, err
	}
	return group.Id, nil
}

func (gr *GroupRepositoryImpl) GetGroup(tx *gorm.DB, groupId int64) (*models.Group, error) {
	var group models.Group
	err := tx.Where("id = ?", groupId).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (gr *GroupRepositoryImpl) DeleteGroup(tx *gorm.DB, groupId int64) error {
	err := tx.Where("id = ?", groupId).Delete(&models.Group{}).Error
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
	return &GroupRepositoryImpl{}, nil
}
