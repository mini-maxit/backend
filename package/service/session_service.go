package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	myerrors "github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const tokenDurationHours = 24

type SessionService interface {
	// Create generates a new session for a given user.
	Create(tx *gorm.DB, userID int64) (*schemas.Session, error)
	// Invalidate deletes a session by its ID.
	Invalidate(tx *gorm.DB, sessionID string) error
	// Validate checks if a session is valid based on its ID.
	Validate(tx *gorm.DB, sessionID string) (schemas.ValidateSessionResponse, error)
}

type sessionService struct {
	sessionRepository repository.SessionRepository
	userRepository    repository.UserRepository
	logger            *zap.SugaredLogger
}

// Generates a new session token.
func (s *sessionService) generateSessionToken() string {
	return uuid.New().String()
}

// Converts a session model to a session schema.
func (s *sessionService) modelToSchema(session *models.Session) *schemas.Session {
	return &schemas.Session{
		ID:        session.ID,
		UserID:    session.UserID,
		ExpiresAt: session.ExpiresAt,
		UserRole:  string(session.User.Role),
	}
}

// Creates a new session for a given user.
func (s *sessionService) Create(tx *gorm.DB, userID int64) (*schemas.Session, error) {
	user, err := s.userRepository.Get(tx, userID)
	if err != nil {
		s.logger.Errorf("Error getting user by id: %v", err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.ErrNotFound
		}
		return nil, err
	}

	session, err := s.sessionRepository.GetByUserID(tx, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Errorf("Error getting session by user id: %v", err.Error())
		return nil, err
	} else if err == nil {
		// If session exists but is expired remove record and create new session
		if session.ExpiresAt.Before(time.Now()) {
			err = s.sessionRepository.Delete(tx, session.ID)
			if err != nil {
				s.logger.Errorf("Error deleting session: %v", err.Error())
				return nil, err
			}
		}
		return s.modelToSchema(session), nil
	}

	sessionToken := s.generateSessionToken()
	session = &models.Session{
		ID:        sessionToken,
		UserID:    userID,
		ExpiresAt: time.Now().Add(time.Hour * tokenDurationHours),
	}

	err = s.sessionRepository.Create(tx, session)
	if err != nil {
		s.logger.Errorf("Error creating session: %v", err.Error())
		return nil, err
	}

	resp := s.modelToSchema(session)
	resp.UserRole = string(user.Role)
	return resp, nil
}

func (s *sessionService) Validate(tx *gorm.DB, sessionID string) (schemas.ValidateSessionResponse, error) {
	session, err := s.sessionRepository.Get(tx, sessionID)
	if err != nil {
		s.logger.Errorf("Error getting session by id: %v", err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return schemas.ValidateSessionResponse{Valid: false, User: InvalidUser}, myerrors.ErrSessionNotFound
		}
		return schemas.ValidateSessionResponse{Valid: false, User: InvalidUser}, err
	}
	currentUserModel, err := s.userRepository.Get(tx, session.UserID)
	if err != nil {
		s.logger.Errorf("Error getting user by id: %v", err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return schemas.ValidateSessionResponse{Valid: false, User: InvalidUser}, myerrors.ErrSessionUserNotFound
		}
		return schemas.ValidateSessionResponse{Valid: false, User: InvalidUser}, err
	}

	if session.ExpiresAt.Before(time.Now()) {
		s.logger.Error("Session expired")
		return schemas.ValidateSessionResponse{Valid: false, User: InvalidUser}, myerrors.ErrSessionExpired
	}

	currentUser := schemas.User{
		ID:       currentUserModel.ID,
		Email:    currentUserModel.Email,
		Username: currentUserModel.Username,
		Role:     currentUserModel.Role,
		Name:     currentUserModel.Name,
		Surname:  currentUserModel.Surname,
	}

	return schemas.ValidateSessionResponse{Valid: true, User: currentUser}, nil
}

func (s *sessionService) Invalidate(tx *gorm.DB, sessionID string) error {
	err := s.sessionRepository.Delete(tx, sessionID)
	if err != nil {
		s.logger.Errorf("Error deleting session: %v", err.Error())
		return err
	}
	return nil
}

func NewSessionService(
	sessionRepository repository.SessionRepository,
	userRepository repository.UserRepository,
) SessionService {
	log := utils.NewNamedLogger("session_service")
	return &sessionService{
		sessionRepository: sessionRepository,
		userRepository:    userRepository,
		logger:            log,
	}
}
