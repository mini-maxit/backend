package service_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type sessionServiceTest struct {
	tx *gorm.DB
	ur *testutils.MockUserRepository
	sr *testutils.MockSessionRepository
	ss service.SessionService
}

func newSessionServiceTest() *sessionServiceTest {
	tx := &gorm.DB{}
	ur := testutils.NewMockUserRepository()
	sr := testutils.NewMockSessionRepository()
	ss := service.NewSessionService(sr, ur)
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
		validateSession, err := sst.ss.Validate(sst.tx, "test-session-id")
		require.ErrorIs(t, err, errors.ErrSessionNotFound)
		assert.False(t, validateSession.Valid)
		assert.Equal(t, service.InvalidUser.ID, validateSession.User.ID)
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
		userID, err := sst.ur.Create(sst.tx, user)
		require.NoError(t, err)
		session, err := sst.ss.Create(sst.tx, userID)
		require.NoError(t, err)
		assert.NotNil(t, session)
		validateSession, err := sst.ss.Validate(sst.tx, session.ID)
		require.NoError(t, err)
		assert.True(t, validateSession.Valid)
		assert.Equal(t, userID, validateSession.User.ID)
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
		userID, err := sst.ur.Create(sst.tx, user)
		require.NoError(t, err)
		session := &models.Session{
			ID:        "test-session-id",
			UserID:    userID,
			ExpiresAt: time.Now().Add(-time.Hour),
		}
		err = sst.sr.Create(sst.tx, session)
		require.NoError(t, err)
		validateSession, err := sst.ss.Validate(sst.tx, session.ID)
		require.ErrorIs(t, err, errors.ErrSessionExpired)
		assert.False(t, validateSession.Valid)
		assert.Equal(t, service.InvalidUser.ID, validateSession.User.ID)
	})

	t.Run("Unexpected error during get session", func(t *testing.T) {
		sst.sr.FailNext()
		validateSession, err := sst.ss.Validate(sst.tx, "test-session-id")
		require.Error(t, err)
		assert.False(t, validateSession.Valid)
		assert.Equal(t, service.InvalidUser.ID, validateSession.User.ID)
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
		userID, err := sst.ur.Create(sst.tx, user)
		require.NoError(t, err)
		session, err := sst.ss.Create(sst.tx, userID)
		require.NoError(t, err)
		assert.NotNil(t, session)
		sst.ur.FailNext()
		validateSession, err := sst.ss.Validate(sst.tx, session.ID)
		require.Error(t, err)
		assert.False(t, validateSession.Valid)
		assert.Equal(t, service.InvalidUser.ID, validateSession.User.ID)
	})

	t.Run("User not found", func(t *testing.T) {
		session := &models.Session{
			ID:        "test-session-id",
			UserID:    -1234,
			ExpiresAt: time.Now().Add(time.Hour),
		}
		err := sst.sr.Create(sst.tx, session)
		require.NoError(t, err)
		validateSession, err := sst.ss.Validate(sst.tx, session.ID)
		require.Error(t, err)
		assert.False(t, validateSession.Valid)
		assert.Equal(t, service.InvalidUser.ID, validateSession.User.ID)
	})
}

func TestInvalidateSession(t *testing.T) {
	sst := newSessionServiceTest()
	t.Run("Session not found", func(t *testing.T) {
		err := sst.ss.Invalidate(sst.tx, "test-session-id")
		require.NoError(t, err)
	})
	t.Run("Session found", func(t *testing.T) {
		userID, err := sst.ur.Create(sst.tx, &models.User{
			Name:         "test-name",
			Surname:      "test-surname",
			Email:        "test-email",
			Username:     "test-username",
			PasswordHash: "test-password-hash",
			Role:         "admin",
		})
		require.NoError(t, err)
		session, err := sst.ss.Create(sst.tx, userID)
		require.NoError(t, err)
		assert.NotNil(t, session)
		err = sst.ss.Invalidate(sst.tx, session.ID)
		require.NoError(t, err)
		validateSession, err := sst.ss.Validate(sst.tx, session.ID)
		require.ErrorIs(t, err, errors.ErrSessionNotFound)
		assert.False(t, validateSession.Valid)
		assert.Equal(t, service.InvalidUser.ID, validateSession.User.ID)
		assert.True(t, reflect.DeepEqual(service.InvalidUser, validateSession.User))
	})
}

func TestCreateSession(t *testing.T) {
	sst := newSessionServiceTest()

	t.Run("User not found", func(t *testing.T) {
		session, err := sst.ss.Create(sst.tx, 0)
		require.ErrorIs(t, err, errors.ErrNotFound)
		assert.Nil(t, session)
	})

	t.Run("Unexpected error during get user", func(t *testing.T) {
		sst.ur.FailNext()
		session, err := sst.ss.Create(sst.tx, -1)
		require.Error(t, err)
		assert.Nil(t, session)
	})

	t.Run("Unexpected error during get session", func(t *testing.T) {
		sst.sr.FailNext()
		userID, err := sst.ur.Create(sst.tx, &models.User{
			Name:         "test-name",
			Surname:      "test-surname",
			Email:        "test-email",
			Username:     "test-username",
			PasswordHash: "test-password-hash",
			Role:         types.UserRoleAdmin,
		})
		require.NoError(t, err)
		session, err := sst.ss.Create(sst.tx, userID)
		if !assert.NotEqual(t, err, errors.ErrNotFound) {
			t.Fatalf("Invalid error, exp!=%q, got=%q", errors.ErrNotFound, err)
		}
		require.Error(t, err)
		assert.Nil(t, session)
	})

	t.Run("Session expired", func(t *testing.T) {
		userID, err := sst.ur.Create(sst.tx, &models.User{
			Name:         "test-name",
			Surname:      "test-surname",
			Email:        "test-email",
			Username:     "test-username",
			PasswordHash: "test-password-hash",
			Role:         types.UserRoleAdmin,
		})
		require.NoError(t, err)
		expSession := &models.Session{
			ID:        "test-session-id",
			UserID:    userID,
			ExpiresAt: time.Now().Add(-time.Hour),
		}
		err = sst.sr.Create(sst.tx, expSession)
		require.NoError(t, err)
		session, err := sst.ss.Create(sst.tx, userID)
		require.NoError(t, err)
		assert.NotNil(t, session)
	})

	t.Run("Session exists", func(t *testing.T) {
		userID, err := sst.ur.Create(sst.tx, &models.User{
			Name:         "test-name",
			Surname:      "test-surname",
			Email:        "test-email",
			Username:     "test-username",
			PasswordHash: "test-password-hash",
			Role:         types.UserRoleAdmin,
		})
		require.NoError(t, err)
		existingSession := &models.Session{
			ID:        "test-session-id",
			UserID:    userID,
			ExpiresAt: time.Now().Add(+time.Hour),
		}
		err = sst.sr.Create(sst.tx, existingSession)
		require.NoError(t, err)
		session, err := sst.ss.Create(sst.tx, userID)
		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.IsType(t, &schemas.Session{}, session)
		assert.Equal(t, existingSession.ID, session.ID)
	})
}
