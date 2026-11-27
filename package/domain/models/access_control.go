package models

import "github.com/mini-maxit/backend/package/domain/types"

// AccessControl represents user access permissions to various resources.
// This unified model handles access control for both contests and tasks.
type AccessControl struct {
	// idx_one_owner: at most one live owner per resource.
	ResourceID   int64              `gorm:"primaryKey;not null;uniqueIndex:idx_one_owner,priority:2,where:permission = 'owner' AND deleted_at IS NULL"`
	ResourceType types.ResourceType `gorm:"primaryKey;type:varchar(20);not null;uniqueIndex:idx_one_owner,priority:1"`
	UserID       int64              `gorm:"primaryKey;not null"`
	Permission   types.Permission   `gorm:"type:varchar(20);not null"`
	BaseModel

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}
