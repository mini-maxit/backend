package schemas

import (
	"time"

	"github.com/mini-maxit/backend/package/domain/models"
)

type Session struct {
	Id        string          `json:"session"`
	UserId    int64           `json:"user_id"`
	UserRole  models.UserRole `json:"user_role"`
	ExpiresAt time.Time       `json:"expires_at"`
}

type ValidateSessionResponse struct {
	Valid  bool  `json:"valid"`
	UserId int64 `json:"user_id"`
}
