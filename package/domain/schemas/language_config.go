package schemas

import "github.com/mini-maxit/backend/package/domain/models"

type LanguageConfig struct {
	Id            int64               `json:"id"`
	Type          models.LanguageType `json:"language"`
	Version       string              `json:"version"`
	FileExtension string              `json:"file_extension"`
}
