package schemas

import "github.com/mini-maxit/backend/package/domain/types"

type User struct {
	Id       int64          `json:"id"`
	Name     string         `json:"name"`
	Surname  string         `json:"surname"`
	Email    string         `json:"email"`
	Username string         `json:"username"`
	Role     types.UserRole `json:"role"`
}

type UserEdit struct {
	Name     *string         `json:"name,omitempty"`
	Surname  *string         `json:"surname,omitempty"`
	Email    *string         `json:"email,omitempty"`
	Username *string         `json:"username,omitempty"`
	Role     *types.UserRole `json:"role,omitempty"`
}
