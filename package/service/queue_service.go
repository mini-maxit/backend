package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	publishTimeoutSeconds        = 5
	pendingSubmissionsBatchLimit = 100
)

type QueueService interface {
	// PublishTask publishes a task to the queue
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
	// SetConnection updates the connection used by the service
	SetConnection(conn *amqp.Connection, channel *amqp.Channel) error
	StatusMux() *sync.Mutex
	StatusCond() *sync.Cond
	LastWorkerStatus() schemas.WorkerStatus
}

type queueService struct {
	testCaseRepository         repository.TestCaseRepository
	taskRepository             repository.TaskRepository
	submissionRepository       repository.SubmissionRepository
	submissionResultRepository repository.SubmissionResultRepository
	queueRepository            repository.QueueMessageRepository
	queueName                  string
	responseQueueName          string

	connMux sync.RWMutex
	channel *amqp.Channel
	conn    *amqp.Connection
	queue   amqp.Queue

	statusMux        *sync.Mutex
	statusCond       *sync.Cond
	lastWorkerStatus schemas.WorkerStatus

	logger *zap.SugaredLogger
}

func (qs *queueService) publishMessage(msq schemas.QueueMessage) error {
	qs.connMux.RLock()
	channel := qs.channel
	queueName := qs.queue.Name
	qs.connMux.RUnlock()

	if channel == nil || channel.IsClosed() {
		qs.logger.Warn("Queue channel is not available - message will not be published")
		return errors.New("queue channel is not available")
	}

	msgBytes, err := json.Marshal(msq)
	if err != nil {
		qs.logger.Errorf("Error marshalling message: %v", err.Error())
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), publishTimeoutSeconds*time.Second)
	defer cancel()

	err = channel.PublishWithContext(ctx, "", queueName, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        msgBytes,
		ReplyTo:     qs.responseQueueName,
	})
	if err != nil {
		qs.logger.Errorf("Error publishing message: %v", err.Error())
		return err
	}

	qs.logger.Info("Message published")
	return nil
}

func (qs *queueService) PublishSubmission(tx *gorm.DB, submissionID int64, submissionResultID int64) error {
	submission, err := qs.submissionRepository.Get(tx, submissionID)
	if err != nil {
		qs.logger.Errorf("Error getting submission: %v", err.Error())
		return err
	}

	submissionResult, err := qs.submissionResultRepository.Get(tx, submissionResultID)
	if err != nil {
		qs.logger.Errorf("Error getting submission result: %v", err.Error())
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
		qs.logger.Errorf("Error marshalling payload: %v", err.Error())
		return err
	}

	msq := schemas.QueueMessage{
		MessageID: strconv.FormatInt(submissionID, 10),
		Type:      schemas.MessageTypeTask,
		Payload:   payloadJSON,
	}
	err = qs.publishMessage(msq)
	if err != nil {
		// Don't fail the submission - just keep it in "received" status
		// It will be picked up later when queue becomes available
		qs.logger.Warnf("Queue unavailable - submission %d will be queued later: %v", submissionID, err.Error())
		return nil
	}
	err = qs.submissionRepository.MarkProcessing(tx, submissionID)
	if err != nil {
		qs.logger.Errorf("Error marking submission processing: %v", err.Error())
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
	msq := schemas.QueueMessage{
		MessageID: uuid.New().String(),
		Type:      schemas.MessageTypeHandshake,
		Payload:   nil,
	}
	err := qs.publishMessage(msq)
	if err != nil {
		qs.logger.Errorf("Error publishing message: %v", err.Error())
		return err
	}
	qs.logger.Info("Handshake published")
	return nil
}

func (qs *queueService) PublishWorkerStatus() error {
	msq := schemas.QueueMessage{
		MessageID: uuid.New().String(),
		Type:      schemas.MessageTypeStatus,
		Payload:   nil,
	}
	err := qs.publishMessage(msq)
	if err != nil {
		qs.logger.Errorf("Error publishing message: %v", err.Error())
		return err
	}
	qs.logger.Info("Worker status published")
	return nil
}

func (qs *queueService) UpdateWorkerStatus(recievedStatus schemas.StatusResponsePayload) error {
	qs.statusMux.Lock()
	defer qs.statusMux.Unlock()

	lastStatus := schemas.WorkerStatus{
		BusyWorkers:  recievedStatus.BusyWorkers,
		TotalWorkers: recievedStatus.TotalWorkers,
		WorkerStatus: recievedStatus.WorkerStatus,
		StatusTime:   time.Now(),
	}
	qs.lastWorkerStatus = lastStatus // Update the last worker status

	qs.statusCond.Broadcast() // Notify any waiting goroutines that the status has changed

	return nil
}

func (qs *queueService) LastWorkerStatus() schemas.WorkerStatus {
	return qs.lastWorkerStatus
}

func (qs *queueService) StatusCond() *sync.Cond {
	return qs.statusCond
}

func (qs *queueService) StatusMux() *sync.Mutex {
	return qs.statusMux
}

func (qs *queueService) RetryPendingSubmissions(db *gorm.DB) error {
	if qs.channel == nil {
		qs.logger.Debug("Queue channel not available - skipping retry of pending submissions")
		return errors.New("queue channel not available")
	}

	// Get pending submissions (limit to avoid overwhelming the queue)
	pendingSubmissions, err := qs.submissionRepository.GetPendingSubmissions(db, pendingSubmissionsBatchLimit)
	if err != nil {
		qs.logger.Errorf("Error getting pending submissions: %v", err.Error())
		return err
	}

	if len(pendingSubmissions) == 0 {
		qs.logger.Debug("No pending submissions to retry")
		return nil
	}

	qs.logger.Infof("Found %d pending submissions to queue", len(pendingSubmissions))

	successCount := 0
	for _, submission := range pendingSubmissions {
		// Get the submission result ID
		if submission.Result == nil {
			qs.logger.Warnf("Submission %d has no result - skipping", submission.ID)
			continue
		}

		// Note: db is a non-transaction connection, so each PublishSubmission call
		// will update the submission status independently without transaction conflicts
		err := qs.PublishSubmission(db, submission.ID, submission.Result.ID)
		if err != nil {
			qs.logger.Warnf("Failed to queue submission %d: %v", submission.ID, err)
			// Continue with other submissions even if one fails
			continue
		}
		successCount++
	}

	qs.logger.Infof("Successfully queued %d out of %d pending submissions", successCount, len(pendingSubmissions))
	return nil
}

func NewQueueService(
	taskRepository repository.TaskRepository,
	submissionRepository repository.SubmissionRepository,
	submissionResultRepository repository.SubmissionResultRepository,
	queueMessageRepository repository.QueueMessageRepository,
	queueName string,
	responseQueueName string,
) QueueService {
	log := utils.NewNamedLogger("queue_service")
	log.Info("Queue service initialized - connection will be established by queue listener")

	s := &queueService{
		taskRepository:             taskRepository,
		submissionRepository:       submissionRepository,
		submissionResultRepository: submissionResultRepository,
		queueRepository:            queueMessageRepository,
		queueName:                  queueName,
		responseQueueName:          responseQueueName,

		statusMux:        &sync.Mutex{},
		lastWorkerStatus: schemas.WorkerStatus{},
		logger:           log,
	}
	s.statusCond = sync.NewCond(s.statusMux)
	return s
}

func (qs *queueService) SetConnection(conn *amqp.Connection, channel *amqp.Channel) error {
	qs.connMux.Lock()
	defer qs.connMux.Unlock()

	if channel != nil {
		// Declare the worker queue for publishing submissions
		args := make(amqp.Table)
		args["x-max-priority"] = 3

		q, err := channel.QueueDeclare(
			qs.queueName,
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			args,
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue: %w", err)
		}

		qs.queue = q
		qs.channel = channel
		qs.conn = conn
		qs.logger.Info("Queue service connection established")

		// Try to publish handshake
		go func() {
			if err := qs.PublishHandshake(); err != nil {
				qs.logger.Warnf("Failed to publish handshake: %v", err)
			}
		}()
	} else {
		qs.channel = nil
		qs.conn = nil
		qs.queue = amqp.Queue{}
		qs.logger.Warn("Queue service connection cleared")
	}

	return nil
}

func (qs *queueService) IsConnected() bool {
	qs.connMux.RLock()
	defer qs.connMux.RUnlock()
	return qs.channel != nil && !qs.channel.IsClosed()
}

func (qs *queueService) SetConnectionNotifyCallback(callback func(connected bool)) {
	// Deprecated - connection management now handled by queue listener
}

func (qs *queueService) Reconnect() error {
	qs.connMux.RLock()
	defer qs.connMux.RUnlock()

	if qs.channel == nil || qs.channel.IsClosed() {
		return errors.New("queue is not connected - automatic reconnection is handled by queue listener")
	}

	qs.logger.Info("Queue is connected - ready to process pending submissions")
	return nil
}
