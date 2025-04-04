package repository

import (
	"time"

	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type SessionRepository interface {
	// Create creates a new session in the database.
	Create(tx *gorm.DB, session *models.Session) error
	// Delete deletes a session from the database.
	Delete(tx *gorm.DB, sessionID string) error
	// Get retrieves a session from the database by its ID.
	Get(tx *gorm.DB, sessionID string) (*models.Session, error)
	// GetByUserID retrieves a session from the database by the user ID.
	GetByUserID(tx *gorm.DB, userID int64) (*models.Session, error)
}

type sessionRepository struct {
}

func (s *sessionRepository) Create(tx *gorm.DB, session *models.Session) error {
	err := tx.Model(&models.Session{}).Create(session).Error
	return err
}

func (s *sessionRepository) Get(tx *gorm.DB, sessionID string) (*models.Session, error) {
	session := &models.Session{}
	err := tx.Model(&models.Session{}).Where("id = ?", sessionID).Take(session).Error
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *sessionRepository) GetByUserID(tx *gorm.DB, userID int64) (*models.Session, error) {
	session := &models.Session{}
	err := tx.Model(&models.Session{}).Preload("User").Where("user_id = ?", userID).Take(session).Error
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *sessionRepository) UpdateExpiration(tx *gorm.DB, sessionID string, expiresAt time.Time) error {
	err := tx.Model(&models.Session{}).Where("id = ?", sessionID).Update("expires_at", expiresAt).Error
	return err
}

func (s *sessionRepository) Delete(tx *gorm.DB, sessionID string) error {
	err := tx.Model(&models.Session{}).Where("id = ?", sessionID).Delete(&models.Session{}).Error
	return err
}

func NewSessionRepository(db *gorm.DB) (SessionRepository, error) {
	if !db.Migrator().HasTable(&models.Session{}) {
		err := db.Migrator().CreateTable(&models.Session{})
		if err != nil {
			return nil, err
		}
	}
	return &sessionRepository{}, nil
}
