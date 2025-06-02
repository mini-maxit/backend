package middleware

import (
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"
)

// RecoveryMiddleware recovers from panics and returns a 500 error.
func RecoveryMiddleware(next http.Handler, log *zap.SugaredLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Errorf("Panic recovered: %v", rec)
				log.Error(string(debug.Stack()))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}
