package schemas

import (
	"time"

	"github.com/mini-maxit/backend/package/domain/types"
)

// User represents the user.
type User struct {
	ID        int64          `json:"id"`
	Name      string         `json:"name"`
	Surname   string         `json:"surname"`
	Email     string         `json:"email"`
	Username  string         `json:"username"`
	Role      types.UserRole `json:"role"`
	CreatedAt time.Time      `json:"createdAt"`
}

// UserCreate represents the user create request.
type UserEdit struct {
	Name     *string         `json:"name,omitempty"`
	Surname  *string         `json:"surname,omitempty"`
	Email    *string         `json:"email,omitempty"`
	Username *string         `json:"username,omitempty"`
	Role     *types.UserRole `json:"role,omitempty"`
}

// UserIDs represents the user IDs request.
type UserIDs struct {
	UserIDs []int64 `json:"userIDs"`
}

// UserChangePassword represents the user change password request.
type UserChangePassword struct {
	OldPassword        string `json:"oldPassword"`
	NewPassword        string `json:"newPassword" validate:"required,password,gte=8,lte=50"`
	NewPasswordConfirm string `json:"newPasswordConfirm" validate:"required,eqfield=NewPassword,gte=8,lte=50"`
}

type UsersRequest struct {
	UserIDs []int64 `json:"userIDs"`
}
