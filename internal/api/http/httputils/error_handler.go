package httputils

import (
	"encoding/json"
	"net/http"

	"github.com/mini-maxit/backend/internal/database"
	myerrors "github.com/mini-maxit/backend/package/errors"
	"go.uber.org/zap"
)

// ServiceErrorResponse is the JSON response structure for service errors
type ServiceErrorResponse struct {
	Ok   bool             `json:"ok"`
	Data ServiceErrorData `json:"data"`
}

// ServiceErrorData contains the error details
type ServiceErrorData struct {
	Code    myerrors.ErrorCode `json:"code"`
	Message string             `json:"message"`
}

// errorCodeToHTTPStatus maps error codes to HTTP status codes.
// This mapping is done in the presentation layer, not in the service layer.
//
//nolint:gocyclo // This function intentionally maps all error codes exhaustively
func errorCodeToHTTPStatus(code myerrors.ErrorCode) int {
	switch code {
	// Database errors - 500
	case myerrors.CodeDatabaseConnection:
		return http.StatusInternalServerError

	// Task errors
	case myerrors.CodeTaskExists:
		return http.StatusConflict
	case myerrors.CodeTaskNotFound:
		return http.StatusNotFound
	case myerrors.CodeTaskAlreadyAssigned:
		return http.StatusConflict
	case myerrors.CodeTaskNotAssignedUser, myerrors.CodeTaskNotAssignedGroup:
		return http.StatusBadRequest

	// Authorization errors
	case myerrors.CodeForbidden, myerrors.CodeNotAllowed, myerrors.CodePermissionDenied:
		return http.StatusForbidden
	case myerrors.CodeNotAuthorized:
		return http.StatusUnauthorized

	// User errors
	case myerrors.CodeUserNotFound:
		return http.StatusNotFound
	case myerrors.CodeUserAlreadyExists:
		return http.StatusConflict

	// Access control errors
	case myerrors.CodeAccessAlreadyExists:
		return http.StatusConflict

	// Authentication errors
	case myerrors.CodeInvalidCredentials:
		return http.StatusUnauthorized

	// Data validation errors
	case myerrors.CodeInvalidData, myerrors.CodeInvalidInputOuput:
		return http.StatusBadRequest

	// Generic not found
	case myerrors.CodeNotFound:
		return http.StatusNotFound

	// File operation errors
	case myerrors.CodeFileOpen, myerrors.CodeTempDirCreate:
		return http.StatusInternalServerError
	case myerrors.CodeDecompressArchive, myerrors.CodeNoInputDirectory, myerrors.CodeNoOutputDirectory,
		myerrors.CodeIOCountMismatch, myerrors.CodeInputContainsDir, myerrors.CodeOutputContainsDir,
		myerrors.CodeInvalidInExtention, myerrors.CodeInvalidOutExtention:
		return http.StatusBadRequest

	// FileStorage errors
	case myerrors.CodeWriteTaskID, myerrors.CodeWriteOverwrite, myerrors.CodeCreateFormFile,
		myerrors.CodeCopyFile, myerrors.CodeSendRequest, myerrors.CodeReadResponse:
		return http.StatusInternalServerError
	case myerrors.CodeResponseFromFileStorage:
		return http.StatusBadGateway

	// Group errors
	case myerrors.CodeGroupNotFound:
		return http.StatusNotFound

	// Pagination errors
	case myerrors.CodeInvalidLimitParam, myerrors.CodeInvalidOffsetParam:
		return http.StatusBadRequest

	// Session errors
	case myerrors.CodeSessionNotFound, myerrors.CodeSessionExpired,
		myerrors.CodeSessionUserNotFound, myerrors.CodeSessionRefresh:
		return http.StatusUnauthorized

	// Archive errors
	case myerrors.CodeInvalidArchive:
		return http.StatusBadRequest

	// Internal errors
	case myerrors.CodeExpectedStruct:
		return http.StatusInternalServerError

	// Timeout errors
	case myerrors.CodeTimeout:
		return http.StatusGatewayTimeout

	// Token errors
	case myerrors.CodeInvalidToken, myerrors.CodeTokenExpired,
		myerrors.CodeTokenUserNotFound, myerrors.CodeInvalidTokenType:
		return http.StatusUnauthorized

	// Contest registration errors
	case myerrors.CodeContestRegistrationClosed, myerrors.CodeContestEnded:
		return http.StatusForbidden
	case myerrors.CodeAlreadyRegistered, myerrors.CodeAlreadyParticipant:
		return http.StatusConflict
	case myerrors.CodeNoPendingRegistration:
		return http.StatusNotFound

	// Contest task errors
	case myerrors.CodeTaskNotInContest:
		return http.StatusBadRequest

	// Language errors
	case myerrors.CodeInvalidLanguage:
		return http.StatusBadRequest

	// Contest submission errors
	case myerrors.CodeContestSubmissionClosed, myerrors.CodeTaskSubmissionClosed:
		return http.StatusForbidden

	// Contest timing errors
	case myerrors.CodeContestNotStarted, myerrors.CodeTaskNotStarted, myerrors.CodeTaskSubmissionEnded:
		return http.StatusForbidden

	// Contest participation errors
	case myerrors.CodeNotContestParticipant:
		return http.StatusForbidden

	// Role errors
	case myerrors.CodeCannotAssignOwner:
		return http.StatusForbidden

	// Internal error
	case myerrors.CodeInternalError:
		return http.StatusInternalServerError

	// Default - internal error
	default:
		return http.StatusInternalServerError
	}
}

// returnServiceError writes a ServiceError as JSON response
func returnServiceError(w http.ResponseWriter, serviceErr *myerrors.ServiceError) {
	httpStatus := errorCodeToHTTPStatus(serviceErr.Code)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	response := ServiceErrorResponse{
		Ok: false,
		Data: ServiceErrorData{
			Code:    serviceErr.Code,
			Message: serviceErr.Message,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Fallback to plain text if JSON encoding fails
		http.Error(w, serviceErr.Message, httpStatus)
	}
}

// HandleServiceError is a centralized error handler for service layer errors
// It maps service errors to appropriate HTTP status codes and messages,
// handles database rollback if needed, and logs unexpected errors
//
// Parameters:
//   - w: http.ResponseWriter to write the error response
//   - err: the error returned from the service layer
//   - db: optional database connection for rollback (can be nil)
//   - logger: optional logger for logging unexpected errors (can be nil)
func HandleServiceError(w http.ResponseWriter, err error, db database.Database, logger *zap.SugaredLogger) {
	if err == nil {
		return
	}

	// Rollback the database transaction if db is provided
	if db != nil {
		db.Rollback()
	}

	// Convert to ServiceError (handles both ServiceError and legacy errors)
	serviceErr := myerrors.ToServiceError(err)

	// Get HTTP status from error code
	httpStatus := errorCodeToHTTPStatus(serviceErr.Code)

	// Log unexpected errors (500 level errors)
	if logger != nil && httpStatus >= http.StatusInternalServerError {
		logger.Errorw("Service error", "error", err, "code", serviceErr.Code, "status", httpStatus)
	}

	// Return the error response with error code
	returnServiceError(w, serviceErr)
}
