package routes

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	myerrors "github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
)

type SubmissionRoutes interface {
	// Get requests
	GetAll(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
	GetAllForUser(w http.ResponseWriter, r *http.Request)
	GetAllForUserShort(w http.ResponseWriter, r *http.Request)
	GetAllForGroup(w http.ResponseWriter, r *http.Request)
	GetAllForTask(w http.ResponseWriter, r *http.Request)
	GetAvailableLanguages(w http.ResponseWriter, r *http.Request)

	// Post requests
	SubmitSolution(w http.ResponseWriter, r *http.Request)
}

const submissionBodyLimit = 2 << 20

type SumbissionImpl struct {
	submissionService service.SubmissionService
	taskService       service.TaskService
	fileStorageURL    string
	queueService      service.QueueService
	logger            *zap.SugaredLogger
}

// GetAll godoc
//
//	@Tags			submission
//	@Summary		Get all submissions for the current user
//	@Description	Depending on the user role, this endpoint will return all submissions for the current user.
//
// If user is student, all submissions to owned tasks.
// If user is teacher or admin, and all submissions in database.
//
//	@Produce		json
//	@Param			limit	query		int	false	"Limit the number of returned submissions"
//	@Param			offset	query		int	false	"Offset the returned submissions"
//	@Success		200		{object}	httputils.APIResponse[[]schemas.Submission]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/submissions [get]
func (s *SumbissionImpl) GetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		s.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)

	submissions, err := s.submissionService.GetAll(tx, currentUser, queryParams)
	if err != nil {
		db.Rollback()
		s.logger.Errorw("Failed to get submissions", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Submission service temporarily unavailable")
		return
	}

	if submissions == nil {
		submissions = []schemas.Submission{}
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

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		s.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	submissionIDStr := httputils.GetPathValue(r, "id")
	submissionID, err := strconv.ParseInt(submissionIDStr, 10, 64)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid submission id. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	submission, err := s.submissionService.Get(tx, submissionID, currentUser)
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
//	@Param			id		path		int	true	"User ID"
//	@Param			limit	query		int	false	"Limit the number of returned submissions"
//	@Param			offset	query		int	false	"Offset the returned submissions"
//	@Success		200		{object}	httputils.APIResponse[[]schemas.Submission]
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

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		s.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)
	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)

	submissions, err := s.submissionService.GetAllForUser(tx, userID, currentUser, queryParams)
	if err != nil {
		db.Rollback()
		switch {
		case errors.Is(err, myerrors.ErrPermissionDenied):
			httputils.ReturnError(w,
				http.StatusForbidden,
				"Permission denied. Current user is not allowed to view submissions for this user.",
			)
		default:
			s.logger.Errorw("Failed to get submissions for user", "error", err)
			httputils.ReturnError(w, http.StatusInternalServerError, "Submission service temporarily unavailable")
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
//	@Description	If the user is a student, it fails with 403 Forbidden.
//
// For teacher it returns all submissions from this user for tasks owned by the teacher.
// For admin it returns all submissions for specific user.
//
//	@Produce		json
//	@Param			id		path		int	true	"User ID"
//	@Param			limit	query		int	false	"Limit the number of returned submissions"
//	@Param			offset	query		int	false	"Offset the returned submissions"
//	@Success		200		{object}	httputils.APIResponse[[]schemas.Submission]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/submissions/users/{id}/short [get]
func (s *SumbissionImpl) GetAllForUserShort(w http.ResponseWriter, r *http.Request) {
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

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		s.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)
	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)

	submissions, err := s.submissionService.GetAllForUserShort(tx, userID, currentUser, queryParams)
	if err != nil {
		db.Rollback()
		switch {
		case errors.Is(err, myerrors.ErrPermissionDenied):
			httputils.ReturnError(w,
				http.StatusForbidden,
				"Permission denied. Current user is not allowed to view submissions for this user.",
			)
		default:
			httputils.ReturnError(w,
				http.StatusInternalServerError,
				"Failed to get submissions. "+err.Error(),
			)
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
//	@Description	If the user is a student, it fails with 403 Forbidden.
//
// For teacher it returns all submissions from this group for tasks he created.
// For admin it returns all submissions for specific group.
//
//	@Produce		json
//	@Param			id		path		int	true	"Group ID"
//	@Param			limit	query		int	false	"Limit the number of returned submissions"
//	@Param			offset	query		int	false	"Offset the returned submissions"
//	@Success		200		{object}	httputils.APIResponse[[]schemas.Submission]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/submissions/groups/{id} [get]
func (s *SumbissionImpl) GetAllForGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	groupIDStr := httputils.GetPathValue(r, "id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group id. "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		s.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)

	submissions, err := s.submissionService.GetAllForGroup(tx, groupID, currentUser, queryParams)
	if err != nil {
		db.Rollback()
		s.logger.Errorw("Failed to get submissions for group", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Submission service temporarily unavailable")
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
//	@Description	If the user is a student and has no access to this task, it fails with 403 Forbidden.
//
// For teacher it returns all submissions for this task if he created it.
// For admin it returns all submissions for specific task.
//
//	@Produce		json
//	@Param			id		path		int	true	"Task ID"
//	@Param			limit	query		int	false	"Limit the number of returned submissions"
//	@Param			offset	query		int	false	"Offset the returned submissions"
//	@Success		200		{object}	httputils.APIResponse[[]schemas.Submission]
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

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)
	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		s.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)

	submissions, err := s.submissionService.GetAllForTask(tx, taskID, currentUser, queryParams)
	if err != nil {
		db.Rollback()
		s.logger.Errorw("Failed to get submissions for task", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Submission service temporarily unavailable")
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
//	@Description	Get all available languages for submitting solutions.
//
// Temporary solution, while all tasks have same languages
//
//	@Produce		json
//	@Success		200	{object}	httputils.APIResponse[[]schemas.LanguageConfig]
//	@Failure		500	{object}	httputils.APIError
//	@Router			/submissions/languages [get]
func (s *SumbissionImpl) GetAvailableLanguages(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		s.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}
	languages, err := s.submissionService.GetAvailableLanguages(tx)
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
//	@Param			contestID	formData	int		false "Contest ID"
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

	// Extract user ID
	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

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

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		s.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	// Create the submission with the correct order
	submissionID, err := s.submissionService.Submit(tx, &currentUser, taskID, languageID, contestID, filePath)
	if err != nil {
		db.Rollback()
		switch {
		case errors.Is(err, myerrors.ErrNotFound):
			httputils.ReturnError(w, http.StatusNotFound, "Task not found")
		case errors.Is(err, myerrors.ErrPermissionDenied):
			httputils.ReturnError(w, http.StatusForbidden, "Permission denied. Current user is not allowed to submit solution to this task.")
		case errors.Is(err, myerrors.ErrContestNotStarted):
			httputils.ReturnError(w, http.StatusBadRequest, "Contest has not started yet.")
		case errors.Is(err, myerrors.ErrContestEnded):
			httputils.ReturnError(w, http.StatusBadRequest, "Contest has already ended.")
		case errors.Is(err, myerrors.ErrNotContestParticipant):
			httputils.ReturnError(w, http.StatusBadRequest, "User is not a participant of the contest.")
		case errors.Is(err, myerrors.ErrContestSubmissionClosed):
			httputils.ReturnError(w, http.StatusBadRequest, "Contest submissions are closed.")
		case errors.Is(err, myerrors.ErrTaskSubmissionClosed):
			httputils.ReturnError(w, http.StatusBadRequest, "Submissions for this task are closed.")
		case errors.Is(err, myerrors.ErrTaskNotStarted):
			httputils.ReturnError(w, http.StatusBadRequest, "Task submission period has not started yet.")
		case errors.Is(err, myerrors.ErrTaskSubmissionEnded):
			httputils.ReturnError(w, http.StatusBadRequest, "Submissions for this task have ended.")
		case errors.Is(err, myerrors.ErrTaskNotInContest):
			httputils.ReturnError(w, http.StatusBadRequest, "Task is not part of this contest.")
		case errors.Is(err, myerrors.ErrInvalidLanguage):
			httputils.ReturnError(w, http.StatusBadRequest, "Invalid language for the task.")
		default:
			s.logger.Errorw("Failed to create submission", "error", err)
			httputils.ReturnError(w,
				http.StatusInternalServerError,
				"Submission service temporarily unavailable",
			)
		}
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, map[string]int64{"submissionId": submissionID})
}

// New Instance.
func NewSubmissionRoutes(
	submissionService service.SubmissionService,
	fileStorageURL string,
	queueService service.QueueService,
	taskService service.TaskService,
) SubmissionRoutes {
	route := &SumbissionImpl{
		submissionService: submissionService,
		fileStorageURL:    fileStorageURL,
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
	mux.HandleFunc("/", route.GetAll)
	mux.HandleFunc("/submit", route.SubmitSolution)
	mux.HandleFunc("/languages", route.GetAvailableLanguages)
	mux.HandleFunc("/{id}", route.GetByID)
	mux.HandleFunc("/users/{id}", route.GetAllForUser)
	mux.HandleFunc("/users/{id}/short", route.GetAllForUserShort)
	mux.HandleFunc("/groups/{id}", route.GetAllForGroup)
	mux.HandleFunc("/tasks/{id}", route.GetAllForTask)
}
