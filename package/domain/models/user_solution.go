package models

import "time"

type UserSolution struct {
	ID               uint      `gorm:"primaryKey;autoIncrement"`
	TaskID           uint      `gorm:"not null; foreignKey:TaskID"`
	SolutionFileName string    `gorm:"type:varchar(255);not null"`
	LanguageType     string    `gorm:"type:varchar(255);not null"`
	LanguageVersion  string    `gorm:"type:varchar(50);not null"`
	Status           string    `gorm:"type:varchar(50);not null"`
	SubmittedAt      time.Time `gorm:"autoCreateTime"`
	CheckedAt        *time.Time
	StatusMessage    string `gorm:"type:varchar"`
}

type UserSolutionResult struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	UserSolutionID uint      `gorm:"not null; foreignKey:UserSolutionID"`
	Code           string    `gorm:"not null"`
	Message        string    `gorm:"type:varchar(255);not null"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}
