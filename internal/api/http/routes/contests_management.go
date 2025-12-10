package routes

import (
	"encoding/json"
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

type ContestsManagementRoute interface {
	CreateContest(w http.ResponseWriter, r *http.Request)
	EditContest(w http.ResponseWriter, r *http.Request)
	DeleteContest(w http.ResponseWriter, r *http.Request)
	GetContestTasks(w http.ResponseWriter, r *http.Request)
	GetAssignableTasks(w http.ResponseWriter, r *http.Request)
	AddTaskToContest(w http.ResponseWriter, r *http.Request)
	RemoveTaskFromContest(w http.ResponseWriter, r *http.Request)
	GetRegistrationRequests(w http.ResponseWriter, r *http.Request)
	ApproveRegistrationRequest(w http.ResponseWriter, r *http.Request)
	RejectRegistrationRequest(w http.ResponseWriter, r *http.Request)
	GetContestSubmissions(w http.ResponseWriter, r *http.Request)
	GetCreatedContests(w http.ResponseWriter, r *http.Request)
	GetManageableContests(w http.ResponseWriter, r *http.Request)
	GetAllContests(w http.ResponseWriter, r *http.Request)
	GetContestTaskStats(w http.ResponseWriter, r *http.Request)
	GetContestTaskUserStats(w http.ResponseWriter, r *http.Request)
	GetContestTaskUserSubmissions(w http.ResponseWriter, r *http.Request)
	GetContestUserStats(w http.ResponseWriter, r *http.Request)
	AddGroupToContest(w http.ResponseWriter, r *http.Request)
	RemoveGroupFromContest(w http.ResponseWriter, r *http.Request)
	GetContestGroups(w http.ResponseWriter, r *http.Request)
	GetAssignableGroups(w http.ResponseWriter, r *http.Request)
}

type contestsManagementRouteImpl struct {
	contestService    service.ContestService
	submissionService service.SubmissionService
	logger            *zap.SugaredLogger
}

// CreateContest godoc
//
//	@Tags			contests-management
//	@Summary		Create a contest
//	@Description	Create a new contest
//	@Accept			json
//	@Produce		json
//	@Param			body	body		schemas.CreateContest	true	"Create Contest"
//	@Failure		400		{object}	httputils.ValidationErrorResponse
//	@Failure		403		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[httputils.IDResponse]
//	@Router			/contests-management/contests [post]
func (cr *contestsManagementRouteImpl) CreateContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.CreateContest
	err := httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		httputils.HandleValidationError(w, err)
		return
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	contestID, err := cr.contestService.Create(db, currentUser, &request)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewIDResponse(contestID))
}

// EditContest godoc
//
//	@Tags			contests-management
//	@Summary		Edit a contest
//	@Description	Edit contest details
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Contest ID"
//	@Param			body	body		schemas.EditContest	true	"Edit Contest"
//	@Failure		400		{object}	httputils.ValidationErrorResponse
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[schemas.CreatedContest]
//	@Router			/contests-management/contests/{id} [put]
func (cr *contestsManagementRouteImpl) EditContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.EditContest
	err := httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		httputils.HandleValidationError(w, err)
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

	resp, err := cr.contestService.Edit(db, currentUser, contestID, &request)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, resp)
}

// DeleteContest godoc
//
//	@Tags			contests-management
//	@Summary		Delete a contest
//	@Description	Delete a contest by ID
//	@Produce		json
//	@Param			id	path		int	true	"Contest ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/contests-management/contests/{id} [delete]
func (cr *contestsManagementRouteImpl) DeleteContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
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

	err = cr.contestService.Delete(db, currentUser, contestID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Contest deleted successfully"))
}

// GetAssignableTasks godoc
//
//	@Tags			contests-management
//	@Summary		Get available tasks for a contest
//	@Description	Get all tasks that are NOT yet assigned to the specified contest (admin/teacher only)
//
//	@Produce		json
//	@Param			id	path		int	true	"Contest ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[[]schemas.Task]
//	@Router			/contests-management/contests/{id}/tasks/assignable-tasks [get]
func (cr *contestsManagementRouteImpl) GetAssignableTasks(w http.ResponseWriter, r *http.Request) {
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

	tasks, err := cr.contestService.GetAssignableTasks(db, currentUser, contestID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, tasks)
}

// AddTaskToContest godoc
//
//	@Tags			contests-management
//	@Summary		Add a task to a contest
//	@Description	Add an existing task to a specific contest
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int							true	"Contest ID"
//	@Param			body	body		schemas.AddTaskToContest	true	"Add Task to Contest"
//	@Failure		400		{object}	httputils.ValidationErrorResponse
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/contests-management/contests/{id}/tasks [post]
func (cr *contestsManagementRouteImpl) AddTaskToContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	var request schemas.AddTaskToContest
	err = httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		httputils.HandleValidationError(w, err)
		return
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	err = cr.contestService.AddTaskToContest(db, currentUser, contestID, &request)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task added to contest successfully"))
}

// RemoveTaskFromContest godoc
//
//	@Tags			contests-management
//	@Summary		Remove tasks from a contest
//	@Description	Batch remove tasks from a specific contest. Existing submissions are preserved.
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int		true	"Contest ID"
//	@Param			body	body		[]int	true	"Task IDs"
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/contests-management/contests/{id}/tasks [delete]
func (cr *contestsManagementRouteImpl) RemoveTaskFromContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
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

	// Parse task IDs from request body
	var req struct {
		TaskIDs []int64 `json:"taskIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if len(req.TaskIDs) == 0 {
		httputils.ReturnError(w, http.StatusBadRequest, "Task IDs cannot be empty")
		return
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	// Batch remove tasks
	for _, taskID := range req.TaskIDs {
		if err := cr.contestService.RemoveTaskFromContest(db, currentUser, contestID, taskID); err != nil {
			httputils.HandleServiceError(w, err, db, cr.logger)
			return
		}
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Tasks removed from contest successfully"))
}

// GetRegistrationRequests godoc
//
//	@Tags			contests-management
//	@Summary		Get registration requests for a contest
//	@Description	Get all pending registration requests for a specific contest (only accessible by contest creator or admin)
//	@Produce		json
//	@Param			id		path		int		true	"Contest ID"
//	@Param			status	query		string	false	"Filter by status (pending, approved, rejected)"	default(pending)
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[[]schemas.RegistrationRequest]
//	@Router			/contests-management/contests/{id}/registration-requests [get]
func (cr *contestsManagementRouteImpl) GetRegistrationRequests(w http.ResponseWriter, r *http.Request) {
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

	statusQuery := r.URL.Query().Get("status")
	if statusQuery == "" {
		statusQuery = "pending"
	}
	status, ok := types.ParseRegistrationRequestStatus(statusQuery)
	if !ok {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid status value")
		return
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	requests, err := cr.contestService.GetRegistrationRequests(db, currentUser, contestID, status)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, requests)
}

func (cr *contestsManagementRouteImpl) ApproveRegistrationRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	contestStr := httputils.GetPathValue(r, "id")
	contestID, err := strconv.ParseInt(contestStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid contest ID")
		return
	}

	userStr := httputils.GetPathValue(r, "user_id")
	userID, err := strconv.ParseInt(userStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	err = cr.contestService.ApproveRegistrationRequest(db, currentUser, contestID, userID)
	if err != nil {
		// Fix me. Previously we still commited on ErrAlreadyParticipant
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Registration request approved successfully"))
}

// RejectRegistrationRequest godoc
//
//	@Tags			contests-management
//
//	@Summary		Reject a registration request
//	@Description	Reject a pending registration request for a contest (only accessible by contest creator or admin)
//
//	@Produce		json
//	@Param			id		path		int	true	"Contest ID"
//	@Param			user_id	path		int	true	"User ID"
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
func (cr *contestsManagementRouteImpl) RejectRegistrationRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	contestStr := httputils.GetPathValue(r, "id")
	contestID, err := strconv.ParseInt(contestStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid contest ID")
		return
	}

	userStr := httputils.GetPathValue(r, "user_id")
	userID, err := strconv.ParseInt(userStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	err = cr.contestService.RejectRegistrationRequest(db, currentUser, contestID, userID)
	if err != nil {
		// Fix me. Previously we still commited on ErrAlreadyParticipant
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Registration request rejected successfully"))
}

// GetContestTasks godoc
//
//	@Tags			contests-management
//	@Summary		Get tasks for a contest
//	@Description	Get all tasks associated with a specific contest
//
//	@Produce		json
//	@Param			id	path		int	true	"Contest ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[[]schemas.ContestTask]
//	@Router			/contests-management/contests/{id}/tasks [get]
func (cr *contestsManagementRouteImpl) GetContestTasks(w http.ResponseWriter, r *http.Request) {
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

	tasks, err := cr.contestService.GetTasksForContest(db, currentUser, contestID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, tasks)
}

// GetContestSubmissions godoc
//
//	@Tags			contests-management
//	@Summary		Get submissions for a contest
//	@Description	Get all submissions for a specific contest. Only accessible by teachers (contest creators) and admins.
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int		true	"Contest ID"
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Param			taskId	query		int		false	"Task ID"
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[schemas.PaginatedResult[[]schemas.Submission]]
//	@Router			/contests-management/contests/{id}/submissions [get]
func (cr *contestsManagementRouteImpl) GetContestSubmissions(w http.ResponseWriter, r *http.Request) {
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

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	paginationParams := httputils.ExtractPaginationParams(queryParams)

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	// Optional filtering by taskId
	var taskIDPtr *int64
	if v, ok := queryParams["taskId"]; ok {
		switch t := v.(type) {
		case string:
			if t != "" {
				tid, perr := strconv.ParseInt(t, 10, 64)
				if perr != nil {
					httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID")
					return
				}
				taskIDPtr = &tid
			}
		default:
			httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID")
			return
		}
	}

	var response *schemas.PaginatedResult[[]schemas.Submission]
	if taskIDPtr != nil {
		response, err = cr.submissionService.GetAll(db, currentUser, nil, taskIDPtr, &contestID, paginationParams)
	} else {
		response, err = cr.submissionService.GetAllForContest(db, contestID, currentUser, paginationParams)
	}
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, response)
}

// GetCreatedContests godoc
//
//	@Tags			contests-management
//	@Summary		This endpoint is deprecated and will be removed in future versions. Please use the /contests-management/contests endpoint instead.
//	@Description	Get all contests created by the currently authenticated user with pagination metadata
//	@Deprecated
//	@Produce	json
//	@Param		limit	query		int		false	"Limit"
//	@Param		offset	query		int		false	"Offset"
//	@Param		sort	query		string	false	"Sort"
//	@Failure	400		{object}	httputils.APIError
//	@Failure	403		{object}	httputils.APIError
//	@Failure	500		{object}	httputils.APIError
//	@Success	200		{object}	httputils.APIResponse[schemas.PaginatedResult[[]schemas.ManagedContest]]
//	@Router		/contests-management/contests/managed [get]
func (cr *contestsManagementRouteImpl) GetCreatedContests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	paginationParams := httputils.ExtractPaginationParams(queryParams)

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	response, err := cr.contestService.GetContestsCreatedByUser(db, currentUser.ID, paginationParams)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, response)
}

// GetContestTaskStats godoc
//
//	@Tags			contests-management
//	@Summary		Get task statistics for a contest
//	@Description	Get aggregated statistics for each task in a contest (for contest creators or admins)
//	@Produce		json
//	@Param			id	path		int	true	"Contest ID"
//	@Success		200	{object}	httputils.APIResponse[[]schemas.ContestTaskStats]
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Router			/contests-management/contests/{id}/task-stats [get]
func (cr *contestsManagementRouteImpl) GetContestTaskStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	contestIDStr := httputils.GetPathValue(r, "id")
	contestID, err := strconv.ParseInt(contestIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid contest ID")
		return
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	stats, err := cr.submissionService.GetTaskStatsForContest(db, currentUser, contestID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, stats)
}

// GetContestTaskUserStats godoc
//
//	@Tags			contests-management
//	@Summary		Get user statistics for a specific task in a contest
//	@Description	Get per-user statistics for a task in a contest (for contest/task creators or admins)
//	@Produce		json
//	@Param			id		path		int	true	"Contest ID"
//	@Param			taskId	path		int	true	"Task ID"
//	@Success		200		{object}	httputils.APIResponse[[]schemas.TaskUserStats]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/contests-management/contests/{id}/tasks/{taskId}/user-stats [get]
func (cr *contestsManagementRouteImpl) GetContestTaskUserStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	contestIDStr := httputils.GetPathValue(r, "id")
	contestID, err := strconv.ParseInt(contestIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid contest ID")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "taskId")
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	stats, err := cr.submissionService.GetUserStatsForContestTask(db, currentUser, contestID, taskID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, stats)
}

// GetContestTaskUserSubmissions godoc
//
//	@Tags			contests-management
//	@Summary		Get submissions for a specific user, task, and contest
//	@Description	Get all submissions for a user on a task in a contest with filtering (for contest/task creators or admins)
//	@Produce		json
//	@Param			id		path		int		true	"Contest ID"
//	@Param			taskId	path		int		true	"Task ID"
//	@Param			userId	path		int		true	"User ID"
//	@Param			status	query		string	false	"Filter by submission status"
//	@Param			limit	query		int		false	"Limit the number of returned submissions"
//	@Param			offset	query		int		false	"Offset the returned submissions"
//	@Param			sort	query		string	false	"Sort order"
//	@Success		200		{object}	httputils.APIResponse[schemas.PaginatedResult[[]schemas.Submission]]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/contests-management/contests/{id}/tasks/{taskId}/users/{userId}/submissions [get]
func (cr *contestsManagementRouteImpl) GetContestTaskUserSubmissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	contestIDStr := httputils.GetPathValue(r, "id")
	contestID, err := strconv.ParseInt(contestIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid contest ID")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "taskId")
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	userIDStr := httputils.GetPathValue(r, "userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)
	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any) // TODO: create function to extract query params
	paginationParams := httputils.ExtractPaginationParams(queryParams)

	// Use the existing GetAll with filters for contest, task, and user
	submissions, err := cr.submissionService.GetAll(db, currentUser, &userID, &taskID, &contestID, paginationParams)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, submissions)
}

// GetContestUserStats godoc
//
//	@Tags			contests-management
//	@Summary		Get overall user statistics for a contest
//	@Description	Get overall performance statistics for all users (or a specific user) in a contest (for contest creators or admins)
//	@Produce		json
//	@Param			id		path		int	true	"Contest ID"
//	@Param			userId	query		int	false	"Filter by specific user ID"
//	@Success		200		{object}	httputils.APIResponse[[]schemas.UserContestStats]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/contests-management/contests/{id}/user-stats [get]
func (cr *contestsManagementRouteImpl) GetContestUserStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	contestIDStr := httputils.GetPathValue(r, "id")
	contestID, err := strconv.ParseInt(contestIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid contest ID")
		return
	}

	// Optional user filter
	var userID *int64
	userIDStr := r.URL.Query().Get("userId")
	if userIDStr != "" {
		userIDVal, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			httputils.ReturnError(w, http.StatusBadRequest, "Invalid user ID")
			return
		}
		userID = &userIDVal
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	stats, err := cr.submissionService.GetUserStatsForContest(db, currentUser, contestID, userID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, stats)
}

// AddGroupToContest godoc
//
//	@Tags			contests-management
//	@Summary		Add a group to a contest
//	@Description	Add a group as participants to a specific contest (only accessible by contest collaborators with edit permission)
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int		true	"Contest ID"
//	@Param			body	body		[]int	true	"Group IDs"
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/contests-management/contests/{id}/groups [post]
func (cr *contestsManagementRouteImpl) AddGroupToContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	// Parse group IDs from request body
	var req struct {
		GroupIDs []int64 `json:"groupIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if len(req.GroupIDs) == 0 {
		httputils.ReturnError(w, http.StatusBadRequest, "Group IDs cannot be empty")
		return
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	// Batch add groups
	for _, groupID := range req.GroupIDs {
		if err := cr.contestService.AddGroupToContest(db, currentUser, contestID, groupID); err != nil {
			httputils.HandleServiceError(w, err, db, cr.logger)
			return
		}
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Groups added to contest successfully"))
}

// RemoveGroupFromContest godoc
//
//	@Tags			contests-management
//	@Summary		Remove a group from a contest
//	@Description	Remove a group from contest participants (only accessible by contest collaborators with edit permission)
//	@Produce		json
//	@Param			id		path		int		true	"Contest ID"
//	@Param			body	body		[]int	true	"Group IDs"
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/contests-management/contests/{id}/groups [delete]
func (cr *contestsManagementRouteImpl) RemoveGroupFromContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
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

	// Parse group IDs from request body
	var req struct {
		GroupIDs []int64 `json:"groupIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if len(req.GroupIDs) == 0 {
		httputils.ReturnError(w, http.StatusBadRequest, "Group IDs cannot be empty")
		return
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	// Batch remove groups
	for _, groupID := range req.GroupIDs {
		if err := cr.contestService.RemoveGroupFromContest(db, currentUser, contestID, groupID); err != nil {
			httputils.HandleServiceError(w, err, db, cr.logger)
			return
		}
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Groups removed from contest successfully"))
}

// GetContestGroups godoc
//
//	@Tags			contests-management
//	@Summary		Get contest groups information
//	@Description	Get groups assigned to a contest and groups that can be assigned (only accessible by contest collaborators with edit permission)
//	@Produce		json
//	@Param			id	path		int	true	"Contest ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[[]schemas.Group]
//	@Router			/contests-management/contests/{id}/groups [get]
func (cr *contestsManagementRouteImpl) GetContestGroups(w http.ResponseWriter, r *http.Request) {
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

	groups, err := cr.contestService.GetContestGroups(db, currentUser, contestID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, groups)
}

// GetAssignableGroups godoc
//
//	@Tags			contests-management
//	@Summary		Get assignable groups for a contest
//	@Description	Get groups that can be assigned to a specific contest (only accessible by contest collaborators with edit permission)
//	@Produce		json
//	@Param			id	path		int	true	"Contest ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[[]schemas.Group]
//	@Router			/contests-management/contests/{id}/groups/assignable [get]
func (cr *contestsManagementRouteImpl) GetAssignableGroups(w http.ResponseWriter, r *http.Request) {
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

	groups, err := cr.contestService.GetAssignableGroups(db, currentUser, contestID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, groups)
}

// GetManageableContests godoc
//
//	@Tags			contests-management
//	@Summary		Get contests manageable by the current user
//	@Description	Get all contests where the current user is the creator or has collaborator access (view, edit, or manage) with pagination metadata
//	@Produce		json
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[schemas.PaginatedResult[[]schemas.ManagedContest]]
//	@Router			/contests-management/contests/managed [get]
func (cr *contestsManagementRouteImpl) GetManageableContests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	currentUser := httputils.GetCurrentUser(r)
	db := httputils.GetDatabase(r)
	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	paginationParams := httputils.ExtractPaginationParams(queryParams)

	response, err := cr.contestService.GetManagedContests(db, currentUser.ID, paginationParams)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, response)
}

// GetAllContests godoc
//
//	@Tags			contests-management, admin
//	@Summary		Get all contests (admin only)
//	@Description	Returns all contests in the system. Accessible only to admins.
//	@Produce		json
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[schemas.PaginatedResult[[]schemas.CreatedContest]]
//	@Router			/contests-management/contests [get]
func (cr *contestsManagementRouteImpl) GetAllContests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	currentUser := httputils.GetCurrentUser(r)
	db := httputils.GetDatabase(r)
	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	paginationParams := httputils.ExtractPaginationParams(queryParams)

	response, err := cr.contestService.GetAllContests(db, currentUser, paginationParams)
	if err != nil {
		httputils.HandleServiceError(w, err, db, cr.logger)
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, response)
}

func RegisterContestsManagementRoute(mux *mux.Router, route ContestsManagementRoute) {
	mux.HandleFunc("/contests", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			route.CreateContest(w, r)
		case http.MethodGet:
			route.GetAllContests(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/contests/created", route.GetCreatedContests)
	mux.HandleFunc("/contests/managed", route.GetManageableContests)

	mux.HandleFunc("/contests/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			route.EditContest(w, r)
		case http.MethodDelete:
			route.DeleteContest(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/contests/{id}/submissions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetContestSubmissions(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/contests/{id}/tasks/assignable-tasks", route.GetAssignableTasks)

	mux.HandleFunc("/contests/{id}/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetContestTasks(w, r)
		case http.MethodPost:
			route.AddTaskToContest(w, r)
		case http.MethodDelete:
			route.RemoveTaskFromContest(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/contests/{id}/registration-requests", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetRegistrationRequests(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/contests/{id}/registration-requests/{user_id}/approve", route.ApproveRegistrationRequest)
	mux.HandleFunc("/contests/{id}/registration-requests/{user_id}/reject", route.RejectRegistrationRequest)

	// Group management endpoints
	mux.HandleFunc("/contests/{id}/groups/assignable", route.GetAssignableGroups)
	mux.HandleFunc("/contests/{id}/groups", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetContestGroups(w, r)
		case http.MethodPost:

			route.AddGroupToContest(w, r)
		case http.MethodDelete:
			route.RemoveGroupFromContest(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// New statistics endpoints
	mux.HandleFunc("/contests/{id}/task-stats", route.GetContestTaskStats)
	mux.HandleFunc("/contests/{id}/user-stats", route.GetContestUserStats)
	mux.HandleFunc("/contests/{id}/tasks/{taskId}/user-stats", route.GetContestTaskUserStats)
	mux.HandleFunc("/contests/{id}/tasks/{taskId}/users/{userId}/submissions", route.GetContestTaskUserSubmissions)
}

func NewContestsManagementRoute(contestService service.ContestService, submissionService service.SubmissionService) ContestsManagementRoute {
	route := &contestsManagementRouteImpl{
		contestService:    contestService,
		submissionService: submissionService,
		logger:            utils.NewNamedLogger("contests-management-route"),
	}
	if err := utils.ValidateStruct(*route); err != nil {
		panic("Invalid contests management route parameters: " + err.Error())
	}
	return route
}
