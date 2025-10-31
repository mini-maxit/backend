package repository

import (
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
)

type LanguageRepository interface {
	// Create creates a new language
	Create(tx database.Database, language *models.LanguageConfig) error
	// Delete deletes a language
	Delete(tx database.Database, languageID int64) error
	// GetAll returns all languages
	GetAll(tx database.Database) ([]models.LanguageConfig, error)
	// GetEnabled returns all enabled languages
	GetEnabled(tx database.Database) ([]models.LanguageConfig, error)
	// MarkDisabled marks a language as disabled
	MarkDisabled(tx database.Database, languageID int64) error
	// MarkEnabled marks a language as enabled
	MarkEnabled(tx database.Database, languageID int64) error
}

type languageRepository struct {
}

func (l *languageRepository) GetAll(tx database.Database) ([]models.LanguageConfig, error) {
	tasks := []models.LanguageConfig{}
	err := tx.Model(&models.LanguageConfig{}).Find(&tasks).Error()
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (l *languageRepository) Create(tx database.Database, language *models.LanguageConfig) error {
	err := tx.Model(models.LanguageConfig{}).Create(&language).Error()
	return err
}

func (l *languageRepository) Delete(tx database.Database, languageID int64) error {
	err := tx.Model(&models.LanguageConfig{}).Where("id = ?", languageID).Delete(&models.LanguageConfig{}).Error()
	return err
}

func (l *languageRepository) MarkDisabled(tx database.Database, languageID int64) error {
	err := tx.Model(&models.LanguageConfig{}).Where("id = ?", languageID).Update("is_disabled", true).Error()
	return err
}

func (l *languageRepository) MarkEnabled(tx database.Database, languageID int64) error {
	err := tx.Model(&models.LanguageConfig{}).Where("id = ?", languageID).Update("is_disabled", false).Error()
	return err
}

func (l *languageRepository) GetEnabled(tx database.Database) ([]models.LanguageConfig, error) {
	tasks := []models.LanguageConfig{}
	err := tx.Model(&models.LanguageConfig{}).Where("is_disabled = ?", false).Find(&tasks).Error()
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func NewLanguageRepository() LanguageRepository {
	return &languageRepository{}
}
