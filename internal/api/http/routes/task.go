package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/api/http/middleware"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/service"
)

type TaskRoute interface {
	GetAllTasks(w http.ResponseWriter, r *http.Request)
	GetTask(w http.ResponseWriter, r *http.Request)
	GetAllForGroup(w http.ResponseWriter, r *http.Request)
	GetAllAssingedTasks(w http.ResponseWriter, r *http.Request)
	GetAllCreatedTasks(w http.ResponseWriter, r *http.Request)
	UploadTask(w http.ResponseWriter, r *http.Request)
	UpdateTask(w http.ResponseWriter, r *http.Request)
	DeleteTask(w http.ResponseWriter, r *http.Request)
	AssignTaskToUsers(w http.ResponseWriter, r *http.Request)
	AssignTaskToGroups(w http.ResponseWriter, r *http.Request)
	UnAssignTaskFromUsers(w http.ResponseWriter, r *http.Request)
	UnAssignTaskFromGroups(w http.ResponseWriter, r *http.Request)
}

type TaskRouteImpl struct {
	fileStorageUrl string

	// Service that handles task-related operations
	taskService service.TaskService
}

func (tr *TaskRouteImpl) GetAllAssingedTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	query := r.URL.Query()
	queryParams, err := httputils.GetQueryParams(&query, httputils.TaskDefaultSortField)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusBadRequest, err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.User)

	task, err := tr.taskService.GetAllAssignedTasks(tx, current_user, queryParams)

	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error getting tasks. %s", err.Error()))
		return
	}

	if task == nil {
		task = []schemas.Task{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, task)
}

func (tr *TaskRouteImpl) GetAllCreatedTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	query := r.URL.Query()
	queryParams, err := httputils.GetQueryParams(&query, httputils.TaskDefaultSortField)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusBadRequest, err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.User)

	task, err := tr.taskService.GetAllCreatedTasks(tx, current_user, queryParams)

	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error getting tasks. %s", err.Error()))
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
//	@Failure		500	{object}	httputils.ApiError
//	@Success		200	{object}	httputils.ApiResponse[[]schemas.Task]
//	@Router			/task/ [get]
func (tr *TaskRouteImpl) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	query := r.URL.Query()
	queryParams, err := httputils.GetQueryParams(&query, httputils.TaskDefaultSortField)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusBadRequest, err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.User)

	tasks, err := tr.taskService.GetAll(tx, current_user, queryParams)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if err == errors.ErrNotAuthorized {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, fmt.Sprintf("Error getting tasks. %s", err.Error()))
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
//	@Failure		400	{object}	httputils.ApiError
//	@Failure		403	{object}	httputils.ApiError
//	@Failure		405	{object}	httputils.ApiError
//	@Failure		500	{object}	httputils.ApiError
//	@Success		200	{object}	httputils.ApiResponse[schemas.TaskDetailed]
//	@Router			/task/{id} [get]
func (tr *TaskRouteImpl) GetTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIdStr := r.PathValue("id")
	if taskIdStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.User)

	task, err := tr.taskService.GetTask(tx, current_user, taskId)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if err == errors.ErrNotAuthorized {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, fmt.Sprintf("Error getting tasks. %s", err.Error()))
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, task)
}

func (tr *TaskRouteImpl) GetAllForGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	groupIdStr := r.PathValue("id")

	if groupIdStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group id")
		return
	}

	groupId, err := strconv.ParseInt(groupIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group id")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	query := r.URL.Query()
	queryParams, err := httputils.GetQueryParams(&query, httputils.TaskDefaultSortField)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusBadRequest, err.Error())
		return
	}

	tasks, err := tr.taskService.GetAllForGroup(tx, groupId, queryParams)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error getting tasks. %s", err.Error()))
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
//	@Param			taskName	formData	string	true	"Name of the task"
//	@Param			userId		formData	int		true	"ID of the author"
//	@Param			overwrite	formData	bool	false	"Overwrite flag"
//	@Param			archive		formData	file	true	"Task archive"
//	@Failure		405			{object}	httputils.ApiError
//	@Failure		400			{object}	httputils.ApiError
//	@Failure		500			{object}	httputils.ApiError
//	@Success		200			{object}	httputils.ApiResponse[schemas.TaskCreateResponse]
//	@Router			/task/ [post]
func (tr *TaskRouteImpl) UploadTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Limit the size of the incoming request
	r.Body = http.MaxBytesReader(w, r.Body, 50*1024*1024) // 50 MB limit

	// Parse the multipart form data
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "The uploaded files are too large.")
		return
	}

	overwriteStr := r.FormValue("overwrite")
	overwrite := false
	if overwriteStr != "" {
		var err error
		overwrite, err = strconv.ParseBool(overwriteStr)
		if err != nil {
			httputils.ReturnError(w, http.StatusBadRequest, "Invalid overwrite flag.")
			return
		}
	}
	taskName := r.FormValue("taskName")
	if taskName == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task name is required.")
		return
	}
	userIdStr := r.FormValue("userId")
	if userIdStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "User ID of author is required.")
		return
	}
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid user ID.")
		return
	}

	// Extract the uploaded file
	file, handler, err := r.FormFile("archive")
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Error retrieving the file. No task file found.")
		return
	}
	if !(strings.HasSuffix(handler.Filename, ".zip") || strings.HasSuffix(handler.Filename, ".tar.gz")) {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid file format. Only .zip and .tar.gz files are allowed as task upload. Received: "+handler.Filename)
		return
	}
	defer file.Close()

	// Create a multipart writer for the HTTP request to FileStorage service
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create empty task to get the task ID
	task := schemas.Task{
		Title:     taskName,
		CreatedBy: userId,
	}
	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.User)

	taskId, err := tr.taskService.Create(tx, current_user, &task)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if err == errors.ErrNotAuthorized {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, fmt.Sprintf("Error getting tasks. %s", err.Error()))
		return
	}

	// Add form fields
	err = writer.WriteField("taskID", fmt.Sprintf("%d", taskId))
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error writing task ID to form. %s", err.Error()))
		return
	}
	err = writer.WriteField("overwrite", strconv.FormatBool(overwrite))
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error writing overwrite flag to form. %s", err.Error()))
		return
	}

	// Create a form file field and copy the uploaded file to it
	part, err := writer.CreateFormFile("archive", handler.Filename)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error creating form file for FileStorage. %s", err.Error()))
		return
	}
	if _, err := io.Copy(part, file); err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error copying file to FileStorage request. %s", err.Error()))
		return
	}
	writer.Close()

	// Send the request to FileStorage service
	client := &http.Client{}
	resp, err := client.Post(tr.fileStorageUrl+"/createTask", writer.FormDataContentType(), body)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error sending file to FileStorage service. %s", err.Error()))
		return
	}
	defer resp.Body.Close()

	// Handle response from FileStorage
	buffer := make([]byte, resp.ContentLength)
	bytesRead, err := resp.Body.Read(buffer)
	if err != nil && bytesRead == 0 {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error reading response from FileStorage. %s", err.Error()))
		return
	}
	if resp.StatusCode != http.StatusOK {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to upload file to FileStorage. %s", string(buffer)))
		return
	}

	// TODO: Update the task with the correct directories. Waiting to be implemented on the FileStorage service side
	// updateInfo := schemas.UpdateTask{
	// 	Title: taskName,
	// }
	// if err := tr.taskService.UpdateTask(taskId, updateInfo); err != nil {
	// 	utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error updating task directories. %s", err.Error()))
	// 	return
	// }

	httputils.ReturnSuccess(w, http.StatusOK, schemas.TaskCreateResponse{Id: taskId})
}

func (tr *TaskRouteImpl) UpdateTask(w http.ResponseWriter, r *http.Request) {
	httputils.ReturnError(w, http.StatusNotImplemented, "Not implemented")
}

func (tr *TaskRouteImpl) DeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIdStr := r.PathValue("id")
	if taskIdStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.User)

	err = tr.taskService.DeleteTask(tx, current_user, taskId)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if err == errors.ErrNotAuthorized {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, fmt.Sprintf("Error getting tasks. %s", err.Error()))
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, "Task deleted successfully")
}

func (tr *TaskRouteImpl) AssignTaskToUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIdStr := r.PathValue("id")
	if taskIdStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	var userIds []int64
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&userIds)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid user IDs.")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.User)

	err = tr.taskService.AssignTaskToUsers(tx, current_user, taskId, userIds)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error assigning task. %s", err.Error()))
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, "Task assigned successfully")
}

func (tr *TaskRouteImpl) AssignTaskToGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIdStr := r.PathValue("id")
	if taskIdStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	var groupIds []int64
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&groupIds)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group IDs.")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.User)

	err = tr.taskService.AssignTaskToGroups(tx, current_user, taskId, groupIds)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error assigning task. %s", err.Error()))
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, "Task assigned successfully")
}

func (tr *TaskRouteImpl) UnAssignTaskFromUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIdStr := r.PathValue("id")
	if taskIdStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	var userIds []int64
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&userIds)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid user IDs.")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.User)

	err = tr.taskService.UnAssignTaskFromUsers(tx, current_user, taskId, userIds)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if err == errors.ErrNotAuthorized {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, fmt.Sprintf("Error unassigning task. %s", err.Error()))
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, "Task unassigned successfully")
}

func (tr *TaskRouteImpl) UnAssignTaskFromGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIdStr := r.PathValue("id")
	if taskIdStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	var groupIds []int64
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&groupIds)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group IDs.")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.User)

	err = tr.taskService.UnAssignTaskFromGroups(tx, current_user, taskId, groupIds)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if err == errors.ErrNotAuthorized {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, fmt.Sprintf("Error unassigning task. %s", err.Error()))
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, "Task unassigned successfully")
}

func NewTaskRoute(fileStorageUrl string, taskService service.TaskService) TaskRoute {
	return &TaskRouteImpl{fileStorageUrl: fileStorageUrl, taskService: taskService}
}

func RegisterTaskRoutes(mux *http.ServeMux, route TaskRoute) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetAllTasks(w, r)
		case http.MethodPost:
			route.UploadTask(w, r)
		case http.MethodPut:
			route.UpdateTask(w, r)
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
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	mux.HandleFunc("/{id}/assign/users", route.AssignTaskToUsers)
	mux.HandleFunc("/{id}/assign/groups", route.AssignTaskToGroups)
	mux.HandleFunc("/{id}/unassign/users", route.UnAssignTaskFromUsers)
	mux.HandleFunc("/{id}/usassign/groups", route.UnAssignTaskFromGroups)
	mux.HandleFunc("/{id}/group", route.GetAllForGroup)
	mux.HandleFunc("/assigned", route.GetAllAssingedTasks)
	mux.HandleFunc("/created", route.GetAllCreatedTasks)
}
