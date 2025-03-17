package schemas

import "encoding/json"

const (
	SubmissionStatusQueued = "queued"
)

type QueueMessage struct {
	Type      string          `json:"type"`
	MessageID string          `json:"message_id"`
	Payload   json.RawMessage `json:"payload"`
}

type TaskQueueMessage struct {
	TaskID           int64  `json:"task_id"`
	UserID           int64  `json:"user_id"`
	SubmissionNumber int64  `json:"submission_number"`
	LanguageType     string `json:"language_type"`
	LanguageVersion  string `json:"language_version"`
	TimeLimits       []int64  `json:"time_limits"`
	MemoryLimits     []int64  `json:"memory_limits"`
	ChrootDirPath    string `json:"chroot_dir_path,omitempty"` // Optional for test purposes
	UseChroot        string `json:"use_chroot,omitempty"`      // Optional for test purposes
}
