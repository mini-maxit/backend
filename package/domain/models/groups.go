package models

type Group struct {
	Id   int64  `gorm:"primaryKey;"`
	Name string `gorm:"not null;"`
}

type UserGroup struct {
	UserId  int64 `gorm:"primaryKey;"`
	GroupId int64 `gorm:"primaryKey;"`
}

type TaskGroup struct {
	TaskId  int64 `gorm:"primaryKey;"`
	GroupId int64 `gorm:"primaryKey;"`
}
