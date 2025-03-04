package service

import (
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

// AuthService defines the methods for authentication-related operations.
type AuthService interface {
	// Login logs user in using provided credentials
	//
	// Method should check if such user exists, and if he does create valid session and return it.
	// Possible errors: ErrInvalidCredentials, ErrUserNotFound, validator.ValidationErrors,
	// errors returned by SessionService, repository.UserRepository.
	Login(tx *gorm.DB, userLogin schemas.UserLoginRequest) (*schemas.Session, error)

	// Register registers user if he is not registered yet
	//
	// Method validates provided user data, checks if user with provided email already exists,
	// creates new user, creates session for him and returns it.
	// Possible errors: ErrUserAlreadyExists, errors returned by SessionService,
	// repostiroy.UserRepository, bcrypt lib.
	Register(tx *gorm.DB, userRegister schemas.UserRegisterRequest) (*schemas.Session, error)
}

// authService implements AuthService interface
type authService struct {
	userRepository repository.UserRepository
	sessionService SessionService
	logger         *zap.SugaredLogger
}

// Login implements Login method of [AuthService] interface
func (as *authService) Login(tx *gorm.DB, userLogin schemas.UserLoginRequest) (*schemas.Session, error) {
	user, err := as.userRepository.GetUserByEmail(tx, userLogin.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrUserNotFound
		}
		as.logger.Errorf("Error getting user by email: %v", err.Error())
		return nil, err

	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(userLogin.Password)) != nil {
		as.logger.Errorf("Error comparing password hash: %v", errors.ErrInvalidCredentials.Error())
		return nil, errors.ErrInvalidCredentials
	}

	session, err := as.sessionService.CreateSession(tx, user.Id)
	if err != nil {
		as.logger.Errorf("Error creating session: %v", err.Error())
		return nil, err
	}
	as.logger.Infof("User logged in successfully")
	return session, nil
}

// Register implements Register method of [AuthService] interface
func (as *authService) Register(tx *gorm.DB, userRegister schemas.UserRegisterRequest) (*schemas.Session, error) {
	user, err := as.userRepository.GetUserByEmail(tx, userRegister.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		as.logger.Errorf("Error getting user by email: %v", err.Error())
		return nil, err
	}

	if user != nil {
		as.logger.Errorf("User already exists")
		return nil, errors.ErrUserAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(userRegister.Password), bcrypt.DefaultCost)
	if err != nil {
		as.logger.Errorf("Error generating password hash: %v", err.Error())
		return nil, err
	}
	userModel := &models.User{
		Name:         userRegister.Name,
		Surname:      userRegister.Surname,
		Email:        userRegister.Email,
		Username:     userRegister.Username,
		PasswordHash: string(hash),
		Role:         types.UserRoleStudent,
	}

	userId, err := as.userRepository.CreateUser(tx, userModel)
	if err != nil {
		as.logger.Errorf("Error creating user: %v", err.Error())
		return nil, err
	}

	session, err := as.sessionService.CreateSession(tx, userId)
	if err != nil {
		as.logger.Errorf("Error creating session: %v", err.Error())
		return nil, err
	}

	as.logger.Infof("User registered successfully")
	return session, nil
}

// NewAuthService creates new instance of [authService]
func NewAuthService(userRepository repository.UserRepository, sessionService SessionService) AuthService {
	log := utils.NewNamedLogger("auth_service")
	return &authService{
		userRepository: userRepository,
		sessionService: sessionService,
		logger:         log,
	}
}
