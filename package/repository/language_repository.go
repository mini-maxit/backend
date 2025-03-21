package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type LanguageRepository interface {
	GetLanguages(tx *gorm.DB) ([]models.LanguageConfig, error)
	GetLanguage(tx *gorm.DB, languageId int64) (*models.LanguageConfig, error)
	CreateLanguage(tx *gorm.DB, language *models.LanguageConfig) error
	DeleteLanguage(tx *gorm.DB, languageId int64) error
	MarkLanguageDisabled(tx *gorm.DB, languageId int64) error
	MarkLanguageEnabled(tx *gorm.DB, languageId int64) error
	GetEnabledLanguages(tx *gorm.DB) ([]models.LanguageConfig, error)
}

type languageRepository struct {
}

func (l *languageRepository) GetLanguages(tx *gorm.DB) ([]models.LanguageConfig, error) {
	tasks := []models.LanguageConfig{}
	err := tx.Model(&models.LanguageConfig{}).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (l *languageRepository) GetLanguage(tx *gorm.DB, languageId int64) (*models.LanguageConfig, error) {
	panic("implement me")
}

func (l *languageRepository) CreateLanguage(tx *gorm.DB, language *models.LanguageConfig) error {
	err := tx.Model(models.LanguageConfig{}).Create(&language).Error
	return err
}

func (l *languageRepository) DeleteLanguage(tx *gorm.DB, languageId int64) error {
	err := tx.Model(&models.LanguageConfig{}).Where("id = ?", languageId).Delete(&models.LanguageConfig{}).Error
	return err
}

func (l *languageRepository) MarkLanguageDisabled(tx *gorm.DB, languageId int64) error {
	err := tx.Model(&models.LanguageConfig{}).Where("id = ?", languageId).Update("disabled", true).Error
	return err
}

func (l *languageRepository) MarkLanguageEnabled(tx *gorm.DB, languageId int64) error {
	err := tx.Model(&models.LanguageConfig{}).Where("id = ?", languageId).Update("disabled", false).Error
	return err
}

func (l *languageRepository) GetEnabledLanguages(tx *gorm.DB) ([]models.LanguageConfig, error) {
	tasks := []models.LanguageConfig{}
	err := tx.Model(&models.LanguageConfig{}).Where("disabled = ?", false).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func NewLanguageRepository(db *gorm.DB) (LanguageRepository, error) {
	if !db.Migrator().HasTable(&models.LanguageConfig{}) {
		err := db.Migrator().CreateTable(&models.LanguageConfig{})
		if err != nil {
			return nil, err
		}
	}
	return &languageRepository{}, nil
}
