package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/mini-maxit/backend/internal/api/http/middleware"
	"github.com/mini-maxit/backend/internal/api/http/utils"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/service"
)

type SubmissionRoutes interface {
	// Get requests
	GetAll(w http.ResponseWriter, r *http.Request)
	GetById(w http.ResponseWriter, r *http.Request)
	GetAllForUser(w http.ResponseWriter, r *http.Request)
	GetAllForGroup(w http.ResponseWriter, r *http.Request)
	GetAllForTask(w http.ResponseWriter, r *http.Request)

	// Post requests
	SubmitSolution(w http.ResponseWriter, r *http.Request)
}

type SumbissionImpl struct {
	submissionService service.SubmissionService
	fileStorageUrl	string
	queueService service.QueueService
}

// Get requests
func (s *SumbissionImpl) GetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.UserSession)
	
	filters := map[string][]string{}
	for key, value := range r.URL.Query() {
		filters[key] = value
	}
	
	submissions, err := s.submissionService.GetAll(tx, current_user, filters)
	if err != nil {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, "Failed to get submissions. "+err.Error())
		return
	}

	if submissions == nil {
		submissions = []schemas.Submission{}
	}

	utils.ReturnSuccess(w, http.StatusOK, submissions)
}

func (s *SumbissionImpl) GetById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	submissionIdStr := r.PathValue("id")
	submissionId, err := strconv.ParseInt(submissionIdStr, 10, 64)
	if err != nil {
		db.Rollback()
		utils.ReturnError(w, http.StatusBadRequest, "Invalid submission id. "+err.Error())
		return
	}

	user := r.Context().Value(middleware.UserKey).(schemas.UserSession)

	submission, err := s.submissionService.GetById(tx, submissionId, user)
	if err != nil {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, "Failed to get submission. "+err.Error())
		return
	}

	utils.ReturnSuccess(w, http.StatusOK, submission)
}

func (s *SumbissionImpl) GetAllForUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userIdStr := r.PathValue("id")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid user id. "+err.Error())
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.UserSession)

	filters := map[string][]string{}
	for key, value := range r.URL.Query() {
		filters[key] = value
	}

	submissions, err := s.submissionService.GetAllForUser(tx, userId, current_user, filters)
	if err != nil {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, "Failed to get submissions. "+err.Error())
		return
	}

	if submissions == nil {
		submissions = []schemas.Submission{}
	}

	utils.ReturnSuccess(w, http.StatusOK, submissions)
}

func (s *SumbissionImpl) GetAllForGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	groupIdStr := r.PathValue("id")
	groupId, err := strconv.ParseInt(groupIdStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid group id. "+err.Error())
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.UserSession)

	filters := map[string][]string{}
	for key, value := range r.URL.Query() {
		filters[key] = value
	}

	submissions, err := s.submissionService.GetAllForGroup(tx, groupId, current_user, filters)
	if err != nil {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, "Failed to get submissions. "+err.Error())
		return
	}

	if submissions == nil {
		submissions = []schemas.Submission{}
	}

	utils.ReturnSuccess(w, http.StatusOK, submissions)
}

func (s *SumbissionImpl) GetAllForTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIdStr := r.PathValue("id")
	taskId, err := strconv.ParseInt(taskIdStr, 10, 64)

	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid task id. "+err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.UserSession)
	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	filters := map[string][]string{}
	for key, value := range r.URL.Query() {
		filters[key] = value
	}

	submissions, err := s.submissionService.GetAllForTask(tx, taskId, current_user, filters)
	if err != nil {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, "Failed to get submissions. "+err.Error())
		return
	}

	if submissions == nil {
		submissions = []schemas.Submission{}
	}

	utils.ReturnSuccess(w, http.StatusOK, submissions)
}

// Post requests
func (s *SumbissionImpl) SubmitSolution(w http.ResponseWriter, r *http.Request) {
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
	resp, err := client.Post(s.fileStorageUrl+"/submit", writer.FormDataContentType(), body)
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
	submissionId, err := s.submissionService.CreateSubmission(tx, taskId, userId, languageId, respJson.SubmissionNumber)
	if err != nil {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error creating submission. %s", err.Error()))
		return
	}

	err = s.queueService.PublishSubmission(tx, submissionId)
	if err != nil {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error publishing submission to the queue. %s", err.Error()))
		return
	}

	utils.ReturnSuccess(w, http.StatusOK, "Solution submitted successfully")
}

// New Instance
func NewSubmissionRoutes(submissionService service.SubmissionService, fileStorageUrl string, queueService service.QueueService) SubmissionRoutes {
	return &SumbissionImpl{
		submissionService: submissionService,
		fileStorageUrl: fileStorageUrl,
		queueService: queueService,
	}
}
