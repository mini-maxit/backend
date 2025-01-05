package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type LanguageRepository interface {
	GetLanguages(tx *gorm.DB) ([]models.LanguageConfig, error)
	GetLanguage(tx *gorm.DB, languageId int64) (*models.LanguageConfig, error)
	CreateLanguage(tx *gorm.DB, language models.LanguageConfig) error
	DeleteLanguage(tx *gorm.DB, languageId int64) error
}

type LanguageRepositoryImpl struct {
}

func (l *LanguageRepositoryImpl) GetLanguages(tx *gorm.DB) ([]models.LanguageConfig, error) {
	tasks := []models.LanguageConfig{}
	err := tx.Model(&models.LanguageConfig{}).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (l *LanguageRepositoryImpl) GetLanguage(tx *gorm.DB, languageId int64) (*models.LanguageConfig, error) {
	panic("implement me")
}

func (l *LanguageRepositoryImpl) CreateLanguage(tx *gorm.DB, language models.LanguageConfig) error {
	err := tx.Model(&models.LanguageConfig{}).Create(&language).Error
	return err
}

func (l *LanguageRepositoryImpl) DeleteLanguage(tx *gorm.DB, languageId int64) error {
	err := tx.Model(&models.LanguageConfig{}).Where("id = ?", languageId).Delete(&models.LanguageConfig{}).Error
	return err
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
