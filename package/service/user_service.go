package service

import (
	"errors"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/internal/logger"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserService interface {
	GetUserByEmail(email string) (*schemas.User, error)
	GetAllUsers(limit, offset int64) ([]schemas.User, error)
	GetUserById(userId int64) (*schemas.User, error)
	EditUser(userId int64, updateInfo *schemas.UserEdit) error
}

type UserServiceImpl struct {
	database       database.Database
	userRepository repository.UserRepository
	user_logger   *logger.ServiceLogger
}

func (us *UserServiceImpl) GetUserByEmail(email string) (*schemas.User, error) {
	tx := us.database.Connect()

	if tx == nil {
		logger.Log(us.user_logger, "Error connecting to database:", ErrDatabaseConnection.Error(), logger.Error)
		return nil, ErrDatabaseConnection
	}

	userModel, err := us.userRepository.GetUserByEmail(tx, email)
	if err != nil {
		logger.Log(us.user_logger, "Error getting user by email:", err.Error(), logger.Error)
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user := us.modelToSchema(userModel)

	return user, nil
}

func (us *UserServiceImpl) GetAllUsers(limit, offset int64) ([]schemas.User, error) {
	tx := us.database.Connect()

	if tx == nil {
		logger.Log(us.user_logger, "Error connecting to database:", ErrDatabaseConnection.Error(), logger.Error)
		return nil, ErrDatabaseConnection
	}

	userModels, err := us.userRepository.GetAllUsers(tx)
	if err != nil {
		logger.Log(us.user_logger, "Error getting all users:", err.Error(), logger.Error)
		return nil, err
	}

	var users []schemas.User
	for _, userModel := range userModels {
		users = append(users, *us.modelToSchema(&userModel))
	}

	// Handle pagination
	if offset >= int64(len(users)) {
		return []schemas.User{}, nil
	}

	end := offset + limit
	if end > int64(len(users)) {
		end = int64(len(users))
	}

	return users[offset:end], nil
}

func (us *UserServiceImpl) GetUserById(userId int64) (*schemas.User, error){
	tx := us.database.Connect()

	if tx == nil {
		logger.Log(us.user_logger, "Error connecting to database:", ErrDatabaseConnection.Error(), logger.Error)
		return nil, ErrDatabaseConnection
	}

	userModel, err := us.userRepository.GetUser(tx, userId)
	if err != nil {
		logger.Log(us.user_logger, "Error getting user by id:", err.Error(), logger.Error)
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user := us.modelToSchema(userModel)

	return user, nil
}

func (us *UserServiceImpl) EditUser(userId int64, updateInfo *schemas.UserEdit) error {
	tx := us.database.Connect()

	if tx == nil {
		logger.Log(us.user_logger, "Error connecting to database:", ErrDatabaseConnection.Error(), logger.Error)
		return ErrDatabaseConnection
	}

	tx = tx.Begin()
    if tx.Error != nil {
		logger.Log(us.user_logger, "Error connecting to database:", tx.Error.Error(), logger.Error)
        return tx.Error
    }

	defer utils.TransactionPanicRecover(tx)

	currentModel, err := us.GetUserById(userId)
	if err != nil {
		logger.Log(us.user_logger, "Error getting user by id:", err.Error(), logger.Error)
		tx.Rollback()
		return err
	}

	us.updateModel(currentModel, updateInfo)


	err = us.userRepository.EditUser(tx, currentModel)
	if err != nil {
		logger.Log(us.user_logger, "Error editing user:", err.Error(), logger.Error)
        tx.Rollback()
        return err
    }

    if err := tx.Commit().Error; err != nil {
		logger.Log(us.user_logger, "Error committing transaction:", err.Error(), logger.Error)
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

func NewUserService(database database.Database, userRepository repository.UserRepository) UserService {
	user_logger := logger.NewNamedLogger("user_service")
	return &UserServiceImpl{
		database:       database,
		userRepository: userRepository,
		user_logger:    &user_logger,
	}
}
