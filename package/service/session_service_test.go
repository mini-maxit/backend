package service

import (
	"reflect"
	"testing"
	"time"

	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type sessionServiceTest struct {
	tx *gorm.DB
	ur *testutils.MockUserRepository
	sr *testutils.MockSessionRepository
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
			Role:         types.UserRoleAdmin,
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
		assert.Equal(t, user.Role, validateSession.User.Role)
	})
	t.Run("Session expired", func(t *testing.T) {
		user := &models.User{
			Name:         "test-name",
			Surname:      "test-surname",
			Email:        "test-email",
			Username:     "test-username",
			PasswordHash: "test-password-hash",
			Role:         types.UserRoleAdmin,
		}
		userId, err := sst.ur.CreateUser(sst.tx, user)
		assert.NoError(t, err)
		session := &models.Session{
			Id:        "test-session-id",
			UserId:    userId,
			ExpiresAt: time.Now().Add(-time.Hour),
		}
		err = sst.sr.CreateSession(sst.tx, session)
		assert.NoError(t, err)
		validateSession, err := sst.ss.ValidateSession(sst.tx, session.Id)
		assert.ErrorIs(t, err, ErrSessionExpired)
		assert.False(t, validateSession.Valid)
		assert.Equal(t, InvalidUser.Id, validateSession.User.Id)
	})

	t.Run("Unexpected error during get session", func(t *testing.T) {
		sst.sr.FailNext()
		validateSession, err := sst.ss.ValidateSession(sst.tx, "test-session-id")
		assert.Error(t, err)
		assert.False(t, validateSession.Valid)
		assert.Equal(t, InvalidUser.Id, validateSession.User.Id)
	})

	t.Run("Unexpected error during get user", func(t *testing.T) {
		user := &models.User{
			Name:         "test-name",
			Surname:      "test-surname",
			Email:        "test-email",
			Username:     "test-username",
			PasswordHash: "test-password-hash",
			Role:         types.UserRoleAdmin,
		}
		userId, err := sst.ur.CreateUser(sst.tx, user)
		assert.NoError(t, err)
		session, err := sst.ss.CreateSession(sst.tx, userId)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		sst.ur.FailNext()
		validateSession, err := sst.ss.ValidateSession(sst.tx, session.Id)
		assert.Error(t, err)
		assert.False(t, validateSession.Valid)
		assert.Equal(t, InvalidUser.Id, validateSession.User.Id)
	})

	t.Run("User not found", func(t *testing.T) {
		session := &models.Session{
			Id:        "test-session-id",
			UserId:    -1234,
			ExpiresAt: time.Now().Add(time.Hour),
		}
		err := sst.sr.CreateSession(sst.tx, session)
		assert.NoError(t, err)
		validateSession, err := sst.ss.ValidateSession(sst.tx, session.Id)
		assert.Error(t, err)
		assert.False(t, validateSession.Valid)
		assert.Equal(t, InvalidUser.Id, validateSession.User.Id)
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

func TestCreateSession(t *testing.T) {
	sst := newSessionServiceTest()

	t.Run("User not found", func(t *testing.T) {
		session, err := sst.ss.CreateSession(sst.tx, 0)
		assert.ErrorIs(t, err, errors.ErrNotFound)
		assert.Nil(t, session)
	})

	t.Run("Unexpected error during get user", func(t *testing.T) {
		sst.ur.FailNext()
		session, err := sst.ss.CreateSession(sst.tx, -1)
		assert.Error(t, err)
		assert.Nil(t, session)
	})

	t.Run("Unexpected error during get session", func(t *testing.T) {
		sst.sr.FailNext()
		userId, err := sst.ur.CreateUser(sst.tx, &models.User{
			Name:         "test-name",
			Surname:      "test-surname",
			Email:        "test-email",
			Username:     "test-username",
			PasswordHash: "test-password-hash",
			Role:         types.UserRoleAdmin,
		})
		assert.NoError(t, err)
		session, err := sst.ss.CreateSession(sst.tx, userId)
		if !assert.NotEqual(t, err, errors.ErrNotFound) {
			t.Fatalf("Invalid error, exp!=%q, got=%q", errors.ErrNotFound, err)
		}
		assert.Error(t, err)
		assert.Nil(t, session)
	})

	t.Run("Session expired", func(t *testing.T) {
		userId, err := sst.ur.CreateUser(sst.tx, &models.User{
			Name:         "test-name",
			Surname:      "test-surname",
			Email:        "test-email",
			Username:     "test-username",
			PasswordHash: "test-password-hash",
			Role:         types.UserRoleAdmin,
		})
		assert.NoError(t, err)
		expSession := &models.Session{
			Id:        "test-session-id",
			UserId:    userId,
			ExpiresAt: time.Now().Add(-time.Hour),
		}
		err = sst.sr.CreateSession(sst.tx, expSession)
		assert.NoError(t, err)
		session, err := sst.ss.CreateSession(sst.tx, userId)
		assert.NoError(t, err)
		assert.NotNil(t, session)
	})

	t.Run("Session exists", func(t *testing.T) {
		userId, err := sst.ur.CreateUser(sst.tx, &models.User{
			Name:         "test-name",
			Surname:      "test-surname",
			Email:        "test-email",
			Username:     "test-username",
			PasswordHash: "test-password-hash",
			Role:         types.UserRoleAdmin,
		})
		assert.NoError(t, err)
		existingSession := &models.Session{
			Id:        "test-session-id",
			UserId:    userId,
			ExpiresAt: time.Now().Add(+time.Hour),
		}
		err = sst.sr.CreateSession(sst.tx, existingSession)
		assert.NoError(t, err)
		session, err := sst.ss.CreateSession(sst.tx, userId)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.IsType(t, &schemas.Session{}, session)
		assert.Equal(t, existingSession.Id, session.Id)
	})
}
