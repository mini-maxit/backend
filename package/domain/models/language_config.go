package models

type LanguageType string

const (
	LangTypeC   LanguageType = "c"
	LangTypeCPP LanguageType = "cpp"
)

type LanguageConfig struct {
	Id            int64        `gorm:"primaryKey;"`
	Type          LanguageType `gorm:"not null;"`
	Version       string       `gorm:"not null;"`
	FileExtension string       `gorm:"not null;"`
}
