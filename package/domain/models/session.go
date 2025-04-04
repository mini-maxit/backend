package models

import "time"

type Session struct {
	ID        string `gorm:"primaryKey"`
	UserID    int64
	ExpiresAt time.Time `gorm:"autoUpdateTime:false"`
	User      User      `gorm:"foreignKey:UserID"`
}
