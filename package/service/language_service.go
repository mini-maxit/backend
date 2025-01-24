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
	// InitLanguages initializes languages in the database
	//

	// It should be called during application initialization.
	// Method initializes languages in the database if they are not already present.
	// If language is already present in the database, and is not disabled it skips it. Otherwise, it enables it.
	// If language is not in enabled languages, but is present in the database, it is marked as disabled.
	InitLanguages(tx *gorm.DB, enabledLanguages []schemas.LanguageConfig) error
	modelToSchema(language *models.LanguageConfig) *schemas.LanguageConfig
}

// languageService implements [LanguageService] interface
type languageService struct {
	languageRepository repository.LanguageRepository
	logger             *zap.SugaredLogger
}

// InitLanguages implements InitLanguages method of [LanguageService] interface
func (l *languageService) InitLanguages(tx *gorm.DB, enabledLanguages []schemas.LanguageConfig) error {
	existingLanguages, err := l.languageRepository.GetLanguages(tx)
	if err != nil {
		l.logger.Errorf("Error getting existing languages: %v", err.Error())
		return err
	}
	for _, newLang := range enabledLanguages {
		found := false
		for _, lang := range existingLanguages {
			if lang.Type == newLang.Type && lang.Version == newLang.Version {
				found = true
				break
			}
		}
		if !found {
			langModel := &models.LanguageConfig{
				Type:    newLang.Type,
				Version: newLang.Version,
			}
			err = l.languageRepository.CreateLanguage(tx, langModel)
			if err != nil {
				l.logger.Errorf("Error creating language: %v", err.Error())
				return err
			}
		}

	}
	return nil
}

func (l *languageService) modelToSchema(language *models.LanguageConfig) *schemas.LanguageConfig {
	return &schemas.LanguageConfig{
		Type:    language.Type,
		Version: language.Version,
	}
}

func NewLanguageService(languageRepository repository.LanguageRepository) LanguageService {
	log := utils.NewNamedLogger("language_service")
	return &languageService{
		languageRepository: languageRepository,
		logger:             log,
	}
}
