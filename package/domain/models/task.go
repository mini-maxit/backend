package models

import "github.com/mini-maxit/backend/package/domain/types"

type Task struct {
	ID                int64  `gorm:"primaryKey;autoIncrement"`
	Title             string `gorm:"type:varchar(255)"`
	DescriptionFileID int64  `gorm:"null;default:null"`
	CreatedBy         int64  `gorm:"foreignKey:UserID"`

	BaseModel

	Author          User `gorm:"foreignKey:CreatedBy; references:ID"`
	DescriptionFile File `gorm:"foreignKey:DescriptionFileID; references:ID"`
}

// ManagedTask is a task together with the current user's permission on it.
type ManagedTask struct {
	Task
	PermissionType types.Permission `gorm:"column:permission_type"`
}
