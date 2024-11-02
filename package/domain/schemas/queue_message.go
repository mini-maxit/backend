package schemas

const (
	SubmissionStatusQueued = "queued"
)

type QueueMessage struct {
	MessageId       string    `json:"message_id"`
	TaskId          int64     `json:"task_id"`
	UserId          int64     `json:"user_id"`
	SumissionNumber int64     `json:"submission_number"`
	LanguageType    string    `json:"language_type"`
	LanguageVersion string    `json:"language_version"`
	TimeLimits      []float64 `json:"time_limits"`
	MemoryLimits    []float64 `json:"memory_limits"`
}
