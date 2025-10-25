package models

import "time"

type QueueMessage struct {
	ID           string     `gorm:"primaryKey;"`
	SubmissionID int64      `gorm:"not null;"`
	QueuedAt     time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
	Submission   Submission `gorm:"foreignKey:SubmissionID;references:ID"`
}
