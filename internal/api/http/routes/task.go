package routes

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/mini-maxit/backend/internal/api/http/utils"
	"github.com/mini-maxit/backend/package/service"
)

type TaskRoute interface {
	UploadTask(w http.ResponseWriter, r *http.Request)
}

type TaskRouteImpl struct {
	fileStorageUrl string

	// Service that handles task-related operations
	taskService service.TaskService
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

	// Extract the overwrite flag
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

	// Extract the uploaded file
	file, handler, err := r.FormFile("task")
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Error retrieving the file. No task file found.")
		return
	}
	defer file.Close()

	// Create a multipart writer for the HTTP request to FileStorage service
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create empty task to get the task ID
	taskId, err := tr.taskService.CreateEmpty()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error creating empty task. %s", err.Error()))
		return
	}

	// Add form fields
	writer.WriteField("taskID", fmt.Sprintf("%d", taskId))
	writer.WriteField("overwrite", strconv.FormatBool(overwrite))

	// Create a form file field and copy the uploaded file to it
	part, err := writer.CreateFormFile("task", handler.Filename)
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
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/createTask", tr.fileStorageUrl), body)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error creating request to FileStorage. %s", err.Error()))
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error sending file to FileStorage service. %s", err.Error()))
		return
	}
	defer resp.Body.Close()

	buffer := make([]byte, resp.ContentLength)
	resp.Body.Read(buffer)
	// Handle response from FileStorage
	if resp.StatusCode != http.StatusOK {
		utils.ReturnError(w, resp.StatusCode, fmt.Sprintf("Failed to upload file to FileStorage. %s", string(buffer)))
		return
	}

	// Update the task withw the correct directories

	utils.ReturnSuccess(w, http.StatusOK, "Task uploaded successfully")
}

func NewTaskRoute(fileStorageUrl string, taskService service.TaskService) TaskRoute {
	return &TaskRouteImpl{fileStorageUrl: fileStorageUrl, taskService: taskService}
}
