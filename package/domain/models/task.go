package models

import "time"

type Task struct {
	Id        int64     `gorm:"primaryKey;autoIncrement"`
	Title     string    `gorm:"type:varchar(255)"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	CreatedBy int64     `gorm:"foreignKey:UserID"`
	Author    User      `gorm:"foreignKey:CreatedBy; references:Id"`
}

type TaskUser struct {
	TaskId int64 `gorm:"primaryKey"`
	UserId int64 `gorm:"primaryKey"`
}
