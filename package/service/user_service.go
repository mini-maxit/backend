package service

import (
	"errors"

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
	GetAll(tx *gorm.DB, paginationParams schemas.PaginationParams) ([]schemas.User, error)
	// GetByEmail retrieves a user by their email.
	GetByEmail(tx *gorm.DB, email string) (*schemas.User, error)
	// Get retrieves a user by their ID.
	Get(tx *gorm.DB, userID int64) (*schemas.User, error)
	// IsTaskAssignedToUser checks if a user has access to a specific task.
	// Currently, a user has access if the task is assigned to a contest and the user is a participant in that contest.
	// Future: may support direct task-to-user assignments.
	IsTaskAssignedToUser(tx *gorm.DB, userID, taskID int64) (bool, error)
}

type userService struct {
	userRepository repository.UserRepository
	contestService ContestService
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

func (us *userService) GetAll(tx *gorm.DB, paginationParams schemas.PaginationParams) ([]schemas.User, error) {
	if paginationParams.Sort == "" {
		paginationParams.Sort = "role:desc"
	}

	userModels, err := us.userRepository.GetAll(tx, paginationParams.Limit, paginationParams.Offset, paginationParams.Sort)
	if err != nil {
		us.logger.Errorf("Error getting all users: %v", err.Error())
		return nil, err
	}

	users := make([]schemas.User, len(userModels))
	for i, userModel := range userModels {
		users[i] = *UserToSchema(&userModel)
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
		return myerrors.ErrForbidden
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
		return myerrors.ErrForbidden
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
		return myerrors.ErrForbidden
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

func (us *userService) updateModel(currentModel *models.User, updateInfo *schemas.UserEdit) {
	if updateInfo.Email != nil {
		currentModel.Email = *updateInfo.Email
	}

	if updateInfo.Name != nil {
		currentModel.Name = *updateInfo.Name
	}

	if updateInfo.Surname != nil {
		currentModel.Surname = *updateInfo.Surname
	}

	if updateInfo.Username != nil {
		currentModel.Username = *updateInfo.Username
	}
	if updateInfo.Role != nil {
		currentModel.Role = *updateInfo.Role
	}
}

func (us *userService) IsTaskAssignedToUser(tx *gorm.DB, userID, taskID int64) (bool, error) {
	// TODO: In the future, this should also check for direct task-to-user assignments
	// For now, check if task is assigned to any contest that the user participates in

	// Get all contests that this task is assigned to
	contestIDs, err := us.contestService.GetTaskContests(tx, taskID)
	if err != nil {
		us.logger.Errorf("Error getting task contests: %v", err.Error())
		return false, err
	}

	// If task is not assigned to any contest, user doesn't have access
	if len(contestIDs) == 0 {
		return false, nil
	}

	// Check if user is a participant in any of the contests
	for _, contestID := range contestIDs {
		isParticipant, err := us.contestService.IsUserParticipant(tx, contestID, userID)
		if err != nil {
			us.logger.Errorf("Error checking user participation in contest: %v", err.Error())
			return false, err
		}
		if isParticipant {
			return true, nil
		}
	}

	// User is not a participant in any contest that has this task
	return false, nil
}

func NewUserService(userRepository repository.UserRepository, contestService ContestService) UserService {
	log := utils.NewNamedLogger("user_service")
	return &userService{
		userRepository: userRepository,
		contestService: contestService,
		logger:         log,
	}
}
