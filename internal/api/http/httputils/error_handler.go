package httputils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/errors"
	"go.uber.org/zap"
)

const InvalidRequestBodyMessage = "Invalid request body"

type errorStruct struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type APIError APIResponse[errorStruct]

// ValidationError represents a single field validation error with code and parameters
type ValidationError struct {
	Code validationErrorCode `json:"code"`
}

type ValidationErrorResponse APIResponse[map[string]ValidationError]

func httpToErrorCode(statusCode int) string {
	code := http.StatusText(statusCode)
	code = strings.ReplaceAll(code, "-", "_")
	code = strings.ReplaceAll(code, " ", "_")
	code = strings.ToUpper(code)
	return "ERR_" + code
}

func ReturnError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	code := httpToErrorCode(statusCode)
	response := APIError{
		Ok:   false,
		Data: errorStruct{Code: code, Message: message},
	}
	encoder := json.NewEncoder(w)
	err := encoder.Encode(response)
	if err != nil {
		ReturnError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

// errorCodeToHTTPStatus maps error codes to HTTP status codes.
// This mapping is done in the presentation layer, not in the service layer.
//
//nolint:gocyclo // This function intentionally maps all error codes exhaustively
func errorCodeToHTTPStatus(code errors.ErrorCode) int {
	switch code {
	case errors.CodeCORSNotAllowed:
		return http.StatusForbidden
	// Database errors - 500
	case errors.CodeDatabaseConnection:
		return http.StatusInternalServerError

	// Task errors
	case errors.CodeTaskExists:
		return http.StatusConflict
	case errors.CodeTaskNotFound:
		return http.StatusNotFound
	case errors.CodeTaskAlreadyAssigned:
		return http.StatusConflict
	case errors.CodeTaskNotAssignedUser, errors.CodeTaskNotAssignedGroup:
		return http.StatusBadRequest

	// Authorization errors
	case errors.CodeForbidden, errors.CodeNotAllowed, errors.CodePermissionDenied:
		return http.StatusForbidden
	case errors.CodeNotAuthorized:
		return http.StatusUnauthorized

	// User errors
	case errors.CodeUserNotFound:
		return http.StatusNotFound
	case errors.CodeUserAlreadyExists:
		return http.StatusConflict

	// Access control errors
	case errors.CodeAccessAlreadyExists:
		return http.StatusConflict

	// Authentication errors
	case errors.CodeInvalidCredentials:
		return http.StatusUnauthorized

	// Data validation errors
	case errors.CodeInvalidData, errors.CodeInvalidInputOuput:
		return http.StatusBadRequest

	// Generic not found
	case errors.CodeNotFound:
		return http.StatusNotFound

	// File operation errors
	case errors.CodeFileOpen, errors.CodeTempDirCreate:
		return http.StatusInternalServerError
	case errors.CodeDecompressArchive, errors.CodeNoInputDirectory, errors.CodeNoOutputDirectory,
		errors.CodeIOCountMismatch, errors.CodeInputContainsDir, errors.CodeOutputContainsDir,
		errors.CodeInvalidInExtention, errors.CodeInvalidOutExtention:
		return http.StatusBadRequest

	// FileStorage errors
	case errors.CodeWriteTaskID, errors.CodeWriteOverwrite, errors.CodeCreateFormFile,
		errors.CodeCopyFile, errors.CodeSendRequest, errors.CodeReadResponse:
		return http.StatusInternalServerError
	case errors.CodeResponseFromFileStorage:
		return http.StatusBadGateway

	// Group errors
	case errors.CodeGroupNotFound:
		return http.StatusNotFound
	// Contest group assignment errors
	case errors.CodeGroupAlreadyAssignedToContest:
		return http.StatusConflict
	case errors.CodeGroupNotAssignedToContest:
		return http.StatusBadRequest

	// Pagination errors
	case errors.CodeInvalidLimitParam, errors.CodeInvalidOffsetParam:
		return http.StatusBadRequest

	// Session errors
	case errors.CodeSessionNotFound, errors.CodeSessionExpired,
		errors.CodeSessionUserNotFound, errors.CodeSessionRefresh:
		return http.StatusUnauthorized

	// Archive errors
	case errors.CodeInvalidArchive:
		return http.StatusBadRequest

	// Internal errors
	case errors.CodeExpectedStruct:
		return http.StatusInternalServerError

	// Timeout errors
	case errors.CodeTimeout:
		return http.StatusGatewayTimeout

	// Token errors
	case errors.CodeInvalidToken, errors.CodeTokenExpired,
		errors.CodeTokenUserNotFound, errors.CodeInvalidTokenType:
		return http.StatusUnauthorized

	// Contest registration errors
	case errors.CodeContestRegistrationClosed, errors.CodeContestEnded:
		return http.StatusForbidden
	case errors.CodeAlreadyRegistered, errors.CodeAlreadyParticipant:
		return http.StatusConflict
	case errors.CodeNoPendingRegistration:
		return http.StatusNotFound

	// Contest task errors
	case errors.CodeTaskNotInContest:
		return http.StatusNotFound

	// Language errors
	case errors.CodeInvalidLanguage:
		return http.StatusBadRequest

	// Contest submission errors
	case errors.CodeContestSubmissionClosed, errors.CodeTaskSubmissionClosed:
		return http.StatusForbidden

	// Contest timing errors
	case errors.CodeContestNotStarted, errors.CodeTaskNotStarted, errors.CodeTaskSubmissionEnded:
		return http.StatusForbidden

	// Contest participation errors
	case errors.CodeNotContestParticipant:
		return http.StatusForbidden

	// Role errors
	case errors.CodeCannotAssignOwner:
		return http.StatusForbidden

	// Internal error
	case errors.CodeInternalError:
		return http.StatusInternalServerError
	case errors.CodeEndBeforeStart:
		return http.StatusBadRequest
	case errors.CodeQueueNotConnected:
		return http.StatusServiceUnavailable
	// Default - internal error
	default:
		return http.StatusInternalServerError
	}
}

// ReturnServiceError writes a ServiceError as JSON response
func ReturnServiceError(w http.ResponseWriter, serviceErr *errors.ServiceError) {
	statusCode := errorCodeToHTTPStatus(serviceErr.Code)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := APIError{
		Ok:   false,
		Data: errorStruct{Code: string(serviceErr.Code), Message: serviceErr.Message},
	}
	encoder := json.NewEncoder(w)
	err := encoder.Encode(response)
	if err != nil {
		ReturnError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

// HandleServiceError is a centralized error handler for service layer errors
// It maps service errors to appropriate HTTP status codes and messages,
// handles database rollback if needed, and logs unexpected errors
func HandleServiceError(w http.ResponseWriter, err error, db database.Database, logger *zap.SugaredLogger) {
	if err == nil {
		return
	}

	// Rollback the database transaction if db is provided
	if db != nil {
		db.Rollback()
	}

	// Convert to ServiceError (handles both ServiceError and legacy errors)
	serviceErr := new(errors.ServiceError)
	if ok := errors.AsServiceError(err, &serviceErr); !ok {
		serviceErr = errors.ErrInternalError
	}

	// Get HTTP status from error code
	httpStatus := errorCodeToHTTPStatus(serviceErr.Code)

	// Log unexpected errors (500 level errors)
	if logger != nil && httpStatus >= http.StatusInternalServerError {
		logger.Errorw("Service error", "error", err, "code", serviceErr.Code, "status", httpStatus)
	}

	// Return the error response with error code
	ReturnServiceError(w, serviceErr)
}

// HandleValidationError handles validation errors and returns appropriate HTTP responses
func HandleValidationError(w http.ResponseWriter, err error) {
	var valErrs validator.ValidationErrors
	if errors.As(err, &valErrs) {
		returnValidationError(w, valErrs)
		return
	}
	ReturnError(w, http.StatusBadRequest, InvalidRequestBodyMessage)
}

type validationErrorCode string

const (
	valCodeRequired        validationErrorCode = "FIELD_REQUIRED"
	valCodeInvalidEmail    validationErrorCode = "INVALID_EMAIL"
	valCodeMinLength       validationErrorCode = "MIN_LENGTH_%s"
	valCodeMaxLength       validationErrorCode = "MAX_LENGTH_%s"
	valCodeFieldsMustMatch validationErrorCode = "FIELD_MUST_MATCH_%s"
	valCodeInvalidUsername validationErrorCode = "INVALID_USERNAME_FORMAT"
	valCodeInvalidPassword validationErrorCode = "INVALID_PASSWORD_FORMAT"
	valCodeInvalidField    validationErrorCode = "INVALID_FIELD"
)

// ConvertValidationErrors converts validator.ValidationErrors to a map of field names to error codes with parameters
func ConvertValidationErrors(validationErrors validator.ValidationErrors) map[string]ValidationError {
	errors := make(map[string]ValidationError)

	for _, err := range validationErrors {
		// Use the field name from validator (now uses JSON tags due to RegisterTagNameFunc)
		fieldName := err.Field()

		switch err.Tag() {
		case "required":
			errors[fieldName] = ValidationError{
				Code: valCodeRequired,
			}
		case "email":
			errors[fieldName] = ValidationError{
				Code: valCodeInvalidEmail,
			}
		case "gte":
			errors[fieldName] = ValidationError{
				Code: validationErrorCode(fmt.Sprintf(string(valCodeMinLength), err.Param())),
			}
		case "lte":
			errors[fieldName] = ValidationError{
				Code: validationErrorCode(fmt.Sprintf(string(valCodeMaxLength), err.Param())),
			}
		case "eqfield":
			// Map struct field name to JSON field name for the parameter
			paramFieldName := err.Param()
			if len(paramFieldName) > 0 {
				paramFieldName = strings.ToLower(paramFieldName[:1]) + paramFieldName[1:]
			}
			errors[fieldName] = ValidationError{
				Code: validationErrorCode(fmt.Sprintf(string(valCodeFieldsMustMatch), paramFieldName)),
			}
		case "username":
			errors[fieldName] = ValidationError{
				Code: valCodeInvalidUsername,
			}
		case "password":
			errors[fieldName] = ValidationError{
				Code: valCodeInvalidPassword,
			}
		default:
			errors[fieldName] = ValidationError{
				Code: valCodeInvalidField,
			}
		}
	}

	return errors
}

// returnValidationError returns a structured validation error response
func returnValidationError(w http.ResponseWriter, validationErrors validator.ValidationErrors) {
	fieldErrors := ConvertValidationErrors(validationErrors)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	response := ValidationErrorResponse{
		Ok:   false,
		Data: fieldErrors,
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		ReturnError(w, http.StatusInternalServerError, err.Error())
		return
	}
}
