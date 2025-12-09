package models

import "github.com/mini-maxit/backend/package/domain/types"

// AccessControl represents user access permissions to various resources.
// This unified model handles access control for both contests and tasks.
type AccessControl struct {
	ResourceID   int64              `gorm:"primaryKey;not null"`
	ResourceType types.ResourceType `gorm:"primaryKey;type:varchar(20);not null"`
	UserID       int64              `gorm:"primaryKey;not null"`
	Permission   types.Permission   `gorm:"type:varchar(20);not null"`
	BaseModel

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}
