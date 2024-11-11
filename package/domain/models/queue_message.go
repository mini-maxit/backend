package models

import "time"

type QueueMessage struct {
	Id           string     `gorm:"primaryKey;"`
	SubmissionId int64      `gorm:"not null;"`
	QueuedAt     time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
	Submission   Submission `gorm:"foreignKey:SubmissionId;references:Id"`
}
