package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/mini-maxit/backend/internal/logger"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

type QueueService interface {
	// PublishTask publishes a task to the queue
	PublishSubmission(tx *gorm.DB, submissionId int64) error
	GetSubmissionId(tx *gorm.DB, messageId string) (int64, error)
}

type QueueServiceImpl struct {
	taskRepository       repository.TaskRepository
	submissionRepository repository.SubmissionRepository
	queueRepository      repository.QueueMessageRepository
	channel              *amqp.Channel
	queue                amqp.Queue
	responseQueueName    string
	service_logger       *logger.ServiceLogger
}

func (qs *QueueServiceImpl) publishMessage(msq schemas.QueueMessage) error {
	msgBytes, err := json.Marshal(msq)
	if err != nil {
		logger.Log(qs.service_logger, "Error marshalling message:", err.Error(), logger.Error)
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
		logger.Log(qs.service_logger, "Error publishing message:", err.Error(), logger.Error)
		return err
	}

	logger.Log(qs.service_logger, "Message published", "", logger.Info)
	return nil
}

func (qs *QueueServiceImpl) PublishSubmission(tx *gorm.DB, submissionId int64) error {
	submission, err := qs.submissionRepository.GetSubmission(tx, submissionId)
	if err != nil {
		logger.Log(qs.service_logger, "Error getting submission:", err.Error(), logger.Error)
		return err
	}

	timeLimits, err := qs.taskRepository.GetTaskTimeLimits(tx, submission.TaskId)
	if err != nil {
		logger.Log(qs.service_logger, "Error getting task time limits:", err.Error(), logger.Error)
		return err
	}
	memoryLimits, err := qs.taskRepository.GetTaskMemoryLimits(tx, submission.TaskId)
	if err != nil {
		logger.Log(qs.service_logger, "Error getting task memory limits:", err.Error(), logger.Error)
		return err
	}

	msq := schemas.QueueMessage{
		MessageId:       uuid.New().String(),
		TaskId:          submission.TaskId,
		UserId:          submission.UserId,
		SumissionNumber: submission.Order,
		LanguageType:    string(submission.Language.Type),
		LanguageVersion: submission.Language.Version,
		TimeLimits:      timeLimits,
		MemoryLimits:    memoryLimits,
	}
	qs.queueRepository.CreateQueueMessage(tx, models.QueueMessage{
		Id:           msq.MessageId,
		SubmissionId: submissionId,
	})
	err = qs.publishMessage(msq)
	if err != nil {
		qs.submissionRepository.MarkSubmissionFailed(tx, submissionId, err.Error())
		logger.Log(qs.service_logger, "Error publishing message:", err.Error(), logger.Error)
		return err
	}
	err = qs.submissionRepository.MarkSubmissionProcessing(tx, submissionId)
	if err != nil {

		logger.Log(qs.service_logger, "Error marking submission processing:", err.Error(), logger.Error)
		return err
	}

	logger.Log(qs.service_logger, "Submission published", "", logger.Info)
	return nil
}

func (qs *QueueServiceImpl) GetSubmissionId(tx *gorm.DB, messageId string) (int64, error) {
	queueMessage, err := qs.queueRepository.GetQueueMessage(tx, messageId)
	if err != nil {
		return 0, err
	}

	logger.Log(qs.service_logger, "Submission id retrieved", "", logger.Info)
	return queueMessage.SubmissionId, nil
}

func NewQueueService(taskRepository repository.TaskRepository, submissionRepository repository.SubmissionRepository, queueMessageRepository repository.QueueMessageRepository, conn *amqp.Connection, channel *amqp.Channel, queueName string, responseQueueName string) (*QueueServiceImpl, error) {
	q, err := channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, err
	}
	service_logger := logger.NewNamedLogger("queue_service")
	return &QueueServiceImpl{
		taskRepository:       taskRepository,
		submissionRepository: submissionRepository,
		queueRepository:      queueMessageRepository,
		queue:                q,
		channel:              channel,
		responseQueueName:    responseQueueName,
		service_logger:       &service_logger,
	}, nil
}
