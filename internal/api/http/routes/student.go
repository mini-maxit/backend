package routes

import (
	"errors"
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

const submissionBodyLimit = 2 << 20

// StudentRoute defines all student-specific routes.
type StudentRoute interface {
	// Contest routes
	GetContests(w http.ResponseWriter, r *http.Request)
	GetContestTasks(w http.ResponseWriter, r *http.Request)
	RegisterForContest(w http.ResponseWriter, r *http.Request)
	GetTaskProgress(w http.ResponseWriter, r *http.Request)

	// Submission routes
	GetSubmissions(w http.ResponseWriter, r *http.Request)
	SubmitSolution(w http.ResponseWriter, r *http.Request)
	GetAvailableLanguages(w http.ResponseWriter, r *http.Request)

	// Group routes
	GetGroups(w http.ResponseWriter, r *http.Request)

	// Task routes
	GetAssignedTasks(w http.ResponseWriter, r *http.Request)
}

type studentRouteImpl struct {
	contestService    service.ContestService
	submissionService service.SubmissionService
	groupService      service.GroupService
	taskService       service.TaskService
	queueService      service.QueueService
	fileStorageURL    string
	logger            *zap.SugaredLogger
}

// GetContests godoc
//
//	@Tags			contests
//	@Summary		Get contests for student
//	@Description	Get all contests accessible to the student (ongoing, upcoming, past)
//	@Produce		json
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Param			userID	query		int64	true	"User ID"
//	@Success		200		{object}	httputils.APIResponse[[]schemas.AvailableContest]
//	@Failure		401		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/student/contests [get]
func (sr *studentRouteImpl) GetContests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	query := r.URL.Query()
	queryParams, err := httputils.GetQueryParams(&query)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid query params")
		return
	}
	userIDstr, ok := queryParams["userID"]
	if !ok {
		httputils.ReturnError(w, http.StatusBadRequest, "Status query parameter is required")
		return
	}
	userID, ok := userIDstr.(int64)
	if !ok {
		httputils.ReturnError(w, http.StatusBadRequest, "Status query parameter must be a string")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		sr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	contests, err := sr.contestService.GetUserContests(tx, userID)
	if err != nil {
		db.Rollback()
		if errors.Is(err, myerrors.ErrUserNotFound) {
			httputils.ReturnError(w, http.StatusNotFound, "User not found")
			return
		}
		sr.logger.Errorw("Failed to get user contests", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Contest service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, contests)
}

// GetContestTasks godoc
//
//	@Tags			contests
//	@Summary		Get contest tasks
//	@Description	Get all tasks for a specific contest currently accessible to the participants
//	@Produce		json
//	@Param			contest_id	path		int	true	"Contest ID"
//	@Success		200			{object}	httputils.APIResponse[[]schemas.Task]
//	@Failure		400			{object}	httputils.APIError
//	@Failure		401			{object}	httputils.APIError
//	@Failure		404			{object}	httputils.APIError
//	@Failure		500			{object}	httputils.APIError
//	@Router			/student/contests/{contest_id}/tasks [get]
func (sr *studentRouteImpl) GetContestTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		sr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	contestStr := httputils.GetPathValue(r, "id")
	if contestStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Contest ID cannot be empty")
		return
	}
	contestID, err := strconv.ParseInt(contestStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid contest ID")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	tasks, err := sr.contestService.GetTasksForContest(tx, currentUser, contestID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			sr.logger.Errorw("Failed to get tasks for contest", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to get tasks for contest")
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, tasks)
}

// RegisterForContest godoc
//
//	@Tags			contests
//	@Summary		Register for contest
//	@Description	Register the student for a specific contest
//	@Produce		json
//	@Param			contest_id	path		int	true	"Contest ID"
//	@Success		200			{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Failure		400			{object}	httputils.APIError
//	@Failure		401			{object}	httputils.APIError
//	@Failure		404			{object}	httputils.APIError
//	@Failure		500			{object}	httputils.APIError
//	@Router			/student/contests/{contest_id}/register [post]
func (sr *studentRouteImpl) RegisterForContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		sr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	contestStr := httputils.GetPathValue(r, "id")
	if contestStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Contest ID cannot be empty")
		return
	}
	contestID, err := strconv.ParseInt(contestStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid contest ID")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = sr.contestService.RegisterForContest(tx, currentUser, contestID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, myerrors.ErrNotAuthorized):
			status = http.StatusForbidden
		case errors.Is(err, myerrors.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(err, myerrors.ErrContestRegistrationClosed):
			status = http.StatusForbidden
		case errors.Is(err, myerrors.ErrContestEnded):
			status = http.StatusForbidden
		case errors.Is(err, myerrors.ErrAlreadyRegistered) || errors.Is(err, myerrors.ErrAlreadyParticipant):
			status = http.StatusConflict
		}
		if status == http.StatusInternalServerError {
			sr.logger.Errorw("Failed to register for contest", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to register for contest")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Registration request submitted successfully"))
}

// GetTaskProgress godoc
//
//	@Tags			contests
//	@Summary		Get task progress
//	@Description	Get student's progress on tasks in a contest
//	@Produce		json
//	@Param			contest_id	path		int	true	"Contest ID"
//	@Success		200			{object}	httputils.APIResponse[[]schemas.TaskWithContestStats]
//	@Failure		400			{object}	httputils.APIError
//	@Failure		401			{object}	httputils.APIError
//	@Failure		404			{object}	httputils.APIError
//	@Failure		500			{object}	httputils.APIError
//	@Router			/student/contests/{contest_id}/task-progress [get]
func (sr *studentRouteImpl) GetTaskProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	contestStr := httputils.GetPathValue(r, "id")
	if contestStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Contest ID cannot be empty")
		return
	}
	contestID, err := strconv.ParseInt(contestStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid contest ID")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)
	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		sr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}
	tasks, err := sr.contestService.GetTaskProgressForContest(tx, currentUser, contestID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			sr.logger.Errorw("Failed to get task progress for contest", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to get task progress for contest")
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, tasks)
}

// GetSubmissions godoc
//
//	@Tags			submissions
//	@Summary		Get student submissions
//	@Description	Get all submissions for the authenticated student
//	@Produce		json
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Success		200		{object}	httputils.APIResponse[[]schemas.Submission]
//	@Failure		401		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/student/submissions [get]
func (sr *studentRouteImpl) GetSubmissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		sr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)
	paginationParams := httputils.ExtractPaginationParams(queryParams)

	submissions, err := sr.submissionService.GetAllForUser(tx, currentUser.ID, currentUser, paginationParams)
	if err != nil {
		db.Rollback()
		sr.logger.Errorw("Failed to get submissions", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Submission service temporarily unavailable")
		return
	}

	if submissions == nil {
		submissions = []schemas.Submission{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, submissions)
}

func (sr *studentRouteImpl) GetGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		sr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	groups, err := sr.groupService.GetAll(tx, currentUser, queryParams)
	if err != nil {
		db.Rollback()
		sr.logger.Errorw("Failed to get groups", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Group service temporarily unavailable")
		return
	}

	if groups == nil {
		groups = []schemas.Group{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, groups)
}

// GetAssignedTasks godoc
//
//	@Tags			tasks
//	@Summary		Get assigned tasks
//	@Description	Get all tasks assigned to the student
//	@Produce		json
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Success		200		{object}	httputils.APIResponse[[]schemas.Task]
//	@Failure		401		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/student/tasks [get]
func (sr *studentRouteImpl) GetAssignedTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		sr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	tasks, err := sr.taskService.GetAllAssigned(tx, currentUser, queryParams)
	if err != nil {
		db.Rollback()
		sr.logger.Errorw("Failed to get assigned tasks", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Task service temporarily unavailable")
		return
	}

	if tasks == nil {
		tasks = []schemas.Task{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, tasks)
}

// SubmitSolution godoc
//
//	@Tags			submissions
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
//	@Router			/teacher/submissions/submit [post]
func (sr *studentRouteImpl) SubmitSolution(w http.ResponseWriter, r *http.Request) {
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
		sr.logger.Errorw("Failed to save multipart file", "error", err)
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
		sr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	// Create the submission with the correct order
	submissionID, err := sr.submissionService.Submit(tx, &currentUser, taskID, languageID, contestID, filePath)
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
			sr.logger.Errorw("Failed to create submission", "error", err)
			httputils.ReturnError(w,
				http.StatusInternalServerError,
				"Submission service temporarily unavailable",
			)
		}
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, map[string]int64{"submissionId": submissionID})
}

// GetAvailableLanguages godoc
//
//	@Tags			submissions
//	@Summary		Get all available languages
//	@Description	Get all available languages for submitting solutions.
//
// Temporary solution, while all tasks have same languages
//
//	@Produce		json
//	@Success		200	{object}	httputils.APIResponse[[]schemas.LanguageConfig]
//	@Failure		500	{object}	httputils.APIError
//	@Router			/student/submissions/languages [get]
func (sr *studentRouteImpl) GetAvailableLanguages(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		sr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}
	languages, err := sr.submissionService.GetAvailableLanguages(tx)
	if err != nil {
		sr.logger.Errorw("Failed to get available languages", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Language service temporarily unavailable")
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, languages)
}

// NewStudentRoute creates a new StudentRoute.
func NewStudentRoute(
	contestService service.ContestService,
	submissionService service.SubmissionService,
	groupService service.GroupService,
	taskService service.TaskService,
	queueService service.QueueService,
) StudentRoute {
	return &studentRouteImpl{
		contestService:    contestService,
		submissionService: submissionService,
		groupService:      groupService,
		taskService:       taskService,
		queueService:      queueService,
		logger:            utils.NewNamedLogger("student-routes"),
	}
}

// RegisterStudentRoutes registers all student routes.
func RegisterStudentRoutes(mux *mux.Router, route StudentRoute) {
	// Contest routes
	mux.HandleFunc("/contests", route.GetContests)
	mux.HandleFunc("/contests/{contest_id}/tasks", route.GetContestTasks)
	mux.HandleFunc("/contests/{contest_id}/register", route.RegisterForContest)
	mux.HandleFunc("/contests/{contest_id}/task-progress", route.GetTaskProgress)

	// Submission routes
	mux.HandleFunc("/submissions", route.GetSubmissions)
	mux.HandleFunc("/submissions/languages", route.GetAvailableLanguages)
	mux.HandleFunc("/submissions/submit", route.SubmitSolution)

	// Task routes
	mux.HandleFunc("/tasks", route.GetAssignedTasks)
}
