package service

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type LanguageService interface {
	InitLanguages(tx *gorm.DB) error
}

type LanguageServiceImpl struct {
	languageRepository repository.LanguageRepository
	logger             *zap.SugaredLogger
}

func (l *LanguageServiceImpl) InitLanguages(tx *gorm.DB) error {
	languages := []models.LanguageConfig{
		{
			Type:    models.LangTypeC,
			Version: "99",
		},
		{
			Type:    models.LangTypeC,
			Version: "11",
		},
		{
			Type:    models.LangTypeC,
			Version: "18",
		},
		{
			Type:    models.LangTypeCPP,
			Version: "11",
		},
		{
			Type:    models.LangTypeCPP,
			Version: "14",
		},
		{
			Type:    models.LangTypeCPP,
			Version: "17",
		},
		{
			Type:    models.LangTypeCPP,
			Version: "20",
		},
		{
			Type:    models.LangTypeCPP,
			Version: "23",
		},
	}

	existingLanguages, err := l.languageRepository.GetLanguages(tx)
	if err != nil {
		l.logger.Errorf("Error getting existing languages: %v", err.Error())
		return err
	}
	for _, newLang := range languages {
		found := false
		for _, lang := range existingLanguages {
			if lang.Type == newLang.Type && lang.Version == newLang.Version {
				found = true
				break
			}
		}
		if !found {
			err = l.languageRepository.CreateLanguage(tx, newLang)
			if err != nil {
				l.logger.Errorf("Error creating language: %v", err.Error())
				return err
			}
		}

	}
	return nil
}

func NewLanguageService(languageRepository repository.LanguageRepository) LanguageService {
	log := utils.NewNamedLogger("language_service")
	return &LanguageServiceImpl{
		languageRepository: languageRepository,
		logger:             log,
	}
}
