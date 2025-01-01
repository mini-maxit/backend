package middleware

import (
	"log"
	"net/http"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/package/domain/models"
)

func RBACMiddleware(next http.Handler, minRole models.UserRole) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole := r.Context().Value(UserRoleKey).(models.UserRole)
		log.Printf("User role: %d, Min role: %d", userRole, minRole)
		if userRole < minRole {
			httputils.ReturnError(w, http.StatusForbidden, "User is not authorized to access this resource")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RBACHandle(handle func(w http.ResponseWriter, r *http.Request), minRole models.UserRole) http.Handler {
	return RBACMiddleware(http.HandlerFunc(handle), minRole)
}
