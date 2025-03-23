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
	Delete(tx *gorm.DB, sessionId string) error
	// Get retrieves a session from the database by its Id.
	Get(tx *gorm.DB, sessionId string) (*models.Session, error)
	// GetByUserId retrieves a session from the database by the user Id.
	GetByUserId(tx *gorm.DB, userId int64) (*models.Session, error)
}

type sessionRepository struct {
}

func (s *sessionRepository) Create(tx *gorm.DB, session *models.Session) error {
	err := tx.Model(&models.Session{}).Create(session).Error
	return err
}

func (s *sessionRepository) Get(tx *gorm.DB, sessionId string) (*models.Session, error) {
	session := &models.Session{}
	err := tx.Model(&models.Session{}).Where("id = ?", sessionId).Take(session).Error
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *sessionRepository) GetByUserId(tx *gorm.DB, userId int64) (*models.Session, error) {
	session := &models.Session{}
	err := tx.Model(&models.Session{}).Preload("User").Where("user_id = ?", userId).Take(session).Error
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *sessionRepository) UpdateExpiration(tx *gorm.DB, sessionId string, expires_at time.Time) error {
	err := tx.Model(&models.Session{}).Where("id = ?", sessionId).Update("expires_at", expires_at).Error
	return err
}

func (s *sessionRepository) Delete(tx *gorm.DB, sessionId string) error {
	err := tx.Model(&models.Session{}).Where("id = ?", sessionId).Delete(&models.Session{}).Error
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
