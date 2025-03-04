package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type QueueMessageRepository interface {
	// CreateQueueMessage creates a new queue message and returns the message ID
	CreateQueueMessage(tx *gorm.DB, queueMessage *models.QueueMessage) (string, error)
	GetQueueMessage(tx *gorm.DB, messageId string) (*models.QueueMessage, error)
	DeleteQueueMessage(tx *gorm.DB, messageId string) error
}

type queueMessageRepository struct {
}

func (qm *queueMessageRepository) CreateQueueMessage(tx *gorm.DB, queueMessage *models.QueueMessage) (string, error) {
	err := tx.Create(queueMessage).Error
	if err != nil {
		return "", err
	}
	return queueMessage.Id, nil
}

func (qm *queueMessageRepository) GetQueueMessage(tx *gorm.DB, messageId string) (*models.QueueMessage, error) {
	queueMessage := &models.QueueMessage{}
	err := tx.Model(&models.QueueMessage{}).Where("id = ?", messageId).First(queueMessage).Error
	if err != nil {
		return nil, err
	}
	return queueMessage, nil
}

func (qm *queueMessageRepository) DeleteQueueMessage(tx *gorm.DB, messageId string) error {
	err := tx.Model(&models.QueueMessage{}).Where("id = ?", messageId).Delete(&models.QueueMessage{}).Error
	if err != nil {
		return err
	}
	return nil
}

func NewQueueMessageRepository(db *gorm.DB) (QueueMessageRepository, error) {
	if !db.Migrator().HasTable(&models.QueueMessage{}) {
		err := db.Migrator().CreateTable(&models.QueueMessage{})
		if err != nil {
			return nil, err
		}
	}

	return &queueMessageRepository{}, nil
}
