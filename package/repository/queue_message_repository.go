package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type QueueMessageRepository interface {
	// Create creates a new queue message and returns the message ID
	Create(tx *gorm.DB, queueMessage *models.QueueMessage) (string, error)
	// Delete deletes a queue message by ID
	Delete(tx *gorm.DB, messageID string) error
	// Get returns a queue message by ID
	Get(tx *gorm.DB, messageID string) (*models.QueueMessage, error)
}

type queueMessageRepository struct {
}

func (qm *queueMessageRepository) Create(tx *gorm.DB, queueMessage *models.QueueMessage) (string, error) {
	err := tx.Create(queueMessage).Error
	if err != nil {
		return "", err
	}
	return queueMessage.ID, nil
}

func (qm *queueMessageRepository) Get(tx *gorm.DB, messageID string) (*models.QueueMessage, error) {
	queueMessage := &models.QueueMessage{}
	err := tx.Model(&models.QueueMessage{}).Where("id = ?", messageID).First(queueMessage).Error
	if err != nil {
		return nil, err
	}
	return queueMessage, nil
}

func (qm *queueMessageRepository) Delete(tx *gorm.DB, messageID string) error {
	err := tx.Model(&models.QueueMessage{}).Where("id = ?", messageID).Delete(&models.QueueMessage{}).Error
	if err != nil {
		return err
	}
	return nil
}

func NewQueueMessageRepository() QueueMessageRepository {
	return &queueMessageRepository{}
}
