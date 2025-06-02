package schemas

import "encoding/json"

type QueueResponseMessage struct {
	MessageID string          `json:"message_id"`
	Type      string          `json:"type"`
	Ok        bool            `json:"ok"`
	Payload   json.RawMessage `json:"payload"`
}

type TaskResponsePayload struct {
	StatusCode  int64             `json:"status_code"`
	Message     string            `json:"message"`
	TestResults []QueueTestResult `json:"test_results"`
}

type HandShakeResponsePayload struct {
	Languages []struct {
		Name     string   `json:"name"`
		Versions []string `json:"versions"`
	} `json:"languages"`
}

type StatusResponsePayload struct {
	BusyWorkers  int               `json:"busy_workers"`
	TotalWorkers int               `json:"total_workers"`
	WorkerStatus map[string]string `json:"worker_status"`
}

type QueueTestResult struct {
	Passed        bool    `json:"passed"`
	ExecutionTime float64 `json:"execution_time"`
	StatusCode    int     `json:"status_code"`
	ErrorMessage  string  `json:"error_message"`
	Order         int     `json:"order"`
}
