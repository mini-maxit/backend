package service

import (
	"testing"

	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/stretchr/testify/assert"
)

func TestCreateSession(t *testing.T) {
	cfg := testutils.NewTestConfig()
	database, err := testutils.NewTestPostgresDB(cfg)

	if err != nil {
		t.Fatalf("failed to create a new database: %v", err)
	}
	tx, err := database.Connect()
	if err != nil {
		t.Fatalf("failed to create a new database connection: %v", err)
	}
	userRepo, err := repository.NewUserRepository(tx)
	if err != nil {
		t.Fatalf("failed to create a new user repository: %v", err)
	}
	sessionRepo, err := repository.NewSessionRepository(tx)
	if err != nil {
		t.Fatalf("failed to create a new session repository: %v", err)
	}
	sessionService := NewSessionService(sessionRepo, userRepo)
	t.Run("User not found", func(t *testing.T) {
		session, err := sessionService.CreateSession(tx, 1)
		assert.ErrorIs(t, err, ErrSessionUserNotFound)
		assert.Nil(t, session)
	})
	userId, err := userRepo.CreateUser(tx, &models.User{
		Name:         "test-name",
		Surname:      "test-surname",
		Email:        "test-email",
		Username:     "test-username",
		PasswordHash: "test-password-hash",
		Role:         "test-role",
	})
	assert.NoError(t, err)
	t.Run("Create session successfuly", func(t *testing.T) {
		session, err := sessionService.CreateSession(tx, userId)
		assert.NoError(t, err)
		assert.NotNil(t, session)
	})
	t.Run("Session already exists", func(t *testing.T) {
		firstSession, err := sessionService.CreateSession(tx, userId)
		assert.NoError(t, err)
		assert.NotNil(t, firstSession)
		session, err := sessionService.CreateSession(tx, userId)
		assert.NoError(t, err)
		// Due to the fact how go stores time, we can't compare the expiresat of the sessions
		assert.Equal(t, firstSession.Id, session.Id)
		assert.Equal(t, firstSession.UserId, session.UserId)
		assert.InEpsilon(t, firstSession.ExpiresAt.Unix(), session.ExpiresAt.Unix(), 1)

	})
	tx.Rollback()
}

func TestValidateSession(t *testing.T) {
	cfg := testutils.NewTestConfig()
	database, err := testutils.NewTestPostgresDB(cfg)

	if err != nil {
		t.Fatalf("failed to create a new database: %v", err)
	}

	tx, err := database.Connect()
	if err != nil {
		t.Fatalf("failed to create a new database connection: %v", err)
	}
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
			Role:         "test-role",
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
	cfg := testutils.NewTestConfig()
	database, err := testutils.NewTestPostgresDB(cfg)

	if err != nil {
		t.Fatalf("failed to create a new database: %v", err)
	}
	tx, err := database.Connect()
	if err != nil {
		t.Fatalf("failed to create a new database connection: %v", err)
	}

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
			Role:         "test-role",
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
