package service

import (
	"time"

	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/utils"
)

type WorkerService interface {
	GetStatus(currentUser schemas.User) (*schemas.WorkerStatus, error)
}

type workerService struct {
	queueService QueueService
}

func (ws *workerService) GetStatus(currentUser schemas.User) (*schemas.WorkerStatus, error) {
	if err := utils.ValidateRoleAccess(
		currentUser.Role,
		[]types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher},
	); err != nil {
		return nil, errors.ErrNotAuthorized
	}

	err := ws.queueService.PublishWorkerStatus()
	if err != nil {
		return nil, err
	}
	// Set a timeout for waiting
	timeout := time.After(10 * time.Second)
	statusUpdated := make(chan bool, 1)
	var result schemas.WorkerStatus

	go func() {
		ws.queueService.StatusMux().Lock()
		defer ws.queueService.StatusMux().Unlock()

		// Wait for the condition to be signaled
		ws.queueService.StatusCond().Wait()

		// When signaled, capture the result while holding the lock
		result = ws.queueService.LastWorkerStatus()
		statusUpdated <- true
	}()

	// Wait for either the status update or timeout
	select {
	case <-statusUpdated:
		return &result, nil
	case <-timeout:
		return nil, errors.ErrTimeout
	}
}

func NewWorkerService(queueService QueueService) WorkerService {
	return &workerService{
		queueService: queueService,
	}
}
