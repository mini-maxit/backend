package middleware

import (
	"context"
	"net/http"

	"github.com/mini-maxit/backend/internal/api/http/utils"
	"github.com/mini-maxit/backend/package/service"
)

type ContextKey string

func SessionValidationMiddleware(next http.Handler, sessionService service.SessionService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionHeader := r.Header.Get("Session")
		if sessionHeader == "" {
			utils.ReturnError(w, http.StatusUnauthorized, utils.CodeUnauthorized, "Session header is not set, could not authorize")
			return
		}
		sessionResponse, err := sessionService.ValidateSession(sessionHeader)
		if err != nil {
			if err == service.ErrSessionNotFound {
				utils.ReturnError(w, http.StatusUnauthorized, utils.CodeUnauthorized, "Session not found")
				return
			}
			if err == service.ErrSessionExpired {
				utils.ReturnError(w, http.StatusUnauthorized, utils.CodeUnauthorized, "Session expired")
				return
			}
			utils.ReturnError(w, http.StatusInternalServerError, utils.CodeInternalServerError, "Failed to validate session. "+err.Error())
			return
		}

		ctx := context.WithValue(context.Background(), ContextKey("session"), sessionHeader)
		ctx = context.WithValue(ctx, ContextKey("userId"), sessionResponse.UserId)
		rWithSession := r.WithContext(ctx)

		next.ServeHTTP(w, rWithSession)
	})
}
