package routes

import (
	"errors"
	"log"
	"net/http"
	"strconv"

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
	GetContests(w http.ResponseWriter, r *http.Request)
}

type contestsRouteImpl struct {
	contestService service.ContestService
	logger         *zap.SugaredLogger
}

// GetContests godoc
//
//	@Tags			global-contests
//	@Summary		Get global contests
//	@Description	Get global contests accessible to the (ongoing, upcoming, past)
//	@Produce		json
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Param			status	query		string	true	"Contest status"	Enums(ongoing, upcoming, past)
//	@Success		200		{object}	httputils.APIResponse[[]schemas.AvailableContest]
//	@Failure		401		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/contests [get]
func (cr *contestsRouteImpl) GetContests(w http.ResponseWriter, r *http.Request) {
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
	var contests []schemas.AvailableContest
	switch statusStr {
	case "ongoing":
		contests, err = cr.contestService.GetOngoingContests(tx, currentUser, paginationParams)
	case "upcoming":
		contests, err = cr.contestService.GetUpcomingContests(tx, currentUser, paginationParams)
	case "past":
		contests, err = cr.contestService.GetPastContests(tx, currentUser, paginationParams)
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

	if contests == nil {
		contests = []schemas.AvailableContest{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, contests)
}

// GetContest godoc
//
//	@Tags			contests
//	@Summary		Get contest details
//	@Description	Get details of a specific contest
//	@Produce		json
//	@Param			id	path		int	true	"Contest ID"
//	@Success		200	{object}	httputils.APIResponse[schemas.Contest]
//	@Failure		400	{object}	httputils.APIError
//	@Failure		401	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Router			/contests/{id} [get]
func (cr *contestsRouteImpl) GetContest(w http.ResponseWriter, r *http.Request) {
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

func RegisterContestRoute(mux *http.ServeMux, route ContestRoute) {
	mux.HandleFunc("/contests", route.GetContests)
	mux.HandleFunc("/contests/{id}", route.GetContest)
}

func NewContestRoute(contestService service.ContestService) ContestRoute {
	route := &contestsRouteImpl{
		contestService: contestService,
		logger:         utils.NewNamedLogger("contests"),
	}
	err := utils.ValidateStruct(*route)
	if err != nil {
		log.Panicf("ContestRoute struct is not valid: %s", err.Error())
	}
	return route
}
