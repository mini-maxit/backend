package service_test

import (
	"testing"
	"time"

	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	mock_repository "github.com/mini-maxit/backend/package/repository/mocks"
	"github.com/mini-maxit/backend/package/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestValidateSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	tx := &gorm.DB{}
	sr := mock_repository.NewMockSessionRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	ss := service.NewSessionService(sr, ur)

	sessionID := "test-session-id"
	user := &models.User{
		ID:           1,
		Name:         "test-name",
		Surname:      "test-surname",
		Email:        "test-email",
		Username:     "test-username",
		PasswordHash: "test-password-hash",
		Role:         types.UserRoleAdmin,
	}
	session := &models.Session{
		ID:        sessionID,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	t.Run("Session not found", func(t *testing.T) {
		sr.EXPECT().Get(tx, sessionID).Return(nil, gorm.ErrRecordNotFound).Times(1)
		validateSession, err := ss.Validate(tx, sessionID)
		require.ErrorIs(t, err, errors.ErrSessionNotFound)
		assert.False(t, validateSession.Valid)
		assert.Equal(t, service.InvalidUser.ID, validateSession.User.ID)
	})

	t.Run("Session found", func(t *testing.T) {
		sr.EXPECT().Get(tx, session.ID).Return(session, nil).Times(1)
		ur.EXPECT().Get(tx, user.ID).Return(user, nil).Times(1)
		validateSession, err := ss.Validate(tx, session.ID)
		require.NoError(t, err)
		assert.True(t, validateSession.Valid)
		assert.Equal(t, user.ID, validateSession.User.ID)
		assert.Equal(t, user.Name, validateSession.User.Name)
		assert.Equal(t, user.Surname, validateSession.User.Surname)
		assert.Equal(t, user.Email, validateSession.User.Email)
		assert.Equal(t, user.Username, validateSession.User.Username)
		assert.Equal(t, user.Role, validateSession.User.Role)
	})
	t.Run("Session expired", func(t *testing.T) {
		session := &models.Session{
			ID:        "test-session-id",
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(-time.Hour),
		}
		sr.EXPECT().Get(tx, session.ID).Return(session, nil).Times(1)
		ur.EXPECT().Get(tx, user.ID).Return(user, nil).Times(1)
		validateSession, err := ss.Validate(tx, session.ID)
		require.ErrorIs(t, err, errors.ErrSessionExpired)
		assert.False(t, validateSession.Valid)
		assert.Equal(t, service.InvalidUser.ID, validateSession.User.ID)
	})

	t.Run("Unexpected error during get session", func(t *testing.T) {
		sr.EXPECT().Get(tx, sessionID).Return(nil, gorm.ErrInvalidDB).Times(1)
		validateSession, err := ss.Validate(tx, sessionID)
		require.Error(t, err)
		assert.False(t, validateSession.Valid)
		assert.Equal(t, service.InvalidUser.ID, validateSession.User.ID)
	})

	t.Run("Unexpected error during get user", func(t *testing.T) {
		sr.EXPECT().Get(tx, sessionID).Return(session, nil).Times(1)
		ur.EXPECT().Get(tx, session.UserID).Return(nil, gorm.ErrInvalidDB).Times(1)
		validateSession, err := ss.Validate(tx, sessionID)
		require.Error(t, err)
		assert.False(t, validateSession.Valid)
		assert.Equal(t, service.InvalidUser.ID, validateSession.User.ID)
	})

	t.Run("User not found", func(t *testing.T) {
		sr.EXPECT().Get(tx, sessionID).Return(session, nil).Times(1)
		ur.EXPECT().Get(tx, session.UserID).Return(nil, gorm.ErrRecordNotFound).Times(1)
		validateSession, err := ss.Validate(tx, sessionID)
		require.Error(t, err)
		assert.False(t, validateSession.Valid)
		assert.Equal(t, service.InvalidUser.ID, validateSession.User.ID)
	})
}

func TestInvalidateSession(t *testing.T) {
	tx := &gorm.DB{}
	ctrl := gomock.NewController(t)
	sr := mock_repository.NewMockSessionRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	ss := service.NewSessionService(sr, ur)
	sessionID := "test-session-id"

	t.Run("Session not found", func(t *testing.T) {
		sr.EXPECT().Delete(tx, sessionID).Return(gorm.ErrRecordNotFound).Times(1)
		err := ss.Invalidate(tx, sessionID)
		require.ErrorIs(t, err, errors.ErrSessionNotFound)
	})
	t.Run("Session found", func(t *testing.T) {
		sr.EXPECT().Delete(tx, sessionID).Return(nil).Times(1)
		err := ss.Invalidate(tx, sessionID)
		require.NoError(t, err)
	})

	t.Run("Unexpected error durig delete", func(t *testing.T) {
		sr.EXPECT().Delete(tx, sessionID).Return(gorm.ErrInvalidDB).Times(1)
		err := ss.Invalidate(tx, sessionID)
		require.Error(t, err)
	})
}

func TestCreateSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	tx := &gorm.DB{}
	sr := mock_repository.NewMockSessionRepository(ctrl)
	ur := mock_repository.NewMockUserRepository(ctrl)
	ss := service.NewSessionService(sr, ur)
	userID := int64(1)
	user := &models.User{
		ID:           userID,
		Name:         "test-name",
		Surname:      "test-surname",
		Email:        "test-email",
		Username:     "test-username",
		PasswordHash: "test-password-hash",
		Role:         types.UserRoleAdmin,
	}

	expSession := &models.Session{
		ID:        "test-session-id",
		UserID:    userID,
		ExpiresAt: time.Now().Add(-time.Hour),
	}

	t.Run("User not found", func(t *testing.T) {
		ur.EXPECT().Get(tx, userID).Return(nil, gorm.ErrRecordNotFound).Times(1)
		session, err := ss.Create(tx, userID)
		require.ErrorIs(t, err, errors.ErrNotFound)
		assert.Nil(t, session)
	})

	t.Run("Unexpected error during get user", func(t *testing.T) {
		ur.EXPECT().Get(tx, userID).Return(nil, gorm.ErrInvalidDB).Times(1)
		session, err := ss.Create(tx, userID)
		require.Error(t, err)
		assert.Nil(t, session)
	})

	t.Run("Unexpected error during get session", func(t *testing.T) {
		ur.EXPECT().Get(tx, userID).Return(user, nil).Times(1)
		sr.EXPECT().GetByUserID(tx, userID).Return(nil, gorm.ErrInvalidDB).Times(1)
		session, err := ss.Create(tx, userID)
		require.Error(t, err)
		assert.Nil(t, session)
	})

	t.Run("Session expired", func(t *testing.T) {
		ur.EXPECT().Get(tx, userID).Return(user, nil).Times(1)
		sr.EXPECT().GetByUserID(tx, userID).Return(expSession, nil).Times(1)
		sr.EXPECT().Delete(tx, expSession.ID).Return(nil).Times(1)
		sr.EXPECT().Create(tx, gomock.Any()).Return(nil).Times(1)
		session, err := ss.Create(tx, userID)
		require.NoError(t, err)
		assert.NotNil(t, session)
	})

	t.Run("Session exists", func(t *testing.T) {
		existingSession := &models.Session{
			ID:        "test-session-id",
			UserID:    userID,
			ExpiresAt: time.Now().Add(time.Hour),
		}
		ur.EXPECT().Get(tx, userID).Return(user, nil).Times(1)
		sr.EXPECT().GetByUserID(tx, userID).Return(existingSession, nil).Times(1)
		session, err := ss.Create(tx, userID)
		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.IsType(t, &schemas.Session{}, session)
		assert.Equal(t, existingSession.ID, session.ID)
	})

	t.Run("Unexpected error during deleteion of expired sessoin", func(t *testing.T) {
		ur.EXPECT().Get(tx, userID).Return(user, nil).Times(1)
		sr.EXPECT().GetByUserID(tx, userID).Return(expSession, nil).Times(1)
		sr.EXPECT().Delete(tx, expSession.ID).Return(gorm.ErrInvalidDB).Times(1)
		session, err := ss.Create(tx, userID)
		require.Error(t, err)
		assert.Nil(t, session)
	})

	t.Run("Unexpected error during session create", func(t *testing.T) {
		ur.EXPECT().Get(tx, userID).Return(user, nil).Times(1)
		sr.EXPECT().GetByUserID(tx, userID).Return(nil, gorm.ErrRecordNotFound).Times(1)
		sr.EXPECT().Create(tx, gomock.Any()).Return(gorm.ErrInvalidDB).Times(1)
		session, err := ss.Create(tx, userID)
		require.Error(t, err)
		assert.Nil(t, session)
	})
}
