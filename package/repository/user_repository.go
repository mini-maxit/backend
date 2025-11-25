package repository

import (
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/utils"
)

type UserRepository interface {
	// Create creates a new user and returns the user ID
	Create(db database.Database, user *models.User) (int64, error)
	// Edit updates the user information by setting the new values
	Edit(db database.Database, user *models.User) error
	// GetAll returns all users with pagination and sorting and total count
	GetAll(db database.Database, limit, offset int, sort string) ([]models.User, int64, error)
	// Get returns a user by ID
	Get(db database.Database, userID int64) (*models.User, error)
	// GetByEmail returns a user by email
	GetByEmail(db database.Database, email string) (*models.User, error)
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

func (ur *userRepository) GetAll(db database.Database, limit, offset int, sort string) ([]models.User, int64, error) {
	tx := db.GetInstance()
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

func (ur *userRepository) Edit(db database.Database, user *models.User) error {
	tx := db.GetInstance()
	return tx.Save(user).Error
}

func NewUserRepository() UserRepository {
	return &userRepository{}
}
