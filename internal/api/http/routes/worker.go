package routes

import (
	"net/http"

	"errors"

	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/package/domain/schemas"
	myerrors "github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/service"
)

type WorkerRoute interface {
	// GetStatus returns current status of workers
	GetStatus(w http.ResponseWriter, r *http.Request)
}

type workerRoute struct {
	workserService service.WorkerService
}

// GetStatus godoc
//
//	@Tags			worker
//	@Summary		Get worker status
//	@Description	Returns the current status of all worker nodes
//	@Produce		json
//	@Failure		401	{object}	httputils.APIError	"Not authorized - requires teacher or admin role"
//	@Failure		504	{object}	httputils.APIError	"Gateway timeout - worker status request timed out"
//	@Failure		500	{object}	httputils.APIError	"Internal server error"
//	@Success		200	{object}	httputils.APIResponse[schemas.WorkerStatus]
//	@Router			/worker/status [get]
func (wr *workerRoute) GetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	status, err := wr.workserService.GetStatus(currentUser)
	if err != nil {
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			httputils.ReturnError(w, http.StatusUnauthorized, err.Error())
			return
		} else if errors.Is(err, myerrors.ErrTimeout) {
			httputils.ReturnError(w, http.StatusGatewayTimeout, err.Error())
			return
		}
		httputils.ReturnError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, status)
}

func RegisterWorkerRoutes(mux *mux.Router, wr WorkerRoute) {
	mux.HandleFunc("/status", wr.GetStatus)
}

func NewWorkerRoute(workserService service.WorkerService) WorkerRoute {
	return &workerRoute{
		workserService: workserService,
	}
}
