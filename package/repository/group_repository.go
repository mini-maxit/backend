package repository

import (
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
)

type GroupRepository interface {
	// AddUser adds user to group. If user already belongs to group, returns error
	AddUser(tx database.Database, groupID int64, userID int64) error
	// Create new group
	Create(tx database.Database, group *models.Group) (int64, error)
	// Delete group by id
	Delete(tx database.Database, groupID int64) error
	// DeleteUser deletes user from group. If user does not belong to group, returns error
	DeleteUser(tx database.Database, groupID int64, userID int64) error
	// Edit group by id, replace all fields with values from function argument
	Edit(tx database.Database, groupID int64, group *models.Group) (*models.Group, error)
	// Get group by id
	Get(tx database.Database, groupID int64) (*models.Group, error)
	// GetAll returns all groups with pagination and sorting
	GetAll(tx database.Database, offset int, limit int, sort string) ([]models.Group, error)
	// GetAllForTeacher returns all groups created by teacher with pagination and sorting
	GetAllForTeacher(tx database.Database, teacherID int64, offset int, limit int, sort string) ([]models.Group, error)
	// GetTasks returns all tasks assigned to group
	GetTasks(tx database.Database, groupID int64) ([]models.Task, error)
	// GetUsers returns all users that belong to group
	GetUsers(tx database.Database, groupID int64) ([]models.User, error)
	// UserBelongsTo checks if user belongs to group
	UserBelongsTo(tx database.Database, groupID int64, userID int64) (bool, error)
}

type groupRepository struct {
}

func (gr *groupRepository) Create(tx database.Database, group *models.Group) (int64, error) {
	err := tx.Create(group).Error()
	if err != nil {
		return 0, err
	}
	return group.ID, nil
}

func (gr *groupRepository) Get(tx database.Database, groupID int64) (*models.Group, error) {
	var group models.Group
	err := tx.Where("id = ?", groupID).Preload("Tasks").Preload("Users").First(&group).Error()
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (gr *groupRepository) Delete(tx database.Database, groupID int64) error {
	err := tx.Where("id = ?", groupID).Delete(&models.Group{}).Error()
	if err != nil {
		return err
	}
	return nil
}

func (gr *groupRepository) Edit(tx database.Database, groupID int64, group *models.Group) (*models.Group, error) {
	err := tx.Model(&models.Group{}).Where("id = ?", groupID).Updates(group).Error()
	if err != nil {
		return nil, err
	}

	return gr.Get(tx, groupID)
}

func (gr *groupRepository) GetAll(tx database.Database, offset int, limit int, sort string) ([]models.Group, error) {
	var groups []models.Group
	tx = tx.ApplyPaginationAndSort(limit, offset, sort)
	err := tx.Model(&models.Group{}).Find(&groups).Error()
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (gr *groupRepository) GetAllForTeacher(
	tx database.Database,
	userID int64,
	offset, limit int,
	sort string,
) ([]models.Group, error) {
	tx = tx.ApplyPaginationAndSort(limit, offset, sort)

	var groups []models.Group
	err := tx.Model(&models.Group{}).
		Where("created_by = ?", userID).
		Preload("Tasks").
		Preload("Users").
		Find(&groups).Error()
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (gr *groupRepository) GetUsers(tx database.Database, groupID int64) ([]models.User, error) {
	var users []models.User
	err := tx.Model(&models.User{}).
		Join("JOIN", &models.UserGroup{}, "user_groups.user_id = users.id").
		Where("user_groups.group_id = ?", groupID).
		Find(&users).Error()
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (gr *groupRepository) AddUser(tx database.Database, groupID int64, userID int64) error {
	userGroup := &models.UserGroup{
		GroupID: groupID,
		UserID:  userID,
	}
	err := tx.Create(userGroup).Error()
	if err != nil {
		return err
	}
	return nil
}

func (gr *groupRepository) DeleteUser(tx database.Database, groupID int64, userID int64) error {
	err := tx.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&models.UserGroup{}).Error()
	return err
}

func (gr *groupRepository) UserBelongsTo(tx database.Database, groupID int64, userID int64) (bool, error) {
	var count int64
	err := tx.Model(&models.UserGroup{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// TODO: refactor this, maybe this should be in task repository.
func (gr *groupRepository) GetTasks(tx database.Database, groupID int64) ([]models.Task, error) {
	group := &models.Group{}
	err := tx.
		Where("id = ?", groupID).
		Preload("Tasks").
		First(group).Error()
	if err != nil {
		return nil, err
	}
	return group.Tasks, nil
}

func NewGroupRepository() GroupRepository {
	return &groupRepository{}
}
