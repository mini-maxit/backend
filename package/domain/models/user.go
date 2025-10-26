package models

import (
	"github.com/mini-maxit/backend/package/domain/types"
)

type User struct {
	ID           int64          `gorm:"primaryKey;autoIncrement"`
	Name         string         `gorm:"NOT NULL"`
	Surname      string         `gorm:"NOT NULL"`
	Email        string         `gorm:"NOT NULL;UNIQUE"`
	Username     string         `gorm:"NOT NULL;UNIQUE"`
	PasswordHash string         `gorm:"NOT NULL"`
	Role         types.UserRole `gorm:"NOT NULL;default:'student'"` // student, teacher, admin
	BaseModel
}
