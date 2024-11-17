package service

import (
	"errors"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserService interface {
	GetUserByEmail(email string) (*schemas.User, error)
}

type UserServiceImpl struct {
	database       database.Database
	userRepository repository.UserRepository
}

func (us *UserServiceImpl) GetUserByEmail(email string) (*schemas.User, error) {
	tx := us.database.Connect()

	userModel, err := us.userRepository.GetUserByEmail(tx, email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user := us.modelToSchema(userModel)

	return user, nil
}

func (us *UserServiceImpl) modelToSchema(user *models.User) *schemas.User {
	return &schemas.User{
		Id:       user.Id,
		Name:     user.Name,
		Surname:  user.Surname,
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
	}
}

func NewUserService(database database.Database, userRepository repository.UserRepository) UserService {
	return &UserServiceImpl{
		database:       database,
		userRepository: userRepository,
	}
}
