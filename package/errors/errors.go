package errors

import "errors"

var ErrDatabaseConnection = errors.New("failed to connect to the database")
var ErrTaskExists = errors.New("task with this title already exists")
var ErrTaskNotFound = errors.New("task not found")
var ErrNotAuthorized = errors.New("not authorized to perform this action")
var ErrTaskAlreadyAssigned = errors.New("task is already assigned to the user")
var ErrTaskNotAssignedToUser = errors.New("task is not assigned to the user")
var ErrTaskNotAssignedToGroup = errors.New("task is not assigned to the group")
var ErrNotAllowed = errors.New("not allowed to perform this action")
var ErrUserNotFound = errors.New("user not found")
var ErrUserAlreadyExists = errors.New("user already exists")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrPermissionDenied = errors.New("user is not allowed to view this submission")
var ErrInvalidData = errors.New("invalid data")
var ErrInvalidInputOuput = errors.New("invalid input or output")
var ErrNotFound = errors.New("request resource not found")

var ErrFileOpen = errors.New("failed to open file")
var ErrTempDirCreate = errors.New("failed to create temp directory")
var ErrDecompressArchive = errors.New("failed to decompress archive")
var ErrNoInputDirectory = errors.New("no input directory found")
var ErrNoOutputDirectory = errors.New("no output directory found")
var ErrIOCountMismatch = errors.New("input and output file count mismatch")
var ErrInputContainsDir = errors.New("input contains a directory")
var ErrOutputContainsDir = errors.New("output contains a directory")
var ErrInvalidInExtention = errors.New("invalid input file extension")
var ErrInvalidOutExtention = errors.New("invalid output file extension")

var ErrWriteTaskID = errors.New("error writing task ID to form")
var ErrWriteOverwrite = errors.New("error writing overwrite to form")
var ErrCreateFormFile = errors.New("error creating form file")
var ErrCopyFile = errors.New("error copying file to form")
var ErrSendRequest = errors.New("error sending request to FileStorage")
var ErrReadResponse = errors.New("error reading response from FileStorage")
var ErrResponseFromFileStorage = errors.New("error response from FileStorage")

var (
	ErrGroupNotFound      = errors.New("group not found")
	ErrInvalidLimitParam  = errors.New("invalid limit parameter")
	ErrInvalidOffsetParam = errors.New("invalid offset parameter")
)

var (
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionExpired      = errors.New("session expired")
	ErrSessionUserNotFound = errors.New("session user not found")
	ErrSessionRefresh      = errors.New("session refresh failed")
)
var ErrInvalidArchive = errors.New(`archive contains a single file, expected a single directory or
[input/ output/ description.pdf]`)
var ErrExpectedStruct = errors.New("input param should be a struct")
