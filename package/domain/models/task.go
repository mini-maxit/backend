package models

import (
	"time"
)

type Task struct {
	Id        int64  `gorm:"primaryKey;autoIncrement"`
	Title     string `gorm:"type:varchar(255)"`
	CreatedBy int64  `gorm:"foreignKey:UserID"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	DeletedAt time.Time `gorm:"index;default:null"`

	Author User    `gorm:"foreignKey:CreatedBy; references:Id"`
	Groups []Group `gorm:"many2many:task_groups;"`
}

type TaskUser struct {
	TaskId int64 `gorm:"primaryKey"`
	UserId int64 `gorm:"primaryKey"`
}
