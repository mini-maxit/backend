package models

import (
	"time"
)

type Task struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	Title     string `gorm:"type:varchar(255)"`
	CreatedBy int64  `gorm:"foreignKey:UserID"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	DeletedAt time.Time `gorm:"index;default:null"`

	Author User    `gorm:"foreignKey:CreatedBy; references:ID"`
	Groups []Group `gorm:"many2many:task_groups;"`
}

type TaskUser struct {
	TaskID int64 `gorm:"primaryKey"`
	UserID int64 `gorm:"primaryKey"`
}
