package routes

import (
	"errors"
	"log"
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

type ContestRoute interface {
	CreateContest(w http.ResponseWriter, r *http.Request)
	GetContest(w http.ResponseWriter, r *http.Request)
	GetOngoingContests(w http.ResponseWriter, r *http.Request)
	GetPastContests(w http.ResponseWriter, r *http.Request)
	GetUpcomingContests(w http.ResponseWriter, r *http.Request)
	EditContest(w http.ResponseWriter, r *http.Request)
	DeleteContest(w http.ResponseWriter, r *http.Request)
	RegisterForContest(w http.ResponseWriter, r *http.Request)
	GetContestTasks(w http.ResponseWriter, r *http.Request)
	GetTaskProgressForContest(w http.ResponseWriter, r *http.Request)
	GetAvailableTasksForContest(w http.ResponseWriter, r *http.Request)
	AddTaskToContest(w http.ResponseWriter, r *http.Request)
	GetRegistrationRequests(w http.ResponseWriter, r *http.Request)
	ApproveRegistrationRequest(w http.ResponseWriter, r *http.Request)
	RejectRegistrationRequest(w http.ResponseWriter, r *http.Request)
	GetContestSubmissions(w http.ResponseWriter, r *http.Request)
}

type ContestRouteImpl struct {
	contestService    service.ContestService
	submissionService service.SubmissionService
	logger            *zap.SugaredLogger
}

// CreateContest godoc
//
//	@Tags			contest
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
//	@Router			/contests/ [post]
func (cr *ContestRouteImpl) CreateContest(w http.ResponseWriter, r *http.Request) {
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
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			cr.logger.Errorw("Failed to create contest", "error", err)
		}
		httputils.ReturnError(w, status, "Contest creation failed")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewIDResponse(contestID))
}

// GetContest godoc
//
//	@Tags			contest
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

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	contest, err := cr.contestService.Get(tx, currentUser, contestID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			cr.logger.Errorw("Failed to get contest", "error", err)
		}
		httputils.ReturnError(w, status, "Contest retrieval failed")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, contest)
}

// GetOngoingContests godoc
//
//	@Tags			contest
//	@Summary		Get ongoing contests
//	@Description	Get contests that are currently running with pagination
//	@Produce		json
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[[]schemas.AvailableContest]
//	@Router			/contests/ongoing [get]
func (cr *ContestRouteImpl) GetOngoingContests(w http.ResponseWriter, r *http.Request) {
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
	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	contests, err := cr.contestService.GetOngoingContests(tx, currentUser, queryParams)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			cr.logger.Errorw("Failed to list ongoing contests", "error", err)
		}
		httputils.ReturnError(w, status, "Contest listing failed")
		return
	}

	if contests == nil {
		contests = []schemas.AvailableContest{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, contests)
}

// GetPastContests godoc
//
//	@Tags			contest
//	@Summary		Get past contests
//	@Description	Get contests that have ended with pagination
//	@Produce		json
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[[]schemas.AvailableContest]
//	@Router			/contests/past [get]
func (cr *ContestRouteImpl) GetPastContests(w http.ResponseWriter, r *http.Request) {
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
	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	contests, err := cr.contestService.GetPastContests(tx, currentUser, queryParams)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			cr.logger.Errorw("Failed to list past contests", "error", err)
		}
		httputils.ReturnError(w, status, "Contest listing failed")
		return
	}

	if contests == nil {
		contests = []schemas.AvailableContest{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, contests)
}

// GetUpcomingContests godoc
//
//	@Tags			contest
//	@Summary		Get upcoming contests
//	@Description	Get contests that haven't started yet with pagination
//	@Produce		json
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[[]schemas.AvailableContest]
//	@Router			/contests/upcoming [get]
func (cr *ContestRouteImpl) GetUpcomingContests(w http.ResponseWriter, r *http.Request) {
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
	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	contests, err := cr.contestService.GetUpcomingContests(tx, currentUser, queryParams)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			cr.logger.Errorw("Failed to list upcoming contests", "error", err)
		}
		httputils.ReturnError(w, status, "Contest listing failed")
		return
	}

	if contests == nil {
		contests = []schemas.AvailableContest{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, contests)
}

// EditContest godoc
//
//	@Tags			contest
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
//	@Success		200		{object}	httputils.APIResponse[schemas.Contest]
//	@Router			/contests/{id} [put]
func (cr *ContestRouteImpl) EditContest(w http.ResponseWriter, r *http.Request) {
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
		if errors.Is(err, myerrors.ErrNotAuthorized) {
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
//	@Tags			contest
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
//	@Router			/contests/{id} [delete]
func (cr *ContestRouteImpl) DeleteContest(w http.ResponseWriter, r *http.Request) {
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
		if errors.Is(err, myerrors.ErrNotAuthorized) {
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

// RegisterForContest godoc
//
//	@Tags			contest
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

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = cr.contestService.RegisterForContest(tx, currentUser, contestID)
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
			cr.logger.Errorw("Failed to register for contest", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to register for contest")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Registration request submitted successfully"))
}

// GetContestTasks godoc
//
//	@Tags			contest
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
//	@Success		200	{object}	httputils.APIResponse[[]schemas.Task]
//	@Router			/contests/{id}/tasks [get]
func (cr *ContestRouteImpl) GetContestTasks(w http.ResponseWriter, r *http.Request) {
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
		} else if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			cr.logger.Errorw("Failed to get tasks for contest", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to get tasks for contest")
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, tasks)
}

// GetTaskProgressForContest godoc
//
//	@Tags			contest
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

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)
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
		} else if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			cr.logger.Errorw("Failed to get task progress for contest", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to get task progress for contest")
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, tasks)
}

// GetAvailableTasksForContest godoc
//
//	@Tags			contest
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
//	@Router			/contests/{id}/available-tasks [get]
func (cr *ContestRouteImpl) GetAvailableTasksForContest(w http.ResponseWriter, r *http.Request) {
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

	tasks, err := cr.contestService.GetAvailableTasksForContest(tx, currentUser, contestID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, myerrors.ErrNotAuthorized) {
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
//	@Tags			contest
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
//	@Router			/contests/{id}/tasks [post]
func (cr *ContestRouteImpl) AddTaskToContest(w http.ResponseWriter, r *http.Request) {
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
		if errors.Is(err, myerrors.ErrNotAuthorized) {
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
//	@Tags			contest
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
//	@Router			/contests/{id}/registration-requests [get]
func (cr *ContestRouteImpl) GetRegistrationRequests(w http.ResponseWriter, r *http.Request) {
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
		} else if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			cr.logger.Errorw("Failed to get registration requests", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to get registration requests")
		return
	}

	if requests == nil {
		requests = []schemas.RegistrationRequest{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, requests)
}

func (cr *ContestRouteImpl) ApproveRegistrationRequest(w http.ResponseWriter, r *http.Request) {
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
		} else if errors.Is(err, myerrors.ErrNotAuthorized) {
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
// @Tags			contest
//
//	@Summary		Reject a registration request
//	@Description	Reject a pending registration request for a contest (only accessible by contest creator or admin)
//
// @Produce		json
// @Param			id		path		int		true	"Contest ID"
// @Param			user_id	path		int		true	"User ID"
// @Failure		400		{object}	httputils.APIError
// @Failure		403		{object}	httputils.APIError
// @Failure		404		{object}	httputils.APIError
// @Failure		405		{object}	httputils.APIError
// @Failure		500		{object}	httputils.APIError
// @Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
func (cr *ContestRouteImpl) RejectRegistrationRequest(w http.ResponseWriter, r *http.Request) {
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
		} else if errors.Is(err, myerrors.ErrNotAuthorized) {
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

func RegisterContestRoutes(mux *mux.Router, contestRoute ContestRoute) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			contestRoute.CreateContest(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/ongoing", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			contestRoute.GetOngoingContests(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/past", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			contestRoute.GetPastContests(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/upcoming", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			contestRoute.GetUpcomingContests(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			contestRoute.GetContest(w, r)
		case http.MethodPut:
			contestRoute.EditContest(w, r)
		case http.MethodDelete:
			contestRoute.DeleteContest(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/{id}/tasks/user-statistics", contestRoute.GetTaskProgressForContest)
	mux.HandleFunc("/{id}/available-tasks", contestRoute.GetAvailableTasksForContest)
	mux.HandleFunc("/{id}/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			contestRoute.GetContestTasks(w, r)
		case http.MethodPost:
			contestRoute.AddTaskToContest(w, r)
		}
	})

	mux.HandleFunc("/{id}/register", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			contestRoute.RegisterForContest(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/{id}/registration-requests", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			contestRoute.GetRegistrationRequests(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/{id}/registration-requests/{user_id}/approve", contestRoute.ApproveRegistrationRequest)
	mux.HandleFunc("/{id}/registration-requests/{user_id}/reject", contestRoute.RejectRegistrationRequest)

	mux.HandleFunc("/{id}/submissions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			contestRoute.GetContestSubmissions(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
}

// GetContestSubmissions godoc
//
//	@Tags			contest
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
//	@Success		200		{object}	httputils.APIResponse[[]schemas.Submission]
//	@Router			/contests/{id}/submissions [get]
func (cr *ContestRouteImpl) GetContestSubmissions(w http.ResponseWriter, r *http.Request) {
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

	submissions, err := cr.submissionService.GetAllForContest(tx, contestID, currentUser, queryParams)
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

	httputils.ReturnSuccess(w, http.StatusOK, submissions)
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
