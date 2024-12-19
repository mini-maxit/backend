package service

import (
	"errors"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
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
}

func (as *AuthServiceImpl) Login(email, password string) (*schemas.Session, error) {
	tx := as.database.Connect().Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	user, err := as.userRepository.GetUserByEmail(tx, email)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, ErrInvalidCredentials
	}

	session, err := as.sessionService.CreateSession(tx, user.Id)
	if err != nil {
		return nil, err
	}
	tx.Commit()
	return session, nil
}

func (as *AuthServiceImpl) Register(userRegister schemas.UserRegisterRequest) (*schemas.Session, error) {
	if err := validator.Validate(userRegister); err != nil {
		return nil, err

	}
	tx := as.database.Connect().Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	user, err := as.userRepository.GetUserByEmail(tx, userRegister.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if user != nil {
		return nil, ErrUserAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(userRegister.Password), bcrypt.DefaultCost)
	if err != nil {
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
		return nil, err
	}

	session, err := as.sessionService.CreateSession(tx, userId)
	if err != nil {
		return nil, err
	}
	tx.Commit()
	return session, nil
}

func NewAuthService(database database.Database, userRepository repository.UserRepository, sessionService SessionService) AuthService {
	return &AuthServiceImpl{
		database:       database,
		userRepository: userRepository,
		sessionService: sessionService,
	}
}
