package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"gorm.io/gorm"
)

var (
	ErrSessionNotFound     = fmt.Errorf("session not found")
	ErrSessionExpired      = fmt.Errorf("session expired")
	ErrSessionUserNotFound = fmt.Errorf("session user not found")
)

type SessionService interface {
	CreateSession(tx *gorm.DB, userId int64) (*schemas.Session, error)
	ValidateSession(sessionId string) (bool, error)
	InvalidateSession(sessionId string) error
}

type SessionServiceImpl struct {
	database          database.Database
	sessionRepository repository.SessionRepository
	userRepository    repository.UserRepository
}

// Generates a new session token
func (s *SessionServiceImpl) generateSessionToken() string {
	return uuid.New().String()
}

// Converts a session model to a session schema
func (s *SessionServiceImpl) modelToSchema(session *models.Session) *schemas.Session {
	return &schemas.Session{
		Id:        session.Id,
		UserId:    session.UserId,
		ExpiresAt: session.ExpiresAt,
	}
}

func (s *SessionServiceImpl) CreateSession(tx *gorm.DB, userId int64) (*schemas.Session, error) {
	if tx == nil {
		tx = s.database.Connect().Begin()
		if tx.Error != nil {
			return nil, tx.Error
		}
	}
	_, err := s.userRepository.GetUser(tx, userId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			tx.Rollback()
			return nil, ErrSessionUserNotFound
		}
		tx.Rollback()
		return nil, err
	}

	session, err := s.sessionRepository.GetSessionByUserId(tx, userId)
	if err == nil {
		if tx.Commit().Error != nil {
			return nil, err
		}
		return s.modelToSchema(session), nil
	}
	if err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return nil, err
	}

	sessionToken := s.generateSessionToken()
	session = &models.Session{
		Id:        sessionToken,
		UserId:    userId,
		ExpiresAt: time.Now().Add(time.Hour * 24),
	}

	err = s.sessionRepository.CreateSession(tx, session)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}
	return s.modelToSchema(session), nil
}

func (s *SessionServiceImpl) ValidateSession(sessionId string) (bool, error) {
	tx := s.database.Connect().Begin()

	session, err := s.sessionRepository.GetSession(tx, sessionId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, ErrSessionNotFound
		}
		return false, err
	}
	_, err = s.userRepository.GetUser(tx, session.UserId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			tx.Rollback()
			return false, ErrSessionUserNotFound
		}
		tx.Rollback()
		return false, err
	}

	if session.ExpiresAt.Before(time.Now()) {
		tx.Rollback()
		return false, ErrSessionExpired
	}

	if tx.Commit().Error != nil {
		return false, err
	}
	return true, nil
}

func (s *SessionServiceImpl) InvalidateSession(sessionId string) error {
	tx := s.database.Connect().Begin()

	err := s.sessionRepository.DeleteSession(tx, sessionId)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		return err
	}
	return nil
}

func NewSessionService(db database.Database, sessionRepository repository.SessionRepository, userRepository repository.UserRepository) SessionService {
	return &SessionServiceImpl{
		database:          db,
		sessionRepository: sessionRepository,
		userRepository:    userRepository,
	}
}
