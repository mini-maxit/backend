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
	DeleteTask(w http.ResponseWriter, r *http.Request)
	EditTask(w http.ResponseWriter, r *http.Request)
	GetAllCreatedTasks(w http.ResponseWriter, r *http.Request)
	GetLimits(w http.ResponseWriter, r *http.Request)
	PutLimits(w http.ResponseWriter, r *http.Request)
	UploadTask(w http.ResponseWriter, r *http.Request)

	// Task collaborator management
	AddTaskCollaborator(w http.ResponseWriter, r *http.Request)
	GetTaskCollaborators(w http.ResponseWriter, r *http.Request)
	UpdateTaskCollaborator(w http.ResponseWriter, r *http.Request)
	RemoveTaskCollaborator(w http.ResponseWriter, r *http.Request)
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
		if errors.Is(err, myerrors.ErrForbidden) {
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
		case errors.Is(err, myerrors.ErrForbidden):
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
		if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else {
			tr.logger.Errorw("Failed to delete task", "error", err)
		}
		httputils.ReturnError(w, status, "Task service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task deleted successfully"))
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
		case errors.Is(err, myerrors.ErrForbidden):
			httputils.ReturnError(w, http.StatusForbidden, "You are not authorized to update limits for this task.")
		default:
			tr.logger.Errorw("Failed to put task limits", "error", err)
			httputils.ReturnError(w, http.StatusInternalServerError, "Task service temporarily unavailable")
		}
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task limits updated successfully"))
}

// GetAllCreatedTasks godoc
//
//	@Tags			tasks-management
//	@Summary		Get all created tasks
//	@Description	Gets all tasks created by the current user with pagination metadata
//	@Produce		json
//	@Param			offset	query		int		false	"Offset for pagination"
//	@Param			limit	query		int		false	"Limit for pagination"
//	@Param			sort	query		string	false	"Field to sort by"
//	@Failure		403		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[schemas.PaginatedResult[[]schemas.Task]]
//	@Router			/tasks-management/tasks/created [get]
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
	paginationParams := httputils.ExtractPaginationParams(queryParams)
	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	response, err := tr.taskService.GetAllCreated(tx, currentUser, paginationParams)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else {
			tr.logger.Errorw("Failed to get created tasks", "error", err)
		}
		httputils.ReturnError(w, status, "Task service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, response)
}

// AddTaskCollaborator godoc
//
//	@Tags			tasks-management
//	@Summary		Add a collaborator to a task
//	@Description	Add a user as a collaborator to a task with specified permissions (view, edit, or manage). Only users with manage permission can add collaborators.
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"Task ID"
//	@Param			body	body		schemas.AddCollaborator	true	"Collaborator details"
//	@Failure		400		{object}	httputils.ValidationErrorResponse
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/tasks-management/tasks/{id}/collaborators [post]
func (tr *tasksManagementRoute) AddTaskCollaborator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID cannot be empty")
		return
	}
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var request schemas.AddCollaborator
	err = httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			httputils.ReturnValidationError(w, valErrs)
			return
		}
		httputils.ReturnError(w, http.StatusBadRequest, "Could not validate request data.")
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

	err = tr.taskService.AddTaskCollaborator(tx, currentUser, taskID, &request)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else {
			tr.logger.Errorw("Failed to add collaborator", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to add collaborator")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Collaborator added successfully"))
}

// GetTaskCollaborators godoc
//
//	@Tags			tasks-management
//	@Summary		Get collaborators for a task
//	@Description	Get all collaborators for a specific task. Users with view permission or higher can see collaborators.
//	@Produce		json
//	@Param			id	path		int	true	"Task ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[[]schemas.Collaborator]
//	@Router			/tasks-management/tasks/{id}/collaborators [get]
func (tr *tasksManagementRoute) GetTaskCollaborators(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID cannot be empty")
		return
	}
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID")
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

	collaborators, err := tr.taskService.GetTaskCollaborators(tx, currentUser, taskID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else {
			tr.logger.Errorw("Failed to get collaborators", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to get collaborators")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, collaborators)
}

// UpdateTaskCollaborator godoc
//
//	@Tags			tasks-management
//	@Summary		Update a collaborator's permission
//	@Description	Update the permission level of a collaborator for a task. Only users with manage permission can update collaborators.
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int							true	"Task ID"
//	@Param			user_id	path		int							true	"User ID"
//	@Param			body	body		schemas.UpdateCollaborator	true	"New permission"
//	@Failure		400		{object}	httputils.ValidationErrorResponse
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/tasks-management/tasks/{id}/collaborators/{user_id} [put]
func (tr *tasksManagementRoute) UpdateTaskCollaborator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID cannot be empty")
		return
	}
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	userIDStr := httputils.GetPathValue(r, "user_id")
	if userIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "User ID cannot be empty")
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var request schemas.UpdateCollaborator
	err = httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			httputils.ReturnValidationError(w, valErrs)
			return
		}
		httputils.ReturnError(w, http.StatusBadRequest, "Could not validate request data.")
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

	err = tr.taskService.UpdateTaskCollaborator(tx, currentUser, taskID, userID, &request)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else {
			tr.logger.Errorw("Failed to update collaborator", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to update collaborator permission")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Collaborator permission updated successfully"))
}

// RemoveTaskCollaborator godoc
//
//	@Tags			tasks-management
//	@Summary		Remove a collaborator from a task
//	@Description	Remove a user's collaborator access to a task. Only users with manage permission can remove collaborators. Cannot remove the creator.
//	@Produce		json
//	@Param			id		path		int	true	"Task ID"
//	@Param			user_id	path		int	true	"User ID"
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/tasks-management/tasks/{id}/collaborators/{user_id} [delete]
func (tr *tasksManagementRoute) RemoveTaskCollaborator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID cannot be empty")
		return
	}
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	userIDStr := httputils.GetPathValue(r, "user_id")
	if userIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "User ID cannot be empty")
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid user ID")
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

	err = tr.taskService.RemoveTaskCollaborator(tx, currentUser, taskID, userID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else {
			tr.logger.Errorw("Failed to remove collaborator", "error", err)
		}
		httputils.ReturnError(w, status, err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Collaborator removed successfully"))
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
	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			route.UploadTask(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	mux.HandleFunc("/tasks/created", route.GetAllCreatedTasks)
	mux.HandleFunc("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			route.DeleteTask(w, r)
		case http.MethodPatch:
			route.EditTask(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	mux.HandleFunc("/tasks/{id}/limits", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetLimits(w, r)
		case http.MethodPut:
			route.PutLimits(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/tasks/{id}/collaborators", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetTaskCollaborators(w, r)
		case http.MethodPost:
			route.AddTaskCollaborator(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/tasks/{id}/collaborators/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			route.UpdateTaskCollaborator(w, r)
		case http.MethodDelete:
			route.RemoveTaskCollaborator(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
}

func isValidFileFormat(filename string) bool {
	return strings.HasSuffix(filename, ".zip") || strings.HasSuffix(filename, ".tar.gz")
}
