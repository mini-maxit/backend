package service

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// LanguageService defines the methods for language-related operations.
type LanguageService interface {
	// GetAll retrieves all language configurations from the database.
	GetAll(tx *gorm.DB) ([]schemas.LanguageConfig, error)
	// GetAllEnabled retrieves all enabled language configurations from the database.
	GetAllEnabled(tx *gorm.DB) ([]schemas.LanguageConfig, error)
	// Init initializes languages in the database
	//
	// It should be called during application initialization.
	// Method initializes languages in the database if they are not already present.
	// If language is already present in the database, and is not disabled it skips it. Otherwise, it enables it.
	// If language is not in enabled languages, but is present in the database, it is marked as disabled.
	Init(tx *gorm.DB, enabledLanguages schemas.HandShakeResponsePayload) error
}

// languageService implements [LanguageService] interface
type languageService struct {
	languageRepository repository.LanguageRepository
	logger             *zap.SugaredLogger
}

// Init implements Init method of [LanguageService] interface
func (l *languageService) Init(tx *gorm.DB, workerLanguages schemas.HandShakeResponsePayload) error {

	l.logger.Infof("Initializing languages: %v", workerLanguages.Languages)

	existingLanguages, err := l.languageRepository.GetAll(tx)
	if err != nil {
		l.logger.Errorf("Error getting existing languages: %v", err.Error())
		return err
	}
	for _, lang := range workerLanguages.Languages {
		for _, version := range lang.Versions {
			language := models.LanguageConfig{
				Type:    lang.Name,
				Version: version,
			}
			var found bool
			for i, existingLanguage := range existingLanguages {
				if existingLanguage.Type == language.Type && existingLanguage.Version == language.Version {
					found = true
					existingLanguages = append(existingLanguages[:i], existingLanguages[i+1:]...)
					err := l.languageRepository.MarkEnabled(tx, existingLanguage.Id)
					if err != nil {
						l.logger.Errorf("Error marking language enabled: %v", err.Error())
					}
					break
				}
			}
			if !found {
				err := l.languageRepository.Create(tx, &language)
				if err != nil {
					l.logger.Errorf("Error adding language: %v", err.Error())
					return err
				}
			}

		}
	}

	if len(existingLanguages) > 0 {
		for _, lang := range existingLanguages {
			err := l.languageRepository.MarkDisabled(tx, lang.Id)
			if err != nil {
				l.logger.Errorf("Error marking language disabled: %v", err.Error())
				return err
			}
		}
	}
	return nil
}

func (l *languageService) GetAll(tx *gorm.DB) ([]schemas.LanguageConfig, error) {
	languages, err := l.languageRepository.GetAll(tx)
	if err != nil {
		l.logger.Errorf("Error getting languages: %v", err.Error())
		return nil, err
	}
	var result []schemas.LanguageConfig
	for _, language := range languages {
		result = append(result, *LanguageToSchema(&language))
	}
	return result, nil
}

func (l *languageService) GetAllEnabled(tx *gorm.DB) ([]schemas.LanguageConfig, error) {
	languages, err := l.languageRepository.GetEnabled(tx)
	if err != nil {
		l.logger.Errorf("Error getting enabled languages: %v", err.Error())
		return nil, err
	}
	var result []schemas.LanguageConfig
	for _, language := range languages {
		result = append(result, *LanguageToSchema(&language))
	}
	return result, nil
}

func LanguageToSchema(language *models.LanguageConfig) *schemas.LanguageConfig {
	return &schemas.LanguageConfig{
		Id:            language.Id,
		Type:          language.Type,
		Version:       language.Version,
		FileExtension: language.FileExtension,
	}
}

func NewLanguageService(languageRepository repository.LanguageRepository) LanguageService {
	log := utils.NewNamedLogger("language_service")
	return &languageService{
		languageRepository: languageRepository,
		logger:             log,
	}
}
