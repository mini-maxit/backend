package service

import (
	"time"

	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"gorm.io/gorm"
)

type WorkerService interface {
	GetStatus(currentUser schemas.User) (*schemas.WorkerStatus, error)
	GetQueueStatus(currentUser schemas.User) (*schemas.QueueStatus, error)
	ReconnectQueue(currentUser schemas.User) error
}

type workerService struct {
	queueService         QueueService
	submissionRepository repository.SubmissionRepository
	db                   *gorm.DB
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

func NewWorkerService(queueService QueueService, submissionRepository repository.SubmissionRepository, db *gorm.DB) WorkerService {
	return &workerService{
		queueService:         queueService,
		submissionRepository: submissionRepository,
		db:                   db,
	}
}

func (ws *workerService) GetQueueStatus(currentUser schemas.User) (*schemas.QueueStatus, error) {
	// Only admin can check queue status
	if err := utils.ValidateRoleAccess(
		currentUser.Role,
		[]types.UserRole{types.UserRoleAdmin},
	); err != nil {
		return nil, errors.ErrNotAuthorized
	}

	connected := ws.queueService.IsConnected()

	// Get count of pending submissions
	pendingSubmissions, err := ws.submissionRepository.GetPendingSubmissions(ws.db, 1000)
	if err != nil {
		return nil, err
	}

	return &schemas.QueueStatus{
		Connected:          connected,
		PendingSubmissions: len(pendingSubmissions),
		LastChecked:        time.Now(),
	}, nil
}

func (ws *workerService) ReconnectQueue(currentUser schemas.User) error {
	// Only admin can trigger reconnection
	if err := utils.ValidateRoleAccess(
		currentUser.Role,
		[]types.UserRole{types.UserRoleAdmin},
	); err != nil {
		return errors.ErrNotAuthorized
	}

	// Attempt reconnection
	err := ws.queueService.Reconnect()
	if err != nil {
		return err
	}

	// Try to process pending submissions
	err = ws.queueService.RetryPendingSubmissions(ws.db)
	if err != nil {
		return err
	}

	return nil
}
