//nolint:testpackage // Testing internal function errorCodeToHTTPStatus
package httputils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/utils"
	"github.com/stretchr/testify/require"
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

func TestConvertValidationErrors(t *testing.T) {
	validate, err := utils.NewValidator()
	require.NoError(t, err)

	type TestStruct struct {
		RequiredField string `validate:"required" json:"requiredField"`
		EmailField    string `validate:"email" json:"emailField"`
		MinLenField   string `validate:"gte=5" json:"minLenField"`
		MaxLenField   string `validate:"lte=3" json:"maxLenField"`
		MatchField    string `validate:"eqfield=ConfirmField" json:"matchField"`
		ConfirmField  string `json:"confirmField"`
		UsernameField string `validate:"username" json:"usernameField"`
		PasswordField string `validate:"password" json:"passwordField"`
	}

	// Build instance to trigger errors for each tag
	obj := TestStruct{
		RequiredField: "",
		EmailField:    "not-an-email",
		MinLenField:   "123",
		MaxLenField:   "1234",
		MatchField:    "abc",
		ConfirmField:  "abcd",
		UsernameField: "1invalid",
		PasswordField: "weak",
	}

	err = validate.Struct(obj)
	if err == nil {
		t.Fatalf("expected validation errors, got nil")
	}

	var valErrs validator.ValidationErrors
	ok := errors.As(err, &valErrs)
	if !ok {
		t.Fatalf("expected validator.ValidationErrors, got %T", err)
	}

	result := ConvertValidationErrors(valErrs)

	// required
	if v, exists := result["requiredField"]; !exists || string(v.Code) != "FIELD_REQUIRED" {
		t.Errorf("requiredField: expected FIELD_REQUIRED, got %+v", v)
	}
	// email
	if v, exists := result["emailField"]; !exists || string(v.Code) != "INVALID_EMAIL" {
		t.Errorf("emailField: expected INVALID_EMAIL, got %+v", v)
	}
	// gte -> MIN_LENGTH_%s with param
	if v, exists := result["minLenField"]; !exists || string(v.Code) != "MIN_LENGTH_5" {
		t.Errorf("minLenField: expected MIN_LENGTH_5, got %+v", v)
	}
	// lte -> MAX_LENGTH_%s with param
	if v, exists := result["maxLenField"]; !exists || string(v.Code) != "MAX_LENGTH_3" {
		t.Errorf("maxLenField: expected MAX_LENGTH_3, got %+v", v)
	}
	// eqfield -> FIELD_MUST_MATCH_%s where %s is json name of param (confirmField)
	if v, exists := result["matchField"]; !exists || string(v.Code) != "FIELD_MUST_MATCH_confirmField" {
		t.Errorf("matchField: expected FIELD_MUST_MATCH_confirmField, got %+v", v)
	}
	// username
	if v, exists := result["usernameField"]; !exists || string(v.Code) != "INVALID_USERNAME_FORMAT" {
		t.Errorf("usernameField: expected INVALID_USERNAME_FORMAT, got %+v", v)
	}
	// password
	if v, exists := result["passwordField"]; !exists || string(v.Code) != "INVALID_PASSWORD_FORMAT" {
		t.Errorf("passwordField: expected INVALID_PASSWORD_FORMAT, got %+v", v)
	}
}
