package service

import (
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
)

// LanguageService defines the methods for language-related operations.
type LanguageService interface {
	// GetAll retrieves all language configurations from the database.
	GetAll(db database.Database) ([]schemas.LanguageConfig, error)
	// GetAllEnabled retrieves all enabled language configurations from the database.
	GetAllEnabled(db database.Database) ([]schemas.LanguageConfig, error)
	// ToggleLanguageVisibility toggles the visibility (enabled/disabled) state of a language.
	ToggleLanguageVisibility(db database.Database, languageID int64) error
	// Init initializes languages in the database
	//
	// It should be called during application initialization.
	// Method initializes languages in the database if they are not already present.
	// If language is already present in the database skips adding it, otherwise adds it as enabled.
	// If language is not in enabled languages, but is present in the database, it is marked as disabled.
	Init(db database.Database, enabledLanguages schemas.HandShakeResponsePayload) error
}

// languageService implements [LanguageService] interface.
type languageService struct {
	languageRepository repository.LanguageRepository
	logger             *zap.SugaredLogger
}

// Init implements Init method of [LanguageService] interface.
func (l *languageService) Init(db database.Database, workerLanguages schemas.HandShakeResponsePayload) error {
	l.logger.Infof("Initializing languages: %v", workerLanguages.Languages)

	existingLanguages, err := l.languageRepository.GetAll(db)
	if err != nil {
		l.logger.Errorf("Error getting existing languages: %v", err.Error())
		return err
	}
	for _, lang := range workerLanguages.Languages {
		for _, version := range lang.Versions {
			language := models.LanguageConfig{
				Type:          lang.Name,
				Version:       version,
				FileExtension: lang.Extension,
			}
			var found bool
			for i, existingLanguage := range existingLanguages {
				if existingLanguage.Type == language.Type && existingLanguage.Version == language.Version {
					found = true
					existingLanguages = append(existingLanguages[:i], existingLanguages[i+1:]...)
					break
				}
			}
			if !found {
				err := l.languageRepository.Create(db, &language)
				if err != nil {
					l.logger.Errorf("Error adding language: %v", err.Error())
					return err
				}
			}
		}
	}

	if len(existingLanguages) > 0 {
		for _, lang := range existingLanguages {
			err := l.languageRepository.MarkDisabled(db, lang.ID)
			if err != nil {
				l.logger.Errorf("Error marking language disabled: %v", err.Error())
				return err
			}
		}
	}
	return nil
}

func (l *languageService) GetAll(db database.Database) ([]schemas.LanguageConfig, error) {
	languages, err := l.languageRepository.GetAll(db)
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

func (l *languageService) GetAllEnabled(db database.Database) ([]schemas.LanguageConfig, error) {
	languages, err := l.languageRepository.GetEnabled(db)
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

func (l *languageService) ToggleLanguageVisibility(db database.Database, languageID int64) error {
	// Get the language to check its current state
	languages, err := l.languageRepository.GetAll(db)
	if err != nil {
		l.logger.Errorf("Error getting languages: %v", err.Error())
		return err
	}

	var found bool
	var currentlyDisabled bool
	for _, lang := range languages {
		if lang.ID == languageID {
			found = true
			if lang.IsDisabled != nil {
				currentlyDisabled = *lang.IsDisabled
			}
			break
		}
	}

	if !found {
		l.logger.Errorf("Language with ID %d not found", languageID)
		return errors.ErrNotFound
	}

	// Toggle the state
	if currentlyDisabled {
		err = l.languageRepository.MarkEnabled(db, languageID)
	} else {
		err = l.languageRepository.MarkDisabled(db, languageID)
	}

	if err != nil {
		l.logger.Errorf("Error toggling language visibility: %v", err.Error())
		return err
	}

	l.logger.Infof("Language %d visibility toggled successfully", languageID)
	return nil
}

func LanguageToSchema(language *models.LanguageConfig) *schemas.LanguageConfig {
	isDisabled := false
	if language.IsDisabled != nil {
		isDisabled = *language.IsDisabled
	}
	return &schemas.LanguageConfig{
		ID:            language.ID,
		Type:          language.Type,
		Version:       language.Version,
		FileExtension: language.FileExtension,
		IsDisabled:    isDisabled,
	}
}

func NewLanguageService(languageRepository repository.LanguageRepository) LanguageService {
	log := utils.NewNamedLogger("language_service")
	return &languageService{
		languageRepository: languageRepository,
		logger:             log,
	}
}
