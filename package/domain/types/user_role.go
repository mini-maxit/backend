package types

import (
	"fmt"
	"slices"
)

type UserRole string

func (ur UserRole) String() string {
	return string(ur)
}

const (
	UserRoleStudent UserRole = "student"
	UserRoleTeacher UserRole = "teacher"
	UserRoleAdmin   UserRole = "admin"
)

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