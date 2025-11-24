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
		err            error
		expectedStatus int
		expectedMsg    string
	}{
		// Database errors
		{
			name:           "ErrDatabaseConnection",
			err:            errors.ErrDatabaseConnection,
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Database connection error",
		},

		// Task errors
		{
			name:           "ErrTaskExists",
			err:            errors.ErrTaskExists,
			expectedStatus: http.StatusConflict,
			expectedMsg:    "Task with this title already exists",
		},
		{
			name:           "ErrTaskNotFound",
			err:            errors.ErrTaskNotFound,
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "Task not found",
		},
		{
			name:           "ErrTaskAlreadyAssigned",
			err:            errors.ErrTaskAlreadyAssigned,
			expectedStatus: http.StatusConflict,
			expectedMsg:    "Task is already assigned to the user",
		},

		// Authorization errors
		{
			name:           "ErrForbidden",
			err:            errors.ErrForbidden,
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "Not authorized to perform this action",
		},
		{
			name:           "ErrNotAuthorized",
			err:            errors.ErrNotAuthorized,
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Not authorized",
		},
		{
			name:           "ErrNotAllowed",
			err:            errors.ErrNotAllowed,
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "Not allowed to perform this action",
		},
		{
			name:           "ErrPermissionDenied",
			err:            errors.ErrPermissionDenied,
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "Permission denied",
		},

		// User errors
		{
			name:           "ErrUserNotFound",
			err:            errors.ErrUserNotFound,
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "User not found",
		},
		{
			name:           "ErrUserAlreadyExists",
			err:            errors.ErrUserAlreadyExists,
			expectedStatus: http.StatusConflict,
			expectedMsg:    "User already exists",
		},

		// Access control errors
		{
			name:           "ErrAccessAlreadyExists",
			err:            errors.ErrAccessAlreadyExists,
			expectedStatus: http.StatusConflict,
			expectedMsg:    "Access already exists",
		},

		// Authentication errors
		{
			name:           "ErrInvalidCredentials",
			err:            errors.ErrInvalidCredentials,
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Invalid credentials",
		},

		// Data validation errors
		{
			name:           "ErrInvalidData",
			err:            errors.ErrInvalidData,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid data",
		},

		// Generic not found
		{
			name:           "ErrNotFound",
			err:            errors.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "Requested resource not found",
		},

		// File operation errors
		{
			name:           "ErrFileOpen",
			err:            errors.ErrFileOpen,
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Failed to open file",
		},
		{
			name:           "ErrTempDirCreate",
			err:            errors.ErrTempDirCreate,
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Failed to create temp directory",
		},
		{
			name:           "ErrDecompressArchive",
			err:            errors.ErrDecompressArchive,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Failed to decompress archive",
		},
		{
			name:           "ErrNoInputDirectory",
			err:            errors.ErrNoInputDirectory,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "No input directory found",
		},
		{
			name:           "ErrNoOutputDirectory",
			err:            errors.ErrNoOutputDirectory,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "No output directory found",
		},
		{
			name:           "ErrIOCountMismatch",
			err:            errors.ErrIOCountMismatch,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Input and output file count mismatch",
		},
		{
			name:           "ErrInputContainsDir",
			err:            errors.ErrInputContainsDir,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Input contains a directory",
		},
		{
			name:           "ErrOutputContainsDir",
			err:            errors.ErrOutputContainsDir,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Output contains a directory",
		},
		{
			name:           "ErrInvalidInExtention",
			err:            errors.ErrInvalidInExtention,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid input file extension",
		},
		{
			name:           "ErrInvalidOutExtention",
			err:            errors.ErrInvalidOutExtention,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid output file extension",
		},

		// FileStorage errors
		{
			name:           "ErrWriteTaskID",
			err:            errors.ErrWriteTaskID,
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Error writing task ID to form",
		},
		{
			name:           "ErrWriteOverwrite",
			err:            errors.ErrWriteOverwrite,
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Error writing overwrite to form",
		},
		{
			name:           "ErrCreateFormFile",
			err:            errors.ErrCreateFormFile,
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Error creating form file",
		},
		{
			name:           "ErrCopyFile",
			err:            errors.ErrCopyFile,
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Error copying file to form",
		},
		{
			name:           "ErrSendRequest",
			err:            errors.ErrSendRequest,
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Error sending request to FileStorage",
		},
		{
			name:           "ErrReadResponse",
			err:            errors.ErrReadResponse,
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Error reading response from FileStorage",
		},
		{
			name:           "ErrResponseFromFileStorage",
			err:            errors.ErrResponseFromFileStorage,
			expectedStatus: http.StatusBadGateway,
			expectedMsg:    "Error response from FileStorage",
		},

		// Group errors
		{
			name:           "ErrGroupNotFound",
			err:            errors.ErrGroupNotFound,
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "Group not found",
		},

		// Pagination errors
		{
			name:           "ErrInvalidLimitParam",
			err:            errors.ErrInvalidLimitParam,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid limit parameter",
		},
		{
			name:           "ErrInvalidOffsetParam",
			err:            errors.ErrInvalidOffsetParam,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid offset parameter",
		},

		// Session errors
		{
			name:           "ErrSessionNotFound",
			err:            errors.ErrSessionNotFound,
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Session not found",
		},
		{
			name:           "ErrSessionExpired",
			err:            errors.ErrSessionExpired,
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Session expired",
		},
		{
			name:           "ErrSessionUserNotFound",
			err:            errors.ErrSessionUserNotFound,
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Session user not found",
		},
		{
			name:           "ErrSessionRefresh",
			err:            errors.ErrSessionRefresh,
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Session refresh failed",
		},

		// Archive validation
		{
			name:           "ErrInvalidArchive",
			err:            errors.ErrInvalidArchive,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid archive format",
		},

		// Internal validation
		{
			name:           "ErrExpectedStruct",
			err:            errors.ErrExpectedStruct,
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Expected struct parameter",
		},

		// Timeout
		{
			name:           "ErrTimeout",
			err:            errors.ErrTimeout,
			expectedStatus: http.StatusGatewayTimeout,
			expectedMsg:    "Operation timeout",
		},

		// Token errors
		{
			name:           "ErrInvalidToken",
			err:            errors.ErrInvalidToken,
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Invalid token",
		},
		{
			name:           "ErrTokenExpired",
			err:            errors.ErrTokenExpired,
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Token expired",
		},
		{
			name:           "ErrTokenUserNotFound",
			err:            errors.ErrTokenUserNotFound,
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Token user not found",
		},
		{
			name:           "ErrInvalidTokenType",
			err:            errors.ErrInvalidTokenType,
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Invalid token type",
		},

		// Contest registration errors
		{
			name:           "ErrContestRegistrationClosed",
			err:            errors.ErrContestRegistrationClosed,
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "Contest registration is closed",
		},
		{
			name:           "ErrContestEnded",
			err:            errors.ErrContestEnded,
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "Contest has ended",
		},
		{
			name:           "ErrAlreadyRegistered",
			err:            errors.ErrAlreadyRegistered,
			expectedStatus: http.StatusConflict,
			expectedMsg:    "Already registered for this contest",
		},
		{
			name:           "ErrAlreadyParticipant",
			err:            errors.ErrAlreadyParticipant,
			expectedStatus: http.StatusConflict,
			expectedMsg:    "User is already a participant of this contest",
		},
		{
			name:           "ErrNoPendingRegistration",
			err:            errors.ErrNoPendingRegistration,
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "No pending registration for this contest",
		},

		// Contest task errors
		{
			name:           "ErrTaskNotInContest",
			err:            errors.ErrTaskNotInContest,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Task is not part of the contest",
		},

		// Language validation
		{
			name:           "ErrInvalidLanguage",
			err:            errors.ErrInvalidLanguage,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid language for the task",
		},

		// Contest submission errors
		{
			name:           "ErrContestSubmissionClosed",
			err:            errors.ErrContestSubmissionClosed,
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "Contest submissions are closed",
		},
		{
			name:           "ErrTaskSubmissionClosed",
			err:            errors.ErrTaskSubmissionClosed,
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "Task submissions are closed for this contest task",
		},

		// Contest timing errors
		{
			name:           "ErrContestNotStarted",
			err:            errors.ErrContestNotStarted,
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "Contest has not started yet",
		},
		{
			name:           "ErrTaskNotStarted",
			err:            errors.ErrTaskNotStarted,
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "Task submission period has not started yet",
		},
		{
			name:           "ErrTaskSubmissionEnded",
			err:            errors.ErrTaskSubmissionEnded,
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "Task submission period has ended",
		},

		// Contest participation
		{
			name:           "ErrNotContestParticipant",
			err:            errors.ErrNotContestParticipant,
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "User is not a participant of this contest",
		},

		// Role assignment
		{
			name:           "ErrCannotAssignOwner",
			err:            errors.ErrCannotAssignOwner,
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "Cannot assign owner role to another user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert the legacy error to a ServiceError
			serviceErr := new(errors.ServiceError)
			if ok := errors.AsServiceError(tt.err, &serviceErr); !ok {
				t.Fatalf("Failed to convert error to ServiceError")
			}

			// Check the HTTP status mapping
			httpStatus := errorCodeToHTTPStatus(serviceErr.Code)
			if httpStatus != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, httpStatus)
			}

			// Check the message
			if serviceErr.Message != tt.expectedMsg {
				t.Errorf("Expected message %q, got %q", tt.expectedMsg, serviceErr.Message)
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
