package service

import (
	"reflect"
	"testing"

	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type sessionServiceTest struct {
	tx *gorm.DB
	ur repository.UserRepository
	sr repository.SessionRepository
	ss SessionService
}

func newSessionServiceTest() *sessionServiceTest {
	tx := &gorm.DB{}
	ur := testutils.NewMockUserRepository()
	sr := testutils.NewMockSessionRepository()
	ss := NewSessionService(sr, ur)
	return &sessionServiceTest{
		tx: tx,
		ur: ur,
		sr: sr,
		ss: ss,
	}
}

func TestValidateSession(t *testing.T) {
	sst := newSessionServiceTest()
	t.Run("Session not found", func(t *testing.T) {
		validateSession, err := sst.ss.ValidateSession(sst.tx, "test-session-id")
		assert.ErrorIs(t, err, ErrSessionNotFound)
		assert.False(t, validateSession.Valid)
		assert.Equal(t, InvalidUser.Id, validateSession.User.Id)
	})
	t.Run("Session found", func(t *testing.T) {
		user := &models.User{
			Name:         "test-name",
			Surname:      "test-surname",
			Email:        "test-email",
			Username:     "test-username",
			PasswordHash: "test-password-hash",
			Role:         models.UserRoleAdmin,
		}
		userId, err := sst.ur.CreateUser(sst.tx, user)
		assert.NoError(t, err)
		session, err := sst.ss.CreateSession(sst.tx, userId)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		validateSession, err := sst.ss.ValidateSession(sst.tx, session.Id)
		assert.NoError(t, err)
		assert.True(t, validateSession.Valid)
		assert.Equal(t, userId, validateSession.User.Id)
		assert.Equal(t, user.Name, validateSession.User.Name)
		assert.Equal(t, user.Surname, validateSession.User.Surname)
		assert.Equal(t, user.Email, validateSession.User.Email)
		assert.Equal(t, user.Username, validateSession.User.Username)
		assert.Equal(t, string(user.Role), validateSession.User.Role)
	})
}

func TestInvalidateSession(t *testing.T) {
	sst := newSessionServiceTest()
	t.Run("Session not found", func(t *testing.T) {
		err := sst.ss.InvalidateSession(sst.tx, "test-session-id")
		assert.NoError(t, err)
	})
	t.Run("Session found", func(t *testing.T) {
		userId, err := sst.ur.CreateUser(sst.tx, &models.User{
			Name:         "test-name",
			Surname:      "test-surname",
			Email:        "test-email",
			Username:     "test-username",
			PasswordHash: "test-password-hash",
			Role:         "admin",
		})
		assert.NoError(t, err)
		session, err := sst.ss.CreateSession(sst.tx, userId)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		err = sst.ss.InvalidateSession(sst.tx, session.Id)
		assert.NoError(t, err)
		validateSession, err := sst.ss.ValidateSession(sst.tx, session.Id)
		assert.ErrorIs(t, err, ErrSessionNotFound)
		assert.False(t, validateSession.Valid)
		assert.Equal(t, InvalidUser.Id, validateSession.User.Id)
		assert.True(t, reflect.DeepEqual(InvalidUser, validateSession.User))
	})
}
