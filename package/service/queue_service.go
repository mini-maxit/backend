package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type QueueService interface {
	// PublishTask publishes a task to the queue
	GetSubmissionId(tx *gorm.DB, messageId string) (int64, error)
	// PublishHandshake publishes a handshake message to the queue
	PublishHandshake() error
	// PublishSubmission publishes a submission message to the queue
	PublishSubmission(tx *gorm.DB, submissionId int64) error
	// PublishWorkerStatus publishes a worker status message to the queue
	PublishWorkerStatus() error
	// UpdateWorkerStatus updates the worker status in the database
	UpdateWorkerStatus(tx *gorm.DB, statusResponse schemas.StatusResponsePayload) error
}

type queueService struct {
	taskRepository       repository.TaskRepository
	submissionRepository repository.SubmissionRepository
	queueRepository      repository.QueueMessageRepository
	channel              *amqp.Channel
	queue                amqp.Queue
	responseQueueName    string
	logger               *zap.SugaredLogger
}

func (qs *queueService) publishMessage(msq schemas.QueueMessage) error {
	msgBytes, err := json.Marshal(msq)
	if err != nil {
		qs.logger.Errorf("Error marshalling message: %v", err.Error())
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

func (qs *queueService) PublishSubmission(tx *gorm.DB, submissionId int64) error {
	submission, err := qs.submissionRepository.Get(tx, submissionId)
	if err != nil {
		qs.logger.Errorf("Error getting submission: %v", err.Error())
		return err
	}

	timeLimits, err := qs.taskRepository.GetTimeLimits(tx, submission.TaskId)
	if err != nil {
		qs.logger.Errorf("Error getting task time limits: %v", err.Error())
		return err
	}
	memoryLimits, err := qs.taskRepository.GetMemoryLimits(tx, submission.TaskId)
	if err != nil {
		qs.logger.Errorf("Error getting task memory limits: %v", err.Error())
		return err
	}

	payload := schemas.TaskQueueMessage{
		TaskID:           submission.TaskId,
		UserID:           submission.UserId,
		SubmissionNumber: submission.Order,
		LanguageType:     string(submission.Language.Type),
		LanguageVersion:  submission.Language.Version,
		TimeLimits:       timeLimits,
		MemoryLimits:     memoryLimits,
	}

	payloadJson, err := json.Marshal(payload)
	if err != nil {
		qs.logger.Errorf("Error marshalling payload: %v", err.Error())
		return err
	}

	msq := schemas.QueueMessage{
		MessageID: uuid.New().String(),
		Type:      schemas.MessageTypeTask,
		Payload:   payloadJson,
	}
	qm := &models.QueueMessage{
		Id:           msq.MessageID,
		SubmissionId: submissionId,
	}
	_, err = qs.queueRepository.Create(tx, qm)
	if err != nil {
		qs.logger.Errorf("Error creating queue message: %v", err.Error())
		return err
	}
	err = qs.publishMessage(msq)
	if err != nil {
		err2 := qs.submissionRepository.MarkFailed(tx, submissionId, err.Error())
		if err2 != nil {
			qs.logger.Errorf("Error marking submission failed: %v. When error occured publishing message: %s", err2.Error(), err.Error())
			return err
		}
		qs.logger.Errorf("Error publishing message: %v", err.Error())
		return err
	}
	err = qs.submissionRepository.MarkProcessing(tx, submissionId)
	if err != nil {
		qs.logger.Errorf("Error marking submission processing: %v", err.Error())
		return err
	}
	qs.logger.Info("Submission published")
	return nil
}

func (qs *queueService) GetSubmissionId(tx *gorm.DB, messageId string) (int64, error) {
	queueMessage, err := qs.queueRepository.Get(tx, messageId)
	if err != nil {
		return 0, err
	}

	return queueMessage.SubmissionId, nil
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

func (q *queueService) UpdateWorkerStatus(tx *gorm.DB, statusResponse schemas.StatusResponsePayload) error {
	panic("implement me")
}

func NewQueueService(taskRepository repository.TaskRepository, submissionRepository repository.SubmissionRepository, queueMessageRepository repository.QueueMessageRepository, conn *amqp.Connection, channel *amqp.Channel, queueName string, responseQueueName string) (*queueService, error) {
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
	return &queueService{
		taskRepository:       taskRepository,
		submissionRepository: submissionRepository,
		queueRepository:      queueMessageRepository,
		queue:                q,
		channel:              channel,
		responseQueueName:    responseQueueName,
		logger:               log,
	}, nil
}
