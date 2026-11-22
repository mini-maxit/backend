package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
)

type DashboardRoute interface {
	GetAdminStats(w http.ResponseWriter, r *http.Request)
	GetContestActivity(w http.ResponseWriter, r *http.Request)
}

type dashboardRoute struct {
	dashboardService service.DashboardService
	logger           *zap.SugaredLogger
}

// GetAdminStats godoc
//
//	@Tags			dashboard
//	@Summary		Get platform usage statistics
//	@Description	User, contest, task and submission counts, plus hourly submissions and outcomes for the last 24 hours (admin only)
//	@Produce		json
//	@Failure		403	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[schemas.AdminDashboardStats]
//	@Router			/dashboard/admin [get]
func (dr *dashboardRoute) GetAdminStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := httputils.GetDatabase(r)
	stats, err := dr.dashboardService.GetAdminStats(db, httputils.GetCurrentUser(r))
	if err != nil {
		httputils.HandleServiceError(w, err, db, dr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, stats)
}

// GetContestActivity godoc
//
//	@Tags			dashboard
//	@Summary		Get recent submission activity of managed contests
//	@Description	For contests the user manages that are running or ended in the last 7 days: submissions in the last 24 hours and the latest submission time (teacher or admin)
//	@Produce		json
//	@Failure		403	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[[]schemas.ContestActivity]
//	@Router			/dashboard/teacher/contest-activity [get]
func (dr *dashboardRoute) GetContestActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := httputils.GetDatabase(r)
	activity, err := dr.dashboardService.GetContestActivity(db, httputils.GetCurrentUser(r))
	if err != nil {
		httputils.HandleServiceError(w, err, db, dr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, activity)
}

func RegisterDashboardRoutes(mux *mux.Router, dr DashboardRoute) {
	mux.HandleFunc("/admin", dr.GetAdminStats)
	mux.HandleFunc("/teacher/contest-activity", dr.GetContestActivity)
}

func NewDashboardRoute(dashboardService service.DashboardService) DashboardRoute {
	return &dashboardRoute{
		dashboardService: dashboardService,
		logger:           utils.NewNamedLogger("dashboard"),
	}
}
