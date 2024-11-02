package models

type QueueMessage struct {
	Id           string       `gorm:"primaryKey;"`
	SubmissionId int64        `gorm:"not null;"`
	UserSolution UserSolution `gorm:"foreignKey:SubmissionId;references:Id"`
}
