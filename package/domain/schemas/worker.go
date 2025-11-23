package schemas

import "time"

type WorkerStatus struct {
	ID                  int     `json:"id"`
	Status              string  `json:"status"`
	ProcessingMessageID *string `json:"processingMessageId,omitempty"`
}

type WorkersStatus struct {
	BusyWorkers  int            `json:"busyWorkers"`
	TotalWorkers int            `json:"totalWorkers"`
	Statuses     []WorkerStatus `json:"workerStatus"`
	StatusTime   time.Time      `json:"statusTime"`
}

type QueueStatus struct {
	Connected          bool      `json:"connected"`
	PendingSubmissions int       `json:"pendingSubmissions"`
	LastChecked        time.Time `json:"lastChecked"`
}
