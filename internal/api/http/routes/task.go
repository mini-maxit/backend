package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/mini-maxit/backend/internal/api/http/utils"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/service"
	"github.com/sirupsen/logrus"
)

type TaskRoute interface {
	UploadTask(w http.ResponseWriter, r *http.Request)
	SubmitSolution(w http.ResponseWriter, r *http.Request)
}

type TaskRouteImpl struct {
	fileStorageUrl string

	// Service that handles task-related operations
	taskService  service.TaskService
	queueService service.QueueService
}

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
		utils.ReturnError(w, http.StatusBadRequest, "User ID is required.")
		return
	}
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid user ID.")
		return
	}

	// Extract the uploaded file
	file, handler, err := r.FormFile("task")
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Error retrieving the file. No task file found.")
		return
	}
	if !strings.HasSuffix(handler.Filename, ".zip") || !strings.HasSuffix(handler.Filename, ".tar.gz") {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid file format. Only .zip and .tar.gz files are allowed as task upload.")
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
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error creating empty task. %s", err.Error()))
		return
	}

	// Add form fields
	writer.WriteField("taskID", fmt.Sprintf("%d", taskId))
	writer.WriteField("overwrite", strconv.FormatBool(overwrite))

	// Create a form file field and copy the uploaded file to it
	part, err := writer.CreateFormFile("archive", handler.Filename)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error creating form file for FileStorage. %s", err.Error()))
		return
	}
	if _, err := io.Copy(part, file); err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error copying file to FileStorage request. %s", err.Error()))
		return
	}
	writer.Close()

	// Send the request to FileStorage service
	client := &http.Client{}
	resp, err := client.Post(tr.fileStorageUrl+"/createTask", writer.FormDataContentType(), body)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error sending file to FileStorage service. %s", err.Error()))
		return
	}
	defer resp.Body.Close()

	// Handle response from FileStorage
	buffer := make([]byte, resp.ContentLength)
	bytesRead, err := resp.Body.Read(buffer)
	if err != nil && bytesRead == 0 {
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error reading response from FileStorage. %s", err.Error()))
		return
	}
	if resp.StatusCode != http.StatusOK {
		utils.ReturnError(w, resp.StatusCode, fmt.Sprintf("Failed to upload file to FileStorage. %s", string(buffer)))
		return
	}

	// Update the task with the correct directories
	updateInfo := schemas.UpdateTask{
		Title: taskName,
	}
	if err := tr.taskService.UpdateTask(taskId, updateInfo); err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error updating task directories. %s", err.Error()))
		return
	}

	utils.ReturnSuccess(w, http.StatusOK, "Task uploaded successfully")
}

func (tr *TaskRouteImpl) SubmitSolution(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit the size of the incoming request to 10 MB
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)

	// Parse the multipart form data
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "The uploaded file is too large.", http.StatusBadRequest)
		return
	}

	// Extract the task ID
	taskIdStr := r.FormValue("taskID")
	if taskIdStr == "" {
		http.Error(w, "Task ID is required.", http.StatusBadRequest)
		return
	}
	taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid task ID.", http.StatusBadRequest)
		return
	}

	// Extract the uploaded file
	file, handler, err := r.FormFile("solution")
	if err != nil {
		http.Error(w, "Error retrieving the file. No solution file found.", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Extract user ID
	userIDStr := r.FormValue("userID")
	if userIDStr == "" {
		http.Error(w, "User ID is required.", http.StatusBadRequest)
		return
	}
	userId, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID.", http.StatusBadRequest)
		return
	}

	// Extract language
	languageStr := r.FormValue("language")
	if languageStr == "" {
		http.Error(w, "Language config is required.", http.StatusBadRequest)
		return
	}
	logrus.Info(languageStr)
	var language schemas.LanguageConfig
	if err := json.Unmarshal([]byte(languageStr), &language); err != nil {
		http.Error(w, "Invalid language config.", http.StatusBadRequest)
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
		http.Error(w, "Error creating form file for FileStorage.", http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(part, file); err != nil {
		http.Error(w, "Error copying file to FileStorage request.", http.StatusInternalServerError)
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

	buffer := make([]byte, resp.ContentLength)
	bytesRead, err := resp.Body.Read(buffer)
	if err != nil && bytesRead == 0 {
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error reading response from FileStorage. %s", err.Error()))
		return
	}
	// Handle response from FileStorage
	if resp.StatusCode != http.StatusOK {
		utils.ReturnError(w, resp.StatusCode, fmt.Sprintf("Failed to upload file to FileStorage. %s", string(buffer)))
		return
	}
	// Waiting to be implemented on the FileStorage service side
	// order, error := strconv.ParseInt(string(buffer), 10, 64)
	// if error != nil {
	// 	http.Error(w, "Error parsing response from FileStorage.", http.StatusInternalServerError)
	// 	return
	// }
	order := rand.Int64N(30)

	// Create the submission with the correct order
	submissionId, err := tr.taskService.CreateSubmission(taskId, userId, language, order)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error creating submission. %s", err.Error()))
		return
	}

	err = tr.queueService.PublishSubmission(submissionId)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error publishing submission to the queue. %s", err.Error()))
		return
	}

	utils.ReturnSuccess(w, http.StatusOK, "Solution submitted successfully")
}

func NewTaskRoute(fileStorageUrl string, taskService service.TaskService, queueService service.QueueService) TaskRoute {
	return &TaskRouteImpl{fileStorageUrl: fileStorageUrl, taskService: taskService, queueService: queueService}
}
