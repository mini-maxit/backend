package middleware

import (
	"context"
	"net/http"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/service"
)

func SessionValidationMiddleware(next http.Handler, db database.Database, sessionService service.SessionService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionHeader := r.Header.Get("Session")
		if sessionHeader == "" {
			httputils.ReturnError(w, http.StatusUnauthorized, "Session header is not set, could not authorize")
			return
		}
		session := db.NewSession()
		tx, err := session.Connect()
		if err != nil {
			httputils.ReturnError(w, http.StatusInternalServerError, "Failed to start transaction. "+err.Error())
			return
		}
		sessionResponse, err := sessionService.ValidateSession(tx, sessionHeader)
		if err != nil {
			if err == service.ErrSessionNotFound {
				httputils.ReturnError(w, http.StatusUnauthorized, "Session not found")
				return
			}
			if err == service.ErrSessionExpired {
				httputils.ReturnError(w, http.StatusUnauthorized, "Session expired")
				return
			}
			httputils.ReturnError(w, http.StatusInternalServerError, "Failed to validate session. "+err.Error())
			return
		}
		tx.Rollback()

		ctx := r.Context()
		ctx = context.WithValue(ctx, SessionKey, sessionHeader)
		ctx = context.WithValue(ctx, UserKey, sessionResponse.User)
		rWithSession := r.WithContext(ctx)

		next.ServeHTTP(w, rWithSession)
	})
}
