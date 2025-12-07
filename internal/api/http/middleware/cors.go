package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/mini-maxit/backend/internal/config"
)

// CORSMiddleware handles Cross-Origin Resource Sharing (CORS) headers
func CORSMiddleware(next http.Handler, cfg *config.CORSConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// Determine allowed origin
		allowedOrigin := cfg.AllowedOrigins
		if cfg.AllowedOrigins != "*" && origin != "" {
			// Check if the origin is in the allowed list
			allowedOrigins := strings.Split(cfg.AllowedOrigins, ",")
			for _, allowed := range allowedOrigins {
				if strings.TrimSpace(allowed) == origin {
					allowedOrigin = origin
					break
				}
			}
		}
		
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", strconv.FormatBool(cfg.AllowCredentials))
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Pass to next handler
		next.ServeHTTP(w, r)
	})
}
