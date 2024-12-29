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
	Role         UserRole `gorm:"NOT NULL;default:'student'"` // student, teacher, admin
}

type UserRole string

func (ur *UserRole) Scan(value interface{}) error {
	valueString, ok := value.(string)
	if !ok {
		return fmt.Errorf("UserRole must be a string")
	}
	role := UserRole(valueString)
	availableRoles := []UserRole{UserRoleStudent, UserRoleTeacher, UserRoleAdmin}
	if !slices.Contains(availableRoles, role) {
		return fmt.Errorf("invalid UserRole: %s. Available roles: %s", role, availableRoles)
	}
	*ur = role
	return nil
}

func (ur UserRole) Value() (driver.Value, error) {
	if ur == "" {
		return nil, nil
	}
	return string(ur), nil
}

const (
	UserRoleStudent UserRole = "student"
	UserRoleTeacher UserRole = "teacher"
	UserRoleAdmin   UserRole = "admin"
)
