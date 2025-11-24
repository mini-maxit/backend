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

// NewServiceError creates a new ServiceError
func NewServiceError(code ErrorCode, message string) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
	}
}

// WithError wraps an underlying error
func (e *ServiceError) WithError(err error) *ServiceError {
	return &ServiceError{
		Code:    e.Code,
		Message: e.Message,
		Err:     err,
	}
}

// Predefined service errors - these can be returned directly or wrapped with WithError
var (
	// Database errors
	ErrServiceDatabaseConnection = NewServiceError(CodeDatabaseConnection, "Database connection error")

	// Task errors
	ErrServiceTaskExists           = NewServiceError(CodeTaskExists, "Task with this title already exists")
	ErrServiceTaskNotFound         = NewServiceError(CodeTaskNotFound, "Task not found")
	ErrServiceTaskAlreadyAssigned  = NewServiceError(CodeTaskAlreadyAssigned, "Task is already assigned to the user")
	ErrServiceTaskNotAssignedUser  = NewServiceError(CodeTaskNotAssignedUser, "Task is not assigned to the user")
	ErrServiceTaskNotAssignedGroup = NewServiceError(CodeTaskNotAssignedGroup, "Task is not assigned to the group")

	// Authorization errors
	ErrServiceForbidden        = NewServiceError(CodeForbidden, "Not authorized to perform this action")
	ErrServiceNotAuthorized    = NewServiceError(CodeNotAuthorized, "Not authorized")
	ErrServiceNotAllowed       = NewServiceError(CodeNotAllowed, "Not allowed to perform this action")
	ErrServicePermissionDenied = NewServiceError(CodePermissionDenied, "Permission denied")

	// User errors
	ErrServiceUserNotFound      = NewServiceError(CodeUserNotFound, "User not found")
	ErrServiceUserAlreadyExists = NewServiceError(CodeUserAlreadyExists, "User already exists")

	// Access control errors
	ErrServiceAccessAlreadyExists = NewServiceError(CodeAccessAlreadyExists, "Access already exists")

	// Authentication errors
	ErrServiceInvalidCredentials = NewServiceError(CodeInvalidCredentials, "Invalid credentials")

	// Data validation errors
	ErrServiceInvalidData        = NewServiceError(CodeInvalidData, "Invalid data")
	ErrServiceInvalidInputOutput = NewServiceError(CodeInvalidInputOuput, "Invalid input or output")

	// Generic not found
	ErrServiceNotFound = NewServiceError(CodeNotFound, "Requested resource not found")

	// File operation errors
	ErrServiceFileOpen            = NewServiceError(CodeFileOpen, "Failed to open file")
	ErrServiceTempDirCreate       = NewServiceError(CodeTempDirCreate, "Failed to create temp directory")
	ErrServiceDecompressArchive   = NewServiceError(CodeDecompressArchive, "Failed to decompress archive")
	ErrServiceNoInputDirectory    = NewServiceError(CodeNoInputDirectory, "No input directory found")
	ErrServiceNoOutputDirectory   = NewServiceError(CodeNoOutputDirectory, "No output directory found")
	ErrServiceIOCountMismatch     = NewServiceError(CodeIOCountMismatch, "Input and output file count mismatch")
	ErrServiceInputContainsDir    = NewServiceError(CodeInputContainsDir, "Input contains a directory")
	ErrServiceOutputContainsDir   = NewServiceError(CodeOutputContainsDir, "Output contains a directory")
	ErrServiceInvalidInExtention  = NewServiceError(CodeInvalidInExtention, "Invalid input file extension")
	ErrServiceInvalidOutExtention = NewServiceError(CodeInvalidOutExtention, "Invalid output file extension")

	// FileStorage errors
	ErrServiceWriteTaskID             = NewServiceError(CodeWriteTaskID, "Error writing task ID to form")
	ErrServiceWriteOverwrite          = NewServiceError(CodeWriteOverwrite, "Error writing overwrite to form")
	ErrServiceCreateFormFile          = NewServiceError(CodeCreateFormFile, "Error creating form file")
	ErrServiceCopyFile                = NewServiceError(CodeCopyFile, "Error copying file to form")
	ErrServiceSendRequest             = NewServiceError(CodeSendRequest, "Error sending request to FileStorage")
	ErrServiceReadResponse            = NewServiceError(CodeReadResponse, "Error reading response from FileStorage")
	ErrServiceResponseFromFileStorage = NewServiceError(CodeResponseFromFileStorage, "Error response from FileStorage")

	// Group errors
	ErrServiceGroupNotFound = NewServiceError(CodeGroupNotFound, "Group not found")

	// Pagination errors
	ErrServiceInvalidLimitParam  = NewServiceError(CodeInvalidLimitParam, "Invalid limit parameter")
	ErrServiceInvalidOffsetParam = NewServiceError(CodeInvalidOffsetParam, "Invalid offset parameter")

	// Session errors
	ErrServiceSessionNotFound     = NewServiceError(CodeSessionNotFound, "Session not found")
	ErrServiceSessionExpired      = NewServiceError(CodeSessionExpired, "Session expired")
	ErrServiceSessionUserNotFound = NewServiceError(CodeSessionUserNotFound, "Session user not found")
	ErrServiceSessionRefresh      = NewServiceError(CodeSessionRefresh, "Session refresh failed")

	// Archive errors
	ErrServiceInvalidArchive = NewServiceError(CodeInvalidArchive, "Invalid archive format")

	// Internal errors
	ErrServiceExpectedStruct = NewServiceError(CodeExpectedStruct, "Expected struct parameter")

	// Timeout errors
	ErrServiceTimeout = NewServiceError(CodeTimeout, "Operation timeout")

	// Token errors
	ErrServiceInvalidToken      = NewServiceError(CodeInvalidToken, "Invalid token")
	ErrServiceTokenExpired      = NewServiceError(CodeTokenExpired, "Token expired")
	ErrServiceTokenUserNotFound = NewServiceError(CodeTokenUserNotFound, "Token user not found")
	ErrServiceInvalidTokenType  = NewServiceError(CodeInvalidTokenType, "Invalid token type")

	// Contest registration errors
	ErrServiceContestRegistrationClosed = NewServiceError(CodeContestRegistrationClosed, "Contest registration is closed")
	ErrServiceContestEnded              = NewServiceError(CodeContestEnded, "Contest has ended")
	ErrServiceAlreadyRegistered         = NewServiceError(CodeAlreadyRegistered, "Already registered for this contest")
	ErrServiceAlreadyParticipant        = NewServiceError(CodeAlreadyParticipant, "User is already a participant of this contest")
	ErrServiceNoPendingRegistration     = NewServiceError(CodeNoPendingRegistration, "No pending registration for this contest")

	// Contest task errors
	ErrServiceTaskNotInContest = NewServiceError(CodeTaskNotInContest, "Task is not part of the contest")

	// Language errors
	ErrServiceInvalidLanguage = NewServiceError(CodeInvalidLanguage, "Invalid language for the task")

	// Contest submission errors
	ErrServiceContestSubmissionClosed = NewServiceError(CodeContestSubmissionClosed, "Contest submissions are closed")
	ErrServiceTaskSubmissionClosed    = NewServiceError(CodeTaskSubmissionClosed, "Task submissions are closed for this contest task")

	// Contest timing errors
	ErrServiceContestNotStarted   = NewServiceError(CodeContestNotStarted, "Contest has not started yet")
	ErrServiceTaskNotStarted      = NewServiceError(CodeTaskNotStarted, "Task submission period has not started yet")
	ErrServiceTaskSubmissionEnded = NewServiceError(CodeTaskSubmissionEnded, "Task submission period has ended")

	// Contest participation errors
	ErrServiceNotContestParticipant = NewServiceError(CodeNotContestParticipant, "User is not a participant of this contest")

	// Role errors
	ErrServiceCannotAssignOwner = NewServiceError(CodeCannotAssignOwner, "Cannot assign owner role to another user")

	// Internal/unknown error
	ErrServiceInternalError = NewServiceError(CodeInternalError, "Internal server error")
)

// errorToServiceError maps legacy errors to ServiceErrors for backward compatibility
var errorToServiceError = map[error]*ServiceError{
	ErrDatabaseConnection:        ErrServiceDatabaseConnection,
	ErrTaskExists:                ErrServiceTaskExists,
	ErrTaskNotFound:              ErrServiceTaskNotFound,
	ErrTaskAlreadyAssigned:       ErrServiceTaskAlreadyAssigned,
	ErrTaskNotAssignedToUser:     ErrServiceTaskNotAssignedUser,
	ErrTaskNotAssignedToGroup:    ErrServiceTaskNotAssignedGroup,
	ErrForbidden:                 ErrServiceForbidden,
	ErrNotAuthorized:             ErrServiceNotAuthorized,
	ErrNotAllowed:                ErrServiceNotAllowed,
	ErrPermissionDenied:          ErrServicePermissionDenied,
	ErrUserNotFound:              ErrServiceUserNotFound,
	ErrUserAlreadyExists:         ErrServiceUserAlreadyExists,
	ErrAccessAlreadyExists:       ErrServiceAccessAlreadyExists,
	ErrInvalidCredentials:        ErrServiceInvalidCredentials,
	ErrInvalidData:               ErrServiceInvalidData,
	ErrInvalidInputOuput:         ErrServiceInvalidInputOutput,
	ErrNotFound:                  ErrServiceNotFound,
	ErrFileOpen:                  ErrServiceFileOpen,
	ErrTempDirCreate:             ErrServiceTempDirCreate,
	ErrDecompressArchive:         ErrServiceDecompressArchive,
	ErrNoInputDirectory:          ErrServiceNoInputDirectory,
	ErrNoOutputDirectory:         ErrServiceNoOutputDirectory,
	ErrIOCountMismatch:           ErrServiceIOCountMismatch,
	ErrInputContainsDir:          ErrServiceInputContainsDir,
	ErrOutputContainsDir:         ErrServiceOutputContainsDir,
	ErrInvalidInExtention:        ErrServiceInvalidInExtention,
	ErrInvalidOutExtention:       ErrServiceInvalidOutExtention,
	ErrWriteTaskID:               ErrServiceWriteTaskID,
	ErrWriteOverwrite:            ErrServiceWriteOverwrite,
	ErrCreateFormFile:            ErrServiceCreateFormFile,
	ErrCopyFile:                  ErrServiceCopyFile,
	ErrSendRequest:               ErrServiceSendRequest,
	ErrReadResponse:              ErrServiceReadResponse,
	ErrResponseFromFileStorage:   ErrServiceResponseFromFileStorage,
	ErrGroupNotFound:             ErrServiceGroupNotFound,
	ErrInvalidLimitParam:         ErrServiceInvalidLimitParam,
	ErrInvalidOffsetParam:        ErrServiceInvalidOffsetParam,
	ErrSessionNotFound:           ErrServiceSessionNotFound,
	ErrSessionExpired:            ErrServiceSessionExpired,
	ErrSessionUserNotFound:       ErrServiceSessionUserNotFound,
	ErrSessionRefresh:            ErrServiceSessionRefresh,
	ErrInvalidArchive:            ErrServiceInvalidArchive,
	ErrExpectedStruct:            ErrServiceExpectedStruct,
	ErrTimeout:                   ErrServiceTimeout,
	ErrInvalidToken:              ErrServiceInvalidToken,
	ErrTokenExpired:              ErrServiceTokenExpired,
	ErrTokenUserNotFound:         ErrServiceTokenUserNotFound,
	ErrInvalidTokenType:          ErrServiceInvalidTokenType,
	ErrContestRegistrationClosed: ErrServiceContestRegistrationClosed,
	ErrContestEnded:              ErrServiceContestEnded,
	ErrAlreadyRegistered:         ErrServiceAlreadyRegistered,
	ErrAlreadyParticipant:        ErrServiceAlreadyParticipant,
	ErrNoPendingRegistration:     ErrServiceNoPendingRegistration,
	ErrTaskNotInContest:          ErrServiceTaskNotInContest,
	ErrInvalidLanguage:           ErrServiceInvalidLanguage,
	ErrContestSubmissionClosed:   ErrServiceContestSubmissionClosed,
	ErrTaskSubmissionClosed:      ErrServiceTaskSubmissionClosed,
	ErrContestNotStarted:         ErrServiceContestNotStarted,
	ErrTaskNotStarted:            ErrServiceTaskNotStarted,
	ErrTaskSubmissionEnded:       ErrServiceTaskSubmissionEnded,
	ErrNotContestParticipant:     ErrServiceNotContestParticipant,
	ErrCannotAssignOwner:         ErrServiceCannotAssignOwner,
}

// ToServiceError converts any error to a ServiceError.
// If the error is already a ServiceError, it returns it directly.
// If it's a legacy error, it maps it to the corresponding ServiceError.
// Otherwise, it returns an internal error.
func ToServiceError(err error) *ServiceError {
	if err == nil {
		return nil
	}

	// Check if it's already a ServiceError
	var serviceErr *ServiceError
	if AsServiceError(err, &serviceErr) {
		return serviceErr
	}

	// Check if it's a legacy error we can map
	for legacyErr, svcErr := range errorToServiceError {
		if Is(err, legacyErr) {
			return svcErr.WithError(err)
		}
	}

	// Unknown error - return internal error
	return ErrServiceInternalError.WithError(err)
}

// Is reports whether any error in err's chain matches target.
// This is a convenience wrapper around errors.Is from the standard library.
func Is(err, target error) bool {
	return stderrors.Is(err, target)
}

// AsServiceError checks if err is a ServiceError and assigns it to target.
// Returns true if successful.
func AsServiceError(err error, target **ServiceError) bool {
	return stderrors.As(err, target)
}
