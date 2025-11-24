// Package errors contains the error types and variables used in the application.
package errors

import (
	stderrors "errors"
)

// ErrorCode represents a unique error code for frontend internationalization
type ErrorCode string

// Error codes for all service layer errors
const (
	// Database errors
	CodeDatabaseConnection ErrorCode = "ERR_DATABASE_CONNECTION"

	CodeEndBeforeStart ErrorCode = "ERR_END_BEFORE_START"

	CodeQueueNotConnected ErrorCode = "ERR_QUEUE_NOT_CONNECTED"

	// Task errors
	CodeTaskExists           ErrorCode = "ERR_TASK_EXISTS"
	CodeTaskNotFound         ErrorCode = "ERR_TASK_NOT_FOUND"
	CodeTaskAlreadyAssigned  ErrorCode = "ERR_TASK_ALREADY_ASSIGNED"
	CodeTaskNotAssignedUser  ErrorCode = "ERR_TASK_NOT_ASSIGNED_USER"
	CodeTaskNotAssignedGroup ErrorCode = "ERR_TASK_NOT_ASSIGNED_GROUP"

	// Authorization errors
	CodeForbidden        ErrorCode = "ERR_FORBIDDEN"
	CodeNotAuthorized    ErrorCode = "ERR_NOT_AUTHORIZED"
	CodeNotAllowed       ErrorCode = "ERR_NOT_ALLOWED"
	CodePermissionDenied ErrorCode = "ERR_PERMISSION_DENIED"

	// User errors
	CodeUserNotFound      ErrorCode = "ERR_USER_NOT_FOUND"
	CodeUserAlreadyExists ErrorCode = "ERR_USER_ALREADY_EXISTS"

	// Access control errors
	CodeAccessAlreadyExists ErrorCode = "ERR_ACCESS_ALREADY_EXISTS"

	// Authentication errors
	CodeInvalidCredentials ErrorCode = "ERR_INVALID_CREDENTIALS"

	// Data validation errors
	CodeInvalidData       ErrorCode = "ERR_INVALID_DATA"
	CodeInvalidInputOuput ErrorCode = "ERR_INVALID_INPUT_OUTPUT"

	// Generic not found
	CodeNotFound ErrorCode = "ERR_NOT_FOUND"

	// File operation errors
	CodeFileOpen            ErrorCode = "ERR_FILE_OPEN"
	CodeTempDirCreate       ErrorCode = "ERR_TEMP_DIR_CREATE"
	CodeDecompressArchive   ErrorCode = "ERR_DECOMPRESS_ARCHIVE"
	CodeNoInputDirectory    ErrorCode = "ERR_NO_INPUT_DIRECTORY"
	CodeNoOutputDirectory   ErrorCode = "ERR_NO_OUTPUT_DIRECTORY"
	CodeIOCountMismatch     ErrorCode = "ERR_IO_COUNT_MISMATCH"
	CodeInputContainsDir    ErrorCode = "ERR_INPUT_CONTAINS_DIR"
	CodeOutputContainsDir   ErrorCode = "ERR_OUTPUT_CONTAINS_DIR"
	CodeInvalidInExtention  ErrorCode = "ERR_INVALID_IN_EXTENSION"
	CodeInvalidOutExtention ErrorCode = "ERR_INVALID_OUT_EXTENSION"

	// FileStorage errors
	CodeWriteTaskID             ErrorCode = "ERR_WRITE_TASK_ID"
	CodeWriteOverwrite          ErrorCode = "ERR_WRITE_OVERWRITE"
	CodeCreateFormFile          ErrorCode = "ERR_CREATE_FORM_FILE"
	CodeCopyFile                ErrorCode = "ERR_COPY_FILE"
	CodeSendRequest             ErrorCode = "ERR_SEND_REQUEST"
	CodeReadResponse            ErrorCode = "ERR_READ_RESPONSE"
	CodeResponseFromFileStorage ErrorCode = "ERR_RESPONSE_FROM_FILE_STORAGE"

	// Group errors
	CodeGroupNotFound ErrorCode = "ERR_GROUP_NOT_FOUND"

	// Pagination errors
	CodeInvalidLimitParam  ErrorCode = "ERR_INVALID_LIMIT_PARAM"
	CodeInvalidOffsetParam ErrorCode = "ERR_INVALID_OFFSET_PARAM"

	// Session errors
	CodeSessionNotFound     ErrorCode = "ERR_SESSION_NOT_FOUND"
	CodeSessionExpired      ErrorCode = "ERR_SESSION_EXPIRED"
	CodeSessionUserNotFound ErrorCode = "ERR_SESSION_USER_NOT_FOUND"
	CodeSessionRefresh      ErrorCode = "ERR_SESSION_REFRESH"

	// Archive errors
	CodeInvalidArchive ErrorCode = "ERR_INVALID_ARCHIVE"

	// Internal errors
	CodeExpectedStruct ErrorCode = "ERR_EXPECTED_STRUCT"

	// Timeout errors
	CodeTimeout ErrorCode = "ERR_TIMEOUT"

	// Token errors
	CodeInvalidToken      ErrorCode = "ERR_INVALID_TOKEN"
	CodeTokenExpired      ErrorCode = "ERR_TOKEN_EXPIRED"
	CodeTokenUserNotFound ErrorCode = "ERR_TOKEN_USER_NOT_FOUND"
	CodeInvalidTokenType  ErrorCode = "ERR_INVALID_TOKEN_TYPE"

	// Contest registration errors
	CodeContestRegistrationClosed ErrorCode = "ERR_CONTEST_REGISTRATION_CLOSED"
	CodeContestEnded              ErrorCode = "ERR_CONTEST_ENDED"
	CodeAlreadyRegistered         ErrorCode = "ERR_ALREADY_REGISTERED"
	CodeAlreadyParticipant        ErrorCode = "ERR_ALREADY_PARTICIPANT"
	CodeNoPendingRegistration     ErrorCode = "ERR_NO_PENDING_REGISTRATION"

	// Contest task errors
	CodeTaskNotInContest ErrorCode = "ERR_TASK_NOT_IN_CONTEST"

	// Language errors
	CodeInvalidLanguage ErrorCode = "ERR_INVALID_LANGUAGE"

	// Contest submission errors
	CodeContestSubmissionClosed ErrorCode = "ERR_CONTEST_SUBMISSION_CLOSED"
	CodeTaskSubmissionClosed    ErrorCode = "ERR_TASK_SUBMISSION_CLOSED"

	// Contest timing errors
	CodeContestNotStarted   ErrorCode = "ERR_CONTEST_NOT_STARTED"
	CodeTaskNotStarted      ErrorCode = "ERR_TASK_NOT_STARTED"
	CodeTaskSubmissionEnded ErrorCode = "ERR_TASK_SUBMISSION_ENDED"

	// Contest participation errors
	CodeNotContestParticipant ErrorCode = "ERR_NOT_CONTEST_PARTICIPANT"

	// Role errors
	CodeCannotAssignOwner ErrorCode = "ERR_CANNOT_ASSIGN_OWNER"

	// Internal/unknown error
	CodeInternalError ErrorCode = "ERR_INTERNAL_ERROR"
)

// ServiceError is a custom error type for service layer errors.
// It implements the error interface and provides structured error information
// for the HTTP presentation layer. HTTPStatus is NOT included here as that
// is a presentation layer concern.
type ServiceError struct {
	// Code is the unique error code for frontend internationalization
	Code ErrorCode
	// Message is the human-readable error message
	Message string
	// Err is the underlying error (if any)
	Err error
}

// Error implements the error interface
func (e *ServiceError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error for errors.Is/As support
func (e *ServiceError) Unwrap() error {
	return e.Err
}

// newServiceError creates a new ServiceError
func newServiceError(code ErrorCode, message string) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
	}
}

// Predefined service errors - these can be returned directly or wrapped with WithError
var (
	// Database errors
	ErrDatabaseConnection = newServiceError(CodeDatabaseConnection, "Database connection error")

	// Task errors
	ErrTaskExists           = newServiceError(CodeTaskExists, "Task with this title already exists")
	ErrTaskNotFound         = newServiceError(CodeTaskNotFound, "Task not found")
	ErrTaskAlreadyAssigned  = newServiceError(CodeTaskAlreadyAssigned, "Task is already assigned to the user")
	ErrTaskNotAssignedUser  = newServiceError(CodeTaskNotAssignedUser, "Task is not assigned to the user")
	ErrTaskNotAssignedGroup = newServiceError(CodeTaskNotAssignedGroup, "Task is not assigned to the group")

	// Authorization errors
	ErrForbidden        = newServiceError(CodeForbidden, "Not authorized to perform this action")
	ErrNotAuthorized    = newServiceError(CodeNotAuthorized, "Not authorized")
	ErrNotAllowed       = newServiceError(CodeNotAllowed, "Not allowed to perform this action")
	ErrPermissionDenied = newServiceError(CodePermissionDenied, "Permission denied")

	// User errors
	ErrUserNotFound      = newServiceError(CodeUserNotFound, "User not found")
	ErrUserAlreadyExists = newServiceError(CodeUserAlreadyExists, "User already exists")

	// Access control errors
	ErrAccessAlreadyExists = newServiceError(CodeAccessAlreadyExists, "Access already exists")

	// Authentication errors
	ErrInvalidCredentials = newServiceError(CodeInvalidCredentials, "Invalid credentials")

	// Data validation errors
	ErrInvalidData        = newServiceError(CodeInvalidData, "Invalid data")
	ErrInvalidInputOutput = newServiceError(CodeInvalidInputOuput, "Invalid input or output")

	// Generic not found
	ErrNotFound = newServiceError(CodeNotFound, "Requested resource not found")

	// File operation errors
	ErrFileOpen            = newServiceError(CodeFileOpen, "Failed to open file")
	ErrTempDirCreate       = newServiceError(CodeTempDirCreate, "Failed to create temp directory")
	ErrDecompressArchive   = newServiceError(CodeDecompressArchive, "Failed to decompress archive")
	ErrNoInputDirectory    = newServiceError(CodeNoInputDirectory, "No input directory found")
	ErrNoOutputDirectory   = newServiceError(CodeNoOutputDirectory, "No output directory found")
	ErrIOCountMismatch     = newServiceError(CodeIOCountMismatch, "Input and output file count mismatch")
	ErrInputContainsDir    = newServiceError(CodeInputContainsDir, "Input contains a directory")
	ErrOutputContainsDir   = newServiceError(CodeOutputContainsDir, "Output contains a directory")
	ErrInvalidInExtention  = newServiceError(CodeInvalidInExtention, "Invalid input file extension")
	ErrInvalidOutExtention = newServiceError(CodeInvalidOutExtention, "Invalid output file extension")

	// FileStorage errors
	ErrWriteTaskID             = newServiceError(CodeWriteTaskID, "Error writing task ID to form")
	ErrWriteOverwrite          = newServiceError(CodeWriteOverwrite, "Error writing overwrite to form")
	ErrCreateFormFile          = newServiceError(CodeCreateFormFile, "Error creating form file")
	ErrCopyFile                = newServiceError(CodeCopyFile, "Error copying file to form")
	ErrSendRequest             = newServiceError(CodeSendRequest, "Error sending request to FileStorage")
	ErrReadResponse            = newServiceError(CodeReadResponse, "Error reading response from FileStorage")
	ErrResponseFromFileStorage = newServiceError(CodeResponseFromFileStorage, "Error response from FileStorage")

	// Group errors
	ErrGroupNotFound = newServiceError(CodeGroupNotFound, "Group not found")

	// Pagination errors
	ErrInvalidLimitParam  = newServiceError(CodeInvalidLimitParam, "Invalid limit parameter")
	ErrInvalidOffsetParam = newServiceError(CodeInvalidOffsetParam, "Invalid offset parameter")

	// Session errors
	ErrSessionNotFound     = newServiceError(CodeSessionNotFound, "Session not found")
	ErrSessionExpired      = newServiceError(CodeSessionExpired, "Session expired")
	ErrSessionUserNotFound = newServiceError(CodeSessionUserNotFound, "Session user not found")
	ErrSessionRefresh      = newServiceError(CodeSessionRefresh, "Session refresh failed")

	// Archive errors
	ErrInvalidArchive = newServiceError(CodeInvalidArchive, "Invalid archive format")

	// Internal errors
	ErrExpectedStruct = newServiceError(CodeExpectedStruct, "Expected struct parameter")

	// Timeout errors
	ErrTimeout = newServiceError(CodeTimeout, "Operation timeout")

	// Token errors
	ErrInvalidToken      = newServiceError(CodeInvalidToken, "Invalid token")
	ErrTokenExpired      = newServiceError(CodeTokenExpired, "Token expired")
	ErrTokenUserNotFound = newServiceError(CodeTokenUserNotFound, "Token user not found")
	ErrInvalidTokenType  = newServiceError(CodeInvalidTokenType, "Invalid token type")

	// Contest registration errors
	ErrContestRegistrationClosed = newServiceError(CodeContestRegistrationClosed, "Contest registration is closed")
	ErrContestEnded              = newServiceError(CodeContestEnded, "Contest has ended")
	ErrAlreadyRegistered         = newServiceError(CodeAlreadyRegistered, "Already registered for this contest")
	ErrAlreadyParticipant        = newServiceError(CodeAlreadyParticipant, "User is already a participant of this contest")
	ErrNoPendingRegistration     = newServiceError(CodeNoPendingRegistration, "No pending registration for this contest")

	// Contest task errors
	ErrTaskNotInContest = newServiceError(CodeTaskNotInContest, "Task is not part of the contest")

	// Language errors
	ErrInvalidLanguage = newServiceError(CodeInvalidLanguage, "Invalid language for the task")

	// Contest submission errors
	ErrContestSubmissionClosed = newServiceError(CodeContestSubmissionClosed, "Contest submissions are closed")
	ErrTaskSubmissionClosed    = newServiceError(CodeTaskSubmissionClosed, "Task submissions are closed for this contest task")

	// Contest timing errors
	ErrContestNotStarted   = newServiceError(CodeContestNotStarted, "Contest has not started yet")
	ErrTaskNotStarted      = newServiceError(CodeTaskNotStarted, "Task submission period has not started yet")
	ErrTaskSubmissionEnded = newServiceError(CodeTaskSubmissionEnded, "Task submission period has ended")

	// Contest participation errors
	ErrNotContestParticipant = newServiceError(CodeNotContestParticipant, "User is not a participant of this contest")

	// Role errors
	ErrCannotAssignOwner = newServiceError(CodeCannotAssignOwner, "Cannot assign owner role to another user")

	// Internal/unknown error
	ErrInternalError = newServiceError(CodeInternalError, "Internal server error")

	ErrEndBeforeStart = newServiceError(CodeEndBeforeStart, "End time cannot be before start time")

	ErrQueueNotConnected = newServiceError(CodeQueueNotConnected, "Worker queue is not connected")
)

// This is a convenience wrapper around errors.Is from the standard library.
func Is(err, target error) bool {
	return stderrors.Is(err, target)
}

// This is a convenience wrapper around errors.As from the standard library.
func As(err error, target any) bool {
	return stderrors.As(err, target)
}

// AsServiceError checks if err is a ServiceError and assigns it to target.
// Returns true if successful.
func AsServiceError(err error, target **ServiceError) bool {
	return stderrors.As(err, target)
}
