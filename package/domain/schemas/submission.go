package schemas

import "time"

type Submission struct {
	Id            int64          `json:"id"`
	TaskId        int64          `json:"task_id"`
	UserId        int64          `json:"user_id"`
	Order         int64          `json:"order"`
	LanguageId    int64          `json:"language_id"`
	Status        string         `json:"status"`
	StatusMessage string         `json:"status_message"`
	SubmittedAt   time.Time      `json:"submitted_at"`
	CheckedAt     time.Time      `json:"checked_at"`
	Language      LanguageConfig `json:"language"`
	Task          Task           `json:"task"`
	User          User           `json:"user"`
}
