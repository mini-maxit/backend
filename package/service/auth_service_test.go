package service

import (
	"testing"

	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestRegister(t *testing.T) {
	ur := testutils.NewMockUserRepository()
	sr := testutils.NewMockSessionRepository()
	ss := NewSessionService(sr, ur)
	as := NewAuthService(ur, ss)
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
			Password: "password",
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
			Password: "password",
		}
		response, err := as.Register(tx, userRegister)
		assert.NoError(t, err)
		assert.NotNil(t, response)
	})

	t.Run("unexpected repostiroy error", func(t *testing.T) {
		userRegister := schemas.UserRegisterRequest{
			Name:     "name",
			Surname:  "surname",
			Email:    "email4@email.com",
			Username: "username4",
			Password: "password",
		}
		response, err := as.Register(nil, userRegister)
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestLogin(t *testing.T) {
	ur := testutils.NewMockUserRepository()
	sr := testutils.NewMockSessionRepository()
	ss := NewSessionService(sr, ur)
	as := NewAuthService(ur, ss)
	tx := &gorm.DB{}

	t.Run("get user by email when user does not exist", func(t *testing.T) {
		userLogin := schemas.UserLoginRequest{
			Email:    "email@email.com",
			Password: "password",
		}
		response, err := as.Login(tx, userLogin)
		assert.ErrorIs(t, err, errors.ErrUserNotFound)
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
		ur.CreateUser(tx, user)
		userLogin := schemas.UserLoginRequest{
			Email:    user.Email,
			Password: user.PasswordHash,
		}
		response, err := as.Login(tx, userLogin)
		assert.ErrorIs(t, err, errors.ErrInvalidCredentials)
		assert.Nil(t, response)
	})

	t.Run("successful user login", func(t *testing.T) {
		password := "supersecretpassword"
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		assert.NoError(t, err)
		user := &models.User{
			Name:         "name",
			Surname:      "surname",
			Email:        "email3@email.com",
			Username:     "username",
			PasswordHash: string(passwordHash),
		}
		ur.CreateUser(tx, user)
		userLogin := schemas.UserLoginRequest{
			Email:    user.Email,
			Password: password,
		}
		response, err := as.Login(tx, userLogin)
		assert.NoError(t, err)
		assert.NotNil(t, response)
	})
}
