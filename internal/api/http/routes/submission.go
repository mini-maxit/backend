package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/api/http/middleware"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/service"
)

type SubmissionRoutes interface {
	// Get requests
	GetAll(w http.ResponseWriter, r *http.Request)
	GetById(w http.ResponseWriter, r *http.Request)
	GetAllForUser(w http.ResponseWriter, r *http.Request)
	GetAllForUserShort(w http.ResponseWriter, r *http.Request)
	GetAllForGroup(w http.ResponseWriter, r *http.Request)
	GetAllForTask(w http.ResponseWriter, r *http.Request)
	GetAvailableLanguages(w http.ResponseWriter, r *http.Request)

	// Post requests
	SubmitSolution(w http.ResponseWriter, r *http.Request)
}

type SumbissionImpl struct {
	submissionService service.SubmissionService
	fileStorageUrl    string
	queueService      service.QueueService
}

// GetAll godoc
//
//	@Tags			submission
//	@Summary		Get all submissions for the current user
//	@Description	Depending on the user role, this endpoint will return all submissions for the current user if user is student, all submissions to owned tasks if user is teacher, and all submissions in database if user is admin
//	@Produce		json
//	@Param			limit	query		int		false	"Limit the number of returned submissions"
//	@Param			offset	query		int		false	"Offset the returned submissions"
//	@Param			Session	header		string	true	"Session Token"
//	@Success		200		{object}	httputils.ApiResponse[[]schemas.Submission]
//	@Failure		400		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Router			/submission [get]
func (s *SumbissionImpl) GetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.User)

	queryParams := r.Context().Value(middleware.QueryParamsKey).(map[string]interface{})

	submissions, err := s.submissionService.GetAll(tx, current_user, queryParams)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to get submissions. "+err.Error())
		return
	}

	if submissions == nil {
		submissions = []schemas.Submission{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, submissions)
}

// GetById godoc
//
//	@Tags			submission
//	@Summary		Get a submission by ID
//	@Description	Get a submission by its ID, if the user is a student, the submission must belong to the user, if the user is a teacher, the submission must belong to a task owned by the teacher, if the user is an admin, the submission can be any submission
//	@Produce		json
//	@Param			id		path		int		true	"Submission ID"
//	@Param			Session	header		string	true	"Session Token"
//	@Success		200		{object}	httputils.ApiResponse[schemas.Submission]
//	@Failure		400		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Router			/submission/{id} [get]
func (s *SumbissionImpl) GetById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	submissionIdStr := r.PathValue("id")
	submissionId, err := strconv.ParseInt(submissionIdStr, 10, 64)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid submission id. "+err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.User)

	submission, err := s.submissionService.GetById(tx, submissionId, current_user)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to get submission. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, submission)
}

// GetAllForUser godoc
//
//	@Tags			submission
//	@Summary		Get all submissions for a user
//	@Description	Gets all submissions for specific user. If the user is a student, it fails with 403 Forbidden. For teacher it returns all submissions from this user for tasks owned by the teacher. For admin it returns all submissions for specific user.
//	@Produce		json
//	@Param			id		path		int		true	"User ID"
//	@Param			limit	query		int		false	"Limit the number of returned submissions"
//	@Param			offset	query		int		false	"Offset the returned submissions"
//	@Param			Session	header		string	true	"Session Token"
//	@Success		200		{object}	httputils.ApiResponse[[]schemas.Submission]
//	@Failure		400		{object}	httputils.ApiError
//	@Failure		403		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Router			/submission/user/{id} [get]
func (s *SumbissionImpl) GetAllForUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userIdStr := r.PathValue("id")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid user id. "+err.Error())
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.User)
	queryParams := r.Context().Value(middleware.QueryParamsKey).(map[string]interface{})

	submissions, err := s.submissionService.GetAllForUser(tx, userId, current_user, queryParams)
	if err != nil {
		db.Rollback()
		switch err {
		case service.ErrPermissionDenied:
			httputils.ReturnError(w, http.StatusForbidden, "Permission denied. Current user is not allowed to view submissions for this user.")
		default:
			httputils.ReturnError(w, http.StatusInternalServerError, "Failed to get submissions. "+err.Error())
		}
		return
	}

	if submissions == nil {
		submissions = []schemas.Submission{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, submissions)
}

// GetAllForUserShort godoc
//
//	@Tags			submission
//	@Summary		Get all submissions for a user
//	@Description	Gets all submissions for specific user. If the user is a student, it fails with 403 Forbidden. For teacher it returns all submissions from this user for tasks owned by the teacher. For admin it returns all submissions for specific user.
//	@Produce		json
//	@Param			id		path		int		true	"User ID"
//	@Param			limit	query		int		false	"Limit the number of returned submissions"
//	@Param			offset	query		int		false	"Offset the returned submissions"
//	@Param			Session	header		string	true	"Session Token"
//	@Success		200		{object}	httputils.ApiResponse[[]schemas.Submission]
//	@Failure		400		{object}	httputils.ApiError
//	@Failure		403		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Router			/submission/user/{id}/short [get]
func (s *SumbissionImpl) GetAllForUserShort(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userIdStr := r.PathValue("id")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid user id. "+err.Error())
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.User)
	queryParams := r.Context().Value(middleware.QueryParamsKey).(map[string]interface{})

	submissions, err := s.submissionService.GetAllForUserShort(tx, userId, current_user, queryParams)
	if err != nil {
		db.Rollback()
		switch err {
		case service.ErrPermissionDenied:
			httputils.ReturnError(w, http.StatusForbidden, "Permission denied. Current user is not allowed to view submissions for this user.")
		default:
			httputils.ReturnError(w, http.StatusInternalServerError, "Failed to get submissions. "+err.Error())
		}
		return
	}

	if submissions == nil {
		submissions = []schemas.SubmissionShort{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, submissions)
}

// GetAllForUser godoc
//
//	@Tags			submission
//	@Summary		Get all submissions for a group
//	@Description	Gets all submissions for specific group. If the user is a student, it fails with 403 Forbidden. For teacher it returns all submissions from this group for tasks he created. For admin it returns all submissions for specific group.
//	@Produce		json
//	@Param			id		path		int		true	"Group ID"
//	@Param			limit	query		int		false	"Limit the number of returned submissions"
//	@Param			offset	query		int		false	"Offset the returned submissions"
//	@Param			Session	header		string	true	"Session Token"
//	@Success		200		{object}	httputils.ApiResponse[[]schemas.Submission]
//	@Failure		400		{object}	httputils.ApiError
//	@Failure		403		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Router			/submission/user/{id} [get]
func (s *SumbissionImpl) GetAllForGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	groupIdStr := r.PathValue("id")
	groupId, err := strconv.ParseInt(groupIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group id. "+err.Error())
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.User)

	queryParams := r.Context().Value(middleware.QueryParamsKey).(map[string]interface{})

	submissions, err := s.submissionService.GetAllForGroup(tx, groupId, current_user, queryParams)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to get submissions. "+err.Error())
		return
	}

	if submissions == nil {
		submissions = []schemas.Submission{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, submissions)
}

// GetAllForTask godoc
//
//	@Tags			submission
//	@Summary		Get all submissions for a task
//	@Description	Gets all submissions for specific task. If the user is a student and has no access to this task, it fails with 403 Forbidden. For teacher it returns all submissions for this task if he created it. For admin it returns all submissions for specific task.
//	@Produce		json
//	@Param			id		path		int		true	"Task ID"
//	@Param			limit	query		int		false	"Limit the number of returned submissions"
//	@Param			offset	query		int		false	"Offset the returned submissions"
//	@Param			Session	header		string	true	"Session Token"
//	@Success		200		{object}	httputils.ApiResponse[[]schemas.Submission]
//	@Failure		400		{object}	httputils.ApiError
//	@Failure		403		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Router			/submission/task/{id} [get]
func (s *SumbissionImpl) GetAllForTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIdStr := r.PathValue("id")
	taskId, err := strconv.ParseInt(taskIdStr, 10, 64)

	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task id. "+err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.User)
	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	queryParams := r.Context().Value(middleware.QueryParamsKey).(map[string]interface{})

	submissions, err := s.submissionService.GetAllForTask(tx, taskId, current_user, queryParams)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to get submissions. "+err.Error())
		return
	}

	if submissions == nil {
		submissions = []schemas.Submission{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, submissions)
}

// GetAvailableLanguages godoc
//
//	@Tags			submission
//	@Summary		Get all available languages
//	@Description	Get all available languages for submitting solutions. Temporary solution, while all tasks have same languages
//	@Produce		json
//	@Success		200	{object}	httputils.ApiResponse[[]schemas.LanguageConfig]
//	@Failure		500	{object}	httputils.ApiError
//	@Router			/submission/languages [get]
func (s *SumbissionImpl) GetAvailableLanguages(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}
	languages, err := s.submissionService.GetAvailableLanguages(tx)
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to get available languages. "+err.Error())
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, languages)
}

// SubmitSolution godoc
//
//	@Tags			submission
//	@Summary		Submit a solution
//	@Description	Submit a solution to a task, the solution is uploaded to the FileStorage service and a submission is created in the database. The submission is then published to the queue for processing. The response contains the submission ID. Fails if user has no access to provided task.
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			taskID		formData	int		true	"Task ID"
//	@Param			solution	formData	file	true	"Solution file"
//	@Param			languageID	formData	int		true	"Language ID"
//	@Param			Session		header		string	true	"Session Token"
//	@Success		200			{object}	httputils.ApiResponse[schemas.SubmitResponse]
//	@Failure		400			{object}	httputils.ApiError
//	@Failure		403			{object}	httputils.ApiError
//	@Failure		500			{object}	httputils.ApiError
func (s *SumbissionImpl) SubmitSolution(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Limit the size of the incoming request to 10 MB
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)

	// Parse the multipart form data
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "The uploaded files are too large.")
		return
	}

	// Extract the task ID
	taskIdStr := r.FormValue("taskID")
	if taskIdStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	// Extract the uploaded file
	file, handler, err := r.FormFile("solution")
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Error retrieving the file. No solution file found.")
		return
	}
	defer file.Close()

	// Extract user ID
	current_user := r.Context().Value(middleware.UserKey).(schemas.User)
	userId := current_user.Id
	userIDStr := strconv.FormatInt(userId, 10)

	// Extract language
	languageStr := r.FormValue("languageID")
	if languageStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Language ID is required.")
		return
	}
	languageId, err := strconv.ParseInt(languageStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid language ID.")
		return
	}

	// Create a multipart writer for the HTTP request to FileStorage service
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	err = writer.WriteField("taskID", taskIdStr)
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Error writing taskID to FileStorage request. "+err.Error())
		return
	}
	err = writer.WriteField("userID", userIDStr)
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Error writing userID to FileStorage request. "+err.Error())
		return
	}

	// Create a form file field and copy the uploaded file to it
	part, err := writer.CreateFormFile("submissionFile", handler.Filename)
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Error creating form file for FileStorage. "+err.Error())
		return
	}
	if _, err := io.Copy(part, file); err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Error copying file to FileStorage request. "+err.Error())
		return
	}

	writer.Close()

	// Send the request to FileStorage service
	client := &http.Client{}
	resp, err := client.Post(s.fileStorageUrl+"/submit", writer.FormDataContentType(), body)
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error sending file to FileStorage service. %s", err.Error()))
		return
	}
	defer resp.Body.Close()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil && len(resBody) == 0 {
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error reading response from FileStorage. %s", err.Error()))
		return
	}
	// Handle response from FileStorage
	if resp.StatusCode != http.StatusOK {
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to upload file to FileStorage. %s", string(resBody)))
		return
	}

	respJson := schemas.SubmitResponse{}
	err = json.Unmarshal(resBody, &respJson)
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error parsing response from FileStorage. %s", err.Error()))
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	// Create the submission with the correct order
	submissionId, err := s.submissionService.CreateSubmission(tx, taskId, userId, languageId, respJson.SubmissionNumber)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error creating submission. %s", err.Error()))
		return
	}

	err = s.queueService.PublishSubmission(tx, submissionId)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error publishing submission to the queue. %s", err.Error()))
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, "Solution submitted successfully")
}

// New Instance
func NewSubmissionRoutes(submissionService service.SubmissionService, fileStorageUrl string, queueService service.QueueService) SubmissionRoutes {
	return &SumbissionImpl{
		submissionService: submissionService,
		fileStorageUrl:    fileStorageUrl,
		queueService:      queueService,
	}
}

// RegisterSubmissionRoutes registers handlers for the submission routes
func RegisterSubmissionRoutes(mux *http.ServeMux, route SubmissionRoutes) {
	mux.HandleFunc("/", route.GetAll)
	mux.HandleFunc("/{id}", route.GetById)
	mux.HandleFunc("/user/{id}", route.GetAllForUser)
	mux.HandleFunc("/user/{id}/short", route.GetAllForUserShort)
	mux.HandleFunc("/group/{id}", route.GetAllForGroup)
	mux.HandleFunc("/task/{id}", route.GetAllForTask)
	mux.HandleFunc("/submit", route.SubmitSolution)
	mux.HandleFunc("/languages", route.GetAvailableLanguages)
}
