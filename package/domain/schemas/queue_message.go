package schemas

import "encoding/json"

const (
	SubmissionStatusQueued = "queued"
	MessageTypeTask        = "task"
	MessageTypeHandshake   = "handshake"
	MessageTypeStatus      = "status"
)

type QueueMessage struct {
	Type      string          `json:"type"`
	MessageID string          `json:"message_id"`
	Payload   json.RawMessage `json:"payload"`
}

type FileLocation struct {
	ServerType string `json:"server_type"` // Currently only "filestorage" is supported
	Bucket     string `json:"bucket"`      // e.g., "submissions", "tasks", "results"
	Path       string `json:"path"`        // e.g., "42/main.cpp"
}

type TestCase struct {
	Order          int          `json:"order"`
	InputFile      FileLocation `json:"input_file"`
	ExpectedOutput FileLocation `json:"expected_output"`
	StdoutResult   FileLocation `json:"stdout_result"`
	StderrResult   FileLocation `json:"stderr_result"`
	DiffResult     FileLocation `json:"diff_result"`
	TimeLimitMs    int64        `json:"time_limit_ms"`
	MemoryLimitKB  int64        `json:"memory_limit_kb"`
}

type TaskQueueMessage struct {
	Order           int          `json:"order"`
	LanguageType    string       `json:"language_type"`
	LanguageVersion string       `json:"language_version"`
	SubmissionFile  FileLocation `json:"submission_file"`
	TestCases       []TestCase   `json:"test_cases"`
}
