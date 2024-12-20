package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
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
	ValidateSession(sessionId string) (schemas.ValidateSessionResponse, error)
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

	defer utils.TransactionPanicRecover(tx)

	_, err := s.userRepository.GetUser(tx, userId)
	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSessionUserNotFound
		}
		return nil, err
	}

	session, err := s.sessionRepository.GetSessionByUserId(tx, userId)
	if err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return nil, err
	} else if err == nil {
		// If session exists but is expired remove record and create new session
		if session.ExpiresAt.Before(time.Now()) {
			err = s.sessionRepository.DeleteSession(tx, session.Id)
			if err != nil {
				tx.Rollback()
				return nil, err
			}
		} else {
			err = tx.Commit().Error
			if err != nil {
				return nil, err
			}
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
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}
	return s.modelToSchema(session), nil
}

func (s *SessionServiceImpl) ValidateSession(sessionId string) (schemas.ValidateSessionResponse, error) {
	tx := s.database.Connect().Begin()
	if tx.Error != nil {
		return schemas.ValidateSessionResponse{Valid: false, UserId: -1}, tx.Error
	}

	defer utils.TransactionPanicRecover(tx)

	session, err := s.sessionRepository.GetSession(tx, sessionId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return schemas.ValidateSessionResponse{Valid: false, UserId: -1}, ErrSessionNotFound
		}
		return schemas.ValidateSessionResponse{Valid: false, UserId: -1}, err
	}
	_, err = s.userRepository.GetUser(tx, session.UserId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			tx.Rollback()
			return schemas.ValidateSessionResponse{Valid: false, UserId: -1}, ErrSessionUserNotFound
		}
		tx.Rollback()
		return schemas.ValidateSessionResponse{Valid: false, UserId: -1}, err
	}

	if session.ExpiresAt.Before(time.Now()) {
		tx.Rollback()
		return schemas.ValidateSessionResponse{Valid: false, UserId: -1}, ErrSessionExpired
	}

	if tx.Commit().Error != nil {
		return schemas.ValidateSessionResponse{Valid: false, UserId: -1}, err
	}
	return schemas.ValidateSessionResponse{Valid: true, UserId: session.UserId}, nil
}

func (s *SessionServiceImpl) RefreshSession(sessionId string) (*schemas.Session, error) {
	tx := s.database.Connect().Begin()
	err := s.sessionRepository.UpdateExpiration(tx, sessionId, time.Now().Add(time.Hour*24))
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	session, err := s.sessionRepository.GetSession(tx, sessionId)
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

func (s *SessionServiceImpl) InvalidateSession(sessionId string) error {
	tx := s.database.Connect().Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer utils.TransactionPanicRecover(tx)

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
