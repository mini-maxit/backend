package routes

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	myerrors "github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
)

type TasksManagementRoute interface {
	AssignTaskToGroups(w http.ResponseWriter, r *http.Request)
	AssignTaskToUsers(w http.ResponseWriter, r *http.Request)
	DeleteTask(w http.ResponseWriter, r *http.Request)
	EditTask(w http.ResponseWriter, r *http.Request)
	GetAllCreatedTasks(w http.ResponseWriter, r *http.Request)
	GetLimits(w http.ResponseWriter, r *http.Request)
	PutLimits(w http.ResponseWriter, r *http.Request)
	UnAssignTaskFromGroups(w http.ResponseWriter, r *http.Request)
	UnAssignTaskFromUsers(w http.ResponseWriter, r *http.Request)
	UploadTask(w http.ResponseWriter, r *http.Request)
}

type tasksManagementRoute struct {
	taskService service.TaskService
	logger      *zap.SugaredLogger
}

// UploadTask godoc
//
//	@Tags			tasks-management
//	@Summary		Upload a task
//	@Description	Uploads a task to the FileStorage service
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			title	formData	string	true	"Name of the task"
//	@Param			archive	formData	file	true	"Task archive"
//	@Failure		405		{object}	httputils.APIError
//	@Failure		400		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[schemas.TaskCreateResponse]
//	@Router			/tasks-management/tasks/ [post]
func (tr *tasksManagementRoute) UploadTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Limit the size of the incoming request
	r.Body = http.MaxBytesReader(w, r.Body, taskBodyLimit) // 50 MB limit

	// Parse the multipart form data
	if err := r.ParseMultipartForm(taskBodyLimit); err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "The uploaded files are too large.")
		return
	}

	title := r.FormValue("title")
	if title == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task name is required.")
		return
	}

	// Extract the uploaded file
	file, handler, err := r.FormFile("archive")
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "No task file found in request")
		return
	}
	if !strings.HasSuffix(handler.Filename, ".zip") && !strings.HasSuffix(handler.Filename, ".tar.gz") {
		httputils.ReturnError(w,
			http.StatusBadRequest,
			"Invalid file format. Only .zip and .tar.gz files are allowed as task upload. Received: "+handler.Filename,
		)
		return
	}
	defer file.Close()
	// Save the file
	filePath, err := httputils.SaveMultiPartFile(file, handler)
	if err != nil {
		tr.logger.Errorw("Failed to save multipart file", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "File upload service temporarily unavailable")
		return
	}

	// Create a multipart writer for the HTTP request to FileStorage service
	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	// Create empty task to get the task ID
	task := schemas.Task{
		Title:     title,
		CreatedBy: currentUser.ID,
	}
	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	taskID, err := tr.taskService.Create(tx, currentUser, &task)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			tr.logger.Errorw("Failed to upload task", "error", err)
		}
		httputils.ReturnError(w, status, "Task upload service temporarily unavailable")
		return
	}

	err = tr.taskService.ProcessAndUpload(tx, currentUser, taskID, filePath)
	if err != nil {
		db.Rollback()
		tr.logger.Errorw("Failed to process and upload task", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Task processing service temporarily unavailable")
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, schemas.TaskCreateResponse{ID: taskID})
}

// EditTask godoc
//
//	@Tags			tasks-management
//	@Summary		Update a task
//	@Description	Updates a task by ID
//	@Produce		json
//	@Param			id		path		int		true	"Task ID"
//	@Param			title	formData	string	false	"New title for the task"
//
//	@Param			archive	formData	file	false	"New archive for the task"
//
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/tasks-management/tasks/{id} [patch]
func (tr *tasksManagementRoute) EditTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	request := schemas.EditTask{}
	newTitle := r.FormValue("title")
	if newTitle != "" {
		request.Title = &newTitle
	}

	err = tr.taskService.Edit(tx, currentUser, taskID, &request)
	if err != nil {
		db.Rollback()
		switch {
		case errors.Is(err, myerrors.ErrNotAuthorized):
			httputils.ReturnError(w, http.StatusForbidden, "You are not authorized to edit this task.")
		case errors.Is(err, myerrors.ErrNotFound):
			httputils.ReturnError(w, http.StatusNotFound, "Task not found.")
		default:
			tr.logger.Errorw("Failed to edit task", "error", err)
			httputils.ReturnError(w, http.StatusInternalServerError, "Task service temporarily unavailable")
		}
		return
	}

	file, handler, err := r.FormFile("archive")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task updated successfully"))
			return
		}
		httputils.ReturnError(w, http.StatusBadRequest, "Error retrieving the file. No task file found.")
		return
	}

	defer file.Close()

	// Validate file format
	if !isValidFileFormat(handler.Filename) {
		httputils.ReturnError(w, http.StatusBadRequest,
			"Invalid file format. Only .zip and .tar.gz files are allowed as task upload. Received: "+handler.Filename)
		return
	}

	// Save the file
	filePath, err := httputils.SaveMultiPartFile(file, handler)
	if err != nil {
		tr.logger.Errorw("Failed to save multipart file", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "File upload service temporarily unavailable")
		return
	}

	// Process and upload
	if err := tr.taskService.ProcessAndUpload(tx, currentUser, taskID, filePath); err != nil {
		db.Rollback()
		tr.logger.Errorw("Failed to process and upload task", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Task processing service temporarily unavailable")
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task updated successfully"))
}

// DeleteTask godoc
//
//	@Tags			tasks-management
//	@Summary		Delete a task
//	@Description	Deletes a task by ID
//	@Produce		json
//	@Param			id	path		int	true	"Task ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/tasks-management/tasks/{id} [delete]
func (tr *tasksManagementRoute) DeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = tr.taskService.Delete(tx, currentUser, taskID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			tr.logger.Errorw("Failed to delete task", "error", err)
		}
		httputils.ReturnError(w, status, "Task service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task deleted successfully"))
}

// AssignTaskToUsers godoc
//
//	@Tags			tasks-management
//	@Summary		Assign a task to users
//	@Description	Assigns a task to users by task ID and user IDs
//	@Produce		json
//	@Param			id		path		int		true	"Task ID"
//	@Param			userIDs	body		[]int	true	"User IDs"
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/tasks-management/tasks/{id}/assign/users [post]
func (tr *tasksManagementRoute) AssignTaskToUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	request := usersRequest{}
	err = httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			httputils.ReturnValidationError(w, valErrs)
			return
		}
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = tr.taskService.AssignToUsers(tx, currentUser, taskID, request.UserIDs)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			tr.logger.Errorw("Failed to assign task to users", "error", err)
		}
		httputils.ReturnError(w, status, "Task assignment service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task assigned successfully"))
}

// AssignTaskToGroups godoc
//
//	@Tags			tasks-management
//	@Summary		Assign a task to groups
//	@Description	Assigns a task to groups by task ID and group IDs
//	@Produce		json
//	@Param			id			path		int		true	"Task ID"
//	@Param			groupIDs	body		[]int	true	"Group IDs"
//	@Failure		400			{object}	httputils.APIError
//	@Failure		403			{object}	httputils.APIError
//	@Failure		405			{object}	httputils.APIError
//	@Failure		500			{object}	httputils.APIError
//	@Success		200			{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/tasks-management/tasks/{id}/assign/groups [post]
func (tr *tasksManagementRoute) AssignTaskToGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	request := groupsRequest{}
	err = httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			httputils.ReturnValidationError(w, valErrs)
			return
		}
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = tr.taskService.AssignToGroups(tx, currentUser, taskID, request.GroupIDs)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			tr.logger.Errorw("Failed to assign task to groups", "error", err)
		}
		httputils.ReturnError(w, status, "Task assignment service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task assigned successfully"))
}

// UnAssignTaskFromUsers godoc
//
//	@Tags			tasks-management
//	@Summary		Unassign a task from users
//	@Description	Unassigns a task from users by task ID and user IDs
//	@Produce		json
//	@Param			id		path		int		true	"Task ID"
//	@Param			userIDs	body		[]int	true	"User IDs"
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/tasks-management/tasks/{id}/unassign/users [post]
func (tr *tasksManagementRoute) UnAssignTaskFromUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	request := usersRequest{}
	err = httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			httputils.ReturnValidationError(w, valErrs)
			return
		}
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = tr.taskService.UnassignFromUsers(tx, currentUser, taskID, request.UserIDs)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			tr.logger.Errorw("Failed to unassign task from users", "error", err)
		}
		httputils.ReturnError(w, status, "Task unassignment service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task unassigned successfully"))
}

// UnAssignTaskFromGroups godoc
//
//	@Tags			tasks-management
//	@Summary		Unassign a task from groups
//	@Description	Unassigns a task from groups by task ID and group IDs
//	@Produce		json
//	@Param			id			path		int		true	"Task ID"
//	@Param			groupIDs	body		[]int	true	"Group IDs"
//	@Failure		400			{object}	httputils.APIError
//	@Failure		403			{object}	httputils.APIError
//	@Failure		405			{object}	httputils.APIError
//	@Failure		500			{object}	httputils.APIError
//	@Success		200			{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/tasks-management/tasks/{id}/unassign/groups [post]
func (tr *tasksManagementRoute) UnAssignTaskFromGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	request := groupsRequest{}
	err = httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			httputils.ReturnValidationError(w, valErrs)
			return
		}
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = tr.taskService.UnassignFromGroups(tx, currentUser, taskID, request.GroupIDs)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			tr.logger.Errorw("Failed to unassign task from groups", "error", err)
		}
		httputils.ReturnError(w, status, "Task unassignment service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task unassigned successfully"))
}

// GetLimits godoc
//
//	@Tags			tasks-management
//	@Summary		Gets task limits
//	@Description	Gets task limits by task ID
//	@Produce		json
//	@Param			id	path		int	true	"Task ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[[]schemas.TestCase]
//	@Router			/tasks-management/tasks/{id}/limits [get]
func (tr *tasksManagementRoute) GetLimits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	limits, err := tr.taskService.GetLimits(tx, currentUser, taskID)
	if err != nil {
		switch {
		case errors.Is(err, myerrors.ErrNotFound):
			httputils.ReturnError(w, http.StatusNotFound, "Task not found.")
		default:
			tr.logger.Errorw("Failed to get task limits", "error", err)
			httputils.ReturnError(w, http.StatusInternalServerError, "Task service temporarily unavailable")
		}
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, limits)
}

// PutLimits godoc
//
//	@Tags			tasks-management
//	@Summary		Updates task limits
//	@Description	Updates task limits by task ID
//	@Produce		json
//	@Param			id		path		int									true	"Task ID"
//	@Param			limits	body		schemas.PutTestCaseLimitsRequest	true	"Task limits"
//	@Failure		400		{object}	httputils.ValidationErrorResponse
//	@Failure		404		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/tasks-management/tasks/{id}/limits [put]
func (tr *tasksManagementRoute) PutLimits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	request := schemas.PutTestCaseLimitsRequest{}
	err = httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			httputils.ReturnValidationError(w, valErrs)
			return
		}
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	err = tr.taskService.PutLimits(tx, currentUser, taskID, request)
	if err != nil {
		db.Rollback()
		switch {
		case errors.Is(err, myerrors.ErrNotFound):
			httputils.ReturnError(w, http.StatusNotFound, "Task not found.")
		case errors.Is(err, myerrors.ErrNotAuthorized):
			httputils.ReturnError(w, http.StatusForbidden, "You are not authorized to update limits for this task.")
		default:
			tr.logger.Errorw("Failed to put task limits", "error", err)
			httputils.ReturnError(w, http.StatusInternalServerError, "Task service temporarily unavailable")
		}
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task limits updated successfully"))
}

func (tr *tasksManagementRoute) GetAllCreatedTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	task, err := tr.taskService.GetAllCreated(tx, currentUser, queryParams)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			tr.logger.Errorw("Failed to get created tasks", "error", err)
		}
		httputils.ReturnError(w, status, "Task service temporarily unavailable")
		return
	}

	if task == nil {
		task = []schemas.Task{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, task)
}

func NewTasksManagementRoute(taskService service.TaskService) TasksManagementRoute {
	route := &tasksManagementRoute{
		taskService: taskService,
		logger:      utils.NewNamedLogger("tasks-management-route"),
	}

	if err := utils.ValidateStruct(*route); err != nil {
		panic(err)
	}
	return route
}

func RegisterTasksManagementRoutes(mux *mux.Router, route TasksManagementRoute) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			route.UploadTask(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	mux.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			route.DeleteTask(w, r)
		case http.MethodPatch:
			route.EditTask(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	mux.HandleFunc("/{id}/assign/users", route.AssignTaskToUsers)
	mux.HandleFunc("/{id}/assign/groups", route.AssignTaskToGroups)
	mux.HandleFunc("/{id}/unassign/users", route.UnAssignTaskFromUsers)
	mux.HandleFunc("/{id}/unassign/groups", route.UnAssignTaskFromGroups)
	mux.HandleFunc("/{id}/limits", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetLimits(w, r)
		case http.MethodPut:
			route.PutLimits(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	mux.HandleFunc("/created", route.GetAllCreatedTasks)
}

func isValidFileFormat(filename string) bool {
	return strings.HasSuffix(filename, ".zip") || strings.HasSuffix(filename, ".tar.gz")
}
