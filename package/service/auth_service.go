package service

import (
	"errors"

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

// AuthService defines the methods for authentication-related operations.
type AuthService interface {
	// Login logs user in using provided credentials
	//
	// Method should check if such user exists, and if he does create valid JWT tokens and return them.
	// Possible errors: ErrInvalidCredentials, ErrUserNotFound, validator.ValidationErrors,
	// errors returned by JWTService, repository.UserRepository.
	Login(tx *gorm.DB, userLogin schemas.UserLoginRequest) (*schemas.JWTTokens, error)

	// Register registers user if he is not registered yet
	//
	// Method validates provided user data, checks if user with provided email already exists,
	// creates new user, creates JWT tokens for him and returns them.
	// Possible errors: ErrUserAlreadyExists, errors returned by JWTService,
	// repository.UserRepository, bcrypt lib.
	Register(tx *gorm.DB, userRegister schemas.UserRegisterRequest) (*schemas.JWTTokens, error)

	// RefreshTokens refreshes JWT tokens using a valid refresh token
	RefreshTokens(tx *gorm.DB, refreshRequest schemas.RefreshTokenRequest) (*schemas.JWTTokens, error)
}

// authService implements AuthService interface.
type authService struct {
	userRepository repository.UserRepository
	jwtService     JWTService
	logger         *zap.SugaredLogger
}

// Login implements Login method of [AuthService] interface.
func (as *authService) Login(tx *gorm.DB, userLogin schemas.UserLoginRequest) (*schemas.JWTTokens, error) {
	user, err := as.userRepository.GetByEmail(tx, userLogin.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrUserNotFound
		}
		as.logger.Errorf("Error getting user by email: %v", err.Error())
		return nil, err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(userLogin.Password)) != nil {
		as.logger.Errorf("Error comparing password hash: %v", myerrors.ErrInvalidCredentials.Error())
		return nil, myerrors.ErrInvalidCredentials
	}

	tokens, err := as.jwtService.GenerateTokens(tx, user.ID)
	if err != nil {
		as.logger.Errorf("Error generating JWT tokens: %v", err.Error())
		return nil, err
	}
	return tokens, nil
}

// Register implements Register method of [AuthService] interface.
func (as *authService) Register(tx *gorm.DB, userRegister schemas.UserRegisterRequest) (*schemas.JWTTokens, error) {
	user, err := as.userRepository.GetByEmail(tx, userRegister.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		as.logger.Errorf("Error getting user by email: %v", err.Error())
		return nil, err
	}

	if user != nil {
		as.logger.Errorf("User already exists")
		return nil, myerrors.ErrUserAlreadyExists
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

	userID, err := as.userRepository.Create(tx, userModel)
	if err != nil {
		as.logger.Errorf("Error creating user: %v", err.Error())
		return nil, err
	}

	tokens, err := as.jwtService.GenerateTokens(tx, userID)
	if err != nil {
		as.logger.Errorf("Error generating JWT tokens: %v", err.Error())
		return nil, err
	}

	as.logger.Infof("User registered successfully")
	return tokens, nil
}

// RefreshTokens implements RefreshTokens method of [AuthService] interface
func (as *authService) RefreshTokens(
	tx *gorm.DB,
	refreshRequest schemas.RefreshTokenRequest,
) (*schemas.JWTTokens, error) {
	tokens, err := as.jwtService.RefreshTokens(tx, refreshRequest.RefreshToken)
	if err != nil {
		as.logger.Errorf("Error refreshing JWT tokens: %v", err.Error())
		return nil, err
	}

	as.logger.Infof("Tokens refreshed successfully")
	return tokens, nil
}

// NewAuthService creates new instance of [authService]
func NewAuthService(userRepository repository.UserRepository, jwtService JWTService) AuthService {
	log := utils.NewNamedLogger("auth_service")
	return &authService{
		userRepository: userRepository,
		jwtService:     jwtService,
		logger:         log,
	}
}
