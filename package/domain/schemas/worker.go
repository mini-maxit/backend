package schemas

import "time"

type WorkerStatus struct {
	BusyWorkers  int               `json:"busyWorkers"`
	TotalWorkers int               `json:"totalWorkers"`
	WorkerStatus map[string]string `json:"workerStatus"`
	StatusTime   time.Time         `json:"statusTime"`
}

type QueueStatus struct {
	Connected          bool      `json:"connected"`
	PendingSubmissions int       `json:"pendingSubmissions"`
	LastChecked        time.Time `json:"lastChecked"`
}
