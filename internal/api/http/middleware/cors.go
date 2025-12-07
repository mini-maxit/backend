package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/config"
	"github.com/mini-maxit/backend/package/errors"
)

// CORSMiddleware handles Cross-Origin Resource Sharing (CORS) headers
func CORSMiddleware(next http.Handler, cfg *config.CORSConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Determine allowed origin
		allowedOrigin := ""
		if cfg.AllowedOrigins == "*" {
			allowedOrigin = "*"
		} else if origin != "" {
			// Check if the origin is in the allowed list
			allowedOrigins := strings.SplitSeq(cfg.AllowedOrigins, ",")
			for allowed := range allowedOrigins {
				if strings.TrimSpace(allowed) == origin {
					allowedOrigin = origin
					break
				}
			}
			if allowedOrigin == "" {
				httputils.ReturnServiceError(w, errors.ErrCORSNotAllowed)
				return
			}
		} else {
			next.ServeHTTP(w, r)
			return
		}

		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", strconv.FormatBool(cfg.AllowCredentials))

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// For non-preflight requests, expose headers
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		// Pass to next handler
		next.ServeHTTP(w, r)
	})
}
