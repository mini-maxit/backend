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

	"github.com/mini-maxit/backend/internal/api/http/utils"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/service"
)

type TaskRoute interface {
	GetAllTasks(w http.ResponseWriter, r *http.Request)
	GetTask(w http.ResponseWriter, r *http.Request)
	UploadTask(w http.ResponseWriter, r *http.Request)
	SubmitSolution(w http.ResponseWriter, r *http.Request)
}

type TaskRouteImpl struct {
	fileStorageUrl string

	// Service that handles task-related operations
	taskService  service.TaskService
	queueService service.QueueService
}

func (tr *TaskRouteImpl) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ReturnError(w, http.StatusMethodNotAllowed, utils.CodeMethodNotAllowed, "Method not allowed")
		return
	}

	tasks, err := tr.taskService.GetAll()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, utils.CodeInternalServerError, fmt.Sprintf("Error getting tasks. %s", err.Error()))
		return
	}
	if tasks == nil {
		tasks = []schemas.Task{}
	}

	utils.ReturnSuccess(w, http.StatusOK, tasks)
}

func (tr *TaskRouteImpl) GetTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ReturnError(w, http.StatusMethodNotAllowed, utils.CodeMethodNotAllowed, "Method not allowed")
		return
	}

	taskIdStr := r.PathValue("id")
	if taskIdStr == "" {
		utils.ReturnError(w, http.StatusBadRequest, utils.CodeBadRequest, "Task ID is required.")
		return
	}
	taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, utils.CodeBadRequest, "Invalid task ID.")
		return
	}

	task, err := tr.taskService.GetTask(taskId)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, utils.CodeInternalServerError, fmt.Sprintf("Error getting task. %s", err.Error()))
		return
	}

	utils.ReturnSuccess(w, http.StatusOK, task)
}

func (tr *TaskRouteImpl) UploadTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ReturnError(w, http.StatusMethodNotAllowed, utils.CodeMethodNotAllowed, "Method not allowed")
		return
	}

	// Limit the size of the incoming request
	r.Body = http.MaxBytesReader(w, r.Body, 50*1024*1024) // 50 MB limit

	// Parse the multipart form data
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		utils.ReturnError(w, http.StatusBadRequest, utils.CodeBadRequest, "The uploaded files are too large.")
		return
	}

	overwriteStr := r.FormValue("overwrite")
	overwrite := false
	if overwriteStr != "" {
		var err error
		overwrite, err = strconv.ParseBool(overwriteStr)
		if err != nil {
			utils.ReturnError(w, http.StatusBadRequest, utils.CodeBadRequest, "Invalid overwrite flag.")
			return
		}
	}
	taskName := r.FormValue("taskName")
	if taskName == "" {
		utils.ReturnError(w, http.StatusBadRequest, utils.CodeBadRequest, "Task name is required.")
		return
	}
	userIdStr := r.FormValue("userId")
	if userIdStr == "" {
		utils.ReturnError(w, http.StatusBadRequest, utils.CodeBadRequest, "User ID of author is required.")
		return
	}
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, utils.CodeBadRequest, "Invalid user ID.")
		return
	}

	// Extract the uploaded file
	file, handler, err := r.FormFile("archive")
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, utils.CodeBadRequest, "Error retrieving the file. No task file found.")
		return
	}
	if !(strings.HasSuffix(handler.Filename, ".zip") || strings.HasSuffix(handler.Filename, ".tar.gz")) {
		utils.ReturnError(w, http.StatusBadRequest, utils.CodeBadRequest, "Invalid file format. Only .zip and .tar.gz files are allowed as task upload. Received: "+handler.Filename)
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
	taskId, err := tr.taskService.Create(task)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, utils.CodeInternalServerError, fmt.Sprintf("Error creating empty task. %s", err.Error()))
		return
	}

	// Add form fields
	writer.WriteField("taskID", fmt.Sprintf("%d", taskId))
	writer.WriteField("overwrite", strconv.FormatBool(overwrite))

	// Create a form file field and copy the uploaded file to it
	part, err := writer.CreateFormFile("archive", handler.Filename)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, utils.CodeInternalServerError, fmt.Sprintf("Error creating form file for FileStorage. %s", err.Error()))
		return
	}
	if _, err := io.Copy(part, file); err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, utils.CodeInternalServerError, fmt.Sprintf("Error copying file to FileStorage request. %s", err.Error()))
		return
	}
	writer.Close()

	// Send the request to FileStorage service
	client := &http.Client{}
	resp, err := client.Post(tr.fileStorageUrl+"/createTask", writer.FormDataContentType(), body)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, utils.CodeInternalServerError, fmt.Sprintf("Error sending file to FileStorage service. %s", err.Error()))
		return
	}
	defer resp.Body.Close()

	// Handle response from FileStorage
	buffer := make([]byte, resp.ContentLength)
	bytesRead, err := resp.Body.Read(buffer)
	if err != nil && bytesRead == 0 {
		utils.ReturnError(w, http.StatusInternalServerError, utils.CodeInternalServerError, fmt.Sprintf("Error reading response from FileStorage. %s", err.Error()))
		return
	}
	if resp.StatusCode != http.StatusOK {
		utils.ReturnError(w, http.StatusInternalServerError, utils.CodeInternalServerError, fmt.Sprintf("Failed to upload file to FileStorage. %s", string(buffer)))
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

	utils.ReturnSuccess(w, http.StatusOK, "Task uploaded successfully")
}

func (tr *TaskRouteImpl) SubmitSolution(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ReturnError(w, http.StatusMethodNotAllowed, utils.CodeMethodNotAllowed, "Method not allowed")
		return
	}

	// Limit the size of the incoming request to 10 MB
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)

	// Parse the multipart form data
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.ReturnError(w, http.StatusBadRequest, utils.CodeBadRequest, "The uploaded files are too large.")
		return
	}

	// Extract the task ID
	taskIdStr := r.FormValue("taskID")
	if taskIdStr == "" {
		utils.ReturnError(w, http.StatusBadRequest, utils.CodeBadRequest, "Task ID is required.")
		return
	}
	taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, utils.CodeBadRequest, "Invalid task ID.")
		return
	}

	// Extract the uploaded file
	file, handler, err := r.FormFile("solution")
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, utils.CodeBadRequest, "Error retrieving the file. No solution file found.")
		return
	}
	defer file.Close()

	// Extract user ID
	userIDStr := r.FormValue("userID")
	if userIDStr == "" {
		utils.ReturnError(w, http.StatusBadRequest, utils.CodeBadRequest, "User ID is required.")
		return
	}
	userId, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, utils.CodeBadRequest, "Invalid user ID.")
		return
	}

	// Extract language
	languageStr := r.FormValue("languageID")
	if languageStr == "" {
		utils.ReturnError(w, http.StatusBadRequest, utils.CodeBadRequest, "Language ID is required.")
		return
	}
	languageId, err := strconv.ParseInt(languageStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, utils.CodeBadRequest, "Invalid language ID.")
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
		utils.ReturnError(w, http.StatusInternalServerError, utils.CodeInternalServerError, "Error creating form file for FileStorage. "+err.Error())
		return
	}
	if _, err := io.Copy(part, file); err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, utils.CodeInternalServerError, "Error copying file to FileStorage request. "+err.Error())
		return
	}

	writer.Close()

	// Send the request to FileStorage service
	client := &http.Client{}
	resp, err := client.Post(tr.fileStorageUrl+"/submit", writer.FormDataContentType(), body)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, utils.CodeInternalServerError, fmt.Sprintf("Error sending file to FileStorage service. %s", err.Error()))
		return
	}
	defer resp.Body.Close()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil && len(resBody) == 0 {
		utils.ReturnError(w, http.StatusInternalServerError, utils.CodeInternalServerError, fmt.Sprintf("Error reading response from FileStorage. %s", err.Error()))
		return
	}
	// Handle response from FileStorage
	if resp.StatusCode != http.StatusOK {
		utils.ReturnError(w, http.StatusInternalServerError, utils.CodeInternalServerError, fmt.Sprintf("Failed to upload file to FileStorage. %s", string(resBody)))
		return
	}

	respJson := schemas.SubmitResponse{}
	err = json.Unmarshal(resBody, &respJson)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, utils.CodeInternalServerError, fmt.Sprintf("Error parsing response from FileStorage. %s", err.Error()))
		return
	}

	// Create the submission with the correct order
	submissionId, err := tr.taskService.CreateSubmission(taskId, userId, languageId, respJson.SubmissionNumber)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, utils.CodeInternalServerError, fmt.Sprintf("Error creating submission. %s", err.Error()))
		return
	}

	err = tr.queueService.PublishSubmission(submissionId)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, utils.CodeInternalServerError, fmt.Sprintf("Error publishing submission to the queue. %s", err.Error()))
		return
	}

	utils.ReturnSuccess(w, http.StatusOK, "Solution submitted successfully")
}

func NewTaskRoute(fileStorageUrl string, taskService service.TaskService, queueService service.QueueService) TaskRoute {
	return &TaskRouteImpl{fileStorageUrl: fileStorageUrl, taskService: taskService, queueService: queueService}
}
