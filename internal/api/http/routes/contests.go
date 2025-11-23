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

type ContestRoute interface {
	GetContest(w http.ResponseWriter, r *http.Request)
	GetMyContests(w http.ResponseWriter, r *http.Request) // legacy combined endpoint
	GetMyActiveContests(w http.ResponseWriter, r *http.Request)
	GetMyPastContests(w http.ResponseWriter, r *http.Request)
	GetContests(w http.ResponseWriter, r *http.Request)
	RegisterForContest(w http.ResponseWriter, r *http.Request)
	GetTaskProgressForContest(w http.ResponseWriter, r *http.Request)
	GetContestTask(w http.ResponseWriter, r *http.Request)
}

type ContestRouteImpl struct {
	contestService    service.ContestService
	submissionService service.SubmissionService
	logger            *zap.SugaredLogger
}

// GetContest godoc
//
//	@Tags			contests
//	@Summary		Get a contest
//	@Description	Get contest details by ID
//	@Produce		json
//	@Param			id	path		int	true	"Contest ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[schemas.Contest]
//	@Router			/contests/{id} [get]
func (cr *ContestRouteImpl) GetContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		cr.logger.Errorw("Failed to begin database transaction", "error", err)
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

	currentUser := httputils.GetCurrentUser(r)

	contest, err := cr.contestService.Get(tx, currentUser, contestID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else {
			cr.logger.Errorw("Failed to get contest", "error", err)
		}
		httputils.ReturnError(w, status, "Contest retrieval failed")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, contest)
}

// GetContests godoc
//
//	@Tags			contests
//	@Summary		Get global contests
//	@Description	Get global contests accessible to the (ongoing, upcoming, past) with pagination metadata
//	@Produce		json
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Param			status	query		string	true	"Contest status"	Enums(ongoing, upcoming, past)
//	@Success		200		{object}	httputils.APIResponse[schemas.PaginatedResult[[]schemas.AvailableContest]]
//	@Failure		401		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/contests [get]
func (cr *ContestRouteImpl) GetContests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		cr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	currentUser := httputils.GetCurrentUser(r)
	paginationParams := httputils.ExtractPaginationParams(queryParams)

	status, ok := queryParams["status"]
	if !ok {
		httputils.ReturnError(w, http.StatusBadRequest, "Status query parameter is required")
		return
	}
	statusStr, ok := status.(string)
	if !ok {
		httputils.ReturnError(w, http.StatusBadRequest, "Status query parameter must be a string")
		return
	}
	var response schemas.PaginatedResult[[]schemas.AvailableContest]
	switch statusStr {
	case "ongoing":
		response, err = cr.contestService.GetOngoingContests(tx, currentUser, paginationParams)
	case "upcoming":
		response, err = cr.contestService.GetUpcomingContests(tx, currentUser, paginationParams)
	case "past":
		response, err = cr.contestService.GetPastContests(tx, currentUser, paginationParams)
	default:
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid status query parameter")
		return
	}
	if err != nil {
		db.Rollback()
		cr.logger.Errorw("Failed to get contests", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Contest service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, response)
}

// RegisterForContest godoc
//
//	@Tags			contests
//	@Summary		Register for a contest
//	@Description	Create a pending registration for a contest (requires contest creator approval)
//	@Produce		json
//	@Param			id	path		int	true	"Contest ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		409	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/contests/{id}/register [post]
func (cr *ContestRouteImpl) RegisterForContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		cr.logger.Errorw("Failed to begin database transaction", "error", err)
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

	currentUser := httputils.GetCurrentUser(r)

	err = cr.contestService.RegisterForContest(tx, currentUser, contestID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, myerrors.ErrForbidden):
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
			cr.logger.Errorw("Failed to register for contest", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to register for contest")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Registration request submitted successfully"))
}

// GetTaskProgressForContest godoc
//
//	@Tags			contests
//	@Summary		Get tasks for a contest with submission statistics
//	@Description	Get tasks associated with a specific contest including best score and attempt count for the current user
//
//	@Produce		json
//	@Param			id	path		int	true	"Contest ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[[]schemas.TaskWithContestStats]
//	@Router			/contests/{id}/tasks/user-statistics [get]
func (cr *ContestRouteImpl) GetTaskProgressForContest(w http.ResponseWriter, r *http.Request) {
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

	currentUser := httputils.GetCurrentUser(r)
	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		cr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}
	tasks, err := cr.contestService.GetTaskProgressForContest(tx, currentUser, contestID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else {
			cr.logger.Errorw("Failed to get task progress for contest", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to get task progress for contest")
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, tasks)
}

// GetMyContests godoc
//
//	@Tags			contests
//	@Summary		Get contests for a user (combined - deprecated; use /contests/my/active or /contests/my/past)
//	@Description	Get contests for a user (returns ongoing, past, upcoming)
//	@Produce		json
//	@Deprecated
//	@Failure	400	{object}	httputils.APIError
//	@Failure	404	{object}	httputils.APIError
//	@Failure	405	{object}	httputils.APIError
//	@Failure	500	{object}	httputils.APIError
//	@Success	200	{object}	httputils.APIResponse[schemas.UserContestsWithStats]
//	@Router		/contests/my [get]
func (cr *ContestRouteImpl) GetMyContests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		cr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := httputils.GetCurrentUser(r)

	contests, err := cr.contestService.GetUserContests(tx, currentUser.ID)
	if err != nil {
		db.Rollback()
		if errors.Is(err, myerrors.ErrUserNotFound) {
			httputils.ReturnError(w, http.StatusNotFound, "User not found")
			return
		}
		cr.logger.Errorw("Failed to get user contests", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Contest service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, contests)
}

// GetMyActiveContests godoc
//
//	@Tags			contests
//	@Summary		Get active contests for the current user
//	@Description	Active contests = started AND not finished (ongoing)
//	@Produce		json
//	@Success		200	{object}	httputils.APIResponse[[]schemas.ContestWithStats]
//	@Failure		404	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Router			/contests/my/active [get]
func (cr *ContestRouteImpl) GetMyActiveContests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		cr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := httputils.GetCurrentUser(r)
	combined, err := cr.contestService.GetUserContests(tx, currentUser.ID)
	if err != nil {
		db.Rollback()
		if errors.Is(err, myerrors.ErrUserNotFound) {
			httputils.ReturnError(w, http.StatusNotFound, "User not found")
			return
		}
		cr.logger.Errorw("Failed to get active contests", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Contest service temporarily unavailable")
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, combined.Ongoing)
}

// GetMyPastContests godoc
//
//	@Tags			contests
//	@Summary		Get past contests for the current user
//	@Description	Past contests = contests whose end time has elapsed
//	@Produce		json
//	@Success		200	{object}	httputils.APIResponse[[]schemas.PastContestWithStats]
//	@Failure		404	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Router			/contests/my/past [get]
func (cr *ContestRouteImpl) GetMyPastContests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		cr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := httputils.GetCurrentUser(r)
	combined, err := cr.contestService.GetUserContests(tx, currentUser.ID)
	if err != nil {
		db.Rollback()
		if errors.Is(err, myerrors.ErrUserNotFound) {
			httputils.ReturnError(w, http.StatusNotFound, "User not found")
			return
		}
		cr.logger.Errorw("Failed to get past contests", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Contest service temporarily unavailable")
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, combined.Past)
}

// GetContestTask godoc
//
//	@Tags			contests
//	@Summary		Get a specific task from a contest
//	@Description	Get detailed information about a specific task within a contest
//	@Produce		json
//	@Param			id		path		int	true	"Contest ID"
//	@Param			task_id	path		int	true	"Task ID"
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[schemas.TaskDetailed]
//	@Router			/contests/{id}/tasks/{task_id} [get]
func (cr *ContestRouteImpl) GetContestTask(w http.ResponseWriter, r *http.Request) {
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
	taskStr := httputils.GetPathValue(r, "task_id")
	if taskStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID cannot be empty")
		return
	}
	taskID, err := strconv.ParseInt(taskStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	currentUser := httputils.GetCurrentUser(r)

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		cr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	task, err := cr.contestService.GetContestTask(tx, currentUser, contestID, taskID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, myerrors.ErrNotFound) || errors.Is(err, myerrors.ErrTaskNotInContest):
			status = http.StatusNotFound
		case errors.Is(err, myerrors.ErrForbidden):
			status = http.StatusForbidden
		default:
			cr.logger.Errorw("Failed to get contest task", "error", err)
		}

		httputils.ReturnError(w, status, "Failed to get contest task")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, task)
}

func RegisterContestRoutes(mux *mux.Router, contestRoute ContestRoute) {
	mux.HandleFunc("/contests", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			contestRoute.GetContests(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// Legacy combined endpoint (optional keep); new split endpoints:
	// /contests/my/active and /contests/my/past
	mux.HandleFunc("/contests/my", contestRoute.GetMyContests)
	mux.HandleFunc("/contests/my/active", contestRoute.GetMyActiveContests)
	mux.HandleFunc("/contests/my/past", contestRoute.GetMyPastContests)

	mux.HandleFunc("/contests/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			contestRoute.GetContest(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/contests/{id}/tasks/user-statistics", contestRoute.GetTaskProgressForContest)
	mux.HandleFunc("/contests/{id}/tasks/{task_id}", contestRoute.GetContestTask)

	mux.HandleFunc("/contests/{id}/register", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			contestRoute.RegisterForContest(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
}

func NewContestRoute(contestService service.ContestService, submissionService service.SubmissionService) ContestRoute {
	route := &ContestRouteImpl{
		contestService:    contestService,
		submissionService: submissionService,
		logger:            utils.NewNamedLogger("contests"),
	}
	err := utils.ValidateStruct(*route)
	if err != nil {
		log.Panicf("ContestRoute struct is not valid: %s", err.Error())
	}
	return route
}
