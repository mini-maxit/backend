package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/utils"
	"gorm.io/gorm"
)

type UserRepository interface {
	// Create creates a new user and returns the user ID
	Create(tx *gorm.DB, user *models.User) (int64, error)
	// Edit updates the user information by setting the new values
	Edit(tx *gorm.DB, user *models.User) error
	// GetAll returns all users with pagination and sorting and total count
	GetAll(tx *gorm.DB, limit, offset int, sort string) ([]models.User, int64, error)
	// Get returns a user by ID
	Get(tx *gorm.DB, userID int64) (*models.User, error)
	// GetByEmail returns a user by email
	GetByEmail(tx *gorm.DB, email string) (*models.User, error)
}

type userRepository struct {
}

func (ur *userRepository) Create(tx *gorm.DB, user *models.User) (int64, error) {
	err := tx.Model(&models.User{}).Create(user).Error
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (ur *userRepository) Get(tx *gorm.DB, userID int64) (*models.User, error) {
	user := &models.User{}
	err := tx.Model(&models.User{}).Where("id = ?", userID).Take(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *userRepository) GetByEmail(tx *gorm.DB, email string) (*models.User, error) {
	user := &models.User{}
	err := tx.Model(&models.User{}).Where("email = ?", email).First(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *userRepository) GetAll(tx *gorm.DB, limit, offset int, sort string) ([]models.User, int64, error) {
	users := &[]models.User{}
	var totalCount int64

	// Get total count (without pagination)
	if err := tx.Model(&models.User{}).Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	paginatedTx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, 0, err
	}

	if err := paginatedTx.Model(&models.User{}).Find(users).Error; err != nil {
		return nil, 0, err
	}

	return *users, totalCount, nil
}

func (ur *userRepository) Edit(tx *gorm.DB, user *models.User) error {
	return tx.Save(user).Error
}

func NewUserRepository() UserRepository {
	return &userRepository{}
}
