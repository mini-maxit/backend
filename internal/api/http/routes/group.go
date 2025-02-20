package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/api/http/middleware"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/service"
)

type GroupRoute interface {
	CreateGroup(w http.ResponseWriter, r *http.Request)
	GetGroup(w http.ResponseWriter, r *http.Request)
	GetAllGroup(w http.ResponseWriter, r *http.Request)
	EditGroup(w http.ResponseWriter, r *http.Request)
}

type GroupRouteImpl struct {
	groupService service.GroupService
}

// CreateGroup godoc
//
// @Tags group
// @Summary Create a group
// @Description Create a group
// @Accept json
// @Produce json
// @Param body body CreateGroup true "Create Group"
// @Success 200 {object} httputils.ApiResponse[int64]
// @Failure 400 {object} httputils.ApiError
// @Failure 401 {object} httputils.ApiError
// @Failure 405 {object} httputils.ApiError
// @Failure 500 {object} httputils.ApiError
// @Router /group/ [post]
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

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	userId := r.Context().Value(middleware.UserIDKey).(int64)

	group := &schemas.Group{
		Name:      request.Name,
		CreatedBy: userId,
	}
	groupId, err := gr.groupService.CreateGroup(tx, group)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to create group. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, groupId)
}

// GetGroup godoc
// @Tags group
// @Summary Get a group
// @Description Get a group
// @Produce json
// @Param id path int true "Group ID"
// @Success 200 {object} httputils.ApiResponse[schemas.Group]
// @Failure 400 {object} httputils.ApiError
// @Failure 401 {object} httputils.ApiError
// @Failure 404 {object} httputils.ApiError
// @Failure 405 {object} httputils.ApiError
// @Failure 500 {object} httputils.ApiError
// @Router /group/{id} [get]
func (gr *GroupRouteImpl) GetGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
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

	group, err := gr.groupService.GetGroup(tx, groupId)
	if err != nil {
		if err == service.ErrGroupNotFound {
			httputils.ReturnError(w, http.StatusNotFound, "Group not found")
			return
		}
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to get group. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, group)
}

// GetAllGroup godoc
// @Tags group
// @Summary Get all groups
// @Description Get all groups
// @Produce json
// @Param limit query string false "Limit"
// @Param offset query string false "Offset"
// @Success 200 {object} httputils.ApiResponse[[]schemas.Group]
// @Failure 400 {object} httputils.ApiError
// @Failure 401 {object} httputils.ApiError
// @Failure 405 {object} httputils.ApiError
// @Failure 500 {object} httputils.ApiError
// @Router /group/ [get]
func (gr *GroupRouteImpl) GetAllGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Error connecting to database. "+err.Error())
		return
	}

	queryParams := r.Context().Value(middleware.QueryParamsKey).(map[string]interface{})
	groups, err := gr.groupService.GetAllGroup(tx, queryParams)
	if err != nil {
		db.Rollback()
		if err == service.ErrInvalidLimitParam {
			httputils.ReturnError(w, http.StatusBadRequest, "Invalid limit")
			return
		} else if err == service.ErrInvalidOffsetParam {
			httputils.ReturnError(w, http.StatusBadRequest, "Invalid offset")
			return
		}
		httputils.ReturnError(w, http.StatusInternalServerError, "Error getting groups. "+err.Error())
		return
	}

	if groups == nil {
		groups = []schemas.Group{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, groups)
}

// EditGroup godoc
// @Tags group
// @Summary Edit a group
// @Description Edit a group
// @Accept json
// @Produce json
// @Param id path int true "Group ID"
// @Param body body EditGroup true "Edit Group"
// @Success 200 {object} httputils.ApiResponse[schemas.Group]
// @Failure 400 {object} httputils.ApiError
// @Failure 401 {object} httputils.ApiError
// @Failure 404 {object} httputils.ApiError
// @Failure 405 {object} httputils.ApiError
// @Failure 500 {object} httputils.ApiError
// @Router /group/{id} [put]
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

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
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

	resp, err := gr.groupService.Edit(tx, groupId, &request)
	if err != nil {
		db.Rollback()
		if err == service.ErrGroupNotFound {
			httputils.ReturnError(w, http.StatusNotFound, "Group not found")
			return
		}
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to edit group. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, resp)
}

func NewGroupRoute(groupService service.GroupService) GroupRoute {
	return &GroupRouteImpl{
		groupService: groupService,
	}
}
