package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
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
	logger            *zap.SugaredLogger
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
		UserRole:  "invalid",
	}
}

// Creates a new session for a given user
func (s *SessionServiceImpl) CreateSession(tx *gorm.DB, userId int64) (*schemas.Session, error) {
	user, err := s.userRepository.GetUser(tx, userId)
	if err != nil {
		s.logger.Errorf("Error getting user by id: %v", err.Error())
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSessionUserNotFound
		}
		return nil, err
	}

	session, err := s.sessionRepository.GetSessionByUserId(tx, userId)
	if err != nil && err != gorm.ErrRecordNotFound {
		s.logger.Errorf("Error getting session by user id: %v", err.Error())
		return nil, err
	} else if err == nil {
		// If session exists but is expired remove record and create new session
		if session.ExpiresAt.Before(time.Now()) {
			err = s.sessionRepository.DeleteSession(tx, session.Id)
			if err != nil {
				s.logger.Errorf("Error deleting session: %v", err.Error())
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
		s.logger.Errorf("Error creating session: %v", err.Error())
		return nil, err
	}

	resp := s.modelToSchema(session)
	resp.UserRole = string(user.Role)
	return resp, nil
}

func (s *SessionServiceImpl) ValidateSession(tx *gorm.DB, sessionId string) (schemas.ValidateSessionResponse, error) {
	session, err := s.sessionRepository.GetSession(tx, sessionId)
	if err != nil {
		s.logger.Errorf("Error getting session by id: %v", err.Error())
		if err == gorm.ErrRecordNotFound {
			return schemas.ValidateSessionResponse{Valid: false, User: InvalidUser}, ErrSessionNotFound
		}
		return schemas.ValidateSessionResponse{Valid: false, User: InvalidUser}, err
	}
	current_user_model, err := s.userRepository.GetUser(tx, session.UserId)
	if err != nil {
		s.logger.Errorf("Error getting user by id: %v", err.Error())
		if err == gorm.ErrRecordNotFound {
			return schemas.ValidateSessionResponse{Valid: false, User: InvalidUser}, ErrSessionUserNotFound
		}
		return schemas.ValidateSessionResponse{Valid: false, User: InvalidUser}, err
	}

	if session.ExpiresAt.Before(time.Now()) {
		s.logger.Error("Session expired")
		return schemas.ValidateSessionResponse{Valid: false, User: InvalidUser}, ErrSessionExpired
	}

	current_user := schemas.User{
		Id: current_user_model.Id,
		Email: current_user_model.Email,
		Username: current_user_model.Username,
		Role: string(current_user_model.Role),
		Name: current_user_model.Name,
		Surname: current_user_model.Surname,
	}

	return schemas.ValidateSessionResponse{Valid: true, User: current_user}, nil
}

func (s *SessionServiceImpl) RefreshSession(tx *gorm.DB, sessionId string) (*schemas.Session, error) {
	err := s.sessionRepository.UpdateExpiration(tx, sessionId, time.Now().Add(time.Hour*24))
	if err != nil {
		s.logger.Errorf("Error updating session expiration: %v", err.Error())
		return nil, err
	}
	session, err := s.sessionRepository.GetSession(tx, sessionId)
	if err != nil {
		s.logger.Errorf("Error getting session by id: %v", err.Error())
		return nil, err
	}
	return s.modelToSchema(session), nil
}

func (s *SessionServiceImpl) InvalidateSession(tx *gorm.DB, sessionId string) error {
	err := s.sessionRepository.DeleteSession(tx, sessionId)
	if err != nil {
		s.logger.Errorf("Error deleting session: %v", err.Error())
		return err
	}
	return nil
}

func NewSessionService(sessionRepository repository.SessionRepository, userRepository repository.UserRepository) SessionService {
	log := utils.NewNamedLogger("session_service")
	return &SessionServiceImpl{
		sessionRepository: sessionRepository,
		userRepository:    userRepository,
		logger:            log,
	}
}
