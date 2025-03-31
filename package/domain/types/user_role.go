package types

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"slices"
)

type UserRole string

func (ur UserRole) Value() (driver.Value, error) {
	if ur == "" {
		return nil, errors.New("user role is empty")
	}
	return string(ur), nil
}

const (
	UserRoleStudent UserRole = "student"
	UserRoleTeacher UserRole = "teacher"
	UserRoleAdmin   UserRole = "admin"
)

func (ur *UserRole) Scan(value any) error {
	valueString, ok := value.(string)
	if !ok {
		return errors.New("UserRole must be a string")
	}
	role := UserRole(valueString)
	availableRoles := []UserRole{UserRoleStudent, UserRoleTeacher, UserRoleAdmin}
	if !slices.Contains(availableRoles, role) {
		return fmt.Errorf("invalid UserRole: %s. Available roles: %s", role, availableRoles)
	}
	*ur = role
	return nil
}
