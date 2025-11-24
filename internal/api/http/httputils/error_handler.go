package httputils

import (
	"errors"
	"net/http"

	"github.com/mini-maxit/backend/internal/database"
	myerrors "github.com/mini-maxit/backend/package/errors"
	"go.uber.org/zap"
)

// ErrorMapping defines the HTTP status code and message for a specific error
type ErrorMapping struct {
	StatusCode int
	Message    string
}

// getErrorMapping returns the HTTP status code and message for a given error
// This function exhaustively handles all errors defined in the errors package
func getErrorMapping(err error) ErrorMapping {
	// Database and connection errors
	if errors.Is(err, myerrors.ErrDatabaseConnection) {
		return ErrorMapping{http.StatusInternalServerError, "Database connection error"}
	}

	// Task-related errors
	if errors.Is(err, myerrors.ErrTaskExists) {
		return ErrorMapping{http.StatusConflict, "Task with this title already exists"}
	}
	if errors.Is(err, myerrors.ErrTaskNotFound) {
		return ErrorMapping{http.StatusNotFound, "Task not found"}
	}
	if errors.Is(err, myerrors.ErrTaskAlreadyAssigned) {
		return ErrorMapping{http.StatusConflict, "Task is already assigned to the user"}
	}
	if errors.Is(err, myerrors.ErrTaskNotAssignedToUser) {
		return ErrorMapping{http.StatusBadRequest, "Task is not assigned to the user"}
	}
	if errors.Is(err, myerrors.ErrTaskNotAssignedToGroup) {
		return ErrorMapping{http.StatusBadRequest, "Task is not assigned to the group"}
	}

	// Authorization errors
	if errors.Is(err, myerrors.ErrForbidden) {
		return ErrorMapping{http.StatusForbidden, "Not authorized to perform this action"}
	}
	if errors.Is(err, myerrors.ErrNotAuthorized) {
		return ErrorMapping{http.StatusUnauthorized, "Not authorized"}
	}
	if errors.Is(err, myerrors.ErrNotAllowed) {
		return ErrorMapping{http.StatusForbidden, "Not allowed to perform this action"}
	}
	if errors.Is(err, myerrors.ErrPermissionDenied) {
		return ErrorMapping{http.StatusForbidden, "Permission denied"}
	}

	// User-related errors
	if errors.Is(err, myerrors.ErrUserNotFound) {
		return ErrorMapping{http.StatusNotFound, "User not found"}
	}
	if errors.Is(err, myerrors.ErrUserAlreadyExists) {
		return ErrorMapping{http.StatusConflict, "User already exists"}
	}

	// Access control errors
	if errors.Is(err, myerrors.ErrAccessAlreadyExists) {
		return ErrorMapping{http.StatusConflict, "Access already exists"}
	}

	// Authentication errors
	if errors.Is(err, myerrors.ErrInvalidCredentials) {
		return ErrorMapping{http.StatusUnauthorized, "Invalid credentials"}
	}

	// Data validation errors
	if errors.Is(err, myerrors.ErrInvalidData) {
		return ErrorMapping{http.StatusBadRequest, "Invalid data"}
	}
	if errors.Is(err, myerrors.ErrInvalidInputOuput) {
		return ErrorMapping{http.StatusBadRequest, "Invalid input or output"}
	}

	// Generic not found error
	if errors.Is(err, myerrors.ErrNotFound) {
		return ErrorMapping{http.StatusNotFound, "Requested resource not found"}
	}

	// File operation errors
	if errors.Is(err, myerrors.ErrFileOpen) {
		return ErrorMapping{http.StatusInternalServerError, "Failed to open file"}
	}
	if errors.Is(err, myerrors.ErrTempDirCreate) {
		return ErrorMapping{http.StatusInternalServerError, "Failed to create temp directory"}
	}
	if errors.Is(err, myerrors.ErrDecompressArchive) {
		return ErrorMapping{http.StatusBadRequest, "Failed to decompress archive"}
	}
	if errors.Is(err, myerrors.ErrNoInputDirectory) {
		return ErrorMapping{http.StatusBadRequest, "No input directory found"}
	}
	if errors.Is(err, myerrors.ErrNoOutputDirectory) {
		return ErrorMapping{http.StatusBadRequest, "No output directory found"}
	}
	if errors.Is(err, myerrors.ErrIOCountMismatch) {
		return ErrorMapping{http.StatusBadRequest, "Input and output file count mismatch"}
	}
	if errors.Is(err, myerrors.ErrInputContainsDir) {
		return ErrorMapping{http.StatusBadRequest, "Input contains a directory"}
	}
	if errors.Is(err, myerrors.ErrOutputContainsDir) {
		return ErrorMapping{http.StatusBadRequest, "Output contains a directory"}
	}
	if errors.Is(err, myerrors.ErrInvalidInExtention) {
		return ErrorMapping{http.StatusBadRequest, "Invalid input file extension"}
	}
	if errors.Is(err, myerrors.ErrInvalidOutExtention) {
		return ErrorMapping{http.StatusBadRequest, "Invalid output file extension"}
	}

	// FileStorage interaction errors
	if errors.Is(err, myerrors.ErrWriteTaskID) {
		return ErrorMapping{http.StatusInternalServerError, "Error writing task ID to form"}
	}
	if errors.Is(err, myerrors.ErrWriteOverwrite) {
		return ErrorMapping{http.StatusInternalServerError, "Error writing overwrite to form"}
	}
	if errors.Is(err, myerrors.ErrCreateFormFile) {
		return ErrorMapping{http.StatusInternalServerError, "Error creating form file"}
	}
	if errors.Is(err, myerrors.ErrCopyFile) {
		return ErrorMapping{http.StatusInternalServerError, "Error copying file to form"}
	}
	if errors.Is(err, myerrors.ErrSendRequest) {
		return ErrorMapping{http.StatusInternalServerError, "Error sending request to FileStorage"}
	}
	if errors.Is(err, myerrors.ErrReadResponse) {
		return ErrorMapping{http.StatusInternalServerError, "Error reading response from FileStorage"}
	}
	if errors.Is(err, myerrors.ErrResponseFromFileStorage) {
		return ErrorMapping{http.StatusBadGateway, "Error response from FileStorage"}
	}

	// Group-related errors
	if errors.Is(err, myerrors.ErrGroupNotFound) {
		return ErrorMapping{http.StatusNotFound, "Group not found"}
	}

	// Pagination parameter errors
	if errors.Is(err, myerrors.ErrInvalidLimitParam) {
		return ErrorMapping{http.StatusBadRequest, "Invalid limit parameter"}
	}
	if errors.Is(err, myerrors.ErrInvalidOffsetParam) {
		return ErrorMapping{http.StatusBadRequest, "Invalid offset parameter"}
	}

	// Session-related errors
	if errors.Is(err, myerrors.ErrSessionNotFound) {
		return ErrorMapping{http.StatusUnauthorized, "Session not found"}
	}
	if errors.Is(err, myerrors.ErrSessionExpired) {
		return ErrorMapping{http.StatusUnauthorized, "Session expired"}
	}
	if errors.Is(err, myerrors.ErrSessionUserNotFound) {
		return ErrorMapping{http.StatusUnauthorized, "Session user not found"}
	}
	if errors.Is(err, myerrors.ErrSessionRefresh) {
		return ErrorMapping{http.StatusUnauthorized, "Session refresh failed"}
	}

	// Archive validation errors
	if errors.Is(err, myerrors.ErrInvalidArchive) {
		return ErrorMapping{http.StatusBadRequest, "Invalid archive format"}
	}

	// Internal validation errors
	if errors.Is(err, myerrors.ErrExpectedStruct) {
		return ErrorMapping{http.StatusInternalServerError, "Expected struct parameter"}
	}

	// Timeout errors
	if errors.Is(err, myerrors.ErrTimeout) {
		return ErrorMapping{http.StatusRequestTimeout, "Operation timeout"}
	}

	// Token-related errors
	if errors.Is(err, myerrors.ErrInvalidToken) {
		return ErrorMapping{http.StatusUnauthorized, "Invalid token"}
	}
	if errors.Is(err, myerrors.ErrTokenExpired) {
		return ErrorMapping{http.StatusUnauthorized, "Token expired"}
	}
	if errors.Is(err, myerrors.ErrTokenUserNotFound) {
		return ErrorMapping{http.StatusUnauthorized, "Token user not found"}
	}
	if errors.Is(err, myerrors.ErrInvalidTokenType) {
		return ErrorMapping{http.StatusUnauthorized, "Invalid token type"}
	}

	// Contest registration errors
	if errors.Is(err, myerrors.ErrContestRegistrationClosed) {
		return ErrorMapping{http.StatusForbidden, "Contest registration is closed"}
	}
	if errors.Is(err, myerrors.ErrContestEnded) {
		return ErrorMapping{http.StatusForbidden, "Contest has ended"}
	}
	if errors.Is(err, myerrors.ErrAlreadyRegistered) {
		return ErrorMapping{http.StatusConflict, "Already registered for this contest"}
	}
	if errors.Is(err, myerrors.ErrAlreadyParticipant) {
		return ErrorMapping{http.StatusConflict, "User is already a participant of this contest"}
	}
	if errors.Is(err, myerrors.ErrNoPendingRegistration) {
		return ErrorMapping{http.StatusNotFound, "No pending registration for this contest"}
	}

	// Contest task errors
	if errors.Is(err, myerrors.ErrTaskNotInContest) {
		return ErrorMapping{http.StatusBadRequest, "Task is not part of the contest"}
	}

	// Language validation errors
	if errors.Is(err, myerrors.ErrInvalidLanguage) {
		return ErrorMapping{http.StatusBadRequest, "Invalid language for the task"}
	}

	// Contest submission errors
	if errors.Is(err, myerrors.ErrContestSubmissionClosed) {
		return ErrorMapping{http.StatusForbidden, "Contest submissions are closed"}
	}
	if errors.Is(err, myerrors.ErrTaskSubmissionClosed) {
		return ErrorMapping{http.StatusForbidden, "Task submissions are closed for this contest task"}
	}

	// Contest timing errors
	if errors.Is(err, myerrors.ErrContestNotStarted) {
		return ErrorMapping{http.StatusForbidden, "Contest has not started yet"}
	}
	if errors.Is(err, myerrors.ErrTaskNotStarted) {
		return ErrorMapping{http.StatusForbidden, "Task submission period has not started yet"}
	}
	if errors.Is(err, myerrors.ErrTaskSubmissionEnded) {
		return ErrorMapping{http.StatusForbidden, "Task submission period has ended"}
	}

	// Contest participation errors
	if errors.Is(err, myerrors.ErrNotContestParticipant) {
		return ErrorMapping{http.StatusForbidden, "User is not a participant of this contest"}
	}

	// Role assignment errors
	if errors.Is(err, myerrors.ErrCannotAssignOwner) {
		return ErrorMapping{http.StatusForbidden, "Cannot assign owner role to another user"}
	}

	// Default case for unknown errors
	return ErrorMapping{http.StatusInternalServerError, "Internal server error"}
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

	// Get the error mapping
	mapping := getErrorMapping(err)

	// Log unexpected errors (500 level errors)
	if logger != nil && mapping.StatusCode >= 500 {
		logger.Errorw("Service error", "error", err, "status", mapping.StatusCode)
	}

	// Return the error response
	ReturnError(w, mapping.StatusCode, mapping.Message)
}
