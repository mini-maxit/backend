package models

import "time"

type Session struct {
	Id        string `gorm:"primaryKey"`
	UserId    int64
	ExpiresAt time.Time `gorm:"autoUpdateTime:false"`
	User      User      `gorm:"foreignKey:UserId"`
}
