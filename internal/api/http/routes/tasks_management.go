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
	currentUser := httputils.GetCurrentUser(r)

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
		httputils.HandleServiceError(w, err, db, tr.logger)
		return
	}

	err = tr.taskService.ProcessAndUpload(tx, currentUser, taskID, filePath)
	if err != nil {
		httputils.HandleServiceError(w, err, db, tr.logger)
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, schemas.TaskCreateResponse{ID: taskID})
}

// EditTask godoc
//
//	@Tags			tasks-management
//	@Summary		Update a task
//	@Description	Updates a task by ID
//	@Consumes		multipart/form-data
//	@Produce		json
//	@Param			id		path		int		true	"Task ID"
//	@Param			title	formData	string	false	"New title for the task"
//	@Param			archive	formData	file	false	"New archive for the task"
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

	currentUser := httputils.GetCurrentUser(r)

	request := schemas.EditTask{}
	newTitle := r.FormValue("title")
	if newTitle != "" {
		request.Title = &newTitle
	}

	err = tr.taskService.Edit(tx, currentUser, taskID, &request)
	if err != nil {
		httputils.HandleServiceError(w, err, db, tr.logger)
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

	currentUser := httputils.GetCurrentUser(r)

	err = tr.taskService.Delete(tx, currentUser, taskID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, tr.logger)
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

	currentUser := httputils.GetCurrentUser(r)

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

	currentUser := httputils.GetCurrentUser(r)

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
		httputils.HandleServiceError(w, err, db, tr.logger)
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
	currentUser := httputils.GetCurrentUser(r)

	response, err := tr.taskService.GetAllCreated(tx, currentUser, paginationParams)
	if err != nil {
		httputils.HandleServiceError(w, err, db, tr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, response)
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
}

func isValidFileFormat(filename string) bool {
	return strings.HasSuffix(filename, ".zip") || strings.HasSuffix(filename, ".tar.gz")
}
