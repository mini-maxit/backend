package service_test

import (
	"testing"

	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestRegister(t *testing.T) {
	ur := testutils.NewMockUserRepository()
	sr := testutils.NewMockSessionRepository()
	ss := service.NewSessionService(sr, ur)
	as := service.NewAuthService(ur, ss)
	tx := &gorm.DB{}

	t.Run("get user by email when user exists", func(t *testing.T) {
		user := &models.User{
			Name:         "name",
			Surname:      "surname",
			Email:        "email2@email.com",
			Username:     "username2",
			PasswordHash: "password",
		}
		_, err := ur.Create(tx, user)
		require.NoError(t, err)
		userRegister := schemas.UserRegisterRequest{
			Name:            "name",
			Surname:         "surname",
			Email:           "email2@email.com",
			Username:        "username",
			Password:        "Password123!",
			ConfirmPassword: "Password123!",
		}
		response, err := as.Register(tx, userRegister)
		require.ErrorIs(t, err, errors.ErrUserAlreadyExists)
		assert.Nil(t, response)
	})

	t.Run("successful user registration", func(t *testing.T) {
		userRegister := schemas.UserRegisterRequest{
			Name:            "name",
			Surname:         "surname",
			Email:           "email3@email.com",
			Username:        "username3",
			Password:        "Password123!",
			ConfirmPassword: "Password123!",
		}
		response, err := as.Register(tx, userRegister)
		require.NoError(t, err)
		assert.NotNil(t, response)
	})

	t.Run("unexpected repostiroy error", func(t *testing.T) {
		userRegister := schemas.UserRegisterRequest{
			Name:            "name",
			Surname:         "surname",
			Email:           "email4@email.com",
			Username:        "username4",
			Password:        "Password123!",
			ConfirmPassword: "Password123!",
		}
		response, err := as.Register(nil, userRegister)
		require.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestLogin(t *testing.T) {
	ur := testutils.NewMockUserRepository()
	sr := testutils.NewMockSessionRepository()
	ss := service.NewSessionService(sr, ur)
	as := service.NewAuthService(ur, ss)
	tx := &gorm.DB{}

	t.Run("get user by email when user does not exist", func(t *testing.T) {
		userLogin := schemas.UserLoginRequest{
			Email:    "email@email.com",
			Password: "password",
		}
		response, err := as.Login(tx, userLogin)
		require.ErrorIs(t, err, errors.ErrUserNotFound)
		assert.Nil(t, response)
	})

	t.Run("compare password hash", func(t *testing.T) {
		user := &models.User{
			Name:         "name",
			Surname:      "surname",
			Email:        "email@email.com",
			Username:     "username",
			PasswordHash: "password",
		}
		_, err := ur.Create(tx, user)
		require.NoError(t, err)
		userLogin := schemas.UserLoginRequest{
			Email:    user.Email,
			Password: user.PasswordHash,
		}
		response, err := as.Login(tx, userLogin)
		require.ErrorIs(t, err, errors.ErrInvalidCredentials)
		assert.Nil(t, response)
	})

	t.Run("successful user login", func(t *testing.T) {
		password := "supersecretpassword"
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		require.NoError(t, err)
		user := &models.User{
			Name:         "name",
			Surname:      "surname",
			Email:        "email3@email.com",
			Username:     "username",
			PasswordHash: string(passwordHash),
		}
		_, err = ur.Create(tx, user)
		require.NoError(t, err)
		userLogin := schemas.UserLoginRequest{
			Email:    user.Email,
			Password: password,
		}
		response, err := as.Login(tx, userLogin)
		require.NoError(t, err)
		assert.NotNil(t, response)
	})
}
