package service

import (
	"errors"

	"github.com/mini-maxit/backend/internal/logger"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService interface {
	Login(tx *gorm.DB, email, password string) (*schemas.Session, error)
	Register(tx *gorm.DB, userRegister schemas.UserRegisterRequest) (*schemas.Session, error)
}

type AuthServiceImpl struct {
	userRepository repository.UserRepository
	sessionService SessionService
	logger         *zap.SugaredLogger
}

func (as *AuthServiceImpl) Login(tx *gorm.DB, email, password string) (*schemas.Session, error) {
	user, err := as.userRepository.GetUserByEmail(tx, email)
	if err != nil {
		as.logger.Errorf("Error getting user by email: %v", err.Error())
		return nil, ErrUserNotFound
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		as.logger.Errorf("Error comparing password hash: %v", ErrInvalidCredentials.Error())
		return nil, ErrInvalidCredentials
	}

	session, err := as.sessionService.CreateSession(tx, user.Id)
	if err != nil {
		as.logger.Errorf("Error creating session: %v", err.Error())
		return nil, err
	}
	as.logger.Infof("User logged in successfully")
	return session, nil
}

func (as *AuthServiceImpl) Register(tx *gorm.DB, userRegister schemas.UserRegisterRequest) (*schemas.Session, error) {
	if err := validator.Validate(userRegister); err != nil {
		as.logger.Errorf("Error validating user register request: %v", err.Error())
		return nil, err
	}

	user, err := as.userRepository.GetUserByEmail(tx, userRegister.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		as.logger.Errorf("Error getting user by email: %v", err.Error())
		return nil, err
	}

	if user != nil {
		as.logger.Errorf("User already exists")
		return nil, ErrUserAlreadyExists
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
		Role:         string(models.UserRoleStudent),
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

func NewAuthService(userRepository repository.UserRepository, sessionService SessionService) AuthService {
	log := logger.NewNamedLogger("auth_service")
	return &AuthServiceImpl{
		userRepository: userRepository,
		sessionService: sessionService,
		logger:         log,
	}
}
