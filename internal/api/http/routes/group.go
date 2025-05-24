package routes

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	myerrors "github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
)

type GroupRoute interface {
	CreateGroup(w http.ResponseWriter, r *http.Request)
	GetGroup(w http.ResponseWriter, r *http.Request)
	GetAllGroup(w http.ResponseWriter, r *http.Request)
	EditGroup(w http.ResponseWriter, r *http.Request)
	AddUsersToGroup(w http.ResponseWriter, r *http.Request)
	DeleteUsersFromGroup(w http.ResponseWriter, r *http.Request)
	GetGroupUsers(w http.ResponseWriter, r *http.Request)
	GetGroupTasks(w http.ResponseWriter, r *http.Request)
}

type GroupRouteImpl struct {
	groupService service.GroupService
}

// CreateGroup godoc
//
//	@Tags			group
//	@Summary		Create a group
//	@Description	Create a group
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CreateGroup	true	"Create Group"
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[int64]
//	@Router			/group/ [post]
func (gr *GroupRouteImpl) CreateGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.CreateGroup
	err := httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}
	if request.Name == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. Group name cannot be empty")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	group := &schemas.Group{
		Name:      request.Name,
		CreatedBy: currentUser.ID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	groupID, err := gr.groupService.Create(tx, currentUser, group)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, "Failed to create group. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, groupID)
}

// GetGroup godoc
//
//	@Tags			group
//	@Summary		Get a group
//	@Description	Get a group
//	@Produce		json
//	@Param			id	path		int	true	"Group ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[schemas.Group]
//	@Router			/group/{id} [get]
func (gr *GroupRouteImpl) GetGroup(w http.ResponseWriter, r *http.Request) {
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

	groupStr := r.PathValue("id")
	if groupStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Group ID cannot be empty")
		return
	}
	groupID, err := strconv.ParseInt(groupStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	group, err := gr.groupService.Get(tx, currentUser, groupID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrGroupNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, "Failed to get group. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, group)
}

// GetAllGroup godoc
//
//	@Tags			group
//	@Summary		Get all groups
//	@Description	Get all groups
//	@Produce		json
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[[]schemas.Group]
//	@Router			/group/ [get]
func (gr *GroupRouteImpl) GetAllGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Error connecting to database. "+err.Error())
		return
	}

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	groups, err := gr.groupService.GetAll(tx, currentUser, queryParams)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, "Failed to list groups. "+err.Error())
		return
	}

	if groups == nil {
		groups = []schemas.Group{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, groups)
}

// EditGroup godoc
//
//	@Tags			group
//	@Summary		Edit a group
//	@Description	Edit a group
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int			true	"Group ID"
//	@Param			body	body		EditGroup	true	"Edit Group"
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[schemas.Group]
//	@Router			/group/{id} [put]
func (gr *GroupRouteImpl) EditGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.EditGroup
	err := httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	groupStr := r.PathValue("id")
	if groupStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Group ID cannot be empty")
		return
	}
	groupID, err := strconv.ParseInt(groupStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	resp, err := gr.groupService.Edit(tx, currentUser, groupID, &request)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, "Failed to create group. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, resp)
}

// AddUsersToGroup godoc
//
//	@Tags			group
//	@Summary		Add users to a group
//	@Description	Add users to a group
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int				true	"Group ID"
//	@Param			body	body		schemas.UserIDs	true	"User IDs"
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[string]
//	@Router			/group/{id}/users [post]
func (gr *GroupRouteImpl) AddUsersToGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	groupIDStr := r.PathValue("id")
	if groupIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Group ID cannot be empty")
		return
	}

	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	request := &schemas.UserIDs{}
	err = httputils.ShouldBindJSON(r.Body, request)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = gr.groupService.AddUsers(tx, currentUser, groupID, request.UserIDs)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrGroupNotFound) {
			status = http.StatusNotFound
		}
		httputils.ReturnError(w, status, "Failed to add users to group. "+err.Error())
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, "Users added to group successfully")
}

// DeleteUsersFromGroup godoc
//
//	@Tags			group
//	@Summary		Delete users from a group
//	@Description	Delete users from a group
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int				true	"Group ID"
//	@Param			body	body		schemas.UserIDs	true	"User IDs"
//	@Failure		400		{object}	httputils.APIError
//	@Failure		403		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[string]
//	@Router			/group/{id}/users [delete]
func (gr *GroupRouteImpl) DeleteUsersFromGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	groupIDStr := r.PathValue("id")
	if groupIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Group ID cannot be empty")
		return
	}

	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	request := &schemas.UserIDs{}
	err = httputils.ShouldBindJSON(r.Body, request)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = gr.groupService.DeleteUsers(tx, currentUser, groupID, request.UserIDs)
	if err != nil {
		db.Rollback()
		var status int
		switch {
		case errors.Is(err, myerrors.ErrNotAuthorized):
			status = http.StatusForbidden
		case errors.Is(err, myerrors.ErrUserNotFound):
			status = http.StatusBadRequest
		case errors.Is(err, myerrors.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(err, myerrors.ErrGroupNotFound):
			status = http.StatusNotFound
		default:
			status = http.StatusInternalServerError
		}
		httputils.ReturnError(w, status, "Failed to delete users from group. "+err.Error())
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, "Users deleted from group successfully")
}

// GetGroupUsers godoc
//
//	@Tags			group
//	@Summary		Get users in a group
//	@Description	Get users in a group
//	@Produce		json
//	@Param			id	path		int	true	"Group ID"
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[string]
//	@Router			/group/{id}/users [get]
func (gr *GroupRouteImpl) GetGroupUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	groupIDStr := r.PathValue("id")
	if groupIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Group ID cannot be empty")
		return
	}

	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	users, err := gr.groupService.GetUsers(tx, currentUser, groupID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrGroupNotFound) {
			status = http.StatusNotFound
		}
		httputils.ReturnError(w, status, "Failed to get group users. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, users)
}

func (gr *GroupRouteImpl) GetGroupTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	groupIDStr := r.PathValue("id")
	if groupIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Group ID cannot be empty")
		return
	}

	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	tasks, err := gr.groupService.GetTasks(tx, currentUser, groupID)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if errors.Is(err, myerrors.ErrNotAuthorized) {
			status = http.StatusForbidden
		} else if errors.Is(err, myerrors.ErrGroupNotFound) {
			status = http.StatusNotFound
		}
		httputils.ReturnError(w, status, "Failed to get group tasks. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, tasks)
}

func RegisterGroupRoutes(mux *http.ServeMux, groupRoute GroupRoute) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			groupRoute.CreateGroup(w, r)
		case http.MethodGet:
			groupRoute.GetAllGroup(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			groupRoute.GetGroup(w, r)
		case http.MethodPut:
			groupRoute.EditGroup(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/{id}/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			groupRoute.AddUsersToGroup(w, r)
		case http.MethodDelete:
			groupRoute.DeleteUsersFromGroup(w, r)
		case http.MethodGet:
			groupRoute.GetGroupUsers(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/{id}/tasks", groupRoute.GetGroupTasks)
}

func NewGroupRoute(groupService service.GroupService) GroupRoute {
	route := &GroupRouteImpl{
		groupService: groupService,
	}
	err := utils.ValidateStruct(*route)
	if err != nil {
		log.Panicf("GroupRoute struct is not valid: %s", err.Error())
	}
	return route
}
