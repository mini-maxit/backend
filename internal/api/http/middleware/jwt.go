package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/service"
)

func JWTValidationMiddleware(next http.Handler, db database.Database, jwtService service.JWTService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			httputils.ReturnError(w, http.StatusUnauthorized, "Authorization header is not set, could not authorize")
			return
		}

		// Check for Bearer token format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			httputils.ReturnError(w, http.StatusUnauthorized, "Invalid authorization header format. Expected 'Bearer <token>'")
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			httputils.ReturnError(w, http.StatusUnauthorized, "Token is empty")
			return
		}

		session := db.NewSession()
		_, err := session.BeginTransaction()
		if err != nil {
			httputils.ReturnError(w, http.StatusInternalServerError, "Failed to start transaction. "+err.Error())
			return
		}

		tokenResponse, err := jwtService.AuthenticateToken(session, token)
		if err != nil {
			session.Rollback()
			if errors.Is(err, errors.ErrInvalidToken) {
				httputils.ReturnError(w, http.StatusUnauthorized, "Invalid token")
				return
			} else if errors.Is(err, errors.ErrTokenExpired) {
				httputils.ReturnError(w, http.StatusUnauthorized, "Token expired")
				return
			} else if errors.Is(err, errors.ErrTokenUserNotFound) {
				httputils.ReturnError(w, http.StatusUnauthorized, "User associated with token not found")
				return
			} else if errors.Is(err, errors.ErrNotAuthorized) {
				httputils.ReturnError(w, http.StatusUnauthorized, "Malformed or outdated token")
				return
			}

			httputils.ReturnError(w, http.StatusInternalServerError, "Failed to validate token. "+err.Error())
			return
		}
		session.Rollback()

		ctx := r.Context()
		ctx = context.WithValue(ctx, httputils.TokenKey, token)
		ctx = context.WithValue(ctx, httputils.UserKey, tokenResponse.User)
		rWithAuth := r.WithContext(ctx)

		next.ServeHTTP(w, rWithAuth)
	})
}
