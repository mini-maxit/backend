package routes

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
)

type GroupRoute interface {
	CreateGroup(w http.ResponseWriter, r *http.Request)
	GetGroup(w http.ResponseWriter, r *http.Request)
	GetAllGroup(w http.ResponseWriter, r *http.Request)
	EditGroup(w http.ResponseWriter, r *http.Request)
	AddUsersToGroup(w http.ResponseWriter, r *http.Request)
	DeleteUsersFromGroup(w http.ResponseWriter, r *http.Request)
	GetGroupUsers(w http.ResponseWriter, r *http.Request)
}

type GroupRouteImpl struct {
	groupService service.GroupService
	logger       *zap.SugaredLogger
}

// CreateGroup godoc
//
//	@Tags			group
//	@Summary		Create a group
//	@Description	Create a group
//	@Accept			json
//	@Produce		json
//	@Param			body	body		schemas.CreateGroup	true	"Create Group"
//	@Failure		400		{object}	httputils.ValidationErrorResponse
//	@Failure		403		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[httputils.IDResponse]
//	@Router			/groups/ [post]
func (gr *GroupRouteImpl) CreateGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.CreateGroup
	err := httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		httputils.HandleValidationError(w, err)
		return
	}
	if request.Name == "" {
		httputils.ReturnError(w, http.StatusBadRequest, httputils.InvalidRequestBodyMessage)
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		gr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := httputils.GetCurrentUser(r)

	group := &schemas.Group{
		Name:      request.Name,
		CreatedBy: currentUser.ID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	groupID, err := gr.groupService.Create(tx, *currentUser, group)
	if err != nil {
		httputils.HandleServiceError(w, err, db, gr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewIDResponse(groupID))
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
//	@Router			/groups/{id} [get]
func (gr *GroupRouteImpl) GetGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		gr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	groupStr := httputils.GetPathValue(r, "id")
	if groupStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Group ID cannot be empty")
		return
	}
	groupID, err := strconv.ParseInt(groupStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	currentUser := httputils.GetCurrentUser(r)

	group, err := gr.groupService.Get(tx, *currentUser, groupID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, gr.logger)
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
//	@Router			/groups/ [get]
func (gr *GroupRouteImpl) GetAllGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		gr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	paginationParams := httputils.ExtractPaginationParams(queryParams)
	currentUser := httputils.GetCurrentUser(r)

	groups, err := gr.groupService.GetAll(tx, *currentUser, paginationParams)
	if err != nil {
		httputils.HandleServiceError(w, err, db, gr.logger)
		return
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
//	@Param			id		path		int					true	"Group ID"
//	@Param			body	body		schemas.EditGroup	true	"Edit Group"
//	@Failure		400		{object}	httputils.ValidationErrorResponse
//	@Failure		403		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[schemas.Group]
//	@Router			/groups/{id} [put]
func (gr *GroupRouteImpl) EditGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.EditGroup
	err := httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		httputils.HandleValidationError(w, err)
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		gr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	groupStr := httputils.GetPathValue(r, "id")
	if groupStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Group ID cannot be empty")
		return
	}
	groupID, err := strconv.ParseInt(groupStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	currentUser := httputils.GetCurrentUser(r)

	resp, err := gr.groupService.Edit(tx, *currentUser, groupID, &request)
	if err != nil {
		httputils.HandleServiceError(w, err, db, gr.logger)
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
//	@Failure		400		{object}	httputils.ValidationErrorResponse
//	@Failure		403		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/groups/{id}/users [post]
func (gr *GroupRouteImpl) AddUsersToGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	groupIDStr := httputils.GetPathValue(r, "id")
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
		httputils.HandleValidationError(w, err)
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		gr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := httputils.GetCurrentUser(r)

	err = gr.groupService.AddUsers(tx, *currentUser, groupID, request.UserIDs)
	if err != nil {
		httputils.HandleServiceError(w, err, db, gr.logger)
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Users added to group successfully"))
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
//	@Failure		400		{object}	httputils.ValidationErrorResponse
//	@Failure		403		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/groups/{id}/users [delete]
func (gr *GroupRouteImpl) DeleteUsersFromGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	groupIDStr := httputils.GetPathValue(r, "id")
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
		httputils.HandleValidationError(w, err)
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		gr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := httputils.GetCurrentUser(r)

	err = gr.groupService.DeleteUsers(tx, *currentUser, groupID, request.UserIDs)
	if err != nil {
		httputils.HandleServiceError(w, err, db, gr.logger)
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Users deleted from group successfully"))
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
//	@Success		200	{object}	httputils.APIResponse[[]schemas.User]
//	@Router			/groups/{id}/users [get]
func (gr *GroupRouteImpl) GetGroupUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	groupIDStr := httputils.GetPathValue(r, "id")
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
		gr.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := httputils.GetCurrentUser(r)

	users, err := gr.groupService.GetUsers(tx, *currentUser, groupID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, gr.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, users)
}

func RegisterGroupRoutes(mux *mux.Router, groupRoute GroupRoute) {
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
}

func NewGroupRoute(groupService service.GroupService) GroupRoute {
	route := &GroupRouteImpl{
		groupService: groupService,
		logger:       utils.NewNamedLogger("groups"),
	}
	err := utils.ValidateStruct(*route)
	if err != nil {
		log.Panicf("GroupRoute struct is not valid: %s", err.Error())
	}
	return route
}
