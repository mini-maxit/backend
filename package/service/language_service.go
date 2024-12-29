package service

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
)

type LanguageService interface {
	modelToSchema(language *models.LanguageConfig) *schemas.LanguageConfig
}

type LanguageServiceImpl struct {}

func (ls *LanguageServiceImpl) modelToSchema(language *models.LanguageConfig) *schemas.LanguageConfig {
	return &schemas.LanguageConfig{
		Language: string(language.Type),
		Version: language.Version,
	}
}

func NewLanguageService() LanguageService {
	return &LanguageServiceImpl{}
}
