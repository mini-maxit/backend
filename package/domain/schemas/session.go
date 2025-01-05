package schemas

import "time"

type Session struct {
	Id        string    `json:"session"`
	UserId    int64     `json:"user_id"`
	UserRole  string    `json:"user_role"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ValidateSessionResponse struct {
	Valid  bool  `json:"valid"`
	User UserSession `json:"user"`
}
