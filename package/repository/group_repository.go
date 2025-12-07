package repository

import (
	"fmt"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/utils"
)

type GroupRepository interface {
	// AddUser adds user to group. If user already belongs to group, returns error
	AddUser(db database.Database, groupID int64, userID int64) error
	// Create new group
	Create(db database.Database, group *models.Group) (int64, error)
	// Delete group by id
	Delete(db database.Database, groupID int64) error
	// DeleteUser deletes user from group. If user does not belong to group, returns error
	DeleteUser(db database.Database, groupID int64, userID int64) error
	// Edit group by id, replace all fields with values from function argument
	Edit(db database.Database, groupID int64, group *models.Group) (*models.Group, error)
	// Get group by id
	Get(db database.Database, groupID int64) (*models.Group, error)
	// GetAll returns all groups with pagination and sorting
	GetAll(db database.Database, offset int, limit int, sort string) ([]models.Group, error)
	// GetAllForTeacher returns all groups created by teacher with pagination and sorting
	GetAllForTeacher(db database.Database, teacherID int64, offset int, limit int, sort string) ([]models.Group, error)
	// GetUsers returns all users that belong to group
	GetUsers(db database.Database, groupID int64) ([]models.User, error)
	// UserBelongsTo checks if user belongs to group
	UserBelongsTo(db database.Database, groupID int64, userID int64) (bool, error)
}

type groupRepository struct {
}

func (gr *groupRepository) Create(db database.Database, group *models.Group) (int64, error) {
	tx := db.GetInstance()
	err := tx.Create(group).Error
	if err != nil {
		return 0, err
	}
	return group.ID, nil
}

func (gr *groupRepository) Get(db database.Database, groupID int64) (*models.Group, error) {
	tx := db.GetInstance()
	var group models.Group
	err := tx.Where("id = ?", groupID).Preload("Users").First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (gr *groupRepository) Delete(db database.Database, groupID int64) error {
	tx := db.GetInstance()
	err := tx.Where("id = ?", groupID).Delete(&models.Group{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (gr *groupRepository) Edit(db database.Database, groupID int64, group *models.Group) (*models.Group, error) {
	tx := db.GetInstance()
	err := tx.Model(&models.Group{}).Where("id = ?", groupID).Updates(group).Error
	if err != nil {
		return nil, err
	}

	return gr.Get(db, groupID)
}

func (gr *groupRepository) GetAll(db database.Database, offset int, limit int, sort string) ([]models.Group, error) {
	tx := db.GetInstance()
	var groups []models.Group
	paginatedTx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, err
	}
	err = paginatedTx.Model(&models.Group{}).Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (gr *groupRepository) GetAllForTeacher(
	db database.Database,
	userID int64,
	offset, limit int,
	sort string,
) ([]models.Group, error) {
	tx := db.GetInstance()
	paginatedTx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, err
	}

	var groups []models.Group
	err = paginatedTx.Model(&models.Group{}).
		Where("created_by = ?", userID).
		Preload("Tasks").
		Preload("Users").
		Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (gr *groupRepository) GetUsers(db database.Database, groupID int64) ([]models.User, error) {
	tx := db.GetInstance()
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

func (gr *groupRepository) AddUser(db database.Database, groupID int64, userID int64) error {
	tx := db.GetInstance()
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

func (gr *groupRepository) DeleteUser(db database.Database, groupID int64, userID int64) error {
	tx := db.GetInstance()
	err := tx.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&models.UserGroup{}).Error
	return err
}

func (gr *groupRepository) UserBelongsTo(db database.Database, groupID int64, userID int64) (bool, error) {
	tx := db.GetInstance()
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
