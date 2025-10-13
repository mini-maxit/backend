package models

import "time"

type Contest struct {
	Id          int64     `gorm:"primaryKey;autoIncrement"`
	Title       string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:text"`
	StartTime   time.Time `gorm:"type:timestamp"`
	EndTime     time.Time `gorm:"type:timestamp"`
	CreatedBy   int64     `gorm:"not null"`
	Author      User      `gorm:"foreignKey:CreatedBy;references:Id"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
	Tasks       []Task    `gorm:"many2many:contest_tasks;"`
}

type ContestTask struct {
	ContestId int64 `gorm:"primaryKey"`
	TaskId    int64 `gorm:"primaryKey"`
}
