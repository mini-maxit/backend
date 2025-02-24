package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
)

type GroupRoute interface {
	CreateGroup(w http.ResponseWriter, r *http.Request)
	GetGroup(w http.ResponseWriter, r *http.Request)
	GetAllGroup(w http.ResponseWriter, r *http.Request)
	EditGroup(w http.ResponseWriter, r *http.Request)
	AddUsersToGroup(w http.ResponseWriter, r *http.Request)
	GetGroupUsers(w http.ResponseWriter, r *http.Request)
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
//	@Failure		400		{object}	httputils.ApiError
//	@Failure		403		{object}	httputils.ApiError
//	@Failure		405		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Success		200		{object}	httputils.ApiResponse[int64]
//	@Router			/group/ [post]
func (gr *GroupRouteImpl) CreateGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.CreateGroup
	err := json.NewDecoder(r.Body).Decode(&request)
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

	current_user := r.Context().Value(httputils.UserKey).(schemas.User)

	group := &schemas.Group{
		Name:      request.Name,
		CreatedBy: current_user.Id,
	}
	groupId, err := gr.groupService.CreateGroup(tx, current_user, group)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if err == errors.ErrNotAuthorized {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, "Failed to create group. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, groupId)
}

// GetGroup godoc
//
//	@Tags			group
//	@Summary		Get a group
//	@Description	Get a group
//	@Produce		json
//	@Param			id	path		int	true	"Group ID"
//	@Failure		400	{object}	httputils.ApiError
//	@Failure		403	{object}	httputils.ApiError
//	@Failure		405	{object}	httputils.ApiError
//	@Failure		500	{object}	httputils.ApiError
//	@Success		200	{object}	httputils.ApiResponse[schemas.Group]
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
	groupId, err := strconv.ParseInt(groupStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	current_user := r.Context().Value(httputils.UserKey).(schemas.User)

	group, err := gr.groupService.GetGroup(tx, current_user, groupId)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if err == errors.ErrNotAuthorized {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, "Failed to create group. "+err.Error())
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
//	@Failure		400	{object}	httputils.ApiError
//	@Failure		403	{object}	httputils.ApiError
//	@Failure		405	{object}	httputils.ApiError
//	@Failure		500	{object}	httputils.ApiError
//	@Success		200	{object}	httputils.ApiResponse[[]schemas.Group]
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

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]interface{})
	current_user := r.Context().Value(httputils.UserKey).(schemas.User)

	groups, err := gr.groupService.GetAllGroup(tx, current_user, queryParams)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if err == errors.ErrNotAuthorized {
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
//	@Failure		400		{object}	httputils.ApiError
//	@Failure		403		{object}	httputils.ApiError
//	@Failure		405		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Success		200		{object}	httputils.ApiResponse[schemas.Group]
//	@Router			/group/{id} [put]
func (gr *GroupRouteImpl) EditGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.EditGroup
	err := json.NewDecoder(r.Body).Decode(&request)
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
	groupId, err := strconv.ParseInt(groupStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	current_user := r.Context().Value(httputils.UserKey).(schemas.User)

	resp, err := gr.groupService.Edit(tx, current_user, groupId, &request)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if err == errors.ErrNotAuthorized {
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
//	@Param			body	body		schemas.UserIds	true	"User IDs"
//	@Failure		400		{object}	httputils.ApiError
//	@Failure		403		{object}	httputils.ApiError
//	@Failure		405		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Success		200		{object}	httputils.ApiResponse[string]
//	@Router			/group/{id}/users [post]
func (gr *GroupRouteImpl) AddUsersToGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	groupIdStr := r.PathValue("id")
	if groupIdStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Group ID cannot be empty")
		return
	}

	groupId, err := strconv.ParseInt(groupIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	decoder := json.NewDecoder(r.Body)
	request := schemas.UserIds{}
	err = decoder.Decode(&request)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group IDs.")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	current_user := r.Context().Value(httputils.UserKey).(schemas.User)

	err = gr.groupService.AddUsersToGroup(tx, current_user, groupId, request.UserIds)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if err == errors.ErrNotAuthorized {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, "Failed to add users to group. "+err.Error())
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, "Users added to group successfully")
}

// GetGroupUsers godoc
//
//	@Tags			group
//	@Summary		Get users in a group
//	@Description	Get users in a group
//	@Produce		json
//	@Param			id	path		int	true	"Group ID"
//	@Failure		400	{object}	httputils.ApiError
//	@Failure		403	{object}	httputils.ApiError
//	@Failure		405	{object}	httputils.ApiError
//	@Failure		500	{object}	httputils.ApiError
//	@Success		200	{object}	httputils.ApiResponse[string]
//	@Router			/group/{id}/users [get]
func (gr *GroupRouteImpl) GetGroupUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	groupIdStr := r.PathValue("id")
	if groupIdStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Group ID cannot be empty")
		return
	}

	groupId, err := strconv.ParseInt(groupIdStr, 10, 64)
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

	current_user := r.Context().Value(httputils.UserKey).(schemas.User)

	users, err := gr.groupService.GetGroupUsers(tx, current_user, groupId)
	if err != nil {
		db.Rollback()
		status := http.StatusInternalServerError
		if err == errors.ErrNotAuthorized {
			status = http.StatusForbidden
		}
		httputils.ReturnError(w, status, "Failed to get group users. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, users)
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
		case http.MethodGet:
			groupRoute.GetGroupUsers(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
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
