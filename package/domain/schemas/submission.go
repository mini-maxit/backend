package schemas

import "time"

type Submission struct {
	ID          int64             `json:"id"`
	TaskID      int64             `json:"taskId"`
	UserID      int64             `json:"userId"`
	Order       int               `json:"order"`
	LanguageID  int64             `json:"languageId"`
	Status      string            `json:"status"`
	SubmittedAt time.Time         `json:"submittedAt"`
	CheckedAt   time.Time         `json:"checkedAt"`
	Language    LanguageConfig    `json:"language"`
	Task        Task              `json:"task"`
	User        User              `json:"user"`
	Result      *SubmissionResult `json:"result"`
}

type SubmissionShort struct {
	ID            int64 `json:"id"`
	TaskID        int64 `json:"taskId"`
	UserID        int64 `json:"userId"`
	Passed        bool  `json:"passed"`
	HowManyPassed int64 `json:"howManyPassed"`
}

type SubmissionResult struct {
	ID           int64        `json:"id"`
	SubmissionID int64        `json:"submissionId"`
	Code         string       `json:"code"`
	Message      string       `json:"message"`
	CreatedAt    time.Time    `json:"createdAt"`
	TestResults  []TestResult `json:"testResults"`
}

type TestResult struct {
	ID                 int64  `json:"id"`
	SubmissionResultID int64  `json:"submissionResultId"`
	TestCaseID         int64  `json:"testCaseId"`
	Passed             bool   `json:"passed"`
	ErrorMessage       string `json:"-"`
}
