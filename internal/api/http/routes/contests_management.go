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

type ContestsManagementRoute interface {
	CreateContest(w http.ResponseWriter, r *http.Request)
	EditContest(w http.ResponseWriter, r *http.Request)
	DeleteContest(w http.ResponseWriter, r *http.Request)
	GetContestTasks(w http.ResponseWriter, r *http.Request)
	GetAssignableTasks(w http.ResponseWriter, r *http.Request)
	AddTaskToContest(w http.ResponseWriter, r *http.Request)
	GetRegistrationRequests(w http.ResponseWriter, r *http.Request)
	ApproveRegistrationRequest(w http.ResponseWriter, r *http.Request)
	RejectRegistrationRequest(w http.ResponseWriter, r *http.Request)
	GetContestSubmissions(w http.ResponseWriter, r *http.Request)
	GetCreatedContests(w http.ResponseWriter, r *http.Request)
	GetManageableContests(w http.ResponseWriter, r *http.Request)
	GetContestTaskStats(w http.ResponseWriter, r *http.Request)
	GetContestTaskUserStats(w http.ResponseWriter, r *http.Request)
	GetContestTaskUserSubmissions(w http.ResponseWriter, r *http.Request)
	GetContestUserStats(w http.ResponseWriter, r *http.Request)
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
		cr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	contestID, err := cr.contestService.Create(tx, currentUser, &request)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else {
			cr.logger.Errorw("Failed to create contest", "error", err)
		}
		httputils.ReturnError(w, status, "Contest creation failed")
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

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	resp, err := cr.contestService.Edit(tx, currentUser, contestID, &request)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else {
			cr.logger.Errorw("Failed to edit contest", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to edit contest")
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

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = cr.contestService.Delete(tx, currentUser, contestID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else {
			cr.logger.Errorw("Failed to delete contest", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to delete contest")
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

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	tasks, err := cr.contestService.GetAssignableTasks(tx, currentUser, contestID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else {
			cr.logger.Errorw("Failed to get available tasks for contest", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to get available tasks for contest")
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

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)
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

	err = cr.contestService.AddTaskToContest(tx, &currentUser, contestID, &request)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else {
			cr.logger.Errorw("Failed to add task to contest", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to add task to contest")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Task added to contest successfully"))
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

	requests, err := cr.contestService.GetRegistrationRequests(tx, currentUser, contestID, status)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else {
			cr.logger.Errorw("Failed to get registration requests", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to get registration requests")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, requests)
}

func (cr *contestsManagementRouteImpl) ApproveRegistrationRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)
	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		cr.logger.Errorw("Failed to begin database transaction", "error", err)
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

	err = cr.contestService.ApproveRegistrationRequest(tx, currentUser, contestID, userID)
	if err != nil {
		if !errors.Is(err, myerrors.ErrAlreadyParticipant) {
			db.Rollback()
		}
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotFound) || errors.Is(err, myerrors.ErrNoPendingRegistration) {
			status = http.StatusNotFound
		} else if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrAlreadyParticipant) {
			status = http.StatusBadRequest
		} else {
			cr.logger.Errorw("Failed to approve registration request", "error", err)
			httputils.ReturnError(w, status, "Failed to approve registration request")
			return
		}
		httputils.ReturnError(w, status, err.Error())
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
	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)
	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		cr.logger.Errorw("Failed to begin database transaction", "error", err)
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

	err = cr.contestService.RejectRegistrationRequest(tx, currentUser, contestID, userID)
	if err != nil {
		if !errors.Is(err, myerrors.ErrAlreadyParticipant) {
			db.Rollback()
		}
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotFound) || errors.Is(err, myerrors.ErrNoPendingRegistration) {
			status = http.StatusNotFound
		} else if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrAlreadyParticipant) {
			status = http.StatusBadRequest
		} else {
			cr.logger.Errorw("Failed to reject registration request", "error", err)
			httputils.ReturnError(w, status, "Failed to reject registration request")
			return
		}
		httputils.ReturnError(w, status, err.Error())
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

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	tasks, err := cr.contestService.GetTasksForContest(tx, currentUser, contestID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else {
			cr.logger.Errorw("Failed to get tasks for contest", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to get tasks for contest")
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

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	// Only teachers and admins can view all contest submissions
	if err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin, types.UserRoleTeacher}); err != nil {
		httputils.ReturnError(w, http.StatusForbidden, "Only teachers and admins can view contest submissions")
		return
	}

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	paginationParams := httputils.ExtractPaginationParams(queryParams)

	response, err := cr.submissionService.GetAllForContest(tx, contestID, currentUser, paginationParams)
	if err != nil {
		db.Rollback()
		switch {
		case errors.Is(err, myerrors.ErrNotFound):
			httputils.ReturnError(w, http.StatusNotFound, "Contest not found")
		case errors.Is(err, myerrors.ErrPermissionDenied):
			httputils.ReturnError(w, http.StatusForbidden, "Permission denied. You are not the creator of this contest.")
		default:
			cr.logger.Errorw("Failed to get contest submissions", "error", err)
			httputils.ReturnError(w, http.StatusInternalServerError, "Failed to get contest submissions")
		}
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
//	@Success	200		{object}	httputils.APIResponse[schemas.PaginatedResult[[]schemas.CreatedContest]]
//	@Router		/contests-management/contests/created [get]
func (cr *contestsManagementRouteImpl) GetCreatedContests(w http.ResponseWriter, r *http.Request) {
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
	paginationParams := httputils.ExtractPaginationParams(queryParams)

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)
	response, err := cr.contestService.GetContestsCreatedByUser(tx, currentUser.ID, paginationParams)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else {
			cr.logger.Errorw("Failed to get created contests", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to get created contests")
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

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		cr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)
	stats, err := cr.submissionService.GetTaskStatsForContest(tx, currentUser, contestID)
	if err != nil {
		db.Rollback()
		if errors.Is(err, myerrors.ErrPermissionDenied) {
			httputils.ReturnError(w, http.StatusForbidden, "Permission denied")
			return
		}
		cr.logger.Errorw("Failed to get task stats", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to get task statistics")
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

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		cr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)
	stats, err := cr.submissionService.GetUserStatsForContestTask(tx, currentUser, contestID, taskID)
	if err != nil {
		db.Rollback()
		if errors.Is(err, myerrors.ErrPermissionDenied) {
			httputils.ReturnError(w, http.StatusForbidden, "Permission denied")
			return
		}
		cr.logger.Errorw("Failed to get user stats for task", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to get user statistics")
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

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		cr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)
	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	paginationParams := httputils.ExtractPaginationParams(queryParams)

	// Use the existing GetAll with filters for contest, task, and user
	submissions, err := cr.submissionService.GetAll(tx, currentUser, &userID, &taskID, &contestID, paginationParams)
	if err != nil {
		db.Rollback()
		if errors.Is(err, myerrors.ErrPermissionDenied) {
			httputils.ReturnError(w, http.StatusForbidden, "Permission denied")
			return
		}
		cr.logger.Errorw("Failed to get submissions", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to get submissions")
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

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		cr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)
	stats, err := cr.submissionService.GetUserStatsForContest(tx, currentUser, contestID, userID)
	if err != nil {
		db.Rollback()
		if errors.Is(err, myerrors.ErrPermissionDenied) {
			httputils.ReturnError(w, http.StatusForbidden, "Permission denied")
			return
		}
		cr.logger.Errorw("Failed to get user stats", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to get user statistics")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, stats)
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
//	@Router			/contests-management/contests [get]
func (cr *contestsManagementRouteImpl) GetManageableContests(w http.ResponseWriter, r *http.Request) {
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
	paginationParams := httputils.ExtractPaginationParams(queryParams)

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)
	response, err := cr.contestService.GetManagedContests(tx, currentUser.ID, paginationParams)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else {
			cr.logger.Errorw("Failed to get manageable contests", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to get manageable contests")
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
			route.GetManageableContests(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/contests/created", route.GetCreatedContests)

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
