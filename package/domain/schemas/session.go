package schemas

import "time"

type Session struct {
	Id        string    `json:"id"`
	UserId    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ValidateSessionResponse struct {
	Valid  bool  `json:"valid"`
	UserId int64 `json:"user_id"`
}
