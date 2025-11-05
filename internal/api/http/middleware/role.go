package middleware

import (
	"net/http"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
)

// RequireMinimalRole returns a middleware that checks if the authenticated user has the required role.
// This middleware should be used after JWTValidationMiddleware, as it expects the user to be in the context.
func RequireMinimalRole(handler http.Handler, minimalRole types.UserRole) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userVal := r.Context().Value(httputils.UserKey)
		if userVal == nil {
			httputils.ReturnError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		user, ok := userVal.(schemas.User)
		if !ok {
			httputils.ReturnError(w, http.StatusInternalServerError, "Invalid user context")
			return
		}

		if user.Role.HasAccess(minimalRole) {
			httputils.ReturnError(w, http.StatusForbidden, "Insufficient permissions for this resource")
			return
		}

		handler.ServeHTTP(w, r)
	})
}
