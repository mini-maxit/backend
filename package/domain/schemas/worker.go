package schemas

import "time"

type WorkerStatus struct {
	BusyWorkers  int               `json:"busy_workers"`
	TotalWorkers int               `json:"total_workers"`
	WorkerStatus map[string]string `json:"worker_status"`
	StatusTime   time.Time         `json:"status_time"`
}
