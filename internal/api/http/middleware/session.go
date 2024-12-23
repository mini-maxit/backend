package middleware

import (
	"context"
	"net/http"

	"github.com/mini-maxit/backend/internal/api/http/utils"
	"github.com/mini-maxit/backend/package/service"
)

func SessionValidationMiddleware(next http.Handler, sessionService service.SessionService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionHeader := r.Header.Get("Session")
		if sessionHeader == "" {
			utils.ReturnError(w, http.StatusUnauthorized, "Session header is not set, could not authorize")
			return
		}
		sessionResponse, err := sessionService.ValidateSession(sessionHeader)
		if err != nil {
			if err == service.ErrSessionNotFound {
				utils.ReturnError(w, http.StatusUnauthorized, "Session not found")
				return
			}
			if err == service.ErrSessionExpired {
				utils.ReturnError(w, http.StatusUnauthorized, "Session expired")
				return
			}
			utils.ReturnError(w, http.StatusInternalServerError, "Failed to validate session. "+err.Error())
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, SessionKey, sessionHeader)
		ctx = context.WithValue(ctx, UserIDKey, sessionResponse.UserId)
		rWithSession := r.WithContext(ctx)

		next.ServeHTTP(w, rWithSession)
	})
}
