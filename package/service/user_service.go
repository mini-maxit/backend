package service

import (
	"errors"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	myerrors "github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	// ChangePassword changes the password of a user.
	ChangePassword(tx *gorm.DB, currentUser schemas.User, userID int64, data *schemas.UserChangePassword) error
	// ChangeRole changes the role of a user.
	ChangeRole(tx *gorm.DB, currentUser schemas.User, userID int64, role types.UserRole) error
	// Edit updates the information of a user.
	Edit(tx *gorm.DB, currentUser schemas.User, userID int64, updateInfo *schemas.UserEdit) error
	// GetAll retrieves all users based on the provided query parameters.
	GetAll(tx *gorm.DB, queryParams map[string]any) ([]schemas.User, error)
	// GetByEmail retrieves a user by their email.
	GetByEmail(tx *gorm.DB, email string) (*schemas.User, error)
	// Get retrieves a user by their ID.
	Get(tx *gorm.DB, userID int64) (*schemas.User, error)
}

type userService struct {
	userRepository repository.UserRepository
	logger         *zap.SugaredLogger
}

func (us *userService) GetByEmail(tx *gorm.DB, email string) (*schemas.User, error) {
	userModel, err := us.userRepository.GetByEmail(tx, email)
	if err != nil {
		us.logger.Errorf("Error getting user by email: %v", err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrUserNotFound
		}
		return nil, err
	}

	return UserToSchema(userModel), nil
}

func (us *userService) GetAll(tx *gorm.DB, queryParams map[string]any) ([]schemas.User, error) {
	limit := queryParams["limit"].(int)
	offset := queryParams["offset"].(int)
	sort := queryParams["sort"].(string)
	if sort == "" {
		sort = "role:desc"
	}

	userModels, err := us.userRepository.GetAll(tx, limit, offset, sort)
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

func (us *userService) Get(tx *gorm.DB, userID int64) (*schemas.User, error) {
	userModel, err := us.userRepository.Get(tx, userID)
	if err != nil {
		us.logger.Errorf("Error getting user by id: %v", err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrUserNotFound
		}
		return nil, err
	}

	return UserToSchema(userModel), nil
}

func (us *userService) Edit(tx *gorm.DB, currentUser schemas.User, userID int64, updateInfo *schemas.UserEdit) error {
	if currentUser.Role != types.UserRoleAdmin && currentUser.ID != userID {
		return myerrors.ErrNotAuthorized
	}
	user, err := us.userRepository.Get(tx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrUserNotFound
		}
		us.logger.Errorf("Error getting user by id: %v", err.Error())
		return err
	}
	if updateInfo.Role != nil {
		if currentUser.Role != types.UserRoleAdmin {
			return myerrors.ErrNotAllowed
		}
	}

	us.updateModel(user, updateInfo)

	err = us.userRepository.Edit(tx, user)
	if err != nil {
		us.logger.Errorf("Error editing user: %v", err.Error())
		return err
	}
	return nil
}

func (us *userService) ChangeRole(tx *gorm.DB, currentUser schemas.User, userID int64, role types.UserRole) error {
	if currentUser.Role != types.UserRoleAdmin {
		return myerrors.ErrNotAuthorized
	}
	user, err := us.userRepository.Get(tx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrUserNotFound
		}
		us.logger.Errorf("Error getting user by id: %v", err.Error())
		return err
	}
	user.Role = role
	err = us.userRepository.Edit(tx, user)
	if err != nil {
		us.logger.Errorf("Error changing user role: %v", err.Error())
		return err
	}
	return nil
}

func (us *userService) ChangePassword(
	tx *gorm.DB,
	currentUser schemas.User,
	userID int64,
	data *schemas.UserChangePassword,
) error {
	if currentUser.ID != userID && currentUser.Role != types.UserRoleAdmin {
		return myerrors.ErrNotAuthorized
	}
	user, err := us.userRepository.Get(tx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.ErrUserNotFound
		}
		us.logger.Errorf("Error getting user by id: %v", err.Error())
		return err
	}

	validate, err := utils.NewValidator()
	if err != nil {
		return err
	}
	if err := validate.Struct(data); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return validationErrors
		}
		return err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(data.OldPassword)) != nil {
		return myerrors.ErrInvalidCredentials
	}

	if data.NewPassword != data.NewPasswordConfirm {
		return myerrors.ErrInvalidData
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(data.NewPassword), bcrypt.DefaultCost)
	log.Printf("from pass: %s new hash %s\n", data.NewPassword, hash)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hash)
	err = us.userRepository.Edit(tx, user)
	if err != nil {
		us.logger.Errorf("Error changing user password: %v", err.Error())
		return err
	}
	return nil
}

func UserToSchema(user *models.User) *schemas.User {
	return &schemas.User{
		ID:        user.ID,
		Name:      user.Name,
		Surname:   user.Surname,
		Email:     user.Email,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
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
