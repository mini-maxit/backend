package repository

import (
	"fmt"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/utils"
	"gorm.io/gorm"
)

type GroupRepository interface {
	// AddUser adds user to group. If user already belongs to group, returns error
	AddUser(tx *gorm.DB, groupID int64, userID int64) error
	// Create new group
	Create(tx *gorm.DB, group *models.Group) (int64, error)
	// Delete group by id
	Delete(tx *gorm.DB, groupID int64) error
	// DeleteUser deletes user from group. If user does not belong to group, returns error
	DeleteUser(tx *gorm.DB, groupID int64, userID int64) error
	// Edit group by id, replace all fields with values from function argument
	Edit(tx *gorm.DB, groupID int64, group *models.Group) (*models.Group, error)
	// Get group by id
	Get(tx *gorm.DB, groupID int64) (*models.Group, error)
	// GetAll returns all groups with pagination and sorting
	GetAll(tx *gorm.DB, offset int, limit int, sort string) ([]models.Group, error)
	// GetAllForTeacher returns all groups created by teacher with pagination and sorting
	GetAllForTeacher(tx *gorm.DB, teacherID int64, offset int, limit int, sort string) ([]models.Group, error)
	// GetUsers returns all users that belong to group
	GetUsers(tx *gorm.DB, groupID int64) ([]models.User, error)
	// UserBelongsTo checks if user belongs to group
	UserBelongsTo(tx *gorm.DB, groupID int64, userID int64) (bool, error)
}

type groupRepository struct {
}

func (gr *groupRepository) Create(tx *gorm.DB, group *models.Group) (int64, error) {
	err := tx.Create(group).Error
	if err != nil {
		return 0, err
	}
	return group.ID, nil
}

func (gr *groupRepository) Get(tx *gorm.DB, groupID int64) (*models.Group, error) {
	var group models.Group
	err := tx.Where("id = ?", groupID).Preload("Tasks").Preload("Users").First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (gr *groupRepository) Delete(tx *gorm.DB, groupID int64) error {
	err := tx.Where("id = ?", groupID).Delete(&models.Group{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (gr *groupRepository) Edit(tx *gorm.DB, groupID int64, group *models.Group) (*models.Group, error) {
	err := tx.Model(&models.Group{}).Where("id = ?", groupID).Updates(group).Error
	if err != nil {
		return nil, err
	}

	return gr.Get(tx, groupID)
}

func (gr *groupRepository) GetAll(tx *gorm.DB, offset int, limit int, sort string) ([]models.Group, error) {
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

func (gr *groupRepository) GetAllForTeacher(
	tx *gorm.DB,
	userID int64,
	offset, limit int,
	sort string,
) ([]models.Group, error) {
	tx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, err
	}

	var groups []models.Group
	err = tx.Model(&models.Group{}).
		Where("created_by = ?", userID).
		Preload("Tasks").
		Preload("Users").
		Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (gr *groupRepository) GetUsers(tx *gorm.DB, groupID int64) ([]models.User, error) {
	var users []models.User
	err := tx.Model(&models.User{}).
		Joins(fmt.Sprintf("JOIN %s ON user_groups.user_id = users.id", database.ResolveTableName(tx, &models.UserGroup{}))).
		Where("user_groups.group_id = ?", groupID).
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (gr *groupRepository) AddUser(tx *gorm.DB, groupID int64, userID int64) error {
	userGroup := &models.UserGroup{
		GroupID: groupID,
		UserID:  userID,
	}
	err := tx.Create(userGroup).Error
	if err != nil {
		return err
	}
	return nil
}

func (gr *groupRepository) DeleteUser(tx *gorm.DB, groupID int64, userID int64) error {
	err := tx.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&models.UserGroup{}).Error
	return err
}

func (gr *groupRepository) UserBelongsTo(tx *gorm.DB, groupID int64, userID int64) (bool, error) {
	var count int64
	err := tx.Model(&models.UserGroup{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func NewGroupRepository() GroupRepository {
	return &groupRepository{}
}
