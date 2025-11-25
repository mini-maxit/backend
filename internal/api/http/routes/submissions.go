package routes

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	_ "github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
)

type SubmissionRoutes interface {
	// Get requests
	GetAll(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
	GetAllForUser(w http.ResponseWriter, r *http.Request)
	GetAllForTask(w http.ResponseWriter, r *http.Request)
	GetAvailableLanguages(w http.ResponseWriter, r *http.Request)
	GetMySubmissions(w http.ResponseWriter, r *http.Request)

	// Post requests
	SubmitSolution(w http.ResponseWriter, r *http.Request)
}

const submissionBodyLimit = 2 << 20

type SumbissionImpl struct {
	submissionService service.SubmissionService
	taskService       service.TaskService
	queueService      service.QueueService
	logger            *zap.SugaredLogger
}

// GetAll godoc
//
//	@Tags			submission
//	@Summary		Get all submissions with optional filtering
//	@Description	Get submissions with optional filters by userId, contestId, and/or taskId. Students can only view their own submissions. Teachers can view submissions for tasks/contests they created. Admins have full access.
//	@Produce		json
//	@Param			userId		query		int		false	"User ID filter"
//	@Param			contestId	query		int		false	"Contest ID filter"
//	@Param			taskId		query		int		false	"Task ID filter"
//	@Param			limit		query		int		false	"Limit the number of returned submissions"
//	@Param			offset		query		int		false	"Offset the returned submissions"
//	@Param			sort		query		string	false	"Sort order"
//	@Success		200			{object}	httputils.APIResponse[schemas.PaginatedResult[schemas.Submission]]
//	@Failure		400			{object}	httputils.APIError
//	@Failure		403			{object}	httputils.APIError
//	@Failure		500			{object}	httputils.APIError
//	@Router			/submissions [get]
func (s *SumbissionImpl) GetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	query := r.URL.Query()
	queryParams, err := httputils.GetQueryParams(&query)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid query parameters. "+err.Error())
		return
	}
	paginationParams := httputils.ExtractPaginationParams(queryParams)

	// Parse optional query parameters: userId, contestId, and taskId
	userIDStr := r.URL.Query().Get("userId")
	var userID *int64 = nil
	if userIDStr != "" {
		userIDVal, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			httputils.ReturnError(w, http.StatusBadRequest, "Invalid userId")
			return
		}
		userID = &userIDVal
	}

	contestIDStr := r.URL.Query().Get("contestId")
	var contestID *int64 = nil
	if contestIDStr != "" {
		contestIDVal, err := strconv.ParseInt(contestIDStr, 10, 64)
		if err != nil {
			httputils.ReturnError(w, http.StatusBadRequest, "Invalid contestId")
			return
		}
		contestID = &contestIDVal
	}

	taskIDStr := r.URL.Query().Get("taskId")
	var taskID *int64 = nil
	if taskIDStr != "" {
		taskIDVal, err := strconv.ParseInt(taskIDStr, 10, 64)
		if err != nil {
			httputils.ReturnError(w, http.StatusBadRequest, "Invalid taskId")
			return
		}
		taskID = &taskIDVal
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	submissions, err := s.submissionService.GetAll(db, currentUser, userID, taskID, contestID, paginationParams)
	if err != nil {
		httputils.HandleServiceError(w, err, db, s.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, submissions)
}

// GetByID godoc
//
//	@Tags			submission
//	@Summary		Get a submission by ID
//	@Description	If the user is a student, the submission must belong to the user.
//
// If the user is a teacher, the submission must belong to a task owned by the teacher.
// If the user is an admin, the submission can be any submission
//
//	@Produce		json
//	@Param			id	path		int	true	"Submission ID"
//	@Success		200	{object}	httputils.APIResponse[schemas.Submission]
//	@Failure		400	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Router			/submissions/{id} [get]
func (s *SumbissionImpl) GetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	submissionIDStr := httputils.GetPathValue(r, "id")
	submissionID, err := strconv.ParseInt(submissionIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid submission id. "+err.Error())
		return
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	submission, err := s.submissionService.Get(db, submissionID, currentUser)
	if err != nil {
		db.Rollback()
		s.logger.Errorw("Failed to get submission", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Submission service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, submission)
}

// GetAllForUser godoc
//
//	@Tags			submission
//	@Summary		Get all submissions for a user
//	@Description	If the user is a student, it fails with 403 Forbidden.
//
// For teacher it returns all submissions from this user for tasks owned by the teacher.
// For admin it returns all submissions for specific user.
//
//	@Produce		json
//	@Param			id		path		int		true	"User ID"
//	@Param			limit	query		int		false	"Limit the number of returned submissions"
//	@Param			offset	query		int		false	"Offset the returned submissions"
//	@Param			sort	query		string	false	"Sort order"
//	@Success		200		{object}	httputils.APIResponse[schemas.PaginatedResult[[]schemas.Submission]]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/submissions/users/{id} [get]
func (s *SumbissionImpl) GetAllForUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userIDStr := httputils.GetPathValue(r, "id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid user id. "+err.Error())
		return
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)
	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	paginationParams := httputils.ExtractPaginationParams(queryParams)

	response, err := s.submissionService.GetAllForUser(db, userID, currentUser, paginationParams)
	if err != nil {
		httputils.HandleServiceError(w, err, db, s.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, response)
}

// GetAllForTask godoc
//
//	@Tags			submission
//	@Summary		Get all submissions for a task
//	@Description	If the user is a student and has no access to this task, it fails with 403 Forbidden.
//
// For teacher it returns all submissions for this task if he created it.
// For admin it returns all submissions for specific task.
//
//	@Produce		json
//	@Param			id		path		int	true	"Task ID"
//	@Param			limit	query		int	false	"Limit the number of returned submissions"
//	@Param			offset	query		int	false	"Offset the returned submissions"
//	@Success		200		{object}	httputils.APIResponse[schemas.PaginatedResult[[]schemas.Submission]]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/submissions/tasks/{id} [get]
func (s *SumbissionImpl) GetAllForTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)

	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task id. "+err.Error())
		return
	}

	currentUser := httputils.GetCurrentUser(r)
	db := httputils.GetDatabase(r)
	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	paginationParams := httputils.ExtractPaginationParams(queryParams)

	response, err := s.submissionService.GetAllForTask(db, taskID, currentUser, paginationParams)
	if err != nil {
		db.Rollback()
		s.logger.Errorw("Failed to get submissions for task", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Submission service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, response)
}

// GetAvailableLanguages godoc
//
//	@Tags			submission
//	@Summary		Get all available languages
//	@Description	Get all available languages for submitting solutions.
//
// Temporary solution, while all tasks have same languages
//
//	@Produce		json
//	@Success		200	{object}	httputils.APIResponse[[]schemas.LanguageConfig]
//	@Failure		500	{object}	httputils.APIError
//	@Router			/submissions/languages [get]
func (s *SumbissionImpl) GetAvailableLanguages(w http.ResponseWriter, r *http.Request) {
	db := httputils.GetDatabase(r)
	languages, err := s.submissionService.GetAvailableLanguages(db)
	if err != nil {
		s.logger.Errorw("Failed to get available languages", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Language service temporarily unavailable")
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, languages)
}

// SubmitSolution godoc
//
//	@Tags			submission
//	@Summary		Submit a solution
//	@Description	The solution is uploaded to the FileStorage service and a submission is created in the database.
//
// The submission is then published to the queue for processing. The response contains the submission ID.
// Fails if user has no access to provided task.
//
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			taskID		formData	int		true	"Task ID"
//	@Param			contestID	formData	int		false	"Contest ID"
//	@Param			solution	formData	file	true	"Solution file"
//	@Param			languageID	formData	int		true	"Language ID"
//	@Success		200			{object}	map[string]int64
//	@Failure		400			{object}	httputils.APIError
//	@Failure		403			{object}	httputils.APIError
//	@Failure		500			{object}	httputils.APIError
//	@Router			/submissions/submit [post]
func (s *SumbissionImpl) SubmitSolution(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Limit the size of the incoming request to 10 MB
	r.Body = http.MaxBytesReader(w, r.Body, submissionBodyLimit)

	// Parse the multipart form data
	if err := r.ParseMultipartForm(submissionBodyLimit); err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "The uploaded files are too large.")
		return
	}

	// Extract the task ID
	taskIDStr := r.FormValue("taskID")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
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
	filePath, err := httputils.SaveMultiPartFile(file, handler)
	if err != nil {
		s.logger.Errorw("Failed to save multipart file", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "File upload service temporarily unavailable")
		return
	}

	// Extract language
	languageStr := r.FormValue("languageID")
	if languageStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Language ID is required.")
		return
	}
	languageID, err := strconv.ParseInt(languageStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid language ID.")
		return
	}

	contestStr := r.FormValue("contestID")
	var contestID *int64 = nil
	if contestStr != "" {
		parsedContestID, err := strconv.ParseInt(contestStr, 10, 64)
		if err != nil {
			httputils.ReturnError(w, http.StatusBadRequest, "Invalid contest ID.")
			return
		}
		contestID = &parsedContestID
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	// Create the submission with the correct order
	submissionID, err := s.submissionService.Submit(db, currentUser, taskID, languageID, contestID, filePath)
	if err != nil {
		httputils.HandleServiceError(w, err, db, s.logger)
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, map[string]int64{"submissionId": submissionID})
}

// GetMySubmissions godoc
//
//	@Tags			submission
//	@Summary		Get all submissions for a user
//	@Description	If the user is a student, it fails with 403 Forbidden.
//
// For teacher it returns all submissions from this user for tasks owned by the teacher.
// For admin it returns all submissions for specific user.
//
//	@Produce		json
//	@Param			limit	query		int		false	"Limit the number of returned submissions"
//	@Param			offset	query		int		false	"Offset the returned submissions"
//	@Param			sort	query		string	false	"Sort order"
//	@Success		200		{object}	httputils.APIResponse[schemas.PaginatedResult[schemas.Submission]]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/submissions/my [get]
func (s *SumbissionImpl) GetMySubmissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)
	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	paginationParams := httputils.ExtractPaginationParams(queryParams)

	response, err := s.submissionService.GetAllForUser(db, currentUser.ID, currentUser, paginationParams)
	if err != nil {
		httputils.HandleServiceError(w, err, db, s.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, response)
}

// New Instance.
func NewSubmissionRoutes(
	submissionService service.SubmissionService,
	queueService service.QueueService,
	taskService service.TaskService,
) SubmissionRoutes {
	route := &SumbissionImpl{
		submissionService: submissionService,
		queueService:      queueService,
		taskService:       taskService,
		logger:            utils.NewNamedLogger("submissions"),
	}
	err := utils.ValidateStruct(*route)
	if err != nil {
		log.Panicf("SubmissionRoutes struct is not valid: %s", err.Error())
	}
	return route
}

// RegisterSubmissionRoutes registers handlers for the submission routes.
func RegisterSubmissionRoutes(mux *mux.Router, route SubmissionRoutes) {
	mux.HandleFunc("/submissions", route.GetAll)
	mux.HandleFunc("/submissions/submit", route.SubmitSolution)
	mux.HandleFunc("/submissions/languages", route.GetAvailableLanguages)
	mux.HandleFunc("/submissions/my", route.GetMySubmissions)
	mux.HandleFunc("/submissions/users/{id}", route.GetAllForUser)
	mux.HandleFunc("/submissions/tasks/{id}", route.GetAllForTask)
	mux.HandleFunc("/submissions/{id}", route.GetByID)
}
