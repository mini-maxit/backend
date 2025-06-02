package models

import "time"

type Group struct {
	ID        int64     `gorm:"primaryKey;"`
	Name      string    `gorm:"not null;"`
	CreatedBy int64     `gorm:"not null;"`
	Author    User      `gorm:"foreignKey:CreatedBy; references:ID"`
	CreatedAt time.Time `gorm:"autoCreateTime;"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;"`
	Tasks     []Task    `gorm:"many2many:task_groups;"`
	Users     []User    `gorm:"many2many:user_groups;"`
}

type UserGroup struct {
	UserID  int64 `gorm:"primaryKey;"`
	GroupID int64 `gorm:"primaryKey;"`
}

type TaskGroup struct {
	TaskID  int64 `gorm:"primaryKey;"`
	GroupID int64 `gorm:"primaryKey;"`
}
