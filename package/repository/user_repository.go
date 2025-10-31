package repository

import (
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
)

type UserRepository interface {
	// Create creates a new user and returns the user ID
	Create(tx database.Database, user *models.User) (int64, error)
	// Edit updates the user information by setting the new values
	Edit(tx database.Database, user *models.User) error
	// GetAll returns all users with pagination and sorting
	GetAll(tx database.Database, limit, offset int, sort string) ([]models.User, error)
	// Get returns a user by ID
	Get(tx database.Database, userID int64) (*models.User, error)
	//
	GetByEmail(tx database.Database, email string) (*models.User, error)
}

type userRepository struct {
}

func (ur *userRepository) Create(tx database.Database, user *models.User) (int64, error) {
	err := tx.Model(&models.User{}).Create(user).Error()
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (ur *userRepository) Get(tx database.Database, userID int64) (*models.User, error) {
	user := &models.User{}
	err := tx.Model(&models.User{}).Where("id = ?", userID).Take(user).Error()
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *userRepository) GetByEmail(tx database.Database, email string) (*models.User, error) {
	user := &models.User{}
	err := tx.Model(&models.User{}).Where("email = ?", email).First(user).Error()
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *userRepository) GetAll(tx database.Database, limit, offset int, sort string) ([]models.User, error) {
	users := &[]models.User{}
	tx = tx.ApplyPaginationAndSort(limit, offset, sort)

	err := tx.Model(&models.User{}).Find(users).Error()
	if err != nil {
		return nil, err
	}
	return *users, nil
}

func (ur *userRepository) Edit(tx database.Database, user *models.User) error {
	err := tx.Save(user).Error()
	return err
}

func NewUserRepository() UserRepository {
	return &userRepository{}
}
