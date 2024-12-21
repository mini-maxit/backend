package middleware

import (
	"net/http"

	"github.com/mini-maxit/backend/internal/logger"
)

// RecoveryMiddleware recovers from panics and returns a 500 error.
func RecoveryMiddleware(next http.Handler, log *logger.ServiceLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {

				logger.Log(log, "Panic recovered", rec.(string), logger.Error)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}
