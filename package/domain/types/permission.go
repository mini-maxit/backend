package types

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"slices"
)

// Permission represents the access level for collaborators.
type Permission string

// Value implements the driver.Valuer interface.
func (p Permission) Value() (driver.Value, error) {
	if p == "" {
		return nil, errors.New("permission is empty")
	}
	return string(p), nil
}

const (
	// PermissionEdit represents edit access.
	PermissionEdit Permission = "edit"
	// PermissionManage represents full access (edit + manage collaborators).
	PermissionManage Permission = "manage"
	// PermissionOwner represents immutable ownership (highest level).
	PermissionOwner Permission = "owner"
)

// Scan implements the sql.Scanner interface.
func (p *Permission) Scan(value any) error {
	valueString, ok := value.(string)
	if !ok {
		return errors.New("Permission must be a string")
	}
	permission := Permission(valueString)
	availablePermissions := []Permission{PermissionEdit, PermissionManage, PermissionOwner}
	if !slices.Contains(availablePermissions, permission) {
		return fmt.Errorf("invalid Permission: %s. Available permissions: %v", permission, availablePermissions)
	}
	*p = permission
	return nil
}

// HasPermission checks if the current permission level includes the required permission.
// Owner > Manage > Edit
func (p Permission) HasPermission(required Permission) bool {
	permissionLevels := map[Permission]int{
		PermissionEdit:   1,
		PermissionManage: 2,
		PermissionOwner:  3,
	}
	return permissionLevels[p] >= permissionLevels[required]
}
