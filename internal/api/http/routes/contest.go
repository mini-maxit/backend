package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/service"
)

type ContestRoute interface {
	CreateContest(w http.ResponseWriter, r *http.Request)
	GetContest(w http.ResponseWriter, r *http.Request)
	GetAllContests(w http.ResponseWriter, r *http.Request)
	EditContest(w http.ResponseWriter, r *http.Request)
	DeleteContest(w http.ResponseWriter, r *http.Request)
	AssignTasksToContest(w http.ResponseWriter, r *http.Request)
	UnAssignTasksFromContest(w http.ResponseWriter, r *http.Request)
}

type ContestRouteImpl struct {
	contestService service.ContestService
}

type tasksRequest struct {
	TaskIds []int64 `json:"taskIds"`
}

// CreateContest godoc
//
//	@Tags			contest
//	@Summary		Create a new contest
//	@Description	Creates a new contest
//	@Accept			json
//	@Produce		json
//	@Param			request	body		schemas.CreateContest	true	"Contest details"
//	@Param			Session	header		string					true	"Session Token"
//	@Success		200		{object}	httputils.ApiResponse[schemas.ContestCreateResponse]
//	@Failure		400		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Router			/contest/ [post]
func (cr *ContestRouteImpl) CreateContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	var request schemas.CreateContest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	contestId, err := cr.contestService.Create(tx, currentUser, &request)
	if err != nil {
		if err == errors.ErrContestExists {
			httputils.ReturnError(w, http.StatusConflict, err.Error())
			return
		}
		if err == errors.ErrInvalidTimeRange {
			httputils.ReturnError(w, http.StatusBadRequest, err.Error())
			return
		}
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to create contest. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, schemas.ContestCreateResponse{Id: contestId})
}

// GetContest godoc
//
//	@Tags			contest
//	@Summary		Get a contest by ID
//	@Description	Returns a contest with all details including tasks
//	@Produce		json
//	@Param			id		path		int		true	"Contest ID"
//	@Param			Session	header		string	true	"Session Token"
//	@Success		200		{object}	httputils.ApiResponse[schemas.ContestDetailed]
//	@Failure		400		{object}	httputils.ApiError
//	@Failure		404		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Router			/contest/{id} [get]
func (cr *ContestRouteImpl) GetContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	contestIdStr := r.PathValue("id")
	if contestIdStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Contest ID is required.")
		return
	}

	contestId, err := strconv.ParseInt(contestIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid contest ID.")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	contest, err := cr.contestService.GetContest(tx, currentUser, contestId)
	if err != nil {
		if err == errors.ErrContestNotFound {
			httputils.ReturnError(w, http.StatusNotFound, err.Error())
			return
		}
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to get contest. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, contest)
}

// GetAllContests godoc
//
//	@Tags			contest
//	@Summary		Get all contests
//	@Description	Returns all contests
//	@Produce		json
//	@Param			limit	query		int		false	"Limit the number of returned contests"
//	@Param			offset	query		int		false	"Offset the returned contests"
//	@Param			sort	query		string	false	"Sort order"
//	@Param			Session	header		string	true	"Session Token"
//	@Success		200		{object}	httputils.ApiResponse[[]schemas.Contest]
//	@Failure		500		{object}	httputils.ApiError
//	@Router			/contest/ [get]
func (cr *ContestRouteImpl) GetAllContests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)
	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]interface{})

	contests, err := cr.contestService.GetAll(tx, currentUser, queryParams)
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to get contests. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, contests)
}

// EditContest godoc
//
//	@Tags			contest
//	@Summary		Update a contest
//	@Description	Updates a contest by ID
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"Contest ID"
//	@Param			request	body		schemas.EditContest		true	"Contest update details"
//	@Param			Session	header		string					true	"Session Token"
//	@Success		200		{object}	httputils.ApiResponse[string]
//	@Failure		400		{object}	httputils.ApiError
//	@Failure		404		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Router			/contest/{id} [patch]
func (cr *ContestRouteImpl) EditContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	contestIdStr := r.PathValue("id")
	if contestIdStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Contest ID is required.")
		return
	}

	contestId, err := strconv.ParseInt(contestIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid contest ID.")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	var request schemas.EditContest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	err = cr.contestService.EditContest(tx, currentUser, contestId, &request)
	if err != nil {
		if err == errors.ErrContestNotFound {
			httputils.ReturnError(w, http.StatusNotFound, err.Error())
			return
		}
		if err == errors.ErrNotAuthorized {
			httputils.ReturnError(w, http.StatusForbidden, err.Error())
			return
		}
		if err == errors.ErrContestExists {
			httputils.ReturnError(w, http.StatusConflict, err.Error())
			return
		}
		if err == errors.ErrInvalidTimeRange {
			httputils.ReturnError(w, http.StatusBadRequest, err.Error())
			return
		}
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to update contest. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, "Contest updated successfully")
}

// DeleteContest godoc
//
//	@Tags			contest
//	@Summary		Delete a contest
//	@Description	Deletes a contest by ID
//	@Produce		json
//	@Param			id		path		int		true	"Contest ID"
//	@Param			Session	header		string	true	"Session Token"
//	@Success		200		{object}	httputils.ApiResponse[string]
//	@Failure		400		{object}	httputils.ApiError
//	@Failure		404		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Router			/contest/{id} [delete]
func (cr *ContestRouteImpl) DeleteContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	contestIdStr := r.PathValue("id")
	if contestIdStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Contest ID is required.")
		return
	}

	contestId, err := strconv.ParseInt(contestIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid contest ID.")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = cr.contestService.DeleteContest(tx, currentUser, contestId)
	if err != nil {
		if err == errors.ErrContestNotFound {
			httputils.ReturnError(w, http.StatusNotFound, err.Error())
			return
		}
		if err == errors.ErrNotAuthorized {
			httputils.ReturnError(w, http.StatusForbidden, err.Error())
			return
		}
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to delete contest. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, "Contest deleted successfully")
}

// AssignTasksToContest godoc
//
//	@Tags			contest
//	@Summary		Assign tasks to a contest
//	@Description	Assigns multiple tasks to a contest
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int				true	"Contest ID"
//	@Param			request	body		tasksRequest	true	"Task IDs to assign"
//	@Param			Session	header		string			true	"Session Token"
//	@Success		200		{object}	httputils.ApiResponse[string]
//	@Failure		400		{object}	httputils.ApiError
//	@Failure		404		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Router			/contest/{id}/assign/tasks [post]
func (cr *ContestRouteImpl) AssignTasksToContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	contestIdStr := r.PathValue("id")
	if contestIdStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Contest ID is required.")
		return
	}

	contestId, err := strconv.ParseInt(contestIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid contest ID.")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	var request tasksRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	err = cr.contestService.AssignTasksToContest(tx, currentUser, contestId, request.TaskIds)
	if err != nil {
		if err == errors.ErrContestNotFound {
			httputils.ReturnError(w, http.StatusNotFound, err.Error())
			return
		}
		if err == errors.ErrTaskNotFound {
			httputils.ReturnError(w, http.StatusNotFound, err.Error())
			return
		}
		if err == errors.ErrNotAuthorized {
			httputils.ReturnError(w, http.StatusForbidden, err.Error())
			return
		}
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to assign tasks to contest. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, "Tasks assigned to contest successfully")
}

// UnAssignTasksFromContest godoc
//
//	@Tags			contest
//	@Summary		Unassign tasks from a contest
//	@Description	Unassigns multiple tasks from a contest
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int				true	"Contest ID"
//	@Param			request	body		tasksRequest	true	"Task IDs to unassign"
//	@Param			Session	header		string			true	"Session Token"
//	@Success		200		{object}	httputils.ApiResponse[string]
//	@Failure		400		{object}	httputils.ApiError
//	@Failure		404		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Router			/contest/{id}/unassign/tasks [post]
func (cr *ContestRouteImpl) UnAssignTasksFromContest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	contestIdStr := r.PathValue("id")
	if contestIdStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Contest ID is required.")
		return
	}

	contestId, err := strconv.ParseInt(contestIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid contest ID.")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	var request tasksRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	err = cr.contestService.UnAssignTasksFromContest(tx, currentUser, contestId, request.TaskIds)
	if err != nil {
		if err == errors.ErrContestNotFound {
			httputils.ReturnError(w, http.StatusNotFound, err.Error())
			return
		}
		if err == errors.ErrNotAuthorized {
			httputils.ReturnError(w, http.StatusForbidden, err.Error())
			return
		}
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to unassign tasks from contest. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, "Tasks unassigned from contest successfully")
}

func NewContestRoute(contestService service.ContestService) ContestRoute {
	return &ContestRouteImpl{
		contestService: contestService,
	}
}

func RegisterContestRoutes(mux *http.ServeMux, route ContestRoute) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetAllContests(w, r)
		case http.MethodPost:
			route.CreateContest(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	mux.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetContest(w, r)
		case http.MethodPatch:
			route.EditContest(w, r)
		case http.MethodDelete:
			route.DeleteContest(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	mux.HandleFunc("/{id}/assign/tasks", route.AssignTasksToContest)
	mux.HandleFunc("/{id}/unassign/tasks", route.UnAssignTasksFromContest)
}
