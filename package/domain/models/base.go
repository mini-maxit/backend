package models

import (
	"time"

	"gorm.io/gorm"
)

type baseModel struct {
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index;default:null"`
}
