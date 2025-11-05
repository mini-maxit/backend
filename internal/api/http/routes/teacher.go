package routes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	myerrors "github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
)

// TeacherRoute defines all teacher-specific routes.
type TeacherRoute interface {
	// Contest management routes
	GetManageableContests(w http.ResponseWriter, r *http.Request)
	CreateContest(w http.ResponseWriter, r *http.Request)
	EditContest(w http.ResponseWriter, r *http.Request)
	DeleteContest(w http.ResponseWriter, r *http.Request)
	GetContestSubmissions(w http.ResponseWriter, r *http.Request)
	GetContestTasks(w http.ResponseWriter, r *http.Request)
	GetAssignableTasksToContest(w http.ResponseWriter, r *http.Request)
	AddTaskToContest(w http.ResponseWriter, r *http.Request)
	GetContestRegistrationRequests(w http.ResponseWriter, r *http.Request)
	ApproveContestRegistrationRequest(w http.ResponseWriter, r *http.Request)
	RejectContestRegistrationRequest(w http.ResponseWriter, r *http.Request)

	// Task management routes
	GetTasksManagement(w http.ResponseWriter, r *http.Request)
	GetTaskManagement(w http.ResponseWriter, r *http.Request)
	CreateTask(w http.ResponseWriter, r *http.Request)
	EditTask(w http.ResponseWriter, r *http.Request)
	DeleteTask(w http.ResponseWriter, r *http.Request)
	AssignTaskToUsers(w http.ResponseWriter, r *http.Request)
	AssignTaskToGroups(w http.ResponseWriter, r *http.Request)
	UnAssignTaskFromUsers(w http.ResponseWriter, r *http.Request)
	UnAssignTaskFromGroups(w http.ResponseWriter, r *http.Request)
	GetTaskLimits(w http.ResponseWriter, r *http.Request)
	PutTaskLimits(w http.ResponseWriter, r *http.Request)

	// Group management routes
	GetGroupsManagement(w http.ResponseWriter, r *http.Request)
	GetGroupManagement(w http.ResponseWriter, r *http.Request)
	CreateGroup(w http.ResponseWriter, r *http.Request)
	EditGroup(w http.ResponseWriter, r *http.Request)
	AddUsersToGroup(w http.ResponseWriter, r *http.Request)
	DeleteUsersFromGroup(w http.ResponseWriter, r *http.Request)
	GetGroupUsers(w http.ResponseWriter, r *http.Request)
	GetGroupTasks(w http.ResponseWriter, r *http.Request)
}

type teacherRouteImpl struct {
	contestService    service.ContestService
	taskService       service.TaskService
	groupService      service.GroupService
	submissionService service.SubmissionService
	logger            *zap.SugaredLogger
}

// GetManageableContests godoc
//
//	@Tags			contests-management
//	@Summary		Get contests treated by teacher
//	@Destription	Get all contests created or owned by the authenticated teacher
//	@Produce		json
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Success		200		{object}	httputils.APIResponse[[]schemas.Contest]
//	@Failure		401		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/teacher/contests [get]
func (tr *teacherRouteImpl) GetManageableContests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	paginationParams := httputils.ExtractPaginationParams(queryParams)
	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	contests, err := tr.contestService.GetAllManageable(tx, currentUser, paginationParams)
	if err != nil {
		db.Rollback()
		switch {
		case errors.Is(err, myerrors.ErrNotAuthorized):
			httputils.ReturnError(w, http.StatusForbidden, err.Error())
			return
		default:
			tr.logger.Errorw("Failed to get manageable contests", "error", err)
		}
		httputils.ReturnError(w, http.StatusInternalServerError, "Contest service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, contests)
}

// GetContestManagement godoc
//
//	@Tags		    contests-management
//	@Summary		Get contest details
//	@Destription	Get details of a specific contest created by teacher
//	@Produce		json
//	@Param			id	path		int	true	"Contest ID"
//	@Success		200	{object}	httputils.APIResponse[schemas.Contest]
//	@Failure		400	{object}	httputils.APIError
//	@Failure		401	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Router			/teacher/contests/{id} [get]
func (tr *teacherRouteImpl) GetContestManagement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	contestIDStr := httputils.GetPathValue(r, "id")
	if contestIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Contest ID cannot be empty")
		return
	}

	contestID, err := strconv.ParseInt(contestIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid contest ID: "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	contest, err := tr.contestService.Get(tx, currentUser, contestID)
	if err != nil {
		db.Rollback()
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			httputils.ReturnError(w, http.StatusForbidden, "You are not authorized to view this contest")
			return
		}
		tr.logger.Errorw("Failed to get contest", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Contest service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, contest)
}

// CreateContest godoc
//
//	@Tags			contests-management
//	@Summary		treate contest
//	@Destription	Create a new contest
//	@Accept			json
//	@Produce		json
//	@Param			body	body		schemas.CreateContest	true	"Create Contest"
//	@Success		201		{object}	httputils.APIResponse[schemas.Contest]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		401		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/teacher/contests [post]
func (tr *teacherRouteImpl) CreateContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.CreateContest
	err := httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			httputils.ReturnValidationError(w, valErrs)
			return
		}
		httputils.ReturnError(w, http.StatusBadRequest, "Could not validate request data.")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	contestID, err := tr.contestService.Create(tx, currentUser, &request)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			tr.logger.Errorw("Failed to treate contest", "error", err)
		}
		httputils.ReturnError(w, status, "Contest treation failed")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewIDResponse(contestID))
}

// EditContest godoc
//
//	@Tags			contests-management
//	@Summary		Edit contest
//	@Destription	Edit an existing contest
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Contest ID"
//	@Param			body	body		schemas.EditContest	true	"Edit Contest"
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		401		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/teacher/contests/{id} [put]
func (tr *teacherRouteImpl) EditContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.EditContest
	err := httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			httputils.ReturnValidationError(w, valErrs)
			return
		}
		httputils.ReturnError(w, http.StatusBadRequest, "Could not validate request data.")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
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

	resp, err := tr.contestService.Edit(tx, currentUser, contestID, &request)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else {
			tr.logger.Errorw("Failed to edit contest", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to edit contest")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, resp)
}

// DeleteContest godoc
//
//	@Tags			contests-management
//	@Summary		Delete contest
//	@Destription	Delete a contest
//	@Produce		json
//	@Param			id	path		int	true	"Contest ID"
//	@Success		200	{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Failure		400	{object}	httputils.APIError
//	@Failure		401	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Router			/teacher/contests/{id} [delete]
func (tr *teacherRouteImpl) DeleteContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
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

	err = tr.contestService.Delete(tx, currentUser, contestID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else {
			tr.logger.Errorw("Failed to delete contest", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to delete contest")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Contest deleted successfully"))
}

// GetContestSubmissions godoc
//
//	@Tags			contests-management
//	@Summary		Get contest submissions
//	@Destription	Get all submissions for a contest
//	@Produce		json
//	@Param			id		path		int		true	"Contest ID"
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Success		200		{object}	httputils.APIResponse[[]schemas.Submission]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		401		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/teacher/contests/{id}/submissions [get]
func (tr *teacherRouteImpl) GetContestSubmissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
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

	// Only teachers and admins can view all contest submissions
	if err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher}); err != nil {
		httputils.ReturnError(w, http.StatusForbidden, "Only teachers and admins can view contest submissions")
		return
	}

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)

	submissions, err := tr.submissionService.GetAllForContest(tx, contestID, currentUser, queryParams)
	if err != nil {
		db.Rollback()
		switch {
		case errors.Is(err, myerrors.ErrNotFound):
			httputils.ReturnError(w, http.StatusNotFound, "Contest not found")
		case errors.Is(err, myerrors.ErrPermissionDenied):
			httputils.ReturnError(w, http.StatusForbidden, "Permission denied. You are not the creator of this contest.")
		default:
			tr.logger.Errorw("Failed to get contest submissions", "error", err)
			httputils.ReturnError(w, http.StatusInternalServerError, "Failed to get contest submissions")
		}
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, submissions)
}

// GetContestTasks godoc
//
//	@Tags			contests-management
//	@Summary		Get contest tasks
//	@Destription	Get all tasks for a specific contest
//	@Produce		json
//	@Param			id	path		int	true	"Contest ID"
//	@Success		200	{object}	httputils.APIResponse[[]schemas.Task]
//	@Failure		400	{object}	httputils.APIError
//	@Failure		401	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Router			/teacher/contests/{id}/tasks [get]
func (tr *teacherRouteImpl) GetContestTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	contestIDStr := httputils.GetPathValue(r, "id")
	if contestIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Contest ID cannot be empty")
		return
	}

	contestID, err := strconv.ParseInt(contestIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid contest ID: "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	tasks, err := tr.contestService.GetTasksForContest(tx, currentUser, contestID)
	if err != nil {
		db.Rollback()
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			httputils.ReturnError(w, http.StatusForbidden, "You are not authorized to view this contest")
			return
		}
		tr.logger.Errorw("Failed to get contest tasks", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Contest service temporarily unavailable")
		return
	}

	if tasks == nil {
		tasks = []schemas.Task{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, tasks)
}

// AddTaskToContest godoc
//
//	@Tags			contests-management
//	@Summary		Add task to contest
//	@Destription	Add a task to a contest
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int							true	"Contest ID"
//	@Param			request	body		schemas.AddTaskToContest	true	"Add Task Request"
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		401		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/teacher/contests/{id}/tasks [post]
func (tr *teacherRouteImpl) AddTaskToContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)
	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
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

	var request schemas.AddTaskToContest
	err = httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			httputils.ReturnValidationError(w, valErrs)
			return
		}
		httputils.ReturnError(w, http.StatusBadRequest, "Could not validate request data.")
		return
	}

	err = tr.contestService.AddTaskToContest(tx, &currentUser, contestID, &request)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else {
			tr.logger.Errorw("Failed to add task to contest", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to add task to contest")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task added to contest successfully"))
}

// GetContestRegistrationRequests godoc
//
//	@Tags			contests-management
//	@Summary		Get registration requests
//	@Destription	Get all registration requests for a contest
//	@Produce		json
//	@Param			id	path		int	true	"Contest ID"
//	@Success		200	{object}	httputils.APIResponse[[]schemas.RegistrationRequest]
//	@Failure		400	{object}	httputils.APIError
//	@Failure		401	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Router			/teacher/contests/{id}/registration-requests [get]
func (tr *teacherRouteImpl) GetContestRegistrationRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
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
	statusQuery := r.URL.Query().Get("status")
	if statusQuery == "" {
		statusQuery = "pending"
	}
	status, ok := types.ParseRegistrationRequestStatus(statusQuery)
	if !ok {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid status value")
		return
	}

	requests, err := tr.contestService.GetRegistrationRequests(tx, currentUser, contestID, status)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			tr.logger.Errorw("Failed to get registration requests", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to get registration requests")
		return
	}

	if requests == nil {
		requests = []schemas.RegistrationRequest{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, requests)
}

// ApproveContestRegistrationRequest godoc
//
//	@Tags			contests-management
//	@Summary		Approve registration request
//	@Destription	Approve a registration request for a contest
//	@Produce		json
//	@Param			id		path		int	true	"Contest ID"
//	@Param			user_id	path		int	true	"User ID"
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		401		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/teacher/contests/{id}/registration-requests/{user_id}/approve [post]
func (tr *teacherRouteImpl) ApproveContestRegistrationRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)
	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
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

	err = tr.contestService.ApproveRegistrationRequest(tx, currentUser, contestID, userID)
	if err != nil {
		if !errors.Is(err, myerrors.ErrAlreadyParticipant) {
			db.Rollback()
		}
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotFound) || errors.Is(err, myerrors.ErrNoPendingRegistration) {
			status = http.StatusNotFound
		} else if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrAlreadyParticipant) {
			status = http.StatusBadRequest
		} else {
			tr.logger.Errorw("Failed to approve registration request", "error", err)
			httputils.ReturnError(w, status, "Failed to approve registration request")
			return
		}
		httputils.ReturnError(w, status, err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Registration request approved successfully"))
}

// RejectContestRegistrationRequest godoc
//
//	@Tags			contests-management
//	@Summary		Reject registration request
//	@Destription	Reject a registration request for a contest
//	@Produce		json
//	@Param			id		path		int	true	"Contest ID"
//	@Param			user_id	path		int	true	"User ID"
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		401		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/teacher/contests/{id}/registration-requests/{user_id}/reject [post]
func (tr *teacherRouteImpl) RejectContestRegistrationRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)
	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
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

	err = tr.contestService.RejectRegistrationRequest(tx, currentUser, contestID, userID)
	if err != nil {
		if !errors.Is(err, myerrors.ErrAlreadyParticipant) {
			db.Rollback()
		}
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotFound) || errors.Is(err, myerrors.ErrNoPendingRegistration) {
			status = http.StatusNotFound
		} else if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrAlreadyParticipant) {
			status = http.StatusBadRequest
		} else {
			tr.logger.Errorw("Failed to reject registration request", "error", err)
			httputils.ReturnError(w, status, "Failed to reject registration request")
			return
		}
		httputils.ReturnError(w, status, err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Registration request rejected successfully"))
}

// GetTasksManagement godoc
//
//	@Tags			tasks-management
//	@Summary		Get tasks treated by teacher
//	@Destription	Get all tasks created by the authenticated teacher
//	@Produce		json
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Success		200		{object}	httputils.APIResponse[[]schemas.Task]
//	@Failure		401		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/teacher/tasks [get]
func (tr *teacherRouteImpl) GetTasksManagement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	tasks, err := tr.taskService.GetAllCreated(tx, currentUser, queryParams)
	if err != nil {
		db.Rollback()
		tr.logger.Errorw("Failed to get tasks", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Task service temporarily unavailable")
		return
	}

	if tasks == nil {
		tasks = []schemas.Task{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, tasks)
}

// GetTaskManagement godoc
//
//	@Tags			tasks-management
//	@Summary		Get task details
//	@Destription	Get details of a specific task created by teacher
//	@Produce		json
//	@Param			id	path		int	true	"Task ID"
//	@Success		200	{object}	httputils.APIResponse[schemas.Task]
//	@Failure		400	{object}	httputils.APIError
//	@Failure		401	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Router			/teacher/tasks/{id} [get]
func (tr *teacherRouteImpl) GetTaskManagement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID cannot be empty")
		return
	}

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID: "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	task, err := tr.taskService.Get(tx, currentUser, taskID)
	if err != nil {
		db.Rollback()
		if errors.Is(err, myerrors.ErrTaskNotFound) {
			httputils.ReturnError(w, http.StatusNotFound, "Task not found")
			return
		}
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			httputils.ReturnError(w, http.StatusForbidden, "You are not authorized to view this task")
			return
		}
		tr.logger.Errorw("Failed to get task", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Task service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, task)
}

// CreateTask godoc
//
//	@Tags			tasks-management
//	@Summary		treate task
//	@Destription	Create a new task
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"Task file"
//	@Success		201		{object}	httputils.APIResponse[schemas.Task]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		401		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/teacher/tasks [post]
func (tr *teacherRouteImpl) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Task treation logic would be similar to UploadTask from tasks.go
	// For brevity, placeholder implementation
	httputils.ReturnError(w, http.StatusNotImplemented, "Task treation endpoint - implementation needed")
}

// EditTask godoc
//
//	@Tags			tasks-management
//	@Summary		Edit task
//	@Destription	Edit an existing task
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			id		path		int		true	"Task ID"
//	@Param			file	formData	file	false	"Task file"
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		401		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/teacher/tasks/{id} [patch]
func (tr *teacherRouteImpl) EditTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID cannot be empty")
		return
	}

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID: "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	// Task edit logic would be similar to EditTask from tasks.go
	// For brevity, showing authorization check pattern
	_, err = tr.taskService.Get(tx, currentUser, taskID)
	if err != nil {
		db.Rollback()
		if errors.Is(err, myerrors.ErrTaskNotFound) {
			httputils.ReturnError(w, http.StatusNotFound, "Task not found")
			return
		}
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			httputils.ReturnError(w, http.StatusForbidden, "You are not authorized to edit this task")
			return
		}
		tr.logger.Errorw("Failed to get task", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Task service temporarily unavailable")
		return
	}

	httputils.ReturnError(w, http.StatusNotImplemented, "Task edit endpoint - implementation needed")
}

// DeleteTask godoc
//
//	@Tags			tasks-management
//	@Summary		Delete task
//	@Destription	Delete a task
//	@Produce		json
//	@Param			id	path		int	true	"Task ID"
//	@Success		200	{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Failure		400	{object}	httputils.APIError
//	@Failure		401	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Router			/teacher/tasks/{id} [delete]
func (tr *teacherRouteImpl) DeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = tr.taskService.Delete(tx, currentUser, taskID)
	if err != nil {
		db.Rollback()
		switch {
		case errors.Is(err, myerrors.ErrNotAuthorized):
			httputils.ReturnError(w, http.StatusForbidden, "You are not authorized to delete this task.")
			return
		case errors.Is(err, myerrors.ErrTaskNotFound):
			httputils.ReturnError(w, http.StatusNotFound, "Task not found.")
			return
		default:
			tr.logger.Errorw("Failed to delete task", "error", err)
		}
		httputils.ReturnError(w, http.StatusInternalServerError, "Task service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task deleted successfully"))
}

// AssignTaskToUsers godoc
//
//	@Tags			tasks-management
//	@Summary		Assign task to users
//	@Destription	Assign a task to specific users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Task ID"
//	@Param			request	body		schemas.UserIDs		true	"User IDs"
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		401		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/teacher/tasks/{id}/assign/users [post]
func (tr *teacherRouteImpl) AssignTaskToUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	request := schemas.UsersRequest{}
	err = httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			httputils.ReturnValidationError(w, valErrs)
			return
		}
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = tr.taskService.AssignToUsers(tx, currentUser, taskID, request.UserIDs)
	if err != nil {
		db.Rollback()
		switch {
		case errors.Is(err, myerrors.ErrNotAuthorized):
			httputils.ReturnError(w, http.StatusForbidden, "You are not authorized to assign this task")
			return
		case errors.Is(err, myerrors.ErrUserNotFound):
			httputils.ReturnError(w, http.StatusBadRequest, "One or more users not found")
			return
		case errors.Is(err, myerrors.ErrTaskNotFound):
			httputils.ReturnError(w, http.StatusNotFound, "Task not found")
			return
		default:
			tr.logger.Errorw("Failed to assign task to users", "error", err)
		}
		httputils.ReturnError(w, http.StatusInternalServerError, "Task assignment service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task assigned successfully"))
}

// AssignTaskToGroups godoc
//
//	@Tags			tasks-management
//	@Summary		Assign task to groups
//	@Destription	Assign a task to specific groups
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Task ID"
//	@Param			request	body		schemas.GroupsRequest true	"Group IDs"
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		401		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/teacher/tasks/{id}/assign/groups [post]
func (tr *teacherRouteImpl) AssignTaskToGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	request := schemas.GroupsRequest{}
	err = httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			httputils.ReturnValidationError(w, valErrs)
			return
		}
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = tr.taskService.AssignToGroups(tx, currentUser, taskID, request.GroupIDs)
	if err != nil {
		db.Rollback()
		switch {
		case errors.Is(err, myerrors.ErrNotAuthorized):
			httputils.ReturnError(w, http.StatusForbidden, "You are not authorized to assign this task")
			return
		case errors.Is(err, myerrors.ErrGroupNotFound):
			httputils.ReturnError(w, http.StatusBadRequest, "One or more groups not found")
			return
		case errors.Is(err, myerrors.ErrTaskNotFound):
			httputils.ReturnError(w, http.StatusNotFound, "Task not found")
			return
		default:
			tr.logger.Errorw("Failed to assign task to groups", "error", err)
		}
		httputils.ReturnError(w, http.StatusInternalServerError, "Task assignment service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task assigned successfully"))
}

// UnAssignTaskFromUsers godoc
//
//	@Tags			tasks-management
//	@Summary		Unassign task from users
//	@Destription	Unassign a task from specific users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Task ID"
//	@Param			request	body		schemas.UserIDs		true	"User IDs"
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		401		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/teacher/tasks/{id}/unassign/users [post]
func (tr *teacherRouteImpl) UnAssignTaskFromUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	request := schemas.UsersRequest{}
	err = httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			httputils.ReturnValidationError(w, valErrs)
			return
		}
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = tr.taskService.UnassignFromUsers(tx, currentUser, taskID, request.UserIDs)
	if err != nil {
		db.Rollback()
		switch {
		case errors.Is(err, myerrors.ErrNotAuthorized):
			httputils.ReturnError(w, http.StatusForbidden, "You are not authorized to unassign this task")
			return
		case errors.Is(err, myerrors.ErrTaskNotFound):
			httputils.ReturnError(w, http.StatusNotFound, "Task not found")
			return
		default:
			tr.logger.Errorw("Failed to unassign task from users", "error", err)
		}
		httputils.ReturnError(w, http.StatusInternalServerError, "Task unassignment service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task unassigned successfully"))
}

// UnAssignTaskFromGroups godoc
//
//	@Tags			tasks-management
//	@Summary		Unassign task from groups
//	@Destription	Unassign a task from specific groups
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Task ID"
//	@Param			request	body		schemas.GroupsRequest	true	"Group IDs"
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		401		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/teacher/tasks/{id}/unassign/groups [post]
func (tr *teacherRouteImpl) UnAssignTaskFromGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	request := schemas.GroupsRequest{}
	err = httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			httputils.ReturnValidationError(w, valErrs)
			return
		}
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = tr.taskService.UnassignFromGroups(tx, currentUser, taskID, request.GroupIDs)
	if err != nil {
		db.Rollback()
		switch {
		case errors.Is(err, myerrors.ErrNotAuthorized):
			httputils.ReturnError(w, http.StatusForbidden, "You are not authorized to unassign this task")
			return
		case errors.Is(err, myerrors.ErrTaskNotFound):
			httputils.ReturnError(w, http.StatusNotFound, "Task not found")
			return
		default:
			tr.logger.Errorw("Failed to unassign task from groups", "error", err)
		}
		httputils.ReturnError(w, http.StatusInternalServerError, "Task unassignment service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task unassigned successfully"))
}

// GetTaskLimits godoc
//
//	@Tags			tasks-management
//	@Summary		Get task limits
//	@Destription	Get resource limits for a task
//	@Produce		json
//	@Param			id	path		int	true	"Task ID"
//	@Success		200	{object}	httputils.APIResponse[[]schemas.TestCase]
//	@Failure		400	{object}	httputils.APIError
//	@Failure		401	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Router			/teacher/tasks/{id}/limits [get]
func (tr *teacherRouteImpl) GetTaskLimits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	limits, err := tr.taskService.GetLimits(tx, currentUser, taskID)
	if err != nil {
		switch {
		case errors.Is(err, myerrors.ErrNotFound):
			httputils.ReturnError(w, http.StatusNotFound, "Task not found.")
		default:
			tr.logger.Errorw("Failed to get task limits", "error", err)
			httputils.ReturnError(w, http.StatusInternalServerError, "Task service temporarily unavailable")
		}
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, limits)
}

// PutTaskLimits godoc
//
//	@Tags			tasks-management
//	@Summary		Set task limits
//	@Destription	Set resource limits for a task
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int									true	"Task ID"
//	@Param			limits	body		schemas.PutTestCaseLimitsRequest	true	"Task limits"
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		401		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/teacher/tasks/{id}/limits [put]
func (tr *teacherRouteImpl) PutTaskLimits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskIDStr := httputils.GetPathValue(r, "id")
	if taskIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID is required.")
		return
	}
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID.")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	request := schemas.PutTestCaseLimitsRequest{}
	err = httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			httputils.ReturnValidationError(w, valErrs)
			return
		}
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	err = tr.taskService.PutLimits(tx, currentUser, taskID, request)
	if err != nil {
		db.Rollback()
		switch {
		case errors.Is(err, myerrors.ErrNotFound):
			httputils.ReturnError(w, http.StatusNotFound, "Task not found.")
		case errors.Is(err, myerrors.ErrNotAuthorized):
			httputils.ReturnError(w, http.StatusForbidden, "You are not authorized to update limits for this task.")
		default:
			tr.logger.Errorw("Failed to put task limits", "error", err)
			httputils.ReturnError(w, http.StatusInternalServerError, "Task service temporarily unavailable")
		}
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task limits updated successfully"))
}

// GetGroupsManagement godoc
//
//	@Tags			groups-management
//	@Summary		Get groups managed by teacher
//	@Destription	Get all groups created or managed by the authenticated teacher
//	@Produce		json
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Success		200		{object}	httputils.APIResponse[[]schemas.Group]
//	@Failure		401		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/teacher/groups [get]
func (tr *teacherRouteImpl) GetGroupsManagement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	groups, err := tr.groupService.GetAll(tx, currentUser, queryParams)
	if err != nil {
		db.Rollback()
		tr.logger.Errorw("Failed to get groups", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Group service temporarily unavailable")
		return
	}

	if groups == nil {
		groups = []schemas.Group{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, groups)
}

// GetGroupManagement godoc
//
//	@Tags			groups-management
//	@Summary		Get group details
//	@Destription	Get details of a specific group managed by teacher
//	@Produce		json
//	@Param			id	path		int	true	"Group ID"
//	@Success		200	{object}	httputils.APIResponse[schemas.Group]
//	@Failure		400	{object}	httputils.APIError
//	@Failure		401	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Router			/teacher/groups/{id} [get]
func (tr *teacherRouteImpl) GetGroupManagement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	groupIDStr := httputils.GetPathValue(r, "id")
	if groupIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Group ID cannot be empty")
		return
	}

	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group ID: "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	group, err := tr.groupService.Get(tx, currentUser, groupID)
	if err != nil {
		db.Rollback()
		if errors.Is(err, myerrors.ErrGroupNotFound) {
			httputils.ReturnError(w, http.StatusNotFound, "Group not found")
			return
		}
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			httputils.ReturnError(w, http.StatusForbidden, "You are not authorized to view this group")
			return
		}
		tr.logger.Errorw("Failed to get group", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Group service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, group)
}

// CreateGroup godoc
//
//	@Tags			groups-management
//	@Summary		treate group
//	@Destription	Create a new group
//	@Accept			json
//	@Produce		json
//	@Param			request	body		schemas.CreateGroup	true	"Group Create Request"
//	@Success		201		{object}	httputils.APIResponse[schemas.Group]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		401		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/teacher/groups [post]
func (tr *teacherRouteImpl) CreateGroup(w http.ResponseWriter, r *http.Request) {
	httputils.ReturnError(w, http.StatusNotImplemented, "treate group endpoint - implementation needed")
}

// EditGroup godoc
//
//	@Tags			groups-management
//	@Summary		Edit group
//	@Destription	Edit an existing group
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Group ID"
//	@Param			request	body		schemas.EditGroup	true	"Group Edit Request"
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		401		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/teacher/groups/{id} [patch]
func (tr *teacherRouteImpl) EditGroup(w http.ResponseWriter, r *http.Request) {
	httputils.ReturnError(w, http.StatusNotImplemented, "Edit group endpoint - implementation needed")
}

// AddUsersToGroup godoc
//
//	@Tags			groups-management
//	@Summary		Add users to group
//	@Destription	Add users to a group
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Group ID"
//	@Param			request	body		schemas.UserIDs		true	"User IDs"
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		401		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/teacher/groups/{id}/users [post]
func (tr *teacherRouteImpl) AddUsersToGroup(w http.ResponseWriter, r *http.Request) {
	httputils.ReturnError(w, http.StatusNotImplemented, "Add users to group endpoint - implementation needed")
}

// DeleteUsersFromGroup godoc
//
//	@Tags			groups-management
//	@Summary		Delete users from group
//	@Destription	Delete users from a group
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Group ID"
//	@Param			request	body		schemas.UserIDs		true	"User IDs"
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		401		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/teacher/groups/{id}/users [delete]
func (tr *teacherRouteImpl) DeleteUsersFromGroup(w http.ResponseWriter, r *http.Request) {
	httputils.ReturnError(w, http.StatusNotImplemented, "Delete users from group endpoint - implementation needed")
}

// GetGroupUsers godoc
//
//	@Tags			groups-management
//	@Summary		Get group users
//	@Destription	Get all users in a group
//	@Produce		json
//	@Param			id	path		int	true	"Group ID"
//	@Success		200	{object}	httputils.APIResponse[[]schemas.User]
//	@Failure		400	{object}	httputils.APIError
//	@Failure		401	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Router			/teacher/groups/{id}/users [get]
func (tr *teacherRouteImpl) GetGroupUsers(w http.ResponseWriter, r *http.Request) {
	httputils.ReturnError(w, http.StatusNotImplemented, "Get group users endpoint - implementation needed")
}

// GetGroupTasks godoc
//
//	@Tags			groups-management
//	@Summary		Get group tasks
//	@Destription	Get all tasks assigned to a group
//	@Produce		json
//	@Param			id	path		int	true	"Group ID"
//	@Success		200	{object}	httputils.APIResponse[[]schemas.Task]
//	@Failure		400	{object}	httputils.APIError
//	@Failure		401	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Router			/teacher/groups/{id}/tasks [get]
func (tr *teacherRouteImpl) GetGroupTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	groupIDStr := httputils.GetPathValue(r, "id")

	if groupIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group id")
		return
	}

	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group id")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	paginationParams := httputils.ExtractPaginationParams(queryParams)

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	tasks, err := tr.taskService.GetAllForGroup(tx, currentUser, groupID, paginationParams)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			tr.logger.Errorw("Failed to get tasks for group", "error", err)
		}
		httputils.ReturnError(w, status, "Task service temporarily unavailable")
		return
	}

	if tasks == nil {
		tasks = []schemas.Task{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, tasks)
}

// GetAssignableTasksToContest godoc
//
//	@Tags			contests-management
//	@Summary		Get available tasks for a contest
//	@Destription	Get all tasks that are NOT yet assigned to the specified contest (admin/teacher only)
//
//	@Produce		json
//	@Param			id	path		int	true	"Contest ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[[]schemas.Task]
//	@Router			/teacher/contests/{id}/tasks/assignable-tasks [get]
func (tr *teacherRouteImpl) GetAssignableTasksToContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		tr.logger.Errorw("Failed to begin database transaction", "error", err)
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

	tasks, err := tr.contestService.GetAssignableTasks(tx, currentUser, contestID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			tr.logger.Errorw("Failed to get available tasks for contest", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to get available tasks for contest")
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, tasks)
}

// NewTeacherRoute treates a new TeacherRoute.
func NewTeacherRoute(
	contestService service.ContestService,
	taskService service.TaskService,
	groupService service.GroupService,
	submissionService service.SubmissionService,
) TeacherRoute {
	return &teacherRouteImpl{
		contestService:    contestService,
		taskService:       taskService,
		groupService:      groupService,
		submissionService: submissionService,
		logger:            utils.NewNamedLogger("teacher-routes"),
	}
}

// RegisterTeacherRoutes registers all teacher routes.
func RegisterTeacherRoutes(mux *mux.Router, route TeacherRoute) {
	// Contest management routes
	mux.HandleFunc("/contests", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetManageableContests(w, r)
		case http.MethodPost:
			route.CreateContest(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

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

	mux.HandleFunc("/contests/{id}/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetContestTasks(w, r)
		case http.MethodPost:
			route.AddTaskToContest(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	mux.HandleFunc("/contests/{id}/tasks/assignable-tasks", route.GetAssignableTasksToContest)

	mux.HandleFunc("/contests/{id}/registration-requests", route.GetContestRegistrationRequests)
	mux.HandleFunc("/contests/{id}/registration-requests/{user_id}/approve", route.ApproveContestRegistrationRequest)
	mux.HandleFunc("/contests/{id}/registration-requests/{user_id}/reject", route.RejectContestRegistrationRequest)

	mux.HandleFunc("/contests/{id}/submissions", route.GetContestSubmissions)

	// Task management routes
	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetTasksManagement(w, r)
		case http.MethodPost:
			route.CreateTask(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetTaskManagement(w, r)
		case http.MethodPatch:
			route.EditTask(w, r)
		case http.MethodDelete:
			route.DeleteTask(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/tasks/{id}/assign/users", route.AssignTaskToUsers)
	mux.HandleFunc("/tasks/{id}/assign/groups", route.AssignTaskToGroups)
	mux.HandleFunc("/tasks/{id}/unassign/users", route.UnAssignTaskFromUsers)
	mux.HandleFunc("/tasks/{id}/unassign/groups", route.UnAssignTaskFromGroups)

	mux.HandleFunc("/tasks/{id}/limits", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetTaskLimits(w, r)
		case http.MethodPut:
			route.PutTaskLimits(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// Group management routes
	mux.HandleFunc("/groups", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetGroupsManagement(w, r)
		case http.MethodPost:
			route.CreateGroup(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/groups/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetGroupManagement(w, r)
		case http.MethodPatch:
			route.EditGroup(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/groups/{id}/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetGroupUsers(w, r)
		case http.MethodPost:
			route.AddUsersToGroup(w, r)
		case http.MethodDelete:
			route.DeleteUsersFromGroup(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/groups/{id}/tasks", route.GetGroupTasks)
}
