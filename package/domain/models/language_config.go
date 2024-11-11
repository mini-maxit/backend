package models

type LanguageType string

type LanguageConfig struct {
	Id      int64        `gorm:"primaryKey;"`
	Type    LanguageType `gorm:"not null;"`
	Version string       `gorm:"not null;"`
}
