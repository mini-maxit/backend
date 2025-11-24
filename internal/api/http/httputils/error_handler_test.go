//nolint:testpackage // Testing internal function errorCodeToHTTPStatus
package httputils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mini-maxit/backend/package/errors"
	"go.uber.org/zap"
)

func TestErrorCodeToHTTPStatus(t *testing.T) {
	tests := []struct {
		name           string
		err            *errors.ServiceError
		expectedStatus int
	}{
		{
			name:           "ErrDatabaseConnection",
			err:            errors.ErrDatabaseConnection,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "ErrCannotAssignOwner",
			err:            errors.ErrCannotAssignOwner,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpStatus := errorCodeToHTTPStatus(tt.err.Code)
			if httpStatus != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, httpStatus)
			}
		})
	}
}

func TestHandleServiceError(t *testing.T) {
	t.Run("handles error and writes response with error code", func(t *testing.T) {
		w := httptest.NewRecorder()
		logger := zap.NewNop().Sugar()

		HandleServiceError(w, errors.ErrUserNotFound, nil, logger)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
		}

		// Verify the response contains the error code
		var response APIError
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Ok {
			t.Error("Expected ok to be false")
		}
		if response.Data.Code != string(errors.CodeUserNotFound) {
			t.Errorf("Expected code %s, got %s", errors.CodeUserNotFound, response.Data.Code)
		}
		if response.Data.Message != "User not found" {
			t.Errorf("Expected message 'User not found', got %s", response.Data.Message)
		}
	})

	t.Run("does nothing when error is nil", func(t *testing.T) {
		w := httptest.NewRecorder()
		logger := zap.NewNop().Sugar()

		HandleServiceError(w, nil, nil, logger)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d (default), got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("handles unknown error with 500 and internal error code", func(t *testing.T) {
		w := httptest.NewRecorder()
		logger := zap.NewNop().Sugar()
		unknownErr := http.ErrServerClosed

		HandleServiceError(w, unknownErr, nil, logger)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
		}

		// Verify the response contains the internal error code
		var response APIError
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Data.Code != string(errors.CodeInternalError) {
			t.Errorf("Expected code %s, got %s", errors.CodeInternalError, response.Data.Code)
		}
	})

	t.Run("handles ServiceError directly", func(t *testing.T) {
		w := httptest.NewRecorder()
		logger := zap.NewNop().Sugar()

		// Create a ServiceError directly
		serviceErr := errors.ErrForbidden

		HandleServiceError(w, serviceErr, nil, logger)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
		}

		var response APIError
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Data.Code != string(errors.CodeForbidden) {
			t.Errorf("Expected code %s, got %s", errors.CodeForbidden, response.Data.Code)
		}
	})
}
