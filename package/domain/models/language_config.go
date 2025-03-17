package models

type LanguageType string

type LanguageConfig struct {
	Id            int64  `gorm:"primaryKey;"`
	Type          string `gorm:"not null;"`
	Version       string `gorm:"not null;"`
	FileExtension string `gorm:"not null;"`
	Disabled      bool   `gorm:"not null;"`
}
