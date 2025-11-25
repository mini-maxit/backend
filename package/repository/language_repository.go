package repository

import (
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
)

type LanguageRepository interface {
	// Create creates a new language
	Create(db database.Database, language *models.LanguageConfig) error
	// Delete deletes a language
	Delete(db database.Database, languageID int64) error
	// GetAll returns all languages
	GetAll(db database.Database) ([]models.LanguageConfig, error)
	// GetEnabled returns all enabled languages
	GetEnabled(db database.Database) ([]models.LanguageConfig, error)
	// MarkDisabled marks a language as disabled
	MarkDisabled(db database.Database, languageID int64) error
	// MarkEnabled marks a language as enabled
	MarkEnabled(db database.Database, languageID int64) error
}

type languageRepository struct {
}

func (l *languageRepository) GetAll(db database.Database) ([]models.LanguageConfig, error) {
	tx := db.GetInstance()
	tasks := []models.LanguageConfig{}
	err := tx.Model(&models.LanguageConfig{}).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (l *languageRepository) Create(db database.Database, language *models.LanguageConfig) error {
	tx := db.GetInstance()
	err := tx.Model(models.LanguageConfig{}).Create(&language).Error
	return err
}

func (l *languageRepository) Delete(db database.Database, languageID int64) error {
	tx := db.GetInstance()
	err := tx.Model(&models.LanguageConfig{}).Where("id = ?", languageID).Delete(&models.LanguageConfig{}).Error
	return err
}

func (l *languageRepository) MarkDisabled(db database.Database, languageID int64) error {
	tx := db.GetInstance()
	err := tx.Model(&models.LanguageConfig{}).Where("id = ?", languageID).Update("is_disabled", true).Error
	return err
}

func (l *languageRepository) MarkEnabled(db database.Database, languageID int64) error {
	tx := db.GetInstance()
	err := tx.Model(&models.LanguageConfig{}).Where("id = ?", languageID).Update("is_disabled", false).Error
	return err
}

func (l *languageRepository) GetEnabled(db database.Database) ([]models.LanguageConfig, error) {
	tx := db.GetInstance()
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
