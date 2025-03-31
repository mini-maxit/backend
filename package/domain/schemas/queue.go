package schemas

import "encoding/json"

type QueueResponseMessage struct {
	MessageID string          `json:"message_id"`
	Type      string          `json:"type"`
	Ok        bool            `json:"ok"`
	Payload   json.RawMessage `json:"payload"`
}

type QueueResult struct {
	Success     bool              `json:"Success"`
	StatusCode  int64             `json:"StatusCode"`
	Message     string            `json:"Message"`
	TestResults []QueueTestResult `json:"TestResults"`
}

type TaskResponsePayload struct {
	Success     bool              `json:"Success"`
	StatusCode  int64             `json:"StatusCode"`
	Message     string            `json:"Message"`
	TestResults []QueueTestResult `json:"TestResults"`
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
	Order        int64  `json:"Order"`
	Passed       bool   `json:"Passed"`
	ErrorMessage string `json:"ErrorMessage"`
}
