package schemas

import "github.com/mini-maxit/backend/package/domain/models"

type LanguageConfig struct {
	Type    models.LanguageType `json:"language"`
	Version string              `json:"version"`
}
