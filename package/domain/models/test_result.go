package models

type TestResult struct {
	ID                   uint   `gorm:"primaryKey;autoIncrement"`
	UserSolutionResultID uint   `gorm:"not null; foreignKey:UserSolutionResultID"`
	InputOutputID        uint   `gorm:"not null; foreignKey:InputOutputID"`
	OutputFilePath       string `gorm:"type:varchar;not null"`
	Passed               bool   `gorm:"not null"`
	ErrorMessage         string `gorm:"type:varchar"`
}
