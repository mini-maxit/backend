package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type LanguageRepository interface {
	GetLanguages(tx *gorm.DB) ([]models.LanguageConfig, error)
	GetLanguage(tx *gorm.DB, languageId int64) (*models.LanguageConfig, error)
}

type LanguageRepositoryImpl struct {
}

func (l *LanguageRepositoryImpl) GetLanguages(tx *gorm.DB) ([]models.LanguageConfig, error) {
	panic("implement me")
}

func (l *LanguageRepositoryImpl) GetLanguage(tx *gorm.DB, languageId int64) (*models.LanguageConfig, error) {
	panic("implement me")
}

func NewLanguageRepository(db *gorm.DB) (LanguageRepository, error) {
	if !db.Migrator().HasTable(&models.LanguageConfig{}) {
		err := db.Migrator().CreateTable(&models.LanguageConfig{})
		if err != nil {
			return nil, err
		}
	}
	return &LanguageRepositoryImpl{}, nil
}
