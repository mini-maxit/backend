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
		corsConfig     *config.CORSConfig
		expectedStatus int
		checkHeaders   map[string]string
	}{
		{
			name:   "OPTIONS request returns 204 with default config",
			method: http.MethodOptions,
			corsConfig: &config.CORSConfig{
				AllowedOrigins:   "*",
				AllowCredentials: true,
			},
			expectedStatus: http.StatusNoContent,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS, PATCH",
				"Access-Control-Allow-Headers":     "Content-Type, Authorization, X-Requested-With",
				"Access-Control-Expose-Headers":    "Content-Length, Content-Type",
				"Access-Control-Allow-Credentials": "true",
				"Access-Control-Max-Age":           "86400",
			},
		},
		{
			name:   "GET request passes through with custom origins",
			method: http.MethodGet,
			corsConfig: &config.CORSConfig{
				AllowedOrigins:   "http://localhost:3000,http://localhost:5173",
				AllowCredentials: false,
			},
			expectedStatus: http.StatusOK,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://localhost:3000,http://localhost:5173",
				"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS, PATCH",
				"Access-Control-Allow-Headers":     "Content-Type, Authorization, X-Requested-With",
				"Access-Control-Allow-Credentials": "false",
			},
		},
		{
			name:   "POST request with credentials enabled",
			method: http.MethodPost,
			corsConfig: &config.CORSConfig{
				AllowedOrigins:   "https://example.com",
				AllowCredentials: true,
			},
			expectedStatus: http.StatusOK,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "https://example.com",
				"Access-Control-Allow-Credentials": "true",
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
