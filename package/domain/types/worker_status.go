package types

type WorkerStatusType string

const (
	WorkerStatusIdle    WorkerStatusType = "idle"
	WorkerStatusBusy    WorkerStatusType = "busy"
	WorkerStatusInvalid WorkerStatusType = "invalid"
)

func IntToWorkerStatusType(status int) WorkerStatusType {
	switch status {
	case 0:
		return WorkerStatusIdle
	case 1:
		return WorkerStatusBusy
	default:
		return WorkerStatusInvalid
	}
}
