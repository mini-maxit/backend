package service_test

import (
	"testing"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	myerrors "github.com/mini-maxit/backend/package/errors"
	mock_repository "github.com/mini-maxit/backend/package/repository/mocks"
	"github.com/mini-maxit/backend/package/service"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ur := mock_repository.NewMockUserRepository(ctrl)
	js := mock_service.NewMockJWTService(ctrl)
	as := service.NewAuthService(ur, js)
	tx := &gorm.DB{}

	t.Run("get user by email when user exists", func(t *testing.T) {
		ur.EXPECT().GetByEmail(tx, "email2@email.com").Return(&models.User{
			ID:           1,
			Name:         "name",
			Surname:      "surname",
			Email:        "email2@email.com",
			Username:     "username2",
			PasswordHash: "password",
		}, nil).Times(1)

		userRegister := schemas.UserRegisterRequest{
			Name:     "name",
			Surname:  "surname",
			Email:    "email2@email.com",
			Username: "username",
			Password: "Password123!",
		}
		response, err := as.Register(tx, userRegister)
		require.ErrorIs(t, err, myerrors.ErrUserAlreadyExists)
		assert.Nil(t, response)
	})

	t.Run("successful user registration", func(t *testing.T) {
		ur.EXPECT().GetByEmail(tx, "email3@email.com").Return(nil, gorm.ErrRecordNotFound).Times(1)
		ur.EXPECT().Create(tx, gomock.Any()).Return(int64(1), nil).Times(1)
		js.EXPECT().GenerateTokens(tx, int64(1)).Return(&schemas.JWTTokens{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
		}, nil).Times(1)

		userRegister := schemas.UserRegisterRequest{
			Name:     "name",
			Surname:  "surname",
			Email:    "email3@email.com",
			Username: "username3",
			Password: "Password123!",
		}
		response, err := as.Register(tx, userRegister)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.IsType(t, &schemas.JWTTokens{}, response)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
	})

	t.Run("unexpected repository error", func(t *testing.T) {
		ur.EXPECT().GetByEmail(tx, "email4@email.com").Return(nil, gorm.ErrInvalidDB).Times(1)
		userRegister := schemas.UserRegisterRequest{
			Name:     "name",
			Surname:  "surname",
			Email:    "email4@email.com",
			Username: "username4",
			Password: "Password123!",
		}
		response, err := as.Register(tx, userRegister)
		require.ErrorIs(t, err, gorm.ErrInvalidDB)
		assert.Nil(t, response)
	})

	t.Run("failed to create user", func(t *testing.T) {
		ur.EXPECT().GetByEmail(tx, "email5@email.com").Return(nil, gorm.ErrRecordNotFound).Times(1)
		ur.EXPECT().Create(tx, gomock.Any()).Return(int64(0), gorm.ErrInvalidDB).Times(1)

		userRegister := schemas.UserRegisterRequest{
			Name:     "name",
			Surname:  "surname",
			Email:    "email5@email.com",
			Username: "username5",
			Password: "Password123!",
		}
		response, err := as.Register(tx, userRegister)
		require.ErrorIs(t, err, gorm.ErrInvalidDB)
		assert.Nil(t, response)
	})
}

func TestLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ur := mock_repository.NewMockUserRepository(ctrl)
	js := mock_service.NewMockJWTService(ctrl)
	as := service.NewAuthService(ur, js)
	tx := &gorm.DB{}

	password := "Password123!"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user := &models.User{
		ID:           1,
		Name:         "name",
		Surname:      "surname",
		Email:        "email5@email.com",
		Username:     "username",
		PasswordHash: string(hash),
	}

	t.Run("get user by email when user does not exist", func(t *testing.T) {
		ur.EXPECT().GetByEmail(tx, "nonexistent@email.com").Return(nil, gorm.ErrRecordNotFound).Times(1)

		userLogin := schemas.UserLoginRequest{
			Email:    "nonexistent@email.com",
			Password: "password",
		}

		response, err := as.Login(tx, userLogin)
		require.ErrorIs(t, err, myerrors.ErrUserNotFound)
		assert.Nil(t, response)
	})

	t.Run("compare password hash fails", func(t *testing.T) {
		ur.EXPECT().GetByEmail(tx, "email5@email.com").Return(user, nil).Times(1)

		userLogin := schemas.UserLoginRequest{
			Email:    "email5@email.com",
			Password: "wrongpassword",
		}

		response, err := as.Login(tx, userLogin)
		require.ErrorIs(t, err, myerrors.ErrInvalidCredentials)
		assert.Nil(t, response)
	})

	t.Run("successful user login", func(t *testing.T) {
		ur.EXPECT().GetByEmail(tx, "email5@email.com").Return(user, nil).Times(1)
		js.EXPECT().GenerateTokens(tx, user.ID).Return(&schemas.JWTTokens{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
		}, nil).Times(1)

		userLogin := schemas.UserLoginRequest{
			Email:    "email5@email.com",
			Password: password,
		}

		response, err := as.Login(tx, userLogin)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.IsType(t, &schemas.JWTTokens{}, response)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
	})
}

func TestRefreshTokens(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ur := mock_repository.NewMockUserRepository(ctrl)
	js := mock_service.NewMockJWTService(ctrl)
	as := service.NewAuthService(ur, js)
	tx := &gorm.DB{}

	t.Run("successful token refresh", func(t *testing.T) {
		// userID := int64(1)
		js.EXPECT().RefreshTokens(tx, "initial-refresh-token").Return(&schemas.JWTTokens{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
		}, nil).Times(1)

		// Refresh tokens using the refresh token
		refreshRequest := schemas.RefreshTokenRequest{
			RefreshToken: "initial-refresh-token",
		}
		newTokens, err := as.RefreshTokens(tx, refreshRequest)
		require.NoError(t, err)
		assert.NotNil(t, newTokens)
		assert.IsType(t, &schemas.JWTTokens{}, newTokens)
		assert.NotEmpty(t, newTokens.AccessToken)
		assert.NotEmpty(t, newTokens.RefreshToken)

		// New tokens should be different from initial tokens
		assert.NotEqual(t, "initial-access-token", newTokens.AccessToken)
		assert.NotEqual(t, "initial-refresh-token", newTokens.RefreshToken)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		js.EXPECT().RefreshTokens(tx, "invalid-refresh-token").Return(nil, myerrors.ErrInvalidTokenType).Times(1)
		refreshRequest := schemas.RefreshTokenRequest{
			RefreshToken: "invalid-refresh-token",
		}
		newTokens, err := as.RefreshTokens(tx, refreshRequest)
		require.Error(t, err)
		require.Nil(t, newTokens)
	})
}
