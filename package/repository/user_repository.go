package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/utils"
	"gorm.io/gorm"
)

type UserRepository interface {
	// CreateUser creates a new user and returns the user ID
	CreateUser(tx *gorm.DB, user *models.User) (int64, error)
	GetUser(tx *gorm.DB, userId int64) (*models.User, error)
	GetUserByEmail(tx *gorm.DB, email string) (*models.User, error)
	GetAllUsers(tx *gorm.DB, limit, offset int, sort string) ([]models.User, error)
	EditUser(tx *gorm.DB, user *models.User) error
}

type userRepository struct {
}

func (ur *userRepository) CreateUser(tx *gorm.DB, user *models.User) (int64, error) {
	err := tx.Model(&models.User{}).Create(user).Error
	if err != nil {
		return 0, err
	}
	return user.Id, nil
}

func (ur *userRepository) GetUser(tx *gorm.DB, userId int64) (*models.User, error) {
	user := &models.User{}
	err := tx.Model(&models.User{}).Where("id = ?", userId).Take(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *userRepository) GetUserByEmail(tx *gorm.DB, email string) (*models.User, error) {
	user := &models.User{}
	err := tx.Model(&models.User{}).Where("email = ?", email).First(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *userRepository) GetAllUsers(tx *gorm.DB, limit, offset int, sort string) ([]models.User, error) {
	users := &[]models.User{}
	tx, err := utils.ApplyPaginationAndSort(tx, limit, offset, sort)
	if err != nil {
		return nil, err
	}

	err = tx.Model(&models.User{}).Find(users).Error
	if err != nil {
		return nil, err
	}
	return *users, nil
}

func (ur *userRepository) EditUser(tx *gorm.DB, user *models.User) error {
	err := tx.Model(&models.User{}).Where("id = ?", user.Id).Updates(user).Error
	return err
}

func NewUserRepository(db *gorm.DB) (UserRepository, error) {
	if !db.Migrator().HasTable(&models.User{}) {
		err := db.Migrator().CreateTable(&models.User{})
		if err != nil {
			return nil, err
		}
	}

	return &userRepository{}, nil
}
