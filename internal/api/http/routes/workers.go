package routes

import (
	"net/http"

	"errors"

	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	myerrors "github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
)

type WorkerRoute interface {
	// GetStatus returns current status of workers
	GetStatus(w http.ResponseWriter, r *http.Request)
	// GetQueueStatus returns current status of the queue connection
	GetQueueStatus(w http.ResponseWriter, r *http.Request)
	// ReconnectQueue triggers queue reconnection and processes pending submissions
	ReconnectQueue(w http.ResponseWriter, r *http.Request)
}

type workerRoute struct {
	workserService service.WorkerService
	logger         *zap.SugaredLogger
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
//	@Router			/workers/status [get]
func (wr *workerRoute) GetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	currentUser := httputils.GetCurrentUser(r)

	status, err := wr.workserService.GetStatus(currentUser)
	if err != nil {
		if errors.Is(err, myerrors.ErrForbidden) {
			httputils.ReturnError(w, http.StatusForbidden, err.Error())
			return
		} else if errors.Is(err, myerrors.ErrTimeout) {
			httputils.ReturnError(w, http.StatusGatewayTimeout, err.Error())
			return
		}
		wr.logger.Errorw("Failed to get worker status", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Worker service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, status)
}

// GetQueueStatus godoc
//
//	@Tags			worker
//	@Summary		Get queue status
//	@Description	Returns the current status of the queue connection and pending submissions
//	@Produce		json
//	@Failure		401	{object}	httputils.APIError	"Not authorized - requires admin role"
//	@Failure		500	{object}	httputils.APIError	"Internal server error"
//	@Success		200	{object}	httputils.APIResponse[schemas.QueueStatus]
//	@Router			/workers/queue/status [get]
func (wr *workerRoute) GetQueueStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	currentUser := httputils.GetCurrentUser(r)

	status, err := wr.workserService.GetQueueStatus(currentUser)
	if err != nil {
		if errors.Is(err, myerrors.ErrForbidden) {
			httputils.ReturnError(w, http.StatusForbidden, err.Error())
			return
		}
		wr.logger.Errorw("Failed to get queue status", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to get queue status")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, status)
}

// ReconnectQueue godoc
//
//	@Tags			worker
//	@Summary		Reconnect to queue
//	@Description	Triggers queue reconnection and processes pending submissions (admin only)
//	@Produce		json
//	@Failure		401	{object}	httputils.APIError				"Not authorized - requires admin role"
//	@Failure		500	{object}	httputils.APIError				"Internal server error"
//	@Success		200	{object}	httputils.APIResponse[string]	"message"
//	@Router			/workers/queue/reconnect [post]
func (wr *workerRoute) ReconnectQueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	currentUser := httputils.GetCurrentUser(r)

	err := wr.workserService.ReconnectQueue(currentUser)
	if err != nil {
		if errors.Is(err, myerrors.ErrForbidden) {
			httputils.ReturnError(w, http.StatusForbidden, err.Error())
			return
		}
		wr.logger.Errorw("Failed to reconnect queue", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to reconnect queue: "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, "Queue reconnection initiated and pending submissions processed")
}

func RegisterWorkerRoutes(mux *mux.Router, wr WorkerRoute) {
	mux.HandleFunc("/status", wr.GetStatus)
	mux.HandleFunc("/queue/status", wr.GetQueueStatus)
	mux.HandleFunc("/queue/reconnect", wr.ReconnectQueue)
}

func NewWorkerRoute(workserService service.WorkerService) WorkerRoute {
	return &workerRoute{
		workserService: workserService,
		logger:         utils.NewNamedLogger("workers"),
	}
}
