package service

import (
	"errors"

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
	Login(tx *gorm.DB, email, password string) (*schemas.Session, error)
	Register(tx *gorm.DB, userRegister schemas.UserRegisterRequest) (*schemas.Session, error)
}

type AuthServiceImpl struct {
	userRepository repository.UserRepository
	sessionService SessionService
}

func (as *AuthServiceImpl) Login(tx *gorm.DB, email, password string) (*schemas.Session, error) {
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
	return session, nil
}

func (as *AuthServiceImpl) Register(tx *gorm.DB, userRegister schemas.UserRegisterRequest) (*schemas.Session, error) {
	if err := validator.Validate(userRegister); err != nil {
		// TODO: Add  ErrValidationFailed
		return nil, err
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

	return session, nil
}

func NewAuthService(userRepository repository.UserRepository, sessionService SessionService) AuthService {
	return &AuthServiceImpl{
		userRepository: userRepository,
		sessionService: sessionService,
	}
}
