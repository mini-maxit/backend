package routes

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
)

type ContestRoute interface {
	GetContest(w http.ResponseWriter, r *http.Request)
	GetContestTask(w http.ResponseWriter, r *http.Request)
	GetContestTasks(w http.ResponseWriter, r *http.Request)
	GetMyActiveContests(w http.ResponseWriter, r *http.Request)
	GetMyContestResults(w http.ResponseWriter, r *http.Request)
	GetMyPastContests(w http.ResponseWriter, r *http.Request)
	GetTaskProgressForContest(w http.ResponseWriter, r *http.Request)
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
//	@Success		200	{object}	httputils.APIResponse[schemas.ContestDetailed]
//	@Router			/contests/{id} [get]
func (cr *ContestRouteImpl) GetContest(w http.ResponseWriter, r *http.Request) {
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

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	contest, err := cr.contestService.GetDetailed(db, currentUser, contestID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, contest)
}

// GetContestTasks godoc
//
//	@Tags			contests
//	@Summary		Get contest tasks with optional status filter
//	@Description	Get already-started tasks for a contest with optional status filter (past, ongoing, upcoming). Only accessible by participants or users with access policy.
//	@Produce		json
//	@Param			id		path		int		true	"Contest ID"
//	@Param			status	query		string	false	"Task status filter"	Enums(past, ongoing, upcoming)
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[[]schemas.ContestTask]
//	@Router			/contests/{id}/tasks [get]
func (cr *ContestRouteImpl) GetContestTasks(w http.ResponseWriter, r *http.Request) {
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

	// Parse optional status filter
	var status types.ContestStatus
	statusStr := r.URL.Query().Get("status")
	if statusStr != "" {
		switch statusStr {
		case "past":
			status = types.ContestStatusPast
		case "ongoing":
			status = types.ContestStatusOngoing
		case "upcoming":
			status = types.ContestStatusUpcoming
		default:
			httputils.ReturnError(w, http.StatusBadRequest, "Invalid status value. Must be 'past', 'ongoing', or 'upcoming'")
			return
		}
	} else {
		status = types.ContestStatusOngoing
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	tasks, err := cr.contestService.GetStartedTasksForContest(db, currentUser, contestID, status)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, tasks)
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

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	tasks, err := cr.contestService.GetTaskProgressForContest(db, currentUser, contestID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, tasks)
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

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	combined, err := cr.contestService.GetUserContests(db, currentUser.ID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}
	result := make([]schemas.ContestWithStats, 0)
	result = append(result, combined.Ongoing...)
	result = append(result, combined.Upcoming...)
	httputils.ReturnSuccess(w, http.StatusOK, result)
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

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	combined, err := cr.contestService.GetUserContests(db, currentUser.ID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
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

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	task, err := cr.contestService.GetContestTask(db, currentUser, contestID, taskID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, task)
}

// GetMyContestResults godoc
//
//	@Tags			contests
//	@Summary		Get user's results for a contest
//	@Description	Get results of the current user for a given contest, including attempt count, best result, and best submission ID for each task
//	@Produce		json
//	@Param			id	path		int	true	"Contest ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[schemas.ContestResults]
//	@Router			/contests/{id}/results/my [get]
func (cr *ContestRouteImpl) GetMyContestResults(w http.ResponseWriter, r *http.Request) {
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
	db := httputils.GetDatabase(r)

	results, err := cr.contestService.GetMyContestResults(db, currentUser, contestID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, results)
}

func RegisterContestRoutes(mux *mux.Router, contestRoute ContestRoute) {
	mux.HandleFunc("/contests/my/active", contestRoute.GetMyActiveContests)
	mux.HandleFunc("/contests/my/past", contestRoute.GetMyPastContests)

	// Contest tasks endpoint with status filter
	mux.HandleFunc("/contests/{id}/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			contestRoute.GetContestTasks(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

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
	mux.HandleFunc("/contests/{id}/results/my", contestRoute.GetMyContestResults)
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
