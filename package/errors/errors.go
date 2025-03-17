package errors

import (
	"fmt"
)

var ErrDatabaseConnection = fmt.Errorf("failed to connect to the database")
var ErrTaskExists = fmt.Errorf("task with this title already exists")
var ErrTaskNotFound = fmt.Errorf("task not found")
var ErrNotAuthorized = fmt.Errorf("not authorized to perform this action")
var ErrTaskAlreadyAssigned = fmt.Errorf("task is already assigned to the user")
var ErrTaskNotAssignedToUser = fmt.Errorf("task is not assigned to the user")
var ErrTaskNotAssignedToGroup = fmt.Errorf("task is not assigned to the group")
var ErrNotAllowed = fmt.Errorf("not allowed to perform this action")
var ErrUserNotFound = fmt.Errorf("user not found")
var ErrUserAlreadyExists = fmt.Errorf("user already exists")
var ErrInvalidCredentials = fmt.Errorf("invalid credentials")
var ErrPermissionDenied = fmt.Errorf("user is not allowed to view this submission")
var ErrInvalidData = fmt.Errorf("invalid data")
var ErrInvalidInputOuput = fmt.Errorf("invalid input or output")
var ErrNotFound = fmt.Errorf("request resource not found")

var ErrFileOpen = fmt.Errorf("failed to open file")
var ErrTempDirCreate = fmt.Errorf("failed to create temp directory")
var ErrDecompressArchive = fmt.Errorf("failed to decompress archive")
var ErrNoInputDirectory = fmt.Errorf("no input directory found")
var ErrNoOutputDirectory = fmt.Errorf("no output directory found")
var ErrIOCountMismatch = fmt.Errorf("input and output file count mismatch")
var ErrInputContainsDir = fmt.Errorf("input contains a directory")
var ErrOutputContainsDir = fmt.Errorf("output contains a directory")
var ErrInvalidInExtention = fmt.Errorf("invalid input file extension")
var ErrInvalidOutExtention = fmt.Errorf("invalid output file extension")
