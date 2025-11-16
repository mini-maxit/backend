package schemas

import (
	"time"
)

// JWTTokens represents the response containing access and refresh tokens
type JWTTokens struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

// JWTClaims represents the claims stored in JWT tokens
type JWTClaims struct {
	UserID   int64  `json:"userId"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TokenID  string `json:"tokenId"`
	Type     string `json:"type"` // "access" or "refresh"
}

// RefreshTokenRequest represents the request to refresh tokens
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// ValidateTokenResponse represents the response from token validation
type ValidateTokenResponse struct {
	Valid bool `json:"valid"`
	User  User `json:"user"`
}
