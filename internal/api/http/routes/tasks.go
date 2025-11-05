package routes

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	myerrors "github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
)

type TaskRoute interface {
	GetAllAssignedTasks(w http.ResponseWriter, r *http.Request)
	GetAllForGroup(w http.ResponseWriter, r *http.Request)
	GetAllTasks(w http.ResponseWriter, r *http.Request)
	GetTask(w http.ResponseWriter, r *http.Request)
}

const taskBodyLimit = 50 << 20

type taskRoute struct {
	// Service that handles task-related operations
	taskService service.TaskService
	logger      *zap.SugaredLogger
}

type usersRequest struct {
	UserIDs []int64 `json:"userIDs"`
}

type groupsRequest struct {
	GroupIDs []int64 `json:"groupIDs"`
}

func (tr *taskRoute) GetAllAssignedTasks(w http.ResponseWriter, r *http.Request) {
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

	task, err := tr.taskService.GetAllAssigned(tx, currentUser, queryParams)
	if err != nil {
		db.Rollback()
		tr.logger.Errorw("Failed to get assigned tasks", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Task service temporarily unavailable")
		return
	}

	if task == nil {
		task = []schemas.Task{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, task)
}

// GetAllTasks godoc
//
//	@Tags			task
//	@Summary		Get all tasks
//	@Description	Returns all tasks
//	@Produce		json
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[[]schemas.Task]
//	@Router			/tasks/ [get]
func (tr *taskRoute) GetAllTasks(w http.ResponseWriter, r *http.Request) {
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

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)
	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	task, err := tr.taskService.GetAll(tx, currentUser, queryParams)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else {
			tr.logger.Errorw("Failed to get all tasks", "error", err)
		}
		httputils.ReturnError(w, status, "Task service temporarily unavailable")
		return
	}

	if task == nil {
		task = []schemas.Task{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, task)
}

// GetTask godoc
//
//	@Tags			task
//	@Summary		Get a task
//	@Description	Returns a task by ID
//	@Produce		json
//	@Param			id	path		int	true	"Task ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[schemas.TaskDetailed]
//	@Router			/tasks/{id} [get]
func (tr *taskRoute) GetTask(w http.ResponseWriter, r *http.Request) {
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

	task, err := tr.taskService.Get(tx, currentUser, taskID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, myerrors.ErrNotAuthorized):
			status = http.StatusForbidden
		case errors.Is(err, myerrors.ErrNotFound):
			status = http.StatusNotFound
		default:
			tr.logger.Errorw("Failed to get task", "error", err)
		}
		httputils.ReturnError(w, status, "Task service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, task)
}

// GetAllForGroup godoc
//
//	@Tags			task
//	@Summary		Get all tasks for a group
//	@Description	Returns all tasks for a group by ID
//	@Produce		json
//	@Param			id	path		int	true	"Group ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[[]schemas.Task]
//	@Router			/tasks/groups/{id} [get]
func (tr *taskRoute) GetAllForGroup(w http.ResponseWriter, r *http.Request) {
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

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	tasks, err := tr.taskService.GetAllForGroup(tx, currentUser, groupID, queryParams)
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

func NewTaskRoute(taskService service.TaskService) TaskRoute {
	route := &taskRoute{
		taskService: taskService,
		logger:      utils.NewNamedLogger("tasks"),
	}

	err := utils.ValidateStruct(*route)
	if err != nil {
		log.Panicf("TaskRoute struct is not valid: %s", err.Error())
	}

	return route
}

func RegisterTaskRoutes(mux *mux.Router, route TaskRoute) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetAllTasks(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	mux.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetTask(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	// mux.HandleFunc("/group/{id}", route.GetAllForGroup)
	mux.HandleFunc("/assigned", route.GetAllAssignedTasks)
}
