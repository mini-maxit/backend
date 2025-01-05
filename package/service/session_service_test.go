package service

import (
	"testing"

	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/stretchr/testify/assert"
)

func TestValidateSession(t *testing.T) {
	tx := testutils.NewTestTx(t)
	userRepo, err := repository.NewUserRepository(tx)
	if err != nil {
		t.Fatalf("failed to create a new user repository: %v", err)
	}
	sessionRepo, err := repository.NewSessionRepository(tx)
	if err != nil {
		t.Fatalf("failed to create a new session repository: %v", err)
	}
	sessionService := NewSessionService(sessionRepo, userRepo)
	t.Run("Session not found", func(t *testing.T) {
		validateSession, err := sessionService.ValidateSession(tx, "test-session-id")
		assert.ErrorIs(t, err, ErrSessionNotFound)
		assert.False(t, validateSession.Valid)
		assert.Equal(t, int64(-1), validateSession.User.Id)
	})
	t.Run("Session found", func(t *testing.T) {
		userId, err := userRepo.CreateUser(tx, &models.User{
			Name:         "test-name",
			Surname:      "test-surname",
			Email:        "test-email",
			Username:     "test-username",
			PasswordHash: "test-password-hash",
			Role:         "admin",
		})
		assert.NoError(t, err)
		session, err := sessionService.CreateSession(tx, userId)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		validateSession, err := sessionService.ValidateSession(tx, session.Id)
		assert.NoError(t, err)
		assert.True(t, validateSession.Valid)
		assert.Equal(t, userId, validateSession.User.Id)
	})
	tx.Rollback()
}

func TestInvalidateSession(t *testing.T) {
	tx := testutils.NewTestTx(t)
	userRepo, err := repository.NewUserRepository(tx)
	if err != nil {
		t.Fatalf("failed to create a new user repository: %v", err)
	}
	sessionRepo, err := repository.NewSessionRepository(tx)
	if err != nil {
		t.Fatalf("failed to create a new session repository: %v", err)
	}
	sessionService := NewSessionService(sessionRepo, userRepo)
	t.Run("Session not found", func(t *testing.T) {
		err := sessionService.InvalidateSession(tx, "test-session-id")
		assert.NoError(t, err)
	})
	t.Run("Session found", func(t *testing.T) {
		userId, err := userRepo.CreateUser(tx, &models.User{
			Name:         "test-name",
			Surname:      "test-surname",
			Email:        "test-email",
			Username:     "test-username",
			PasswordHash: "test-password-hash",
			Role:         "admin",
		})
		assert.NoError(t, err)
		session, err := sessionService.CreateSession(tx, userId)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		err = sessionService.InvalidateSession(tx, session.Id)
		assert.NoError(t, err)
		validateSession, err := sessionService.ValidateSession(tx, session.Id)
		assert.ErrorIs(t, err, ErrSessionNotFound)
		assert.False(t, validateSession.Valid)
		assert.Equal(t, int64(-1), validateSession.User.Id)
	})
	tx.Rollback()
}
