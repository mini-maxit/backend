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

	"github.com/mini-maxit/backend/internal/api/http/middleware"
	"github.com/mini-maxit/backend/internal/api/http/utils"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/service"
)

type TaskRoute interface {
	GetAllTasks(w http.ResponseWriter, r *http.Request)
	GetTask(w http.ResponseWriter, r *http.Request)
	GetAllForUser(w http.ResponseWriter, r *http.Request)
	GetAllForGroup(w http.ResponseWriter, r *http.Request)
	UploadTask(w http.ResponseWriter, r *http.Request)
	SubmitSolution(w http.ResponseWriter, r *http.Request)
}

type TaskRouteImpl struct {
	fileStorageUrl string

	// Service that handles task-related operations
	taskService  service.TaskService
	queueService service.QueueService
}

// GetAllTasks godoc
//
//	@Tags			task
//	@Summary		Get all tasks
//	@Description	Returns all tasks
//	@Produce		json
//	@Failure		500	{object}	utils.ApiResponse[schemas.ErrorResponse]
//	@Success		200	{object}	utils.ApiResponse[[]schemas.Task]
//	@Router			/tasks [get]
func (tr *TaskRouteImpl) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	query := r.URL.Query()
	limitStr := query.Get("limit")
	if limitStr == "" {
		limitStr = utils.DefaultPaginationLimitStr
	}

	offsetStr := query.Get("offset")
	if offsetStr == "" {
		offsetStr = utils.DefaultPaginationOffsetStr
	}

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid limit.")
		return
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid offset.")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	tasks, err := tr.taskService.GetAll(tx, limit, offset)
	if err != nil {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error getting tasks. %s", err.Error()))
		return
	}

	if tasks == nil {
		tasks = []schemas.Task{}
	}

	utils.ReturnSuccess(w, http.StatusOK, tasks)
}

// GetTask godoc
//
//	@Tags			task
//	@Summary		Get a task
//	@Description	Returns a task by ID
//	@Produce		json
//	@Param			id	path		int	true	"Task ID"
//	@Failure		400	{object}	utils.ApiResponse[schemas.ErrorResponse]
//	@Failure		405	{object}	utils.ApiResponse[schemas.ErrorResponse]
//	@Failure		500	{object}	utils.ApiResponse[schemas.ErrorResponse]
//	@Success		200	{object}	utils.ApiResponse[schemas.TaskDetailed]
//	@Router			/task/{id} [get]
func (tr *TaskRouteImpl) GetTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIdStr := r.PathValue("id")
	if taskIdStr == "" {
		utils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	task, err := tr.taskService.GetTask(tx, taskId)
	if err != nil {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error getting task. %s", err.Error()))
		return
	}

	utils.ReturnSuccess(w, http.StatusOK, task)
}

func (tr *TaskRouteImpl) GetAllForUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userIdStr := r.PathValue("id")

	if userIdStr == "" {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid user id")
		return
	}

	query := r.URL.Query()
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	if limitStr == "" {
		limitStr = utils.DefaultPaginationLimitStr
	}

	if offsetStr == "" {
		offsetStr = utils.DefaultPaginationOffsetStr
	}

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid offset")
		return
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid offset")
		return
	}

	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid user id")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
	}

	tasks, err := tr.taskService.GetAllForUser(tx, userId, limit, offset)
	if err != nil {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error getting tasks. %s", err.Error()))
		return
	}

	if tasks == nil {
		tasks = []schemas.Task{}
	}

	utils.ReturnSuccess(w, http.StatusOK, tasks)
}

func (tr *TaskRouteImpl) GetAllForGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	groupIdStr := r.PathValue("id")

	if groupIdStr == "" {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid group id")
		return
	}

	query := r.URL.Query()
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	if limitStr == "" {
		limitStr = utils.DefaultPaginationLimitStr
	}

	if offsetStr == "" {
		offsetStr = utils.DefaultPaginationOffsetStr
	}

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid limit")
		return
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid offset")
		return
	}

	groupId, err := strconv.ParseInt(groupIdStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid group id")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	tasks, err := tr.taskService.GetAllForGroup(tx, groupId, limit, offset)
	if err != nil {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error getting tasks. %s", err.Error()))
		return
	}

	if tasks == nil {
		tasks = []schemas.Task{}
	}

	utils.ReturnSuccess(w, http.StatusOK, tasks)
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
//	@Failure		405			{object}	utils.ApiResponse[schemas.ErrorResponse]
//	@Failure		400			{object}	utils.ApiResponse[schemas.ErrorResponse]
//	@Failure		500			{object}	utils.ApiResponse[schemas.ErrorResponse]
//	@Success		200			{object}	utils.ApiResponse[schemas.SuccessResponse]
//	@Router			/task [post]
func (tr *TaskRouteImpl) UploadTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Limit the size of the incoming request
	r.Body = http.MaxBytesReader(w, r.Body, 50*1024*1024) // 50 MB limit

	// Parse the multipart form data
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "The uploaded files are too large.")
		return
	}

	overwriteStr := r.FormValue("overwrite")
	overwrite := false
	if overwriteStr != "" {
		var err error
		overwrite, err = strconv.ParseBool(overwriteStr)
		if err != nil {
			utils.ReturnError(w, http.StatusBadRequest, "Invalid overwrite flag.")
			return
		}
	}
	taskName := r.FormValue("taskName")
	if taskName == "" {
		utils.ReturnError(w, http.StatusBadRequest, "Task name is required.")
		return
	}
	userIdStr := r.FormValue("userId")
	if userIdStr == "" {
		utils.ReturnError(w, http.StatusBadRequest, "User ID of author is required.")
		return
	}
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid user ID.")
		return
	}

	// Extract the uploaded file
	file, handler, err := r.FormFile("archive")
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Error retrieving the file. No task file found.")
		return
	}
	if !(strings.HasSuffix(handler.Filename, ".zip") || strings.HasSuffix(handler.Filename, ".tar.gz")) {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid file format. Only .zip and .tar.gz files are allowed as task upload. Received: "+handler.Filename)
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
		utils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}
	taskId, err := tr.taskService.Create(tx, &task)
	if err != nil {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error creating empty task. %s", err.Error()))
		return
	}

	// Add form fields
	writer.WriteField("taskID", fmt.Sprintf("%d", taskId))
	writer.WriteField("overwrite", strconv.FormatBool(overwrite))

	// Create a form file field and copy the uploaded file to it
	part, err := writer.CreateFormFile("archive", handler.Filename)
	if err != nil {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error creating form file for FileStorage. %s", err.Error()))
		return
	}
	if _, err := io.Copy(part, file); err != nil {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error copying file to FileStorage request. %s", err.Error()))
		return
	}
	writer.Close()

	// Send the request to FileStorage service
	client := &http.Client{}
	resp, err := client.Post(tr.fileStorageUrl+"/createTask", writer.FormDataContentType(), body)
	if err != nil {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error sending file to FileStorage service. %s", err.Error()))
		return
	}
	defer resp.Body.Close()

	// Handle response from FileStorage
	buffer := make([]byte, resp.ContentLength)
	bytesRead, err := resp.Body.Read(buffer)
	if err != nil && bytesRead == 0 {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error reading response from FileStorage. %s", err.Error()))
		return
	}
	if resp.StatusCode != http.StatusOK {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to upload file to FileStorage. %s", string(buffer)))
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

	utils.ReturnSuccess(w, http.StatusOK, map[string]interface{}{"taskId": taskId})
}

// SubmitSolution godoc
//
// @Tags	task
// @Summary Submit a solution
// @Description Uploads a solution to the FileSotrage servide
// @Accept multipart/form-data
// @Param taskID formData int true "ID of the task"
// @Param solution formData file true "solution file"
// @Param userID formData int true "ID of the user"
// @Param languageID formData int true "Id of the language used in the soloution"
// @Failure		405		{object}	utils.ApiResponse[schemas.ErrorResponse]
// @Failure		400		{object}	utils.ApiResponse[schemas.ErrorResponse]
// @Failure		500		{object}	utils.ApiResponse[schemas.ErrorResponse]
// @Success 		200 	{object}    utils.ApiResponse[schemas.SuccessResponse]
//
//	@Router			/task/submit [post]
func (tr *TaskRouteImpl) SubmitSolution(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Limit the size of the incoming request to 10 MB
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)

	// Parse the multipart form data
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "The uploaded files are too large.")
		return
	}

	// Extract the task ID
	taskIdStr := r.FormValue("taskID")
	if taskIdStr == "" {
		utils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	// Extract the uploaded file
	file, handler, err := r.FormFile("solution")
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Error retrieving the file. No solution file found.")
		return
	}
	defer file.Close()

	// Extract user ID
	userIDStr := r.FormValue("userID")
	if userIDStr == "" {
		utils.ReturnError(w, http.StatusBadRequest, "User ID is required.")
		return
	}
	userId, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid user ID.")
		return
	}

	// Extract language
	languageStr := r.FormValue("languageID")
	if languageStr == "" {
		utils.ReturnError(w, http.StatusBadRequest, "Language ID is required.")
		return
	}
	languageId, err := strconv.ParseInt(languageStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid language ID.")
		return
	}

	// Create a multipart writer for the HTTP request to FileStorage service
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	writer.WriteField("taskID", taskIdStr)
	writer.WriteField("userID", userIDStr)

	// Create a form file field and copy the uploaded file to it
	part, err := writer.CreateFormFile("submissionFile", handler.Filename)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Error creating form file for FileStorage. "+err.Error())
		return
	}
	if _, err := io.Copy(part, file); err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Error copying file to FileStorage request. "+err.Error())
		return
	}

	writer.Close()

	// Send the request to FileStorage service
	client := &http.Client{}
	resp, err := client.Post(tr.fileStorageUrl+"/submit", writer.FormDataContentType(), body)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error sending file to FileStorage service. %s", err.Error()))
		return
	}
	defer resp.Body.Close()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil && len(resBody) == 0 {
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error reading response from FileStorage. %s", err.Error()))
		return
	}
	// Handle response from FileStorage
	if resp.StatusCode != http.StatusOK {
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to upload file to FileStorage. %s", string(resBody)))
		return
	}

	respJson := schemas.SubmitResponse{}
	err = json.Unmarshal(resBody, &respJson)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error parsing response from FileStorage. %s", err.Error()))
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	// Create the submission with the correct order
	submissionId, err := tr.taskService.CreateSubmission(tx, taskId, userId, languageId, respJson.SubmissionNumber)
	if err != nil {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error creating submission. %s", err.Error()))
		return
	}

	err = tr.queueService.PublishSubmission(tx, submissionId)
	if err != nil {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error publishing submission to the queue. %s", err.Error()))
		return
	}

	utils.ReturnSuccess(w, http.StatusOK, "Solution submitted successfully")
}

func NewTaskRoute(fileStorageUrl string, taskService service.TaskService, queueService service.QueueService) TaskRoute {
	return &TaskRouteImpl{fileStorageUrl: fileStorageUrl, taskService: taskService, queueService: queueService}
}
