package repository

import (
	"github.com/mini-maxit/backend/package/domain/models"
	"gorm.io/gorm"
)

type QueueMessageRepository interface {
	// Create creates a new queue message and returns the message Id
	Create(tx *gorm.DB, queueMessage *models.QueueMessage) (string, error)
	// Delete deletes a queue message by Id
	Delete(tx *gorm.DB, messageId string) error
	// Get returns a queue message by Id
	Get(tx *gorm.DB, messageId string) (*models.QueueMessage, error)
}

type queueMessageRepository struct {
}

func (qm *queueMessageRepository) Create(tx *gorm.DB, queueMessage *models.QueueMessage) (string, error) {
	err := tx.Create(queueMessage).Error
	if err != nil {
		return "", err
	}
	return queueMessage.Id, nil
}

func (qm *queueMessageRepository) Get(tx *gorm.DB, messageId string) (*models.QueueMessage, error) {
	queueMessage := &models.QueueMessage{}
	err := tx.Model(&models.QueueMessage{}).Where("id = ?", messageId).First(queueMessage).Error
	if err != nil {
		return nil, err
	}
	return queueMessage, nil
}

func (qm *queueMessageRepository) Delete(tx *gorm.DB, messageId string) error {
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
