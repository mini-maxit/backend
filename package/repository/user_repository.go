package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type UserRepositorty interface {
	// CreateUser creates a new user and returns the user ID
	CreateUser(tx *gorm.DB, user *models.User) (int64, error)
	GetUser(tx *gorm.DB, userId int64) (*models.User, error)
}

type UserRepositoryImpl struct {
}

func (ur *UserRepositoryImpl) CreateUser(tx *gorm.DB, user *models.User) (int64, error) {
	err := tx.Model(&models.User{}).Create(user).Error
	if err != nil {
		return 0, err
	}
	return user.Id, nil
}

func (ur *UserRepositoryImpl) GetUser(tx *gorm.DB, userId int64) (*models.User, error) {
	user := &models.User{}
	err := tx.Model(&models.User{}).Where("id = ?", userId).First(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func NewUserRepository(db *gorm.DB) (UserRepositorty, error) {
	if !db.Migrator().HasTable(&models.User{}) {
		err := db.Migrator().CreateTable(&models.User{})
		if err != nil {
			return nil, err
		}
	}

	return &UserRepositoryImpl{}, nil
}
