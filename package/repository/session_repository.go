package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type SessionRepository interface {
	CreateSession(tx *gorm.DB, session *models.Session) error
	GetSession(tx *gorm.DB, sessionId string) (*models.Session, error)
	GetSessionByUserId(tx *gorm.DB, userId int64) (*models.Session, error)
	DeleteSession(tx *gorm.DB, sessionId string) error
}

type SessionRepositoryImpl struct {
}

func (s *SessionRepositoryImpl) CreateSession(tx *gorm.DB, session *models.Session) error {
	err := tx.Model(&models.Session{}).Create(session).Error
	return err
}

func (s *SessionRepositoryImpl) GetSession(tx *gorm.DB, sessionId string) (*models.Session, error) {
	session := &models.Session{}
	err := tx.Model(&models.Session{}).Where("id = ?", sessionId).First(session).Error
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *SessionRepositoryImpl) GetSessionByUserId(tx *gorm.DB, userId int64) (*models.Session, error) {
	session := &models.Session{}
	err := tx.Model(&models.Session{}).Where("user_id = ?", userId).First(session).Error
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *SessionRepositoryImpl) DeleteSession(tx *gorm.DB, sessionId string) error {
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
	return &SessionRepositoryImpl{}, nil
}
