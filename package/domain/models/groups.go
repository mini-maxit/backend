package models

import "time"

type Group struct {
	Id        int64     `gorm:"primaryKey;"`
	Name      string    `gorm:"not null;"`
	CreatedBy int64     `gorm:"not null;"`
	Author    User      `gorm:"foreignKey:CreatedBy; references:Id"`
	CreatedAt time.Time `gorm:"autoCreateTime;"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;"`
}

type UserGroup struct {
	UserId  int64 `gorm:"primaryKey;"`
	GroupId int64 `gorm:"primaryKey;"`
}

type TaskGroup struct {
	TaskId  int64 `gorm:"primaryKey;"`
	GroupId int64 `gorm:"primaryKey;"`
}
