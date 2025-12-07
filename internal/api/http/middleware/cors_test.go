package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mini-maxit/backend/internal/config"
)

func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		origin         string
		corsConfig     *config.CORSConfig
		expectedStatus int
		checkHeaders   map[string]string
	}{
		{
			name:   "OPTIONS request returns 204 with wildcard origin",
			method: http.MethodOptions,
			origin: "http://example.com",
			corsConfig: &config.CORSConfig{
				AllowedOrigins:   "*",
				AllowCredentials: false,
			},
			expectedStatus: http.StatusNoContent,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS, PATCH",
				"Access-Control-Allow-Headers":     "Content-Type, Authorization, X-Requested-With",
				"Access-Control-Expose-Headers":    "Content-Length, Content-Type",
				"Access-Control-Allow-Credentials": "false",
				"Access-Control-Max-Age":           "86400",
			},
		},
		{
			name:   "GET request with specific allowed origin",
			method: http.MethodGet,
			origin: "http://localhost:3000",
			corsConfig: &config.CORSConfig{
				AllowedOrigins:   "http://localhost:3000,http://localhost:5173",
				AllowCredentials: true,
			},
			expectedStatus: http.StatusOK,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://localhost:3000",
				"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS, PATCH",
				"Access-Control-Allow-Headers":     "Content-Type, Authorization, X-Requested-With",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			name:   "POST request with credentials and matching origin",
			method: http.MethodPost,
			origin: "http://localhost:5173",
			corsConfig: &config.CORSConfig{
				AllowedOrigins:   "http://localhost:3000,http://localhost:5173",
				AllowCredentials: true,
			},
			expectedStatus: http.StatusOK,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://localhost:5173",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			name:   "Request from non-allowed origin",
			method: http.MethodGet,
			origin: "http://evil.com",
			corsConfig: &config.CORSConfig{
				AllowedOrigins:   "http://localhost:3000,http://localhost:5173",
				AllowCredentials: true,
			},
			expectedStatus: http.StatusOK,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin": "http://localhost:3000,http://localhost:5173",
			},
		},
		{
			name:   "Request with no origin header",
			method: http.MethodGet,
			origin: "",
			corsConfig: &config.CORSConfig{
				AllowedOrigins:   "http://localhost:3000",
				AllowCredentials: true,
			},
			expectedStatus: http.StatusOK,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin": "http://localhost:3000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that returns 200 OK
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Wrap with CORS middleware
			handler := CORSMiddleware(testHandler, tt.corsConfig)

			// Create test request
			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			w := httptest.NewRecorder()

			// Execute request
			handler.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check CORS headers
			for header, expected := range tt.checkHeaders {
				if got := w.Header().Get(header); got != expected {
					t.Errorf("Header %s: expected %q, got %q", header, expected, got)
				}
			}
		})
	}
}
