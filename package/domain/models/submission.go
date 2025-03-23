package models

import "time"

const (
	StatusReceived          string = "received"
	StatusSentForEvaluation string = "sent for evaluation"
	StatusEvaluated         string = "evaluated"
	StatusLost              string = "lost"
)

type Submission struct {
	Id     int64 `gorm:"primaryKey;autoIncrement"`
	TaskId int64 `gorm:"not null; foreignKey:TaskID"`
	UserId int64 `gorm:"not null; foreignKey:UserID"`
	// Order represents the submission attempt number for a specific task by a user.
	// The first submission has Order = 1, the second = 2, and so on.
	// It helps track the sequence of submissions made by a user for a given task.
	Order         int64             `gorm:"not null"`
	LanguageId    int64             `gorm:"not null; foreignKey:LanguageID"`
	Status        string            `gorm:"type:varchar(50);not null"`
	StatusMessage string            `gorm:"type:varchar"`
	SubmittedAt   time.Time         `gorm:"type:timestamp;autoCreateTime"`
	CheckedAt     time.Time         `gorm:"type:timestamp;default:null"`
	Language      LanguageConfig    `gorm:"foreignKey:LanguageId;references:Id"`
	Task          Task              `gorm:"foreignKey:TaskId;references:Id"`
	User          User              `gorm:"foreignKey:UserId;references:Id"`
	Result        *SubmissionResult `gorm:"foreignKey:SubmissionId;references:Id"`
}

type SubmissionResult struct {
	Id           int64        `gorm:"primaryKey;autoIncrement"`
	SubmissionId int64        `gorm:"not null"`
	Code         string       `gorm:"not null"`
	Message      string       `gorm:"type:varchar(255);not null"`
	CreatedAt    time.Time    `gorm:"autoCreateTime"`
	Submission   Submission   `gorm:"foreignKey:SubmissionId;references:Id"`
	TestResult   []TestResult `gorm:"foreignKey:SubmissionResultId;references:Id"`
}

type TestResult struct {
	ID                 int64            `gorm:"primaryKey;autoIncrement"`
	SubmissionResultId int64            `gorm:"not null"`
	InputOutputId      int64            `gorm:"not null"`
	Passed             bool             `gorm:"not null"`
	ErrorMessage       string           `gorm:"type:varchar"`
	InputOutput        InputOutput      `gorm:"foreignKey:InputOutputId;references:Id"`
	SubmissionResult   SubmissionResult `gorm:"foreignKey:SubmissionResultId;references:Id"`
}
