package service

import (
	"strings"
	"testing"

	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister(t *testing.T) {
	tx := testutils.NewTestTx(t)
	ur, err := repository.NewUserRepository(tx)
	assert.NoError(t, err)
	sr, err := repository.NewSessionRepository(tx)
	assert.NoError(t, err)
	ss := NewSessionService(sr, ur)
	as := NewAuthService(ur, ss)
	savePoint := "before"
	tx.SavePoint(savePoint)

	t.Run("validaton of user register request", func(t *testing.T) {
		cases := []struct {
			name     string
			surname  string
			email    string
			username string
			password string

			valid bool
		}{
			{"", "surname", "email@email.com", "username", "password", false},                      // Empty name
			{"a", "surname", "email@email.com", "username", "password", false},                     // Too short name
			{strings.Repeat("a", 51), "surname", "email@email.com", "username", "password", false}, // Too long name
			{"name", "", "email@email.com", "username", "password", false},                         // Empty surname
			{"name", "a", "email@email.com", "username", "password", false},                        // Too short surname
			{"name", strings.Repeat("a", 51), "email@email.com", "username", "password", false},    // Too long surname
			{"name", "surname", "", "username", "password", false},                                 // Empty email
			{"name", "surname", "aaaa", "username", "password", false},                             // Invalid email
			{"name", "surname", "email@email.com", "", "password", false},                          // Empty username
			{"name", "surname", "email@email.com", "a", "password", false},                         // Too short username
			{"name", "surname", "email@email.com", strings.Repeat("a", 31), "password", false},     // Too long username
			{"name", "surname", "email@email.com", "_SuperCoolUsername_", "password", false},       // Invalid username
			{"name", "surname", "email@email.com", "username", "", false},                          // Empty password
			{"name", "surname", "email@email.com", "username", "aaaa", false},                      // Too short password
			{"name", "surname", "email@email.com", "username", strings.Repeat("a", 51), false},     // Too long password
			{"name", "surname", "email@email.com", "username", strings.Repeat("a", 13), true},      // Too long password
		}
		for _, tc := range cases {
			userRegister := schemas.UserRegisterRequest{
				Name:     tc.name,
				Surname:  tc.surname,
				Email:    tc.email,
				Username: tc.username,
				Password: tc.password,
			}
			response, err := as.Register(tx, userRegister)
			if tc.valid {
				assert.NoError(t, err)
				assert.NotNil(t, response)
			} else {
				assert.Error(t, err, tc)
				assert.Nil(t, response, tc)
			}
		}
		tx.RollbackTo(savePoint)
	})
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
		assert.ErrorIs(t, err, ErrUserAlreadyExists)
		assert.Nil(t, response)
		tx.RollbackTo(savePoint)
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
		tx.RollbackTo(savePoint)
	})

	tx.Rollback()
}

func TestLogin(t *testing.T) {
	tx := testutils.NewTestTx(t)
	ur, err := repository.NewUserRepository(tx)
	assert.NoError(t, err)
	sr, err := repository.NewSessionRepository(tx)
	assert.NoError(t, err)
	ss := NewSessionService(sr, ur)
	as := NewAuthService(ur, ss)
	savePoint := "before"
	tx.SavePoint(savePoint)

	t.Run("validation of user login request", func(t *testing.T) {
		cases := []struct {
			email    string
			password string
		}{
			{"", "password"},                             // Empty email
			{"aaaa", "password"},                         // Invalid email
			{"email@email.com", ""},                      // Empty password
			{"email@email.com", "aa"},                    // Short password
			{"email@email.com", strings.Repeat("a", 51)}, // Long password
		}
		for _, tc := range cases {
			userLogin := schemas.UserLoginRequest{
				Email:    tc.email,
				Password: tc.password,
			}
			response, err := as.Login(tx, userLogin)
			assert.Error(t, err, tc)
			assert.Nil(t, response, tc)
		}
		tx.RollbackTo(savePoint)
	})

	t.Run("get user by email when user does not exist", func(t *testing.T) {
		userLogin := schemas.UserLoginRequest{
			Email:    "email@email.com",
			Password: "password",
		}
		response, err := as.Login(tx, userLogin)
		assert.ErrorIs(t, err, ErrUserNotFound)
		assert.Nil(t, response)
		tx.RollbackTo(savePoint)
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
		assert.ErrorIs(t, err, ErrInvalidCredentials)
		assert.Nil(t, response)
		tx.RollbackTo(savePoint)
	})

	t.Run("successful user login", func(t *testing.T) {
		password := "password"
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		assert.NoError(t, err)
		user := &models.User{
			Name:         "name",
			Surname:      "surname",
			Email:        "email@email.com",
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
		tx.RollbackTo(savePoint)
	})

	tx.Rollback()
}
