package repository

import (
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/utils"
	"gorm.io/gorm"
)

type UserRepository interface {
	// Create creates a new user and returns the user ID
	Create(db database.Database, user *models.User) (int64, error)
	// Edit updates the user information by setting the new values
	Edit(db database.Database, user *models.User) error
	// GetAll returns users matching the optional search (name, surname, email) and role filters,
	// with pagination, sorting and the total count of matching users
	// When deleted is true only soft-deleted users are returned, otherwise only active ones.
	GetAll(db database.Database, limit, offset int, sort, search, role string, deleted bool) ([]models.User, int64, error)
	// Get returns a user by ID
	Get(db database.Database, userID int64) (*models.User, error)
	// GetByEmail returns a user by email
	GetByEmail(db database.Database, email string) (*models.User, error)
	// IsEmailDeleted reports whether a soft-deleted user has this email
	IsEmailDeleted(db database.Database, email string) (bool, error)
	// Delete soft-deletes a user by setting deleted_at
	Delete(db database.Database, userID int64) error
	// Restore clears deleted_at of a soft-deleted user; returns gorm.ErrRecordNotFound if there is none
	Restore(db database.Database, userID int64) error
}

type userRepository struct {
}

func (ur *userRepository) Create(db database.Database, user *models.User) (int64, error) {
	tx := db.GetInstance()
	err := tx.Model(&models.User{}).Create(user).Error
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (ur *userRepository) Get(db database.Database, userID int64) (*models.User, error) {
	tx := db.GetInstance()
	user := &models.User{}
	err := tx.Model(&models.User{}).Where("id = ?", userID).Take(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *userRepository) GetByEmail(db database.Database, email string) (*models.User, error) {
	tx := db.GetInstance()
	user := &models.User{}
	err := tx.Model(&models.User{}).Where("email = ?", email).First(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *userRepository) GetAll(db database.Database, limit, offset int, sort, search, role string, deleted bool) ([]models.User, int64, error) {
	tx := db.GetInstance()
	users := &[]models.User{}
	var totalCount int64

	filter := func(q *gorm.DB) *gorm.DB {
		if deleted {
			q = q.Unscoped().Where("deleted_at IS NOT NULL")
		}
		if search != "" {
			pattern := utils.ContainsPattern(search)
			q = q.Where("CONCAT(name, ' ', surname) ILIKE ? OR email ILIKE ?", pattern, pattern)
		}
		if role != "" {
			q = q.Where("role = ?", role)
		}
		return q
	}

	// Get total count of matching users (without pagination)
	if err := tx.Model(&models.User{}).Scopes(filter).Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	paginatedTx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	if err := paginatedTx.Model(&models.User{}).Scopes(filter).Find(users).Error; err != nil {
		return nil, 0, err
	}

	return *users, totalCount, nil
}

func (ur *userRepository) IsEmailDeleted(db database.Database, email string) (bool, error) {
	var count int64
	err := db.GetInstance().Unscoped().Model(&models.User{}).
		Where("email = ? AND deleted_at IS NOT NULL", email).Count(&count).Error
	return count > 0, err
}

func (ur *userRepository) Delete(db database.Database, userID int64) error {
	res := db.GetInstance().Delete(&models.User{}, userID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (ur *userRepository) Restore(db database.Database, userID int64) error {
	res := db.GetInstance().Unscoped().Model(&models.User{}).
		Where("id = ? AND deleted_at IS NOT NULL", userID).Update("deleted_at", nil)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (ur *userRepository) Edit(db database.Database, user *models.User) error {
	tx := db.GetInstance()
	return tx.Save(user).Error
}

func NewUserRepository() UserRepository {
	return &userRepository{}
}
