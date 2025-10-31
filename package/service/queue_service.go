package service

import (
	"context"
	"encoding/json"
	"github.com/mini-maxit/backend/internal/database"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const publishTimeoutSeconds = 5

type QueueService interface {
	// PublishTask publishes a task to the queue
	GetSubmissionID(tx database.Database, messageID string) (int64, error)
	// PublishHandshake publishes a handshake message to the queue
	PublishHandshake() error
	// PublishSubmission publishes a submission message to the queue
	PublishSubmission(tx database.Database, submissionID int64, submissionResultID int64) error
	// PublishWorkerStatus publishes a worker status message to the queue
	PublishWorkerStatus() error
	// UpdateWorkerStatus updates the worker status in the database
	UpdateWorkerStatus(statusResponse schemas.StatusResponsePayload) error
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
	channel                    *amqp.Channel
	queue                      amqp.Queue
	responseQueueName          string

	statusMux        *sync.Mutex
	statusCond       *sync.Cond
	lastWorkerStatus schemas.WorkerStatus

	logger *zap.SugaredLogger
}

func (qs *queueService) publishMessage(msq schemas.QueueMessage) error {
	msgBytes, err := json.Marshal(msq)
	if err != nil {
		qs.logger.Errorf("Error marshalling message: %v", err.Error())
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), publishTimeoutSeconds*time.Second)
	defer cancel()

	err = qs.channel.PublishWithContext(ctx, "", qs.queue.Name, false, false, amqp.Publishing{
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

func (qs *queueService) PublishSubmission(tx database.Database, submissionID int64, submissionResultID int64) error {
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
		err2 := qs.submissionRepository.MarkFailed(tx, submissionID, err.Error())
		if err2 != nil {
			qs.logger.Errorf("Error marking submission failed: %v. When error occured publishing message: %s",
				err2.Error(),
				err.Error(),
			)
			return err
		}
		qs.logger.Errorf("Error publishing message: %v", err.Error())
		return err
	}
	err = qs.submissionRepository.MarkProcessing(tx, submissionID)
	if err != nil {
		qs.logger.Errorf("Error marking submission processing: %v", err.Error())
		return err
	}
	qs.logger.Info("Submission published")
	return nil
}

func (qs *queueService) GetSubmissionID(tx database.Database, messageID string) (int64, error) {
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

func NewQueueService(
	taskRepository repository.TaskRepository,
	submissionRepository repository.SubmissionRepository,
	submissionResultRepository repository.SubmissionResultRepository,
	queueMessageRepository repository.QueueMessageRepository,
	_ *amqp.Connection, // TODO: review if this is needed
	channel *amqp.Channel,
	queueName string,
	responseQueueName string,
) (QueueService, error) {
	args := make(amqp.Table)
	args["x-max-priority"] = 3

	q, err := channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		args,      // arguments
	)
	if err != nil {
		return nil, err
	}
	log := utils.NewNamedLogger("queue_service")
	s := &queueService{
		taskRepository:             taskRepository,
		submissionRepository:       submissionRepository,
		submissionResultRepository: submissionResultRepository,
		queueRepository:            queueMessageRepository,
		queue:                      q,
		channel:                    channel,
		responseQueueName:          responseQueueName,

		statusMux:        &sync.Mutex{},
		lastWorkerStatus: schemas.WorkerStatus{},
		logger:           log,
	}
	s.statusCond = sync.NewCond(s.statusMux)
	return s, nil
}
