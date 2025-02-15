package errors

import "fmt"

var ErrDatabaseConnection = fmt.Errorf("failed to connect to the database")
var ErrTaskExists = fmt.Errorf("task with this title already exists")
var ErrTaskNotFound = fmt.Errorf("task not found")
var ErrNotAuthorized = fmt.Errorf("not authorized to perform this action")
var ErrTaskAlreadyAssigned = fmt.Errorf("task is already assigned to the user")
var ErrTaskNotAssigned = fmt.Errorf("task is not assigned to the user")