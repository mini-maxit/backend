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
	myerrors "github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
)

type AccessControlRoute interface {
	// Contest collaborators
	AddContestCollaborator(w http.ResponseWriter, r *http.Request)
	GetContestCollaborators(w http.ResponseWriter, r *http.Request)
	UpdateContestCollaborator(w http.ResponseWriter, r *http.Request)
	RemoveContestCollaborator(w http.ResponseWriter, r *http.Request)

	// Task collaborators
	AddTaskCollaborator(w http.ResponseWriter, r *http.Request)
	GetTaskCollaborators(w http.ResponseWriter, r *http.Request)
	UpdateTaskCollaborator(w http.ResponseWriter, r *http.Request)
	RemoveTaskCollaborator(w http.ResponseWriter, r *http.Request)
}

type accessControlRoute struct {
	contestService service.ContestService
	taskService    service.TaskService
	logger         *zap.SugaredLogger
}

// Contest Collaborators

// AddContestCollaborator godoc
//
//	@Tags			access-control
//	@Summary		Add a collaborator to a contest
//	@Description	Add a user as a collaborator to a contest with specified permissions (view, edit, or manage). Only users with manage permission can add collaborators.
//	@Accept			json
//	@Produce		json
//	@Param			resource_id	path		int						true	"Contest ID"
//	@Param			body		body		schemas.AddCollaborator	true	"Collaborator details"
//	@Failure		400			{object}	httputils.ValidationErrorResponse
//	@Failure		403			{object}	httputils.APIError
//	@Failure		404			{object}	httputils.APIError
//	@Failure		500			{object}	httputils.APIError
//	@Success		200			{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/access-control/contests/{resource_id}/collaborators [post]
func (ac *accessControlRoute) AddContestCollaborator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	resourceIDStr := httputils.GetPathValue(r, "resource_id")
	if resourceIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Contest ID cannot be empty")
		return
	}
	contestID, err := strconv.ParseInt(resourceIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid contest ID")
		return
	}

	var request schemas.AddCollaborator
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

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		ac.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = ac.contestService.AddContestCollaborator(tx, currentUser, contestID, &request)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else {
			ac.logger.Errorw("Failed to add contest collaborator", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to add collaborator")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Collaborator added successfully"))
}

// GetContestCollaborators godoc
//
//	@Tags			access-control
//	@Summary		Get collaborators for a contest
//	@Description	Get all collaborators for a specific contest. Users with view permission or higher can see collaborators.
//	@Produce		json
//	@Param			resource_id	path		int	true	"Contest ID"
//	@Failure		400			{object}	httputils.APIError
//	@Failure		403			{object}	httputils.APIError
//	@Failure		404			{object}	httputils.APIError
//	@Failure		500			{object}	httputils.APIError
//	@Success		200			{object}	httputils.APIResponse[[]schemas.Collaborator]
//	@Router			/access-control/contests/{resource_id}/collaborators [get]
func (ac *accessControlRoute) GetContestCollaborators(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	resourceIDStr := httputils.GetPathValue(r, "resource_id")
	if resourceIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Contest ID cannot be empty")
		return
	}
	contestID, err := strconv.ParseInt(resourceIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid contest ID")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		ac.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	collaborators, err := ac.contestService.GetContestCollaborators(tx, currentUser, contestID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else {
			ac.logger.Errorw("Failed to get contest collaborators", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to get collaborators")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, collaborators)
}

// UpdateContestCollaborator godoc
//
//	@Tags			access-control
//	@Summary		Update a contest collaborator's permission
//	@Description	Update the permission level of a collaborator for a contest. Only users with manage permission can update collaborators.
//	@Accept			json
//	@Produce		json
//	@Param			resource_id	path		int							true	"Contest ID"
//	@Param			user_id		path		int							true	"User ID"
//	@Param			body		body		schemas.UpdateCollaborator	true	"New permission"
//	@Failure		400			{object}	httputils.ValidationErrorResponse
//	@Failure		403			{object}	httputils.APIError
//	@Failure		404			{object}	httputils.APIError
//	@Failure		500			{object}	httputils.APIError
//	@Success		200			{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/access-control/contests/{resource_id}/collaborators/{user_id} [put]
func (ac *accessControlRoute) UpdateContestCollaborator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	resourceIDStr := httputils.GetPathValue(r, "resource_id")
	if resourceIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Contest ID cannot be empty")
		return
	}
	contestID, err := strconv.ParseInt(resourceIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid contest ID")
		return
	}

	userIDStr := httputils.GetPathValue(r, "user_id")
	if userIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "User ID cannot be empty")
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var request schemas.UpdateCollaborator
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

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		ac.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = ac.contestService.UpdateContestCollaborator(tx, currentUser, contestID, userID, &request)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else {
			ac.logger.Errorw("Failed to update contest collaborator", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to update collaborator permission")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Collaborator permission updated successfully"))
}

// RemoveContestCollaborator godoc
//
//	@Tags			access-control
//	@Summary		Remove a collaborator from a contest
//	@Description	Remove a user's collaborator access to a contest. Only users with manage permission can remove collaborators. Cannot remove the creator.
//	@Produce		json
//	@Param			resource_id	path		int	true	"Contest ID"
//	@Param			user_id		path		int	true	"User ID"
//	@Failure		400			{object}	httputils.APIError
//	@Failure		403			{object}	httputils.APIError
//	@Failure		404			{object}	httputils.APIError
//	@Failure		500			{object}	httputils.APIError
//	@Success		200			{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/access-control/contests/{resource_id}/collaborators/{user_id} [delete]
func (ac *accessControlRoute) RemoveContestCollaborator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	resourceIDStr := httputils.GetPathValue(r, "resource_id")
	if resourceIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Contest ID cannot be empty")
		return
	}
	contestID, err := strconv.ParseInt(resourceIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid contest ID")
		return
	}

	userIDStr := httputils.GetPathValue(r, "user_id")
	if userIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "User ID cannot be empty")
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		ac.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = ac.contestService.RemoveContestCollaborator(tx, currentUser, contestID, userID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else {
			ac.logger.Errorw("Failed to remove contest collaborator", "error", err)
		}
		httputils.ReturnError(w, status, err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Collaborator removed successfully"))
}

// Task Collaborators

// AddTaskCollaborator godoc
//
//	@Tags			access-control
//	@Summary		Add a collaborator to a task
//	@Description	Add a user as a collaborator to a task with specified permissions (view, edit, or manage). Only users with manage permission can add collaborators.
//	@Accept			json
//	@Produce		json
//	@Param			resource_id	path		int						true	"Task ID"
//	@Param			body		body		schemas.AddCollaborator	true	"Collaborator details"
//	@Failure		400			{object}	httputils.ValidationErrorResponse
//	@Failure		403			{object}	httputils.APIError
//	@Failure		404			{object}	httputils.APIError
//	@Failure		500			{object}	httputils.APIError
//	@Success		200			{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/access-control/tasks/{resource_id}/collaborators [post]
func (ac *accessControlRoute) AddTaskCollaborator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	resourceIDStr := httputils.GetPathValue(r, "resource_id")
	if resourceIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID cannot be empty")
		return
	}
	taskID, err := strconv.ParseInt(resourceIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var request schemas.AddCollaborator
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

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		ac.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = ac.taskService.AddTaskCollaborator(tx, currentUser, taskID, &request)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else {
			ac.logger.Errorw("Failed to add task collaborator", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to add collaborator")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Collaborator added successfully"))
}

// GetTaskCollaborators godoc
//
//	@Tags			access-control
//	@Summary		Get collaborators for a task
//	@Description	Get all collaborators for a specific task. Users with view permission or higher can see collaborators.
//	@Produce		json
//	@Param			resource_id	path		int	true	"Task ID"
//	@Failure		400			{object}	httputils.APIError
//	@Failure		403			{object}	httputils.APIError
//	@Failure		404			{object}	httputils.APIError
//	@Failure		500			{object}	httputils.APIError
//	@Success		200			{object}	httputils.APIResponse[[]schemas.Collaborator]
//	@Router			/access-control/tasks/{resource_id}/collaborators [get]
func (ac *accessControlRoute) GetTaskCollaborators(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	resourceIDStr := httputils.GetPathValue(r, "resource_id")
	if resourceIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID cannot be empty")
		return
	}
	taskID, err := strconv.ParseInt(resourceIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		ac.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	collaborators, err := ac.taskService.GetTaskCollaborators(tx, currentUser, taskID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else {
			ac.logger.Errorw("Failed to get task collaborators", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to get collaborators")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, collaborators)
}

// UpdateTaskCollaborator godoc
//
//	@Tags			access-control
//	@Summary		Update a task collaborator's permission
//	@Description	Update the permission level of a collaborator for a task. Only users with manage permission can update collaborators.
//	@Accept			json
//	@Produce		json
//	@Param			resource_id	path		int							true	"Task ID"
//	@Param			user_id		path		int							true	"User ID"
//	@Param			body		body		schemas.UpdateCollaborator	true	"New permission"
//	@Failure		400			{object}	httputils.ValidationErrorResponse
//	@Failure		403			{object}	httputils.APIError
//	@Failure		404			{object}	httputils.APIError
//	@Failure		500			{object}	httputils.APIError
//	@Success		200			{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/access-control/tasks/{resource_id}/collaborators/{user_id} [put]
func (ac *accessControlRoute) UpdateTaskCollaborator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	resourceIDStr := httputils.GetPathValue(r, "resource_id")
	if resourceIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID cannot be empty")
		return
	}
	taskID, err := strconv.ParseInt(resourceIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	userIDStr := httputils.GetPathValue(r, "user_id")
	if userIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "User ID cannot be empty")
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var request schemas.UpdateCollaborator
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

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		ac.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = ac.taskService.UpdateTaskCollaborator(tx, currentUser, taskID, userID, &request)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else {
			ac.logger.Errorw("Failed to update task collaborator", "error", err)
		}
		httputils.ReturnError(w, status, "Failed to update collaborator permission")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Collaborator permission updated successfully"))
}

// RemoveTaskCollaborator godoc
//
//	@Tags			access-control
//	@Summary		Remove a collaborator from a task
//	@Description	Remove a user's collaborator access to a task. Only users with manage permission can remove collaborators. Cannot remove the creator.
//	@Produce		json
//	@Param			resource_id	path		int	true	"Task ID"
//	@Param			user_id		path		int	true	"User ID"
//	@Failure		400			{object}	httputils.APIError
//	@Failure		403			{object}	httputils.APIError
//	@Failure		404			{object}	httputils.APIError
//	@Failure		500			{object}	httputils.APIError
//	@Success		200			{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/access-control/tasks/{resource_id}/collaborators/{user_id} [delete]
func (ac *accessControlRoute) RemoveTaskCollaborator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	resourceIDStr := httputils.GetPathValue(r, "resource_id")
	if resourceIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Task ID cannot be empty")
		return
	}
	taskID, err := strconv.ParseInt(resourceIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	userIDStr := httputils.GetPathValue(r, "user_id")
	if userIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "User ID cannot be empty")
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		ac.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = ac.taskService.RemoveTaskCollaborator(tx, currentUser, taskID, userID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrForbidden) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrNotFound) {
			status = http.StatusNotFound
		} else {
			ac.logger.Errorw("Failed to remove task collaborator", "error", err)
		}
		httputils.ReturnError(w, status, err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Collaborator removed successfully"))
}

func NewAccessControlRoute(contestService service.ContestService, taskService service.TaskService) AccessControlRoute {
	route := &accessControlRoute{
		contestService: contestService,
		taskService:    taskService,
		logger:         utils.NewNamedLogger("access_control_route"),
	}
	if err := utils.ValidateStruct(*route); err != nil {
		panic(err)
	}
	return route
}

func RegisterAccessControlRoutes(mux *mux.Router, route AccessControlRoute) {
	// Contest collaborators
	mux.HandleFunc("/contests/{resource_id}/collaborators", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetContestCollaborators(w, r)
		case http.MethodPost:
			route.AddContestCollaborator(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/contests/{resource_id}/collaborators/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			route.UpdateContestCollaborator(w, r)
		case http.MethodDelete:
			route.RemoveContestCollaborator(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// Task collaborators
	mux.HandleFunc("/tasks/{resource_id}/collaborators", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetTaskCollaborators(w, r)
		case http.MethodPost:
			route.AddTaskCollaborator(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/tasks/{resource_id}/collaborators/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			route.UpdateTaskCollaborator(w, r)
		case http.MethodDelete:
			route.RemoveTaskCollaborator(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
}
