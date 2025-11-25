package repository

import (
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
)

type QueueMessageRepository interface {
	// Create creates a new queue message and returns the message ID
	Create(db database.Database, queueMessage *models.QueueMessage) (string, error)
	// Delete deletes a queue message by ID
	Delete(db database.Database, messageID string) error
	// Get returns a queue message by ID
	Get(db database.Database, messageID string) (*models.QueueMessage, error)
}

type queueMessageRepository struct {
}

func (qm *queueMessageRepository) Create(db database.Database, queueMessage *models.QueueMessage) (string, error) {
	tx := db.GetInstance()
	err := tx.Create(queueMessage).Error
	if err != nil {
		return "", err
	}
	return queueMessage.ID, nil
}

func (qm *queueMessageRepository) Get(db database.Database, messageID string) (*models.QueueMessage, error) {
	tx := db.GetInstance()
	queueMessage := &models.QueueMessage{}
	err := tx.Model(&models.QueueMessage{}).Where("id = ?", messageID).First(queueMessage).Error
	if err != nil {
		return nil, err
	}
	return queueMessage, nil
}

func (qm *queueMessageRepository) Delete(db database.Database, messageID string) error {
	tx := db.GetInstance()
	err := tx.Model(&models.QueueMessage{}).Where("id = ?", messageID).Delete(&models.QueueMessage{}).Error
	if err != nil {
		return err
	}
	return nil
}

func NewQueueMessageRepository() QueueMessageRepository {
	return &queueMessageRepository{}
}
