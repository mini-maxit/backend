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
	PublishSubmission(tx *gorm.DB, submissionId int64) error
	GetSubmissionId(tx *gorm.DB, messageId string) (int64, error)
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
	submission, err := qs.submissionRepository.GetSubmission(tx, submissionId)
	if err != nil {
		qs.logger.Errorf("Error getting submission: %v", err.Error())
		return err
	}

	timeLimits, err := qs.taskRepository.GetTaskTimeLimits(tx, submission.TaskId)
	if err != nil {
		qs.logger.Errorf("Error getting task time limits: %v", err.Error())
		return err
	}
	memoryLimits, err := qs.taskRepository.GetTaskMemoryLimits(tx, submission.TaskId)
	if err != nil {
		qs.logger.Errorf("Error getting task memory limits: %v", err.Error())
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
	qm := &models.QueueMessage{
		Id:           msq.MessageId,
		SubmissionId: submissionId,
	}
	_, err = qs.queueRepository.CreateQueueMessage(tx, qm)
	if err != nil {
		qs.logger.Errorf("Error creating queue message: %v", err.Error())
		return err
	}
	err = qs.publishMessage(msq)
	if err != nil {
		err2 := qs.submissionRepository.MarkSubmissionFailed(tx, submissionId, err.Error())
		if err2 != nil {
			qs.logger.Errorf("Error marking submission failed: %v. When error occured publishing message: %s", err2.Error(), err.Error())
			return err
		}
		qs.logger.Errorf("Error publishing message: %v", err.Error())
		return err
	}
	err = qs.submissionRepository.MarkSubmissionProcessing(tx, submissionId)
	if err != nil {
		qs.logger.Errorf("Error marking submission processing: %v", err.Error())
		return err
	}
	qs.logger.Info("Submission published")
	return nil
}

func (qs *queueService) GetSubmissionId(tx *gorm.DB, messageId string) (int64, error) {
	queueMessage, err := qs.queueRepository.GetQueueMessage(tx, messageId)
	if err != nil {
		return 0, err
	}

	return queueMessage.SubmissionId, nil
}

func NewQueueService(taskRepository repository.TaskRepository, submissionRepository repository.SubmissionRepository, queueMessageRepository repository.QueueMessageRepository, conn *amqp.Connection, channel *amqp.Channel, queueName string, responseQueueName string) (*queueService, error) {
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
