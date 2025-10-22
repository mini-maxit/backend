package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type LanguageRepository interface {
	// Create creates a new language
	Create(tx *gorm.DB, language *models.LanguageConfig) error
	// Delete deletes a language
	Delete(tx *gorm.DB, languageID int64) error
	// GetAll returns all languages
	GetAll(tx *gorm.DB) ([]models.LanguageConfig, error)
	// GetEnabled returns all enabled languages
	GetEnabled(tx *gorm.DB) ([]models.LanguageConfig, error)
	// MarkDisabled marks a language as disabled
	MarkDisabled(tx *gorm.DB, languageID int64) error
	// MarkEnabled marks a language as enabled
	MarkEnabled(tx *gorm.DB, languageID int64) error
}

type languageRepository struct {
}

func (l *languageRepository) GetAll(tx *gorm.DB) ([]models.LanguageConfig, error) {
	tasks := []models.LanguageConfig{}
	err := tx.Model(&models.LanguageConfig{}).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (l *languageRepository) Create(tx *gorm.DB, language *models.LanguageConfig) error {
	err := tx.Model(models.LanguageConfig{}).Create(&language).Error
	return err
}

func (l *languageRepository) Delete(tx *gorm.DB, languageID int64) error {
	err := tx.Model(&models.LanguageConfig{}).Where("id = ?", languageID).Delete(&models.LanguageConfig{}).Error
	return err
}

func (l *languageRepository) MarkDisabled(tx *gorm.DB, languageID int64) error {
	err := tx.Model(&models.LanguageConfig{}).Where("id = ?", languageID).Update("is_disabled", true).Error
	return err
}

func (l *languageRepository) MarkEnabled(tx *gorm.DB, languageID int64) error {
	err := tx.Model(&models.LanguageConfig{}).Where("id = ?", languageID).Update("is_disabled", false).Error
	return err
}

func (l *languageRepository) GetEnabled(tx *gorm.DB) ([]models.LanguageConfig, error) {
	tasks := []models.LanguageConfig{}
	err := tx.Model(&models.LanguageConfig{}).Where("is_disabled = ?", false).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func NewLanguageRepository() LanguageRepository {
	return &languageRepository{}
}
