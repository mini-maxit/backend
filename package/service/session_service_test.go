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

	userRepo, err := repository.NewUserRepository(database.Connect())
	if err != nil {
		t.Fatalf("failed to create a new user repository: %v", err)
	}
	sessionRepo, err := repository.NewSessionRepository(database.Connect())
	if err != nil {
		t.Fatalf("failed to create a new session repository: %v", err)
	}
	sessionService := NewSessionService(database, sessionRepo, userRepo)
	t.Run("User not found", func(t *testing.T) {
		session, err := sessionService.CreateSession(nil, 1)
		assert.ErrorIs(t, err, ErrSessionUserNotFound)
		assert.Nil(t, session)
	})
	userId, err := userRepo.CreateUser(database.Connect(), &models.User{
		Name:         "test-name",
		Surname:      "test-surname",
		Email:        "test-email",
		Username:     "test-username",
		PasswordHash: "test-password-hash",
		Role:         "test-role",
	})
	assert.NoError(t, err)
	t.Run("Create session successfuly", func(t *testing.T) {
		session, err := sessionService.CreateSession(nil, userId)
		assert.NoError(t, err)
		assert.NotNil(t, session)
	})
	t.Run("Session already exists", func(t *testing.T) {
		firstSession, err := sessionService.CreateSession(nil, userId)
		assert.NoError(t, err)
		assert.NotNil(t, firstSession)
		session, err := sessionService.CreateSession(nil, userId)
		assert.NoError(t, err)
		// Due to the fact how go stores time, we can't compare the expiresat of the sessions
		assert.Equal(t, firstSession.Id, session.Id)
		assert.Equal(t, firstSession.UserId, session.UserId)
		assert.InEpsilon(t, firstSession.ExpiresAt.Unix(), session.ExpiresAt.Unix(), 1)

	})
}

func TestValidateSession(t *testing.T) {
	cfg := testutils.NewTestConfig()
	database, err := testutils.NewTestPostgresDB(cfg)

	if err != nil {
		t.Fatalf("failed to create a new database: %v", err)
	}

	userRepo, err := repository.NewUserRepository(database.Connect())
	if err != nil {
		t.Fatalf("failed to create a new user repository: %v", err)
	}
	sessionRepo, err := repository.NewSessionRepository(database.Connect())
	if err != nil {
		t.Fatalf("failed to create a new session repository: %v", err)
	}
	sessionService := NewSessionService(database, sessionRepo, userRepo)
	t.Run("Session not found", func(t *testing.T) {
		valid, err := sessionService.ValidateSession("test-session-id")
		assert.ErrorIs(t, err, ErrSessionNotFound)
		assert.False(t, valid)
	})
	t.Run("Session found", func(t *testing.T) {
		userId, err := userRepo.CreateUser(database.Connect(), &models.User{
			Name:         "test-name",
			Surname:      "test-surname",
			Email:        "test-email",
			Username:     "test-username",
			PasswordHash: "test-password-hash",
			Role:         "test-role",
		})
		assert.NoError(t, err)
		session, err := sessionService.CreateSession(nil, userId)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		valid, err := sessionService.ValidateSession(session.Id)
		assert.NoError(t, err)
		assert.True(t, valid)
	})
}

func TestInvalidateSession(t *testing.T) {
	cfg := testutils.NewTestConfig()
	database, err := testutils.NewTestPostgresDB(cfg)

	if err != nil {
		t.Fatalf("failed to create a new database: %v", err)
	}

	userRepo, err := repository.NewUserRepository(database.Connect())
	if err != nil {
		t.Fatalf("failed to create a new user repository: %v", err)
	}
	sessionRepo, err := repository.NewSessionRepository(database.Connect())
	if err != nil {
		t.Fatalf("failed to create a new session repository: %v", err)
	}
	sessionService := NewSessionService(database, sessionRepo, userRepo)
	t.Run("Session not found", func(t *testing.T) {
		err := sessionService.InvalidateSession("test-session-id")
		assert.NoError(t, err)
	})
	t.Run("Session found", func(t *testing.T) {
		userId, err := userRepo.CreateUser(database.Connect(), &models.User{
			Name:         "test-name",
			Surname:      "test-surname",
			Email:        "test-email",
			Username:     "test-username",
			PasswordHash: "test-password-hash",
			Role:         "test-role",
		})
		assert.NoError(t, err)
		session, err := sessionService.CreateSession(nil, userId)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		err = sessionService.InvalidateSession(session.Id)
		assert.NoError(t, err)
		valid, err := sessionService.ValidateSession(session.Id)
		assert.ErrorIs(t, err, ErrSessionNotFound)
		assert.False(t, valid)
	})
}
