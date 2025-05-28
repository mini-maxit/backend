// Package errors contains the error variables used in the application.
package errors

import (
	"errors"
)

// ErrDatabaseConnection is returned when the database connection fails.
var ErrDatabaseConnection = errors.New("failed to connect to the database")

// ErrTaskExists is returned when a task with the same title already exists.
var ErrTaskExists = errors.New("task with this title already exists")

// ErrTaskNotFound is returned when the specified task is not found.
var ErrTaskNotFound = errors.New("task not found")

// ErrNotAuthorized is returned when the user is not authorized to perform the action.
var ErrNotAuthorized = errors.New("not authorized to perform this action")

// ErrTaskAlreadyAssigned is returned when the task is already assigned to the user.
var ErrTaskAlreadyAssigned = errors.New("task is already assigned to the user")

// ErrTaskNotAssignedToUser is returned when the task is not assigned to the user.
var ErrTaskNotAssignedToUser = errors.New("task is not assigned to the user")

// ErrTaskNotAssignedToGroup is returned when the task is not assigned to the group.
var ErrTaskNotAssignedToGroup = errors.New("task is not assigned to the group")

// ErrNotAllowed is returned when the action is not allowed.
var ErrNotAllowed = errors.New("not allowed to perform this action")

// ErrUserNotFound is returned when the specified user is not found.
var ErrUserNotFound = errors.New("user not found")

// ErrUserAlreadyExists is returned when the user already exists.
var ErrUserAlreadyExists = errors.New("user already exists")

// ErrInvalidCredentials is returned when the provided credentials are invalid.
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrPermissionDenied is returned when the user is not allowed to view the submission.
var ErrPermissionDenied = errors.New("user is not allowed to view this submission")

// ErrInvalidData is returned when the provided data is invalid.
var ErrInvalidData = errors.New("invalid data")

// ErrInvalidInputOuput is returned when the input or output is invalid.
var ErrInvalidInputOuput = errors.New("invalid input or output")

// ErrNotFound is returned when the requested resource is not found.
var ErrNotFound = errors.New("requested was resource not found")

// ErrFileOpen is returned when the file fails to open.
var ErrFileOpen = errors.New("failed to open file")

// ErrTempDirCreate is returned when the temporary directory creation fails.
var ErrTempDirCreate = errors.New("failed to create temp directory")

// ErrDecompressArchive is returned when the archive decompression fails.
var ErrDecompressArchive = errors.New("failed to decompress archive")

// ErrNoInputDirectory is returned when no input directory is found.
var ErrNoInputDirectory = errors.New("no input directory found")

// ErrNoOutputDirectory is returned when no output directory is found.
var ErrNoOutputDirectory = errors.New("no output directory found")

// ErrIOCountMismatch is returned when the input and output file count mismatch.
var ErrIOCountMismatch = errors.New("input and output file count mismatch")

// ErrInputContainsDir is returned when the input contains a directory.
var ErrInputContainsDir = errors.New("input contains a directory")

// ErrOutputContainsDir is returned when the output contains a directory.
var ErrOutputContainsDir = errors.New("output contains a directory")

// ErrInvalidInExtention is returned when the input file extension is invalid.
var ErrInvalidInExtention = errors.New("invalid input file extension")

// ErrInvalidOutExtention is returned when the output file extension is invalid.
var ErrInvalidOutExtention = errors.New("invalid output file extension")

// ErrWriteTaskID is returned when writing the task ID to the form fails.
var ErrWriteTaskID = errors.New("error writing task ID to form")

// ErrWriteOverwrite is returned when writing overwrite to the form fails.
var ErrWriteOverwrite = errors.New("error writing overwrite to form")

// ErrCreateFormFile is returned when creating the form file fails.
var ErrCreateFormFile = errors.New("error creating form file")

// ErrCopyFile is returned when copying the file to the form fails.
var ErrCopyFile = errors.New("error copying file to form")

// ErrSendRequest is returned when sending the request to FileStorage fails.
var ErrSendRequest = errors.New("error sending request to FileStorage")

// ErrReadResponse is returned when reading the response from FileStorage fails.
var ErrReadResponse = errors.New("error reading response from FileStorage")

// ErrResponseFromFileStorage is returned when there is an error response from FileStorage.
var ErrResponseFromFileStorage = errors.New("error response from FileStorage")

var (
	// ErrGroupNotFound is returned when the specified group is not found.
	ErrGroupNotFound = errors.New("group not found")

	// ErrInvalidLimitParam is returned when the limit parameter is invalid.
	ErrInvalidLimitParam = errors.New("invalid limit parameter")

	// ErrInvalidOffsetParam is returned when the offset parameter is invalid.
	ErrInvalidOffsetParam = errors.New("invalid offset parameter")
)

var (
	// ErrSessionNotFound is returned when the session is not found.
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionExpired is returned when the session has expired.
	ErrSessionExpired = errors.New("session expired")

	// ErrSessionUserNotFound is returned when the session user is not found.
	ErrSessionUserNotFound = errors.New("session user not found")

	// ErrSessionRefresh is returned when the session refresh fails.
	ErrSessionRefresh = errors.New("session refresh failed")
)

// ErrInvalidArchive is returned when the archive contains a single file, expected a single directory or.
var ErrInvalidArchive = errors.New(`archive contains a single file, expected a single directory or
[input/ output/ description.pdf]`)

// ErrExpectedStruct is returned when the input parameter should be a struct.
var ErrExpectedStruct = errors.New("input param should be a struct")

// ErrTimeout is returned when the operation times out.
var ErrTimeout = errors.New("timeout waiting for response")

// ErrInvalidToken is returned when the token is invalid.
var ErrInvalidToken = errors.New("invalid token")

// ErrInvalidTokenType is returned when the token type is invalid.
var ErrTokenExpired = errors.New("token expired")

// ErrTokenUserNotFound is returned when the user associated with the token is not found.
var ErrTokenUserNotFound = errors.New("token user not found")

// ErrInvalidTokenType is returned when the token type is invalid.
var ErrInvalidTokenType = errors.New("invalid token type")
