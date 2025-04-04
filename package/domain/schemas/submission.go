package schemas

import "time"

type Submission struct {
	ID            int64             `json:"id"`
	TaskID        int64             `json:"task_id"`
	UserID        int64             `json:"user_id"`
	Order         int64             `json:"order"`
	LanguageID    int64             `json:"language_id"`
	Status        string            `json:"status"`
	StatusMessage string            `json:"status_message"`
	SubmittedAt   time.Time         `json:"submitted_at"`
	CheckedAt     time.Time         `json:"checked_at"`
	Language      LanguageConfig    `json:"language"`
	Task          Task              `json:"task"`
	User          User              `json:"user"`
	Result        *SubmissionResult `json:"result"`
}

type SubmissionShort struct {
	ID            int64 `json:"id"`
	TaskID        int64 `json:"task_id"`
	UserID        int64 `json:"user_id"`
	Passed        bool  `json:"passed"`
	HowManyPassed int64 `json:"how_many_passed"`
}

type SubmissionResult struct {
	ID           int64        `json:"id"`
	SubmissionID int64        `json:"submission_id"`
	Code         string       `json:"code"`
	Message      string       `json:"message"`
	CreatedAt    time.Time    `json:"created_at"`
	TestResults  []TestResult `json:"test_results"`
}

type TestResult struct {
	ID                 int64  `json:"id"`
	SubmissionResultID int64  `json:"submission_result_id"`
	InputOutputID      int64  `json:"input_output_id"`
	Passed             bool   `json:"passed"`
	ErrorMessage       string `json:"-"`
}
