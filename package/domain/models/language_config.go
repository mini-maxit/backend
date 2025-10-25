package models

type LanguageType string

type LanguageConfig struct {
	ID            int64  `gorm:"primaryKey;"`
	Type          string `gorm:"not null;"`
	Version       string `gorm:"not null;"`
	FileExtension string `gorm:"not null;"`
	IsDisabled    *bool  `gorm:"not null;default:false"`
}
