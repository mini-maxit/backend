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

// StudentRoute defines all student-specific routes.
type StudentRoute interface {
	// Contest routes
	GetContests(w http.ResponseWriter, r *http.Request)
	GetContest(w http.ResponseWriter, r *http.Request)
	GetContestTasks(w http.ResponseWriter, r *http.Request)
	RegisterForContest(w http.ResponseWriter, r *http.Request)
	GetTaskProgress(w http.ResponseWriter, r *http.Request)

	// Submission routes
	GetSubmissions(w http.ResponseWriter, r *http.Request)

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
//	@Tags			student
//	@Summary		Get contests for student
//	@Description	Get all contests accessible to the student (ongoing, upcoming, past)
//	@Produce		json
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Success		200		{object}	httputils.APIResponse[[]schemas.AvailableContest]
//	@Failure		401		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/student/contests [get]
func (sr *studentRouteImpl) GetContests(w http.ResponseWriter, r *http.Request) {
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

	status, ok := queryParams["status"]
	if !ok {
		httputils.ReturnError(w, http.StatusBadRequest, "Status query parameter is required")
	}
	statusStr, ok := status.(string)
	if !ok {
		httputils.ReturnError(w, http.StatusBadRequest, "Status query parameter must be a string")
	}
	var contests []schemas.AvailableContest
	switch statusStr {
	case "ongoing":
		contests, err = sr.contestService.GetOngoingContests(tx, currentUser, paginationParams)
	case "upcoming":
		contests, err = sr.contestService.GetUpcomingContests(tx, currentUser, paginationParams)
	case "past":
		contests, err = sr.contestService.GetPastContests(tx, currentUser, paginationParams)
	default:
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid status query parameter")
		return
	}
	if err != nil {
		db.Rollback()
		sr.logger.Errorw("Failed to get contests", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Contest service temporarily unavailable")
		return
	}

	if contests == nil {
		contests = []schemas.AvailableContest{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, contests)
}

// GetContest godoc
//
//	@Tags			student
//	@Summary		Get contest details
//	@Description	Get details of a specific contest
//	@Produce		json
//	@Param			id	path		int	true	"Contest ID"
//	@Success		200	{object}	httputils.APIResponse[schemas.Contest]
//	@Failure		400	{object}	httputils.APIError
//	@Failure		401	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Router			/student/contests/{id} [get]
func (sr *studentRouteImpl) GetContest(w http.ResponseWriter, r *http.Request) {
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

	contest, err := sr.contestService.Get(tx, currentUser, contestID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			sr.logger.Errorw("Failed to get contest", "error", err)
		}
		httputils.ReturnError(w, status, "Contest retrieval failed")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, contest)
}

// GetContestTasks godoc
//
//	@Tags			student
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
//	@Tags			student
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
//	@Tags			student
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
//	@Tags			student
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
//	@Tags			student
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
	mux.HandleFunc("/contests/{id}", route.GetContest)
	mux.HandleFunc("/contests/{contest_id}/tasks", route.GetContestTasks)
	mux.HandleFunc("/contests/{contest_id}/register", route.RegisterForContest)
	mux.HandleFunc("/contests/{contest_id}/task-progress", route.GetTaskProgress)

	// Submission routes
	mux.HandleFunc("/submissions", route.GetSubmissions)

	// Task routes
	mux.HandleFunc("/tasks", route.GetAssignedTasks)
}
