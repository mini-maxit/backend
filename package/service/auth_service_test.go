package service_test

import (
	"testing"
	"time"

	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	myerrors "github.com/mini-maxit/backend/package/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestRegister(t *testing.T) {
	ur := testutils.NewMockUserRepository()
	js := NewJWTService(ur, "test-secret-key")
	as := NewAuthService(ur, js)
	tx := &gorm.DB{}

	t.Run("get user by email when user exists", func(t *testing.T) {
		ur.CreateUser(tx, &models.User{
			Name:         "name",
			Surname:      "surname",
			Email:        "email2@email.com",
			Username:     "username2",
			PasswordHash: "password",
		})
		userRegister := schemas.UserRegisterRequest{
			Name:     "name",
			Surname:  "surname",
			Email:    "email2@email.com",
			Username: "username",
			Password: "Password123!",
		}
		response, err := as.Register(tx, userRegister)
		assert.ErrorIs(t, err, errors.ErrUserAlreadyExists)
		assert.Nil(t, response)
	})

	t.Run("successful user registration", func(t *testing.T) {
		userRegister := schemas.UserRegisterRequest{
			Name:     "name",
			Surname:  "surname",
			Email:    "email3@email.com",
			Username: "username3",
			Password: "Password123!",
		}
		response, err := as.Register(tx, userRegister)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.IsType(t, &schemas.JWTTokens{}, response)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
		assert.Equal(t, "Bearer", response.TokenType)
	})

	t.Run("unexpected repository error", func(t *testing.T) {
		userRegister := schemas.UserRegisterRequest{
			Name:     "name",
			Surname:  "surname",
			Email:    "email4@email.com",
			Username: "username4",
			Password: "Password123!",
		}
		response, err := as.Register(nil, userRegister)
		require.ErrorIs(t, err, gorm.ErrInvalidDB)
		assert.Nil(t, response)
	})

	t.Run("failed to create user", func(t *testing.T) {
		ur.EXPECT().GetByEmail(
			gomock.Any(),
			userRegister.Email,
		).Return(nil, nil).Times(1)

		ur.EXPECT().Create(
			gomock.Any(),
			gomock.Any(),
		).Return(int64(0), gorm.ErrInvalidDB).Times(1)

		response, err := as.Register(nil, userRegister)
		require.ErrorIs(t, err, gorm.ErrInvalidDB)
		assert.Nil(t, response)
	})

	t.Run("failed to create session", func(t *testing.T) {
		ur.EXPECT().GetByEmail(
			gomock.Any(),
			userRegister.Email,
		).Return(nil, nil).Times(1)

		ur.EXPECT().Create(
			gomock.Any(),
			gomock.Any(),
		).Return(int64(1), nil).Times(1)

		ss.EXPECT().Create(
			gomock.Any(),
			int64(1),
		).Return(nil, gorm.ErrInvalidDB).Times(1)

		response, err := as.Register(nil, userRegister)
		require.ErrorIs(t, err, gorm.ErrInvalidDB)
		assert.Nil(t, response)
	})
}

func TestLogin(t *testing.T) {
	ur := testutils.NewMockUserRepository()
	js := NewJWTService(ur, "test-secret-key")
	as := NewAuthService(ur, js)
	tx := &gorm.DB{}

	password := "Password123!"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &models.User{
		Name:         "name",
		Surname:      "surname",
		Email:        "email5@email.com",
		Username:     "username",
		PasswordHash: string(hash),
	}
	t.Run("get user by email when user does not exist", func(t *testing.T) {
		ur.EXPECT().GetByEmail(
			gomock.Any(),
			"nonexistent@email.com",
		).Return(nil, gorm.ErrRecordNotFound).Times(1)

		userLogin := schemas.UserLoginRequest{
			Email:    "nonexistent@email.com",
			Password: "password",
		}

		response, err := as.Login(tx, userLogin)
		require.ErrorIs(t, err, myerrors.ErrUserNotFound)
		assert.Nil(t, response)
	})

	t.Run("compare password hash fails", func(t *testing.T) {
		ur.EXPECT().GetByEmail(
			gomock.Any(),
			user.Email,
		).Return(user, nil).Times(1)

		userLogin := schemas.UserLoginRequest{
			Email:    user.Email,
			Password: "wrongpassword",
		}

		response, err := as.Login(tx, userLogin)
		require.ErrorIs(t, err, myerrors.ErrInvalidCredentials)
		assert.Nil(t, response)
	})

	t.Run("successful user login", func(t *testing.T) {
		ur.EXPECT().GetByEmail(
			gomock.Any(),
			user.Email,
		).Return(user, nil).Times(1)

		userLogin := schemas.UserLoginRequest{
			Email:    user.Email,
			Password: password,
		}

		ss.EXPECT().Create(
			gomock.Any(),
			user.ID,
		).DoAndReturn(func(tx *gorm.DB, userID int64) (*schemas.Session, error) {
			return &schemas.Session{
				ID:        "session-id",
				UserID:    userID,
				UserRole:  string(types.UserRoleStudent),
				ExpiresAt: time.Now().Add(time.Hour * 24),
			}, nil
		}).Times(1)

		response, err := as.Login(tx, userLogin)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.IsType(t, &schemas.JWTTokens{}, response)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
		assert.Equal(t, "Bearer", response.TokenType)
	})
}

func TestRefreshTokens(t *testing.T) {
	ur := testutils.NewMockUserRepository()
	js := NewJWTService(ur, "test-secret-key")
	as := NewAuthService(ur, js)
	tx := &gorm.DB{}

	t.Run("successful token refresh", func(t *testing.T) {
		// First create a user and get initial tokens
		password := "supersecretpassword"
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		assert.NoError(t, err)
		user := &models.User{
			Name:         "name",
			Surname:      "surname",
			Email:        "refresh@email.com",
			Username:     "refresh_username",
			PasswordHash: string(passwordHash),
		}
		ur.CreateUser(tx, user)

		// Generate initial tokens
		initialTokens, err := js.GenerateTokens(tx, user.Id)
		assert.NoError(t, err)
		assert.NotNil(t, initialTokens)

		// Refresh tokens using the refresh token
		refreshRequest := schemas.RefreshTokenRequest{
			RefreshToken: initialTokens.RefreshToken,
		}
		newTokens, err := as.RefreshTokens(tx, refreshRequest)
		assert.NoError(t, err)
		assert.NotNil(t, newTokens)
		assert.IsType(t, &schemas.JWTTokens{}, newTokens)
		assert.NotEmpty(t, newTokens.AccessToken)
		assert.NotEmpty(t, newTokens.RefreshToken)
		assert.Equal(t, "Bearer", newTokens.TokenType)

		// New tokens should be different from initial tokens
		assert.NotEqual(t, initialTokens.AccessToken, newTokens.AccessToken)
		assert.NotEqual(t, initialTokens.RefreshToken, newTokens.RefreshToken)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		refreshRequest := schemas.RefreshTokenRequest{
			RefreshToken: "invalid-refresh-token",
		}
		newTokens, err := as.RefreshTokens(tx, refreshRequest)
		assert.Error(t, err)
		assert.Nil(t, newTokens)
	})
}
