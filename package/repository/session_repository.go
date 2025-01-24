package repository

import (
	"time"

	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type SessionRepository interface {
	CreateSession(tx *gorm.DB, session *models.Session) error
	GetSession(tx *gorm.DB, sessionId string) (*models.Session, error)
	GetSessionByUserId(tx *gorm.DB, userId int64) (*models.Session, error)
	UpdateExpiration(tx *gorm.DB, sessionId string, expires_at time.Time) error
	DeleteSession(tx *gorm.DB, sessionId string) error
}

type sessionRepository struct {
}

func (s *sessionRepository) CreateSession(tx *gorm.DB, session *models.Session) error {
	err := tx.Model(&models.Session{}).Create(session).Error
	return err
}

func (s *sessionRepository) GetSession(tx *gorm.DB, sessionId string) (*models.Session, error) {
	session := &models.Session{}
	err := tx.Model(&models.Session{}).Where("id = ?", sessionId).First(session).Error
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *sessionRepository) GetSessionByUserId(tx *gorm.DB, userId int64) (*models.Session, error) {
	session := &models.Session{}
	err := tx.Model(&models.Session{}).Where("user_id = ?", userId).First(session).Error
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *sessionRepository) UpdateExpiration(tx *gorm.DB, sessionId string, expires_at time.Time) error {
	err := tx.Model(&models.Session{}).Where("id = ?", sessionId).Update("expires_at", expires_at).Error
	return err
}

func (s *sessionRepository) DeleteSession(tx *gorm.DB, sessionId string) error {
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
