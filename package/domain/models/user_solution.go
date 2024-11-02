package models

import "time"

type UserSolution struct {
	Id              int64      `gorm:"primaryKey;autoIncrement"`
	TaskId          int64      `gorm:"not null; foreignKey:TaskID"`
	UserId          int64      `gorm:"not null; foreignKey:UserID"`
	Order           int64      `gorm:"not null"`
	LanguageType    string     `gorm:"type:varchar(255);not null"`
	LanguageVersion string     `gorm:"type:varchar(50);not null"`
	Status          string     `gorm:"type:varchar(50);not null"`
	StatusMessage   string     `gorm:"type:varchar"`
	SubmittedAt     time.Time  `gorm:"type:datetime;autoCreateTime"`
	CheckedAt       *time.Time `gorm:"type:datetime"`
	Task            Task       `gorm:"foreignKey:TaskId;references:Id"`
	User            User       `gorm:"foreignKey:UserId;references:Id"`
}

type UserSolutionResult struct {
	Id             int64        `gorm:"primaryKey;autoIncrement"`
	UserSolutionId int64        `gorm:"not null"`
	Code           string       `gorm:"not null"`
	Message        string       `gorm:"type:varchar(255);not null"`
	CreatedAt      time.Time    `gorm:"autoCreateTime"`
	UserSolution   UserSolution `gorm:"foreignKey:UserSolutionId;references:Id"`
}

type TestResult struct {
	ID                   int64              `gorm:"primaryKey;autoIncrement"`
	UserSolutionResultId int64              `gorm:"not null"`
	InputOutputId        int64              `gorm:"not null"`
	Passed               bool               `gorm:"not null"`
	ErrorMessage         string             `gorm:"type:varchar"`
	InputOutput          InputOutput        `gorm:"foreignKey:InputOutputId;references:Id"`
	UserSolutionResult   UserSolutionResult `gorm:"foreignKey:UserSolutionResultId;references:Id"`
}
