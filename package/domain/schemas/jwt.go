package schemas

import "time"

// JWTTokens represents the response containing access and refresh tokens
type JWTTokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// JWTClaims represents the claims stored in JWT tokens
type JWTClaims struct {
	UserID   int64  `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TokenID  string `json:"token_id"`
	Type     string `json:"type"` // "access" or "refresh"
}

// RefreshTokenRequest represents the request to refresh tokens
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ValidateTokenResponse represents the response from token validation
type ValidateTokenResponse struct {
	Valid bool `json:"valid"`
	User  User `json:"user"`
}
