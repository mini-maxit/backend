package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mini-maxit/backend/internal/logger"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"gorm.io/gorm"
)

var (
	ErrSessionNotFound     = fmt.Errorf("session not found")
	ErrSessionExpired      = fmt.Errorf("session expired")
	ErrSessionUserNotFound = fmt.Errorf("session user not found")
	ErrSessionRefresh      = fmt.Errorf("session refresh failed")
)

type SessionService interface {
	CreateSession(tx *gorm.DB, userId int64) (*schemas.Session, error)
	ValidateSession(tx *gorm.DB, sessionId string) (schemas.ValidateSessionResponse, error)
	InvalidateSession(tx *gorm.DB, sessionId string) error
}

type SessionServiceImpl struct {
	sessionRepository repository.SessionRepository
	userRepository    repository.UserRepository
	session_logger    *logger.ServiceLogger
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

// Creates a new session for a given user
func (s *SessionServiceImpl) CreateSession(tx *gorm.DB, userId int64) (*schemas.Session, error) {
	_, err := s.userRepository.GetUser(tx, userId)
	if err != nil {
		logger.Log(s.session_logger, "Error getting user by id:", err.Error(), logger.Error)
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSessionUserNotFound
		}
		return nil, err
	}

	session, err := s.sessionRepository.GetSessionByUserId(tx, userId)
	if err != nil && err != gorm.ErrRecordNotFound {
		logger.Log(s.session_logger, "Error getting session by user id:", err.Error(), logger.Error)
		return nil, err
	} else if err == nil {
		// If session exists but is expired remove record and create new session
		if session.ExpiresAt.Before(time.Now()) {
			err = s.sessionRepository.DeleteSession(tx, session.Id)
			if err != nil {
				logger.Log(s.session_logger, "Error deleting session:", err.Error(), logger.Error)
				return nil, err
			}
		} else {
			return s.modelToSchema(session), nil
		}
	}

	sessionToken := s.generateSessionToken()
	session = &models.Session{
		Id:        sessionToken,
		UserId:    userId,
		ExpiresAt: time.Now().Add(time.Hour * 24),
	}

	err = s.sessionRepository.CreateSession(tx, session)
	if err != nil {
		logger.Log(s.session_logger, "Error creating session:", err.Error(), logger.Error)
		return nil, err
	}

	return s.modelToSchema(session), nil
}

func (s *SessionServiceImpl) ValidateSession(tx *gorm.DB, sessionId string) (schemas.ValidateSessionResponse, error) {
	session, err := s.sessionRepository.GetSession(tx, sessionId)
	if err != nil {
		logger.Log(s.session_logger, "Error getting session by id:", err.Error(), logger.Error)
		if err == gorm.ErrRecordNotFound {
			return schemas.ValidateSessionResponse{Valid: false, UserId: -1}, ErrSessionNotFound
		}
		return schemas.ValidateSessionResponse{Valid: false, UserId: -1}, err
	}
	_, err = s.userRepository.GetUser(tx, session.UserId)
	if err != nil {
		logger.Log(s.session_logger, "Error getting user by id:", err.Error(), logger.Error)
		if err == gorm.ErrRecordNotFound {
			return schemas.ValidateSessionResponse{Valid: false, UserId: -1}, ErrSessionUserNotFound
		}
		return schemas.ValidateSessionResponse{Valid: false, UserId: -1}, err
	}

	if session.ExpiresAt.Before(time.Now()) {
		logger.Log(s.session_logger, "Session expired", "", logger.Info)
		return schemas.ValidateSessionResponse{Valid: false, UserId: -1}, ErrSessionExpired
	}

	if err := tx.Commit().Error; err != nil {
		logger.Log(s.session_logger, "Error committing transaction:", err.Error(), logger.Error)
		return schemas.ValidateSessionResponse{Valid: false, UserId: -1}, err
	}
	return schemas.ValidateSessionResponse{Valid: true, UserId: session.UserId}, nil
}

func (s *SessionServiceImpl) RefreshSession(tx *gorm.DB, sessionId string) (*schemas.Session, error) {
	err := s.sessionRepository.UpdateExpiration(tx, sessionId, time.Now().Add(time.Hour*24))
	if err != nil {
		logger.Log(s.session_logger, "Error updating session expiration:", err.Error(), logger.Error)
		return nil, err
	}
	session, err := s.sessionRepository.GetSession(tx, sessionId)
	if err != nil {
		logger.Log(s.session_logger, "Error getting session by id:", err.Error(), logger.Error)
		return nil, err
	}
	return s.modelToSchema(session), nil
}

func (s *SessionServiceImpl) InvalidateSession(tx *gorm.DB, sessionId string) error {
	err := s.sessionRepository.DeleteSession(tx, sessionId)
	if err != nil {
		logger.Log(s.session_logger, "Error deleting session:", err.Error(), logger.Error)
		return err
	}
	return nil
}

func NewSessionService(sessionRepository repository.SessionRepository, userRepository repository.UserRepository) SessionService {
	session_logger := logger.NewNamedLogger("session_service")
	return &SessionServiceImpl{
		sessionRepository: sessionRepository,
		userRepository:    userRepository,
		session_logger:    &session_logger,
	}
}
