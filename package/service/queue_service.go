package service

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/queue"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	publishTimeoutSeconds        = 5
	pendingSubmissionsBatchLimit = 100
)

type QueueService interface {
	// GetSubmissionID returns the submission ID for a message ID
	GetSubmissionID(tx *gorm.DB, messageID string) (int64, error)
	// PublishHandshake publishes a handshake message to the queue
	PublishHandshake() error
	// PublishSubmission publishes a submission message to the queue
	PublishSubmission(tx *gorm.DB, submissionID int64, submissionResultID int64) error
	// PublishWorkerStatus publishes a worker status message to the queue
	PublishWorkerStatus() error
	// UpdateWorkerStatus updates the worker status in the database
	UpdateWorkerStatus(statusResponse schemas.StatusResponsePayload) error
	// RetryPendingSubmissions attempts to queue submissions that are in "received" status
	RetryPendingSubmissions(db *gorm.DB) error
	// IsConnected returns true if queue is connected and ready
	IsConnected() bool
	// StatusMux returns the status mutex
	StatusMux() *sync.Mutex
	// StatusCond returns the status condition variable
	StatusCond() *sync.Cond
	// LastWorkerStatus returns the last known worker status
	LastWorkerStatus() schemas.WorkersStatus
}

type queueService struct {
	testCaseRepository         repository.TestCaseRepository
	taskRepository             repository.TaskRepository
	submissionRepository       repository.SubmissionRepository
	submissionResultRepository repository.SubmissionResultRepository
	queueRepository            repository.QueueMessageRepository

	// Queue client for publishing
	queueClient       queue.Publisher
	queueName         string
	responseQueueName string

	statusMux        *sync.Mutex
	statusCond       *sync.Cond
	lastWorkerStatus schemas.WorkersStatus

	logger *zap.SugaredLogger
}

func NewQueueService(
	taskRepository repository.TaskRepository,
	submissionRepository repository.SubmissionRepository,
	submissionResultRepository repository.SubmissionResultRepository,
	queueMessageRepository repository.QueueMessageRepository,
	queueClient queue.Publisher,
	queueName string,
	responseQueueName string,
) QueueService {
	log := utils.NewNamedLogger("queue_service")
	log.Info("Queue service initialized")

	s := &queueService{
		taskRepository:             taskRepository,
		submissionRepository:       submissionRepository,
		submissionResultRepository: submissionResultRepository,
		queueRepository:            queueMessageRepository,
		queueClient:                queueClient,
		queueName:                  queueName,
		responseQueueName:          responseQueueName,
		statusMux:                  &sync.Mutex{},
		lastWorkerStatus:           schemas.WorkersStatus{},
		logger:                     log,
	}
	s.statusCond = sync.NewCond(s.statusMux)
	return s
}

func (qs *queueService) publishMessage(msg schemas.QueueMessage) error {
	if !qs.queueClient.IsConnected() {
		qs.logger.Warn("Queue is not connected - message will not be published")
		return errors.New("queue is not connected")
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		qs.logger.Errorf("Error marshalling message: %v", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), publishTimeoutSeconds*time.Second)
	defer cancel()

	err = qs.queueClient.Publish(ctx, qs.queueName, qs.responseQueueName, msgBytes)
	if err != nil {
		qs.logger.Errorf("Error publishing message: %v", err)
		return err
	}

	qs.logger.Info("Message published")
	return nil
}

func (qs *queueService) PublishSubmission(tx *gorm.DB, submissionID int64, submissionResultID int64) error {
	submission, err := qs.submissionRepository.Get(tx, submissionID)
	if err != nil {
		qs.logger.Errorf("Error getting submission: %v", err)
		return err
	}

	submissionResult, err := qs.submissionResultRepository.Get(tx, submissionResultID)
	if err != nil {
		qs.logger.Errorf("Error getting submission result: %v", err)
		return err
	}

	testCases := make([]schemas.QTestCase, 0, len(submissionResult.TestResults))
	for _, tr := range submissionResult.TestResults {
		testCases = append(testCases, schemas.QTestCase{
			Order: tr.TestCase.Order,
			InputFile: schemas.FileLocation{
				ServerType: tr.TestCase.InputFile.ServerType,
				Bucket:     tr.TestCase.InputFile.Bucket,
				Path:       tr.TestCase.InputFile.Path,
			},
			ExpectedOutput: schemas.FileLocation{
				ServerType: tr.TestCase.OutputFile.ServerType,
				Bucket:     tr.TestCase.OutputFile.Bucket,
				Path:       tr.TestCase.OutputFile.Path,
			},
			StdoutResult: schemas.FileLocation{
				ServerType: tr.StdoutFile.ServerType,
				Bucket:     tr.StdoutFile.Bucket,
				Path:       tr.StdoutFile.Path,
			},
			StderrResult: schemas.FileLocation{
				ServerType: tr.StderrFile.ServerType,
				Bucket:     tr.StderrFile.Bucket,
				Path:       tr.StderrFile.Path,
			},
			DiffResult: schemas.FileLocation{
				ServerType: tr.DiffFile.ServerType,
				Bucket:     tr.DiffFile.Bucket,
				Path:       tr.DiffFile.Path,
			},
			TimeLimitMs:   tr.TestCase.TimeLimit,
			MemoryLimitKB: tr.TestCase.MemoryLimit,
		})
	}

	payload := schemas.TaskQueueMessage{
		Order:           submission.Order,
		LanguageType:    submission.Language.Type,
		LanguageVersion: submission.Language.Version,
		SubmissionFile: schemas.FileLocation{
			ServerType: submission.File.ServerType,
			Bucket:     submission.File.Bucket,
			Path:       submission.File.Path,
		},
		TestCases: testCases,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		qs.logger.Errorf("Error marshalling payload: %v", err)
		return err
	}

	msg := schemas.QueueMessage{
		MessageID: strconv.FormatInt(submissionID, 10),
		Type:      schemas.MessageTypeTask,
		Payload:   payloadJSON,
	}

	err = qs.publishMessage(msg)
	if err != nil {
		// Don't fail the submission - just keep it in "received" status
		// It will be picked up later when queue becomes available
		qs.logger.Warnf("Queue unavailable - submission %d will be queued later: %v", submissionID, err)
		return nil
	}
	err = qs.submissionRepository.MarkProcessing(tx, submissionID)
	if err != nil {
		qs.logger.Errorf("Error marking submission processing: %v", err)
		return err
	}
	qs.logger.Info("Submission published")

	return nil
}

func (qs *queueService) GetSubmissionID(tx *gorm.DB, messageID string) (int64, error) {
	queueMessage, err := qs.queueRepository.Get(tx, messageID)
	if err != nil {
		return 0, err
	}
	return queueMessage.SubmissionID, nil
}

func (qs *queueService) PublishHandshake() error {
	msg := schemas.QueueMessage{
		MessageID: uuid.New().String(),
		Type:      schemas.MessageTypeHandshake,
		Payload:   nil,
	}
	err := qs.publishMessage(msg)
	if err != nil {
		qs.logger.Errorf("Error publishing handshake: %v", err)
		return err
	}
	qs.logger.Info("Handshake published")
	return nil
}

func (qs *queueService) PublishWorkerStatus() error {
	msg := schemas.QueueMessage{
		MessageID: uuid.New().String(),
		Type:      schemas.MessageTypeStatus,
		Payload:   nil,
	}
	err := qs.publishMessage(msg)
	if err != nil {
		qs.logger.Errorf("Error publishing worker status: %v", err)
		return err
	}
	qs.logger.Info("Worker status published")
	return nil
}

func (qs *queueService) UpdateWorkerStatus(receivedStatus schemas.StatusResponsePayload) error {
	qs.statusMux.Lock()
	defer qs.statusMux.Unlock()

	statuses := make([]schemas.WorkerStatus, 0, len(receivedStatus.WorkerStatus))
	for _, ws := range receivedStatus.WorkerStatus {
		statuses = append(statuses, schemas.WorkerStatus{
			ID:                  ws.ID,
			Status:              ws.Status.String(),
			ProcessingMessageID: ws.ProcessingMessageID,
		})
	}
	slices.SortFunc(statuses, func(a schemas.WorkerStatus, b schemas.WorkerStatus) int {
		return a.ID - b.ID
	})
	qs.lastWorkerStatus = schemas.WorkersStatus{
		BusyWorkers:  receivedStatus.BusyWorkers,
		TotalWorkers: receivedStatus.TotalWorkers,
		Statuses:     statuses,

		StatusTime: time.Now(),
	}

	qs.statusCond.Broadcast()
	return nil
}

func (qs *queueService) LastWorkerStatus() schemas.WorkersStatus {
	return qs.lastWorkerStatus
}

func (qs *queueService) StatusCond() *sync.Cond {
	return qs.statusCond
}

func (qs *queueService) StatusMux() *sync.Mutex {
	return qs.statusMux
}

func (qs *queueService) RetryPendingSubmissions(db *gorm.DB) error {
	if !qs.queueClient.IsConnected() {
		qs.logger.Debug("Queue not connected - skipping retry of pending submissions")
		return errors.New("queue not connected")
	}

	pendingSubmissions, err := qs.submissionRepository.GetPendingSubmissions(db, pendingSubmissionsBatchLimit)
	if err != nil {
		qs.logger.Errorf("Error getting pending submissions: %v", err)
		return err
	}

	if len(pendingSubmissions) == 0 {
		qs.logger.Debug("No pending submissions to retry")
		return nil
	}

	qs.logger.Infof("Found %d pending submissions to queue", len(pendingSubmissions))

	successCount := 0
	for _, submission := range pendingSubmissions {
		if submission.Result == nil {
			qs.logger.Warnf("Submission %d has no result - skipping", submission.ID)
			continue
		}

		// Note: db is a non-transaction connection, so each PublishSubmission call
		// will update the submission status independently without transaction conflicts
		err := qs.PublishSubmission(db, submission.ID, submission.Result.ID)
		if err != nil {
			qs.logger.Warnf("Failed to queue submission %d: %v", submission.ID, err)
			continue
		}
		successCount++
	}

	qs.logger.Infof("Successfully queued %d out of %d pending submissions", successCount, len(pendingSubmissions))
	return nil
}

func (qs *queueService) IsConnected() bool {
	return qs.queueClient.IsConnected()
}
