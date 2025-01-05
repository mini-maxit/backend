package service

import (
	"errors"

	"github.com/mini-maxit/backend/internal/logger"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserService interface {
	GetUserByEmail(tx *gorm.DB, email string) (*schemas.User, error)
	GetAllUsers(tx *gorm.DB, queryParams map[string][]string) ([]schemas.User, error)
	GetUserById(tx *gorm.DB, userId int64) (*schemas.User, error)
	EditUser(tx *gorm.DB, userId int64, updateInfo *schemas.UserEdit) error
	modelToSchema(user *models.User) *schemas.User
}

type UserServiceImpl struct {
	userRepository repository.UserRepository
	logger         *zap.SugaredLogger
}

func (us *UserServiceImpl) GetUserByEmail(tx *gorm.DB, email string) (*schemas.User, error) {
	userModel, err := us.userRepository.GetUserByEmail(tx, email)
	if err != nil {
		us.logger.Errorf("Error getting user by email: %v", err.Error())
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user := us.modelToSchema(userModel)
	return user, nil
}

func (us *UserServiceImpl) GetAllUsers(tx *gorm.DB, queryParams map[string][]string) ([]schemas.User, error) {
	userModels, err := us.userRepository.GetAllUsers(tx, queryParams)
	if err != nil {
		us.logger.Errorf("Error getting all users: %v", err.Error())
		return nil, err
	}

	var users []schemas.User
	for _, userModel := range userModels {
		users = append(users, *us.modelToSchema(&userModel))
	}

	return users, nil
}

func (us *UserServiceImpl) GetUserById(tx *gorm.DB, userId int64) (*schemas.User, error) {
	userModel, err := us.userRepository.GetUser(tx, userId)
	if err != nil {
		us.logger.Errorf("Error getting user by id: %v", err.Error())
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user := us.modelToSchema(userModel)

	return user, nil
}

func (us *UserServiceImpl) EditUser(tx *gorm.DB, userId int64, updateInfo *schemas.UserEdit) error {
	currentModel, err := us.GetUserById(tx, userId)
	if err != nil {
		us.logger.Errorf("Error getting user by id: %v", err.Error())
		return err
	}

	us.updateModel(currentModel, updateInfo)

	err = us.userRepository.EditUser(tx, currentModel)
	if err != nil {
		us.logger.Errorf("Error editing user: %v", err.Error())
		return err
	}
	return nil
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

func (us *UserServiceImpl) updateModel(curretnModel *schemas.User, updateInfo *schemas.UserEdit) {
	if updateInfo.Email != nil {
		curretnModel.Email = *updateInfo.Email
	}

	if updateInfo.Name != nil {
		curretnModel.Name = *updateInfo.Name
	}

	if updateInfo.Surname != nil {
		curretnModel.Surname = *updateInfo.Surname
	}

	if updateInfo.Username != nil {
		curretnModel.Username = *updateInfo.Username
	}
}

func NewUserService(userRepository repository.UserRepository) UserService {
	log := logger.NewNamedLogger("user_service")
	return &UserServiceImpl{
		userRepository: userRepository,
		logger:         log,
	}
}
