package schemas

import (
	"encoding/json"

	"github.com/mini-maxit/backend/package/domain/types"
)

type QueueResponseMessage struct {
	MessageID string          `json:"message_id"`
	Type      string          `json:"type"`
	Ok        bool            `json:"ok"`
	Payload   json.RawMessage `json:"payload"`
}

type SubmissionResultWorkerResponse struct {
	Code        types.SubmissionResultCode `json:"status_code"`
	Message     string                     `json:"message"`
	TestResults []QueueTestResult          `json:"test_results"`
}

type HandShakeResponsePayload struct {
	Languages []struct {
		Name      string   `json:"name"`
		Versions  []string `json:"versions"`
		Extension string   `json:"extension"`
	} `json:"languages"`
}

type queueWorkerStatus struct {
	ID                  int     `json:"worker_id"`
	Status              int     `json:"status"`
	ProcessingMessageID *string `json:"processing_message_id,omitempty"`
}

type StatusResponsePayload struct {
	BusyWorkers  int                 `json:"busy_workers"`
	TotalWorkers int                 `json:"total_workers"`
	WorkerStatus []queueWorkerStatus `json:"worker_status"`
}

type QueueTestResult struct {
	Passed        bool                       `json:"passed"`
	ExecutionTime float64                    `json:"execution_time"`
	PeakMem       int64                      `json:"peak_memory"`
	StatusCode    types.TestResultStatusCode `json:"status_code"`
	ErrorMessage  string                     `json:"error_message"`
	Order         int                        `json:"order"`
}
