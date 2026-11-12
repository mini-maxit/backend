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
	GetAll(db database.Database, limit, offset int, sort, search, role string) ([]models.User, int64, error)
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

func (ur *userRepository) GetAll(db database.Database, limit, offset int, sort, search, role string) ([]models.User, int64, error) {
	tx := db.GetInstance()
	users := &[]models.User{}
	var totalCount int64

	filter := func(q *gorm.DB) *gorm.DB {
		if search != "" {
			pattern := "%" + search + "%"
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

func (ur *userRepository) Edit(db database.Database, user *models.User) error {
	tx := db.GetInstance()
	return tx.Save(user).Error
}

func NewUserRepository() UserRepository {
	return &userRepository{}
}
