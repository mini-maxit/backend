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
	// PermissionView represents view-only access.
	PermissionView Permission = "view"
	// PermissionEdit represents edit access (view + edit).
	PermissionEdit Permission = "edit"
	// PermissionManage represents full access (view + edit + delete + manage collaborators).
	PermissionManage Permission = "manage"
)

// Scan implements the sql.Scanner interface.
func (p *Permission) Scan(value any) error {
	valueString, ok := value.(string)
	if !ok {
		return errors.New("Permission must be a string")
	}
	permission := Permission(valueString)
	availablePermissions := []Permission{PermissionView, PermissionEdit, PermissionManage}
	if !slices.Contains(availablePermissions, permission) {
		return fmt.Errorf("invalid Permission: %s. Available permissions: %v", permission, availablePermissions)
	}
	*p = permission
	return nil
}

// HasPermission checks if the current permission level includes the required permission.
// Manage > Edit > View
func (p Permission) HasPermission(required Permission) bool {
	permissionLevels := map[Permission]int{
		PermissionView:   1,
		PermissionEdit:   2,
		PermissionManage: 3,
	}
	return permissionLevels[p] >= permissionLevels[required]
}
