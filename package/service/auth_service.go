package service

import (
	"errors"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/internal/logger"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
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
	logger         *zap.SugaredLogger
}

func (as *AuthServiceImpl) Login(email, password string) (*schemas.Session, error) {
	tx := as.database.Connect().Begin()
	if tx.Error != nil {
		as.logger.Errorf("Error connecting to database: %v", tx.Error.Error())
		return nil, tx.Error
	}

	defer utils.TransactionPanicRecover(tx)

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

	if err := tx.Commit().Error; err != nil {
		as.logger.Errorf("Error committing transaction: %v", err.Error())
		return nil, err
	}

	as.logger.Infof("User logged in successfully")
	return session, nil
}

func (as *AuthServiceImpl) Register(userRegister schemas.UserRegisterRequest) (*schemas.Session, error) {
	if err := validator.Validate(userRegister); err != nil {
		as.logger.Errorf("Error validating user register request: %v", err.Error())
		return nil, err
	}
	tx := as.database.Connect().Begin()
	if tx.Error != nil {
		as.logger.Errorf("Error connecting to database: %v", tx.Error.Error())
		return nil, tx.Error
	}

	defer utils.TransactionPanicRecover(tx)

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

	if err := tx.Commit().Error; err != nil {
		as.logger.Errorf("Error committing transaction: %v", err.Error())
		return nil, err
	}

	as.logger.Infof("User registered successfully")
	return session, nil
}

func NewAuthService(database database.Database, userRepository repository.UserRepository, sessionService SessionService) AuthService {
	log := logger.NewNamedLogger("auth_service")
	return &AuthServiceImpl{
		database:       database,
		userRepository: userRepository,
		sessionService: sessionService,
		logger:         log,
	}
}
