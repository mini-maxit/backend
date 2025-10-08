package models

import "time"

const (
	StatusReceived          string = "received"
	StatusSentForEvaluation string = "sent for evaluation"
	StatusEvaluated         string = "evaluated"
	StatusLost              string = "lost"
)

type Submission struct {
	ID     int64 `gorm:"primaryKey;autoIncrement"`
	TaskID int64 `gorm:"not null; foreignKey:TaskID"`
	UserID int64 `gorm:"not null; foreignKey:UserID"`
	// Order represents the submission attempt number for a specific task by a user.
	// The first submission has Order = 1, the second = 2, and so on.
	// It helps track the sequence of submissions made by a user for a given task.
	Order         int               `gorm:"not null"`
	LanguageID    int64             `gorm:"not null; foreignKey:LanguageID"`
	FileID        int64             `gorm:"not null; foreignKey:FileID"`
	Status        string            `gorm:"type:varchar(50);not null"`
	StatusMessage string            `gorm:"type:varchar(255);default:null"`
	SubmittedAt   time.Time         `gorm:"type:timestamp;autoCreateTime"`
	CheckedAt     time.Time         `gorm:"type:timestamp;default:null"`
	Language      LanguageConfig    `gorm:"foreignKey:LanguageID;references:ID"`
	Task          Task              `gorm:"foreignKey:TaskID;references:ID"`
	User          User              `gorm:"foreignKey:UserID;references:ID"`
	Result        *SubmissionResult `gorm:"foreignKey:SubmissionID;references:ID"`
	File          File              `gorm:"foreignKey:FileID;references:ID"`
}

type SubmissionResult struct {
	ID           int64        `gorm:"primaryKey;autoIncrement"`
	SubmissionID int64        `gorm:"not null"`
	Code         string       `gorm:"not null"`
	Message      string       `gorm:"type:varchar(255);not null"`
	CreatedAt    time.Time    `gorm:"autoCreateTime"`
	Submission   Submission   `gorm:"foreignKey:SubmissionID;references:ID"`
	TestResult   []TestResult `gorm:"foreignKey:SubmissionResultID;references:ID"`
}

type TestResultStatusCode int

const (
	TestResultStatusCodeOK TestResultStatusCode = iota + 1
	TestResultStatusCodeAnswerMismatch
	TestResultStatusCodeTimeLimit
	TestResultStatusCodeMemoryLimit
	TestResultStatusCodeRuntimeError
	TestResultStatusCodeNotExecuted
)

type TestResult struct {
	ID                 int64                `gorm:"primaryKey;autoIncrement"`
	SubmissionResultID int64                `gorm:"not null"`
	InputOutputID      int64                `gorm:"not null"`
	Passed             bool                 `gorm:"not null"`
	ExecutionTime      float64              `gorm:"not null"`
	StatusCode         TestResultStatusCode `gorm:"not null"`
	ErrorMessage       string               `gorm:"type:varchar"`
	StdoutFileID       int64                `gorm:"not null"`
	StderrFileID       int64                `gorm:"not null"`
	InputOutput        TestCase             `gorm:"foreignKey:InputOutputID;references:ID"`
	SubmissionResult   SubmissionResult     `gorm:"foreignKey:SubmissionResultID;references:ID"`
	StdoutFile         File                 `gorm:"foreignKey:StdoutFileID;references:ID"`
	StderrFile         File                 `gorm:"foreignKey:StderrFileID;references:ID"`
}
