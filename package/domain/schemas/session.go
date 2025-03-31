package schemas

import "time"

type Session struct {
	ID        string    `json:"session"`
	UserID    int64     `json:"user_id"`
	UserRole  string    `json:"user_role"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ValidateSessionResponse struct {
	Valid bool `json:"valid"`
	User  User `json:"user"`
}
