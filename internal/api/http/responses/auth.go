package responses

import (
	"net/http"
	"time"

	"github.com/mini-maxit/backend/package/domain/schemas"
)

// AuthResponse represents the response for auth endpoints (excludes refresh token for security)
type AuthResponse struct {
	AccessToken string    `json:"accessToken"`
	ExpiresAt   time.Time `json:"expiresAt"`
	TokenType   string    `json:"tokenType"`
}

// NewAuthResponse creates an AuthResponse from JWTTokens, excluding the refresh token
func NewAuthResponse(tokens *schemas.JWTTokens) AuthResponse {
	return AuthResponse{
		AccessToken: tokens.AccessToken,
		ExpiresAt:   tokens.ExpiresAt,
		TokenType:   tokens.TokenType,
	}
}

// SetRefreshTokenCookie sets the refresh token as an httpOnly cookie
func SetRefreshTokenCookie(w http.ResponseWriter, path, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     path,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})
}
