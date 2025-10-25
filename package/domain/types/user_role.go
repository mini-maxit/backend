package types

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"slices"
)

// UserRole represents the user role.
type UserRole string

// Value implements the driver.Valuer interface.
func (ur UserRole) Value() (driver.Value, error) {
	if ur == "" {
		return nil, errors.New("user role is empty")
	}
	return string(ur), nil
}

const (
	// UserRoleStudent represents the student role.
	UserRoleStudent UserRole = "student"
	// UserRoleTeacher represents the teacher role.
	UserRoleTeacher UserRole = "teacher"
	// UserRoleAdmin represents the admin role.
	UserRoleAdmin UserRole = "admin"
)

// Scan implements the sql.Scanner interface.
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
