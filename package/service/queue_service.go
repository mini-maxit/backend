package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/repository"
	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueService interface {
	// PublishTask publishes a task to the queue
	PublishSubmission(submissionId int64) error
	GetSubmissionId(messageId string) (int64, error)
}

type QueueServiceImpl struct {
	database             database.Database
	taskRepository       repository.TaskRepository
	submissionRepository repository.SubmissionRepository
	queueRepository      repository.QueueMessageRepository
	channel              *amqp.Channel
	queue                amqp.Queue
	responseQueueName    string
}

func (qs *QueueServiceImpl) publishMessage(msq schemas.QueueMessage) error {
	msgBytes, err := json.Marshal(msq)
	if err != nil {
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
		return err
	}

	return nil
}

func (qs *QueueServiceImpl) PublishSubmission(submissionId int64) error {
	db := qs.database.Connect()
	tx := db.Begin()

	submission, err := qs.submissionRepository.GetSubmission(tx, submissionId)
	if err != nil {
		tx.Rollback()
		return err
	}

	timeLimits, err := qs.taskRepository.GetTaskTimeLimits(tx, submission.TaskId)
	if err != nil {
		tx.Rollback()
		return err
	}
	memoryLimits, err := qs.taskRepository.GetTaskMemoryLimits(tx, submission.TaskId)
	if err != nil {
		tx.Rollback()
		return err
	}

	msq := schemas.QueueMessage{
		MessageId:       uuid.New().String(),
		TaskId:          submission.TaskId,
		UserId:          submission.UserId,
		SumissionNumber: submission.Order,
		LanguageType:    submission.LanguageType,
		LanguageVersion: submission.LanguageVersion,
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
		tx.Rollback()
		return err
	}
	err = qs.submissionRepository.MarkSubmissionProcessing(tx, submissionId)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (qs *QueueServiceImpl) GetSubmissionId(messageId string) (int64, error) {
	db := qs.database.Connect()
	tx := db.Begin()

	queueMessage, err := qs.queueRepository.GetQueueMessage(tx, messageId)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	tx.Commit()
	return queueMessage.SubmissionId, nil
}

func NewQueueService(db database.Database, taskRepository repository.TaskRepository, submissionRepository repository.SubmissionRepository, queueMessageRepository repository.QueueMessageRepository, conn *amqp.Connection, channel *amqp.Channel, queueName string, responseQueueName string) (*QueueServiceImpl, error) {
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

	return &QueueServiceImpl{database: db,
		taskRepository:       taskRepository,
		submissionRepository: submissionRepository,
		queueRepository:      queueMessageRepository,
		queue:                q,
		channel:              channel,
		responseQueueName:    responseQueueName}, nil
}
