package service

import (
	"github.com/go-playground/validator/v10"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	GetUserByEmail(tx *gorm.DB, email string) (*schemas.User, error)
	GetAllUsers(tx *gorm.DB, queryParams map[string]interface{}) ([]schemas.User, error)
	GetUserById(tx *gorm.DB, userId int64) (*schemas.User, error)
	EditUser(tx *gorm.DB, currentUser schemas.User, userId int64, updateInfo *schemas.UserEdit) error
	ChangeRole(tx *gorm.DB, currentUser schemas.User, userId int64, role types.UserRole) error
	ChangePassword(tx *gorm.DB, currentUser schemas.User, userId int64, data *schemas.UserChangePassword) error
}

type userService struct {
	userRepository repository.UserRepository
	logger         *zap.SugaredLogger
}

func (us *userService) GetUserByEmail(tx *gorm.DB, email string) (*schemas.User, error) {
	userModel, err := us.userRepository.GetUserByEmail(tx, email)
	if err != nil {
		us.logger.Errorf("Error getting user by email: %v", err.Error())
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrUserNotFound
		}
		return nil, err
	}

	return UserToSchema(userModel), nil
}

func (us *userService) GetAllUsers(tx *gorm.DB, queryParams map[string]interface{}) ([]schemas.User, error) {
	limit := queryParams["limit"].(uint64)
	offset := queryParams["offset"].(uint64)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = "role:desc"
	}

	userModels, err := us.userRepository.GetAllUsers(tx, int(limit), int(offset), sort)
	if err != nil {
		us.logger.Errorf("Error getting all users: %v", err.Error())
		return nil, err
	}

	var users []schemas.User
	for _, userModel := range userModels {
		users = append(users, *UserToSchema(&userModel))
	}

	return users, nil
}

func (us *userService) GetUserById(tx *gorm.DB, userId int64) (*schemas.User, error) {
	userModel, err := us.userRepository.GetUser(tx, userId)
	if err != nil {
		us.logger.Errorf("Error getting user by id: %v", err.Error())
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrUserNotFound
		}
		return nil, err
	}

	return UserToSchema(userModel), nil
}

func (us *userService) EditUser(tx *gorm.DB, currentUser schemas.User, userId int64, updateInfo *schemas.UserEdit) error {
	if currentUser.Role != types.UserRoleAdmin && currentUser.Id != userId {
		return errors.ErrNotAuthorized
	}
	user, err := us.userRepository.GetUser(tx, userId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUserNotFound
		}
		us.logger.Errorf("Error getting user by id: %v", err.Error())
		return err
	}
	if updateInfo.Role != nil {
		if currentUser.Role != types.UserRoleAdmin {
			return errors.ErrNotAllowed
		}
	}

	us.updateModel(user, updateInfo)

	err = us.userRepository.EditUser(tx, user)
	if err != nil {
		us.logger.Errorf("Error editing user: %v", err.Error())
		return err
	}
	return nil
}

func (us *userService) ChangeRole(tx *gorm.DB, currentUser schemas.User, userId int64, role types.UserRole) error {
	if currentUser.Role != types.UserRoleAdmin {
		return errors.ErrNotAuthorized
	}
	user, err := us.userRepository.GetUser(tx, userId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUserNotFound
		}
		us.logger.Errorf("Error getting user by id: %v", err.Error())
		return err
	}
	user.Role = role
	err = us.userRepository.EditUser(tx, user)
	if err != nil {
		us.logger.Errorf("Error changing user role: %v", err.Error())
		return err
	}
	return nil
}

func (us *userService) ChangePassword(tx *gorm.DB, currentUser schemas.User, userId int64, data *schemas.UserChangePassword) error {
	if currentUser.Id != userId && currentUser.Role != types.UserRoleAdmin {
		return errors.ErrNotAuthorized
	}
	user, err := us.userRepository.GetUser(tx, userId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUserNotFound
		}
		us.logger.Errorf("Error getting user by id: %v", err.Error())
		return err
	}

	validate, err := utils.NewValidator()
	if err != nil {
		return err
	}
	if err := validate.Struct(data); err != nil {
		return err.(validator.ValidationErrors)
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(data.OldPassword)) != nil {
		return errors.ErrInvalidCredentials
	}

	if data.NewPassword != data.NewPasswordConfirm {
		return errors.ErrInvalidData
	}

	// user.PasswordHash =
	hash, err := bcrypt.GenerateFromPassword([]byte(data.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hash)
	err = us.userRepository.EditUser(tx, user)
	if err != nil {
		us.logger.Errorf("Error changing user password: %v", err.Error())
		return err
	}
	return nil
}

func UserToSchema(user *models.User) *schemas.User {
	return &schemas.User{
		Id:       user.Id,
		Name:     user.Name,
		Surname:  user.Surname,
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
	}
}

func (us *userService) updateModel(curretnModel *models.User, updateInfo *schemas.UserEdit) {
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
	if updateInfo.Role != nil {
		curretnModel.Role = *updateInfo.Role
	}
}

func NewUserService(userRepository repository.UserRepository) UserService {
	log := utils.NewNamedLogger("user_service")
	return &userService{
		userRepository: userRepository,
		logger:         log,
	}
}
