package routes

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	myerrors "github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
)

type TaskRoute interface {
	AssignTaskToGroups(w http.ResponseWriter, r *http.Request)
	AssignTaskToUsers(w http.ResponseWriter, r *http.Request)
	DeleteTask(w http.ResponseWriter, r *http.Request)
	EditTask(w http.ResponseWriter, r *http.Request)
	GetAllAssignedTasks(w http.ResponseWriter, r *http.Request)
	GetAllCreatedTasks(w http.ResponseWriter, r *http.Request)
	GetAllForGroup(w http.ResponseWriter, r *http.Request)
	GetAllTasks(w http.ResponseWriter, r *http.Request)
	GetTask(w http.ResponseWriter, r *http.Request)
	GetLimits(w http.ResponseWriter, r *http.Request)
	PutLimits(w http.ResponseWriter, r *http.Request)
	UnAssignTaskFromGroups(w http.ResponseWriter, r *http.Request)
	UnAssignTaskFromUsers(w http.ResponseWriter, r *http.Request)
	UploadTask(w http.ResponseWriter, r *http.Request)
}

const taskBodyLimit = 50 << 20

type taskRoute struct {
	fileStorageURL string

	// Service that handles task-related operations
	taskService service.TaskService
}

type usersRequest struct {
	UserIDs []int64 `json:"userIDs"`
}

type groupsRequest struct {
	GroupIDs []int64 `json:"groupIDs"`
}

func (tr *taskRoute) GetAllAssignedTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	task, err := tr.taskService.GetAllAssigned(tx, currentUser, queryParams)

	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, "Error getting tasks. "+err.Error())
		return
	}

	if task == nil {
		task = []schemas.Task{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, task)
}

func (tr *taskRoute) GetAllCreatedTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
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
		}
		httputils.ReturnError(w, status, "Error getting tasks. "+err.Error())
		return
	}

	if task == nil {
		task = []schemas.Task{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, task)
}

// GetAllTasks godoc
//
//	@Tags			task
//	@Summary		Get all tasks
//	@Description	Returns all tasks
//	@Produce		json
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[[]schemas.Task]
//	@Router			/task/ [get]
func (tr *taskRoute) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)
	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	tasks, err := tr.taskService.GetAll(tx, currentUser, queryParams)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, "Error getting tasks. "+err.Error())
		return
	}

	if tasks == nil {
		tasks = []schemas.Task{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, tasks)
}

// GetTask godoc
//
//	@Tags			task
//	@Summary		Get a task
//	@Description	Returns a task by ID
//	@Produce		json
//	@Param			id	path		int	true	"Task ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[schemas.TaskDetailed]
//	@Router			/task/{id} [get]
func (tr *taskRoute) GetTask(w http.ResponseWriter, r *http.Request) {
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
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	task, err := tr.taskService.Get(tx, currentUser, taskID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, myerrors.ErrNotAuthorized):
			status = http.StatusForbidden
		case errors.Is(err, myerrors.ErrNotFound):
			status = http.StatusNotFound
		}
		httputils.ReturnError(w, status, "Error getting task. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, task)
}

// GetAllForGroup godoc
//
//	@Tags			task
//	@Summary		Get all tasks for a group
//	@Description	Returns all tasks for a group by ID
//	@Produce		json
//	@Param			id	path		int	true	"Group ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[[]schemas.Task]
//	@Router			/task/group/{id} [get]
func (tr *taskRoute) GetAllForGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	groupIDStr := httputils.GetPathValue(r, "id")

	if groupIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group id")
		return
	}

	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group id")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	tasks, err := tr.taskService.GetAllForGroup(tx, currentUser, groupID, queryParams)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, "Error getting tasks. "+err.Error())
		return
	}

	if tasks == nil {
		tasks = []schemas.Task{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, tasks)
}

// UploadTask godoc
//
//	@Tags			task
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
//	@Router			/task/ [post]
func (tr *taskRoute) UploadTask(w http.ResponseWriter, r *http.Request) {
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
		httputils.ReturnError(w, http.StatusBadRequest, "Error retrieving the file. No task file found."+err.Error())
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
	filePath, err := httputils.SaveMultiPartFile(file, handler)
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Error saving multipart file. "+err.Error())
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
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	taskID, err := tr.taskService.Create(tx, currentUser, &task)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, "Error getting tasks. "+err.Error())
		return
	}

	err = tr.taskService.ProcessAndUpload(tx, currentUser, taskID, filePath)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, "Error processing and uploading task. "+err.Error())
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, schemas.TaskCreateResponse{ID: taskID})
}

// EditTask godoc
//
//	@Tags			task
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
//	@Router			/task/{id} [patch]
func (tr *taskRoute) EditTask(w http.ResponseWriter, r *http.Request) {
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
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
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
			httputils.ReturnError(w, http.StatusInternalServerError, "Error getting tasks. "+err.Error())
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
		httputils.ReturnError(w, http.StatusInternalServerError, "Error saving multipart file. "+err.Error())
		return
	}

	// Process and upload
	if err := tr.taskService.ProcessAndUpload(tx, currentUser, taskID, filePath); err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, "Error processing and uploading task. "+err.Error())
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task updated successfully"))
}

// DeleteTask godoc
//
//	@Tags			task
//	@Summary		Delete a task
//	@Description	Deletes a task by ID
//	@Produce		json
//	@Param			id	path		int	true	"Task ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/task/{id} [delete]
func (tr *taskRoute) DeleteTask(w http.ResponseWriter, r *http.Request) {
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
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = tr.taskService.Delete(tx, currentUser, taskID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, "Error getting tasks. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task deleted successfully"))
}

// AssignTaskToUsers godoc
//
//	@Tags			task
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
//	@Router			/task/{id}/assign/users [post]
func (tr *taskRoute) AssignTaskToUsers(w http.ResponseWriter, r *http.Request) {
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
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = tr.taskService.AssignToUsers(tx, currentUser, taskID, request.UserIDs)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, "Error assigning task. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task assigned successfully"))
}

// AssignTaskToGroups godoc
//
//	@Tags			task
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
//	@Router			/task/{id}/assign/groups [post]
func (tr *taskRoute) AssignTaskToGroups(w http.ResponseWriter, r *http.Request) {
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
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = tr.taskService.AssignToGroups(tx, currentUser, taskID, request.GroupIDs)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, "Error assigning task. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task assigned successfully"))
}

// UnAssignTaskFromUsers godoc
//
//	@Tags			task
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
//	@Router			/task/{id}/unassign/users [delete]
func (tr *taskRoute) UnAssignTaskFromUsers(w http.ResponseWriter, r *http.Request) {
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

	request := usersRequest{}
	err = httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = tr.taskService.UnassignFromUsers(tx, currentUser, taskID, request.UserIDs)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, "Error unassigning task. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task unassigned successfully"))
}

// UnAssignTaskFromGroups godoc
//
//	@Tags			task
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
//	@Router			/task/{id}/unassign/groups [delete]
func (tr *taskRoute) UnAssignTaskFromGroups(w http.ResponseWriter, r *http.Request) {
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

	request := groupsRequest{}
	err = httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = tr.taskService.UnassignFromGroups(tx, currentUser, taskID, request.GroupIDs)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, "Error unassigning task. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task unassigned successfully"))
}

// GetLimits godoc
//
//	@Tags			task
//	@Summary		Gets task limits
//	@Description	Gets task limits by task ID
//	@Produce		json
//	@Param			id	path		int	true	"Task ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/task/{id}/limits [get]
func (tr *taskRoute) GetLimits(w http.ResponseWriter, r *http.Request) {
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
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	limits, err := tr.taskService.GetLimits(tx, currentUser, taskID)
	if err != nil {
		switch {
		case errors.Is(err, myerrors.ErrNotFound):
			httputils.ReturnError(w, http.StatusNotFound, "Task not found.")
		default:
			httputils.ReturnError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, limits)
}

// PutLimits godoc
//
//	@Tags			task
//	@Summary		Updates task limits
//	@Description	Updates task limits by task ID
//	@Produce		json
//	@Param			id		path		int								true	"Task ID"
//	@Param			limits	body		schemas.PutTestCaseLimitsRequest	true	"Task limits"
//	@Failure		400		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/task/{id}/limits [put]
func (tr *taskRoute) PutLimits(w http.ResponseWriter, r *http.Request) {
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
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	request := schemas.PutTestCaseLimitsRequest{}
	err = httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
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
			httputils.ReturnError(w, http.StatusInternalServerError, "Error puting limits. "+err.Error())
		}
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task limits updated successfully"))
}

func NewTaskRoute(fileStorageURL string, taskService service.TaskService) TaskRoute {
	route := &taskRoute{fileStorageURL: fileStorageURL, taskService: taskService}

	err := utils.ValidateStruct(*route)
	if err != nil {
		log.Panicf("TaskRoute struct is not valid: %s", err.Error())
	}

	return route
}

func isValidFileFormat(filename string) bool {
	return strings.HasSuffix(filename, ".zip") || strings.HasSuffix(filename, ".tar.gz")
}

func RegisterTaskRoutes(mux *mux.Router, route TaskRoute) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetAllTasks(w, r)
		case http.MethodPost:
			route.UploadTask(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	mux.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetTask(w, r)
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
	// mux.HandleFunc("/group/{id}", route.GetAllForGroup)
	mux.HandleFunc("/assigned", route.GetAllAssignedTasks)
	mux.HandleFunc("/created", route.GetAllCreatedTasks)
}
