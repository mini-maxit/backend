package models

import "github.com/mini-maxit/backend/package/domain/types"

// ContestCollaborator represents a user who has been granted access to a contest.
type ContestCollaborator struct {
	ContestID  int64            `gorm:"primaryKey;not null"`
	UserID     int64            `gorm:"primaryKey;not null"`
	Permission types.Permission `gorm:"type:varchar(20);not null"`
	BaseModel

	Contest Contest `gorm:"foreignKey:ContestID;references:ID;constraint:OnDelete:CASCADE"`
	User    User    `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}

// TaskCollaborator represents a user who has been granted access to a task.
type TaskCollaborator struct {
	TaskID     int64            `gorm:"primaryKey;not null"`
	UserID     int64            `gorm:"primaryKey;not null"`
	Permission types.Permission `gorm:"type:varchar(20);not null"`
	BaseModel

	Task Task `gorm:"foreignKey:TaskID;references:ID;constraint:OnDelete:CASCADE"`
	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}
