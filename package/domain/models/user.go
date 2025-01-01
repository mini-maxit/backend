package models

import (
	"database/sql/driver"
	"fmt"
	"slices"
)

type User struct {
	Id           int64    `gorm:"primaryKey;autoIncrement"`
	Name         string   `gorm:"NOT NULL"`
	Surname      string   `gorm:"NOT NULL"`
	Email        string   `gorm:"NOT NULL;UNIQUE"`
	Username     string   `gorm:"NOT NULL;UNIQUE"`
	PasswordHash string   `gorm:"NOT NULL"`
	Role         UserRole `gorm:"NOT NULL;default:3"` // student=1, teacher=2, admin=3
}

type UserRole uint

func (ur *UserRole) Scan(value interface{}) error {
	valueInt, ok := value.(int64)
	if !ok {
		return fmt.Errorf("UserRole must be uint, got %T", value)
	}
	role := UserRole(valueInt)
	availableRoles := []UserRole{UserRoleStudent, UserRoleTeacher, UserRoleAdmin}
	if !slices.Contains(availableRoles, role) {
		return fmt.Errorf("invalid UserRole: %d Available roles: %v", role, availableRoles)
	}
	*ur = role
	return nil
}

func (ur UserRole) Value() (driver.Value, error) {
	if ur == 0 {
		return nil, nil
	}
	return uint(ur), nil
}

const (
	UserRoleStudent UserRole = iota + 1
	UserRoleTeacher
	UserRoleAdmin
)
