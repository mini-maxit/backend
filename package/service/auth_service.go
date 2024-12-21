package service

import (
	"errors"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/internal/logger"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"

)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService interface {
	Login(email, password string) (*schemas.Session, error)
	Register(userRegister schemas.UserRegisterRequest) (*schemas.Session, error)
}

type AuthServiceImpl struct {
	database       database.Database
	userRepository repository.UserRepository
	sessionService SessionService
	service_logger *logger.ServiceLogger
}

func (as *AuthServiceImpl) Login(email, password string) (*schemas.Session, error) {
	tx := as.database.Connect().Begin()
	if tx.Error != nil {
		logger.Log(as.service_logger, "Error connecting to database:", tx.Error.Error(), logger.Error)
		return nil, tx.Error
	}

	defer utils.TransactionPanicRecover(tx)

	user, err := as.userRepository.GetUserByEmail(tx, email)
	if err != nil {
		logger.Log(as.service_logger, "Error getting user by email:", err.Error(), logger.Error)
		return nil, ErrUserNotFound
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		logger.Log(as.service_logger, "Error comparing password hash:", ErrInvalidCredentials.Error(), logger.Error)
		return nil, ErrInvalidCredentials
	}

	session, err := as.sessionService.CreateSession(tx, user.Id)
	if err != nil {
		logger.Log(as.service_logger, "Error creating session:", err.Error(), logger.Error)
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		logger.Log(as.service_logger, "Error committing transaction:", err.Error(), logger.Error)
		return nil, err
	}

	logger.Log(as.service_logger, "User logged in successfully", "", logger.Info)
	return session, nil
}

func (as *AuthServiceImpl) Register(userRegister schemas.UserRegisterRequest) (*schemas.Session, error) {
	if err := validator.Validate(userRegister); err != nil {
		logger.Log(as.service_logger, "Error validating user register request", err.Error(), logger.Error)
		return nil, err
	}
	tx := as.database.Connect().Begin()
	if tx.Error != nil {
		logger.Log(as.service_logger, "Error connecting to database", tx.Error.Error(), logger.Error)
		return nil, tx.Error
	}

	defer utils.TransactionPanicRecover(tx)

	user, err := as.userRepository.GetUserByEmail(tx, userRegister.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		logger.Log(as.service_logger, "Error getting user by email", err.Error(), logger.Error)
		return nil, err
	}
	if user != nil {
		logger.Log(as.service_logger, "User already exists", "", logger.Error)
		return nil, ErrUserAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(userRegister.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Log(as.service_logger, "Error generating password hash", err.Error(), logger.Error)
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
		logger.Log(as.service_logger, "Error creating user", err.Error(), logger.Error)
		return nil, err
	}

	session, err := as.sessionService.CreateSession(tx, userId)
	if err != nil {
		logger.Log(as.service_logger, "Error creating session", err.Error(), logger.Error)
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		logger.Log(as.service_logger, "Error committing transaction", err.Error(), logger.Error)
		return nil, err
	}

	logger.Log(as.service_logger, "User registered successfully", "", logger.Info)
	return session, nil
}

func NewAuthService(database database.Database, userRepository repository.UserRepository, sessionService SessionService) AuthService {
	logger := logger.NewNamedLogger("auth_service")
	return &AuthServiceImpl{
		database:       database,
		userRepository: userRepository,
		sessionService: sessionService,
		service_logger: &logger,
	}
}
