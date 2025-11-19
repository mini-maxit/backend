package models

import "github.com/mini-maxit/backend/package/domain/types"

// ResourceType represents the type of resource in access control.
type ResourceType string

const (
	// ResourceTypeContest represents a contest resource.
	ResourceTypeContest ResourceType = "contest"
	// ResourceTypeTask represents a task resource.
	ResourceTypeTask ResourceType = "task"
)

// AccessControl represents user access permissions to various resources.
// This unified model handles access control for both contests and tasks.
type AccessControl struct {
	ResourceID   int64            `gorm:"primaryKey;not null"`
	ResourceType ResourceType     `gorm:"primaryKey;type:varchar(20);not null"`
	UserID       int64            `gorm:"primaryKey;not null"`
	Permission   types.Permission `gorm:"type:varchar(20);not null"`
	BaseModel

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}

// TableName overrides the table name to be access_control
func (AccessControl) TableName() string {
	return "access_control"
}
