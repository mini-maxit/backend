package schemas

import "time"

type WorkerStatus struct {
	BusyWorkers  int               `json:"busyWorkers"`
	TotalWorkers int               `json:"totalWorkers"`
	WorkerStatus map[string]string `json:"workerStatus"`
	StatusTime   time.Time         `json:"statusTime"`
}
