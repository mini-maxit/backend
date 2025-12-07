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
	require.True(t, ok, "expected validator.ValidationErrors")

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

func TestHttpToErrorCode(t *testing.T) {
	tests := []struct {
		status   int
		expected string
	}{
		{http.StatusNotFound, "ERR_NOT_FOUND"},
		{http.StatusInternalServerError, "ERR_INTERNAL_SERVER_ERROR"},
		{http.StatusBadRequest, "ERR_BAD_REQUEST"},
		{http.StatusNonAuthoritativeInfo, "ERR_NON_AUTHORITATIVE_INFORMATION"},
	}
	for _, tc := range tests {
		got := httpToErrorCode(tc.status)
		if got != tc.expected {
			t.Fatalf("expected %s, got %s", tc.expected, got)
		}
	}
}

const applicationJSON = "application/json"

func TestReturnError(t *testing.T) {
	w := httptest.NewRecorder()
	ReturnError(w, http.StatusBadRequest, "bad req")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != applicationJSON {
		t.Fatalf("expected Content-Type application/json, got %s", ct)
	}

	var resp APIError
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Ok {
		t.Fatalf("expected ok=false")
	}
	if resp.Data.Code != "ERR_BAD_REQUEST" {
		t.Fatalf("expected code ERR_BAD_REQUEST, got %s", resp.Data.Code)
	}
	if resp.Data.Message != "bad req" {
		t.Fatalf("expected message 'bad req', got %s", resp.Data.Message)
	}
}

func TestReturnServiceError(t *testing.T) {
	w := httptest.NewRecorder()
	ReturnServiceError(w, errors.ErrUserAlreadyExists)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != applicationJSON {
		t.Fatalf("expected Content-Type application/json, got %s", ct)
	}

	var resp APIError
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Ok {
		t.Fatalf("expected ok=false")
	}
	if resp.Data.Code != string(errors.CodeUserAlreadyExists) {
		t.Fatalf("expected code %s, got %s", errors.CodeUserAlreadyExists, resp.Data.Code)
	}
	if resp.Data.Message != "User already exists" {
		t.Fatalf("expected message 'User already exists', got %s", resp.Data.Message)
	}
}

func TestReturnValidationError(t *testing.T) {
	validate, err := utils.NewValidator()
	require.NoError(t, err)

	type S struct {
		Name string `json:"name" validate:"required"`
	}
	var s S
	err = validate.Struct(s)
	require.Error(t, err)

	var valErrs validator.ValidationErrors
	ok := errors.As(err, &valErrs)
	require.True(t, ok)

	w := httptest.NewRecorder()
	returnValidationError(w, valErrs)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != applicationJSON {
		t.Fatalf("expected Content-Type application/json, got %s", ct)
	}

	var resp ValidationErrorResponse
	if derr := json.NewDecoder(w.Body).Decode(&resp); derr != nil {
		t.Fatalf("decode error: %v", derr)
	}
	if resp.Ok {
		t.Fatalf("expected ok=false")
	}
	if v, ok := resp.Data["name"]; !ok || string(v.Code) != "FIELD_REQUIRED" {
		t.Fatalf("expected name FIELD_REQUIRED, got %+v", v)
	}
}

func TestHandleValidationError(t *testing.T) {
	t.Run("validation errors mapped", func(t *testing.T) {
		validate, err := utils.NewValidator()
		require.NoError(t, err)

		type S struct {
			Email string `json:"email" validate:"email"`
		}
		s := S{Email: "bad-email"}
		verr := validate.Struct(s)
		require.Error(t, verr)

		w := httptest.NewRecorder()
		HandleValidationError(w, verr)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
		var resp ValidationErrorResponse
		if derr := json.NewDecoder(w.Body).Decode(&resp); derr != nil {
			t.Fatalf("decode error: %v", derr)
		}
		if resp.Ok {
			t.Fatalf("expected ok=false")
		}
		if v, ok := resp.Data["email"]; !ok || string(v.Code) != "INVALID_EMAIL" {
			t.Fatalf("expected email INVALID_EMAIL, got %+v", v)
		}
	})

	t.Run("non-validation error returns generic bad request", func(t *testing.T) {
		w := httptest.NewRecorder()
		HandleValidationError(w, errors.New("boom"))

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
		var resp APIError
		if derr := json.NewDecoder(w.Body).Decode(&resp); derr != nil {
			t.Fatalf("decode error: %v", derr)
		}
		if resp.Ok {
			t.Fatalf("expected ok=false")
		}
		if resp.Data.Code != "ERR_BAD_REQUEST" {
			t.Fatalf("expected code ERR_BAD_REQUEST, got %s", resp.Data.Code)
		}
		if resp.Data.Message != InvalidRequestBodyMessage {
			t.Fatalf("expected message %q, got %q", InvalidRequestBodyMessage, resp.Data.Message)
		}
	})
}
