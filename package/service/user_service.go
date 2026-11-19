package service

import (
	"crypto/rand"
	"math/big"

	"github.com/go-playground/validator/v10"
	"github.com/mini-maxit/backend/internal/database"
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
	// ChangePassword changes the password of a user.
	ChangePassword(db database.Database, currentUser schemas.User, userID int64, data *schemas.UserChangePassword) error
	// ResetPassword sets a new random password for a user (admin only) and returns it in plain text.
	ResetPassword(db database.Database, currentUser schemas.User, userID int64) (string, error)
	// ChangeRole changes the role of a user.
	ChangeRole(db database.Database, currentUser schemas.User, userID int64, role types.UserRole) error
	// Edit updates the information of a user.
	Edit(db database.Database, currentUser schemas.User, userID int64, updateInfo *schemas.UserEdit) error
	// GetAll retrieves users matching the optional search (name, surname, email) and role filters.
	// With deleted set it lists only soft-deleted users instead.
	GetAll(db database.Database, paginationParams schemas.PaginationParams, search, role string, deleted bool) (schemas.PaginatedResult[[]schemas.User], error)
	// GetByEmail retrieves a user by their email.
	GetByEmail(db database.Database, email string) (*schemas.User, error)
	// Get retrieves a user by their ID.
	Get(db database.Database, userID int64) (*schemas.User, error)
	// Delete soft-deletes a user (admin only). Admin accounts, including the caller, cannot be deleted.
	Delete(db database.Database, currentUser schemas.User, userID int64) error
	// Restore restores a soft-deleted user (admin only).
	Restore(db database.Database, currentUser schemas.User, userID int64) error
	// IsTaskAssignedToUser checks if a user has access to a specific task.
	// Currently, a user has access if the task is assigned to a contest and the user is a participant in that contest.
	// Future: may support direct task-to-user assignments.
	IsTaskAssignedToUser(db database.Database, userID, taskID int64) (bool, error)
}

type userService struct {
	userRepository repository.UserRepository
	contestService ContestService
	logger         *zap.SugaredLogger
}

func (us *userService) GetByEmail(db database.Database, email string) (*schemas.User, error) {
	userModel, err := us.userRepository.GetByEmail(db, email)
	if err != nil {
		us.logger.Errorf("Error getting user by email: %v", err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrUserNotFound
		}
		return nil, err
	}

	return UserToSchema(userModel), nil
}

func (us *userService) GetAll(db database.Database, paginationParams schemas.PaginationParams, search, role string, deleted bool) (schemas.PaginatedResult[[]schemas.User], error) {
	if paginationParams.Sort == "" {
		paginationParams.Sort = "role:desc"
	}

	userModels, totalCount, err := us.userRepository.GetAll(db, paginationParams.Limit, paginationParams.Offset, paginationParams.Sort, search, role, deleted)
	if err != nil {
		us.logger.Errorf("Error getting all users: %v", err.Error())
		return schemas.PaginatedResult[[]schemas.User]{}, err
	}

	users := make([]schemas.User, len(userModels))
	for i, userModel := range userModels {
		users[i] = *UserToSchema(&userModel)
	}

	return schemas.NewPaginatedResult(users, paginationParams.Offset, paginationParams.Limit, totalCount), nil
}

func (us *userService) Delete(db database.Database, currentUser schemas.User, userID int64) error {
	if currentUser.Role != types.UserRoleAdmin {
		return errors.ErrForbidden
	}
	target, err := us.userRepository.Get(db, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrUserNotFound
		}
		us.logger.Errorf("Error getting user by id: %v", err.Error())
		return err
	}
	// covers self-deletion too, since only admins get this far
	if target.Role == types.UserRoleAdmin {
		return errors.ErrNotAllowed
	}
	if err := us.userRepository.Delete(db, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrUserNotFound
		}
		us.logger.Errorf("Error deleting user: %v", err.Error())
		return err
	}
	return nil
}

func (us *userService) Restore(db database.Database, currentUser schemas.User, userID int64) error {
	if currentUser.Role != types.UserRoleAdmin {
		return errors.ErrForbidden
	}
	if err := us.userRepository.Restore(db, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrUserNotFound
		}
		us.logger.Errorf("Error restoring user: %v", err.Error())
		return err
	}
	return nil
}

func (us *userService) Get(db database.Database, userID int64) (*schemas.User, error) {
	userModel, err := us.userRepository.Get(db, userID)
	if err != nil {
		us.logger.Errorf("Error getting user by id: %v", err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrUserNotFound
		}
		return nil, err
	}

	return UserToSchema(userModel), nil
}

func (us *userService) Edit(db database.Database, currentUser schemas.User, userID int64, updateInfo *schemas.UserEdit) error {
	if currentUser.Role != types.UserRoleAdmin && currentUser.ID != userID {
		return errors.ErrForbidden
	}
	user, err := us.userRepository.Get(db, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

	err = us.userRepository.Edit(db, user)
	if err != nil {
		us.logger.Errorf("Error editing user: %v", err.Error())
		return err
	}
	return nil
}

func (us *userService) ChangeRole(db database.Database, currentUser schemas.User, userID int64, role types.UserRole) error {
	if currentUser.Role != types.UserRoleAdmin {
		return errors.ErrForbidden
	}
	user, err := us.userRepository.Get(db, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrUserNotFound
		}
		us.logger.Errorf("Error getting user by id: %v", err.Error())
		return err
	}
	user.Role = role
	err = us.userRepository.Edit(db, user)
	if err != nil {
		us.logger.Errorf("Error changing user role: %v", err.Error())
		return err
	}
	return nil
}

func (us *userService) ChangePassword(
	db database.Database,
	currentUser schemas.User,
	userID int64,
	data *schemas.UserChangePassword,
) error {
	if currentUser.ID != userID && currentUser.Role != types.UserRoleAdmin {
		return errors.ErrForbidden
	}
	user, err := us.userRepository.Get(db, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return validationErrors
		}
		return err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(data.OldPassword)) != nil {
		return errors.ErrInvalidCredentials
	}

	if data.NewPassword != data.NewPasswordConfirm {
		return errors.ErrInvalidData
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(data.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hash)
	err = us.userRepository.Edit(db, user)
	if err != nil {
		us.logger.Errorf("Error changing user password: %v", err.Error())
		return err
	}
	return nil
}

// generatePassword returns a random password satisfying utils.PasswordValidator.
func generatePassword() (string, error) {
	groups := []string{"ABCDEFGHJKLMNPQRSTUVWXYZ", "abcdefghijkmnopqrstuvwxyz", "23456789", "#!?@$%^&*-"}
	all := groups[0] + groups[1] + groups[2] + groups[3]
	const length = 16
	pick := func(set string) (byte, error) {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(set))))
		if err != nil {
			return 0, err
		}
		return set[n.Int64()], nil
	}
	pw := make([]byte, length)
	for i := range pw {
		set := all
		if i < len(groups) {
			set = groups[i]
		}
		c, err := pick(set)
		if err != nil {
			return "", err
		}
		pw[i] = c
	}
	// shuffle so the guaranteed characters are not always first
	for i := length - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		pw[i], pw[n.Int64()] = pw[n.Int64()], pw[i]
	}
	return string(pw), nil
}

func (us *userService) ResetPassword(db database.Database, currentUser schemas.User, userID int64) (string, error) {
	if currentUser.Role != types.UserRoleAdmin {
		return "", errors.ErrForbidden
	}
	user, err := us.userRepository.Get(db, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.ErrUserNotFound
		}
		us.logger.Errorf("Error getting user by id: %v", err.Error())
		return "", err
	}
	password, err := generatePassword()
	if err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	user.PasswordHash = string(hash)
	if err := us.userRepository.Edit(db, user); err != nil {
		us.logger.Errorf("Error resetting user password: %v", err.Error())
		return "", err
	}
	return password, nil
}

func UserToInfoSchema(user *models.User) *schemas.UserInfo {
	return &schemas.UserInfo{
		ID:      user.ID,
		Name:    user.Name,
		Surname: user.Surname,
	}
}

func UserToSchema(user *models.User) *schemas.User {
	return &schemas.User{
		ID:        user.ID,
		Name:      user.Name,
		Surname:   user.Surname,
		Email:     user.Email,
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

	if updateInfo.Role != nil {
		currentModel.Role = *updateInfo.Role
	}
}

func (us *userService) IsTaskAssignedToUser(db database.Database, userID, taskID int64) (bool, error) {
	// TODO: In the future, this should also check for direct task-to-user assignments
	// For now, check if task is assigned to any contest that the user participates in

	// Get all contests that this task is assigned to
	contestIDs, err := us.contestService.GetTaskContests(db, taskID)
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
		isParticipant, err := us.contestService.IsUserParticipant(db, contestID, userID)
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
