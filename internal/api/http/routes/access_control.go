package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
)

// parseResourceType extracts resource_type from path and maps it to types.ResourceType.
// Supports contests and tasks; extend switch when adding more types.
func parseResourceType(r *http.Request) (types.ResourceType, bool) {
	resourceTypeStr := httputils.GetPathValue(r, "resource_type")
	switch resourceTypeStr {
	case "contests":
		return types.ResourceTypeContest, true
	case "tasks":
		return types.ResourceTypeTask, true
	default:
		return types.ResourceType(resourceTypeStr), false
	}
}

// parseResourceID extracts resource_id from path and parses it to int64.
func parseResourceID(r *http.Request) (int64, error) {
	resourceIDStr := httputils.GetPathValue(r, "resource_id")
	return strconv.ParseInt(resourceIDStr, 10, 64)
}

// parseUserID extracts user_id from path and parses it to int64.
func parseUserID(r *http.Request) (int64, error) {
	userIDStr := httputils.GetPathValue(r, "user_id")
	return strconv.ParseInt(userIDStr, 10, 64)
}

type AccessControlRoute interface {
	// Generic collaborators
	AddCollaborator(w http.ResponseWriter, r *http.Request)
	GetCollaborators(w http.ResponseWriter, r *http.Request)
	UpdateCollaborator(w http.ResponseWriter, r *http.Request)
	RemoveCollaborator(w http.ResponseWriter, r *http.Request)

	// Assignable users for a resource (teachers without existing access)
	GetAssignableUsers(w http.ResponseWriter, r *http.Request)
}

type accessControlRoute struct {
	accessControlService service.AccessControlService
	logger               *zap.SugaredLogger
}

// Contest Collaborators

// AddCollaborator godoc
//
//	@Tags			access-control
//	@Summary		Add a collaborator to a resource
//	@Description	Add a user as a collaborator to a resource (contest, task, etc.) with specified permissions (edit or manage). Only users with manage permission can add collaborators.
//	@Accept			json
//	@Produce		json
//	@Param			resource_type	path		string					true	"Resource type (contests|tasks)"
//	@Param			resource_id		path		int						true	"Resource ID"
//	@Param			body			body		schemas.AddCollaborator	true	"Collaborator details"
//	@Failure		400				{object}	httputils.ValidationErrorResponse
//	@Failure		403				{object}	httputils.APIError
//	@Failure		404				{object}	httputils.APIError
//	@Failure		500				{object}	httputils.APIError
//	@Success		200				{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/access-control/resources/{resource_type}/{resource_id}/collaborators [post]
func (ac *accessControlRoute) AddCollaborator(w http.ResponseWriter, r *http.Request) {
	// Generic handler retained
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	resourceType, ok := parseResourceType(r)
	if !ok {
		httputils.ReturnError(w, http.StatusBadRequest, "Unsupported resource type")
		return
	}

	resourceID, err := parseResourceID(r)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid resource ID")
		return
	}

	var request schemas.AddCollaborator
	err = httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		httputils.HandleValidationError(w, err)
		return
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	err = ac.accessControlService.AddCollaborator(db, currentUser, resourceType, resourceID, request.UserID, request.Permission)
	if err != nil {
		httputils.HandleServiceError(w, err, db, ac.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Collaborator added successfully"))
}

// GetCollaborators godoc
//
//	@Tags			access-control
//	@Summary		Get collaborators for a resource
//	@Description	Get all collaborators for a specific resource. Users with edit permission or higher can see collaborators.
//	@Produce		json
//	@Param			resource_type	path		string	true	"Resource type (contests|tasks)"
//	@Param			resource_id		path		int		true	"Resource ID"
//	@Failure		400				{object}	httputils.APIError
//	@Failure		403				{object}	httputils.APIError
//	@Failure		404				{object}	httputils.APIError
//	@Failure		500				{object}	httputils.APIError
//	@Success		200				{object}	httputils.APIResponse[[]schemas.Collaborator]
//	@Router			/access-control/resources/{resource_type}/{resource_id}/collaborators [get]
func (ac *accessControlRoute) GetCollaborators(w http.ResponseWriter, r *http.Request) {
	// Generic handler retained
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	resourceType, ok := parseResourceType(r)
	if !ok {
		httputils.ReturnError(w, http.StatusBadRequest, "Unsupported resource type")
		return
	}

	resourceID, err := parseResourceID(r)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid resource ID")
		return
	}

	db := httputils.GetDatabase(r)
	user := httputils.GetCurrentUser(r)

	collaborators, err := ac.accessControlService.GetCollaborators(db, user, resourceType, resourceID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, ac.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, collaborators)
}

// UpdateCollaborator godoc
//
//	@Tags			access-control
//	@Summary		Update a collaborator's permission for a resource
//	@Description	Update the permission level of a collaborator for a resource. Only users with manage permission can update collaborators.
//	@Accept			json
//	@Produce		json
//	@Param			resource_type	path		string						true	"Resource type (contests|tasks)"
//	@Param			resource_id		path		int							true	"Resource ID"
//	@Param			user_id			path		int							true	"User ID"
//	@Param			body			body		schemas.UpdateCollaborator	true	"New permission"
//	@Failure		400				{object}	httputils.ValidationErrorResponse
//	@Failure		403				{object}	httputils.APIError
//	@Failure		404				{object}	httputils.APIError
//	@Failure		500				{object}	httputils.APIError
//	@Success		200				{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/access-control/resources/{resource_type}/{resource_id}/collaborators/{user_id} [put]
func (ac *accessControlRoute) UpdateCollaborator(w http.ResponseWriter, r *http.Request) {
	// Generic handler retained
	if r.Method != http.MethodPut {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	resourceType, ok := parseResourceType(r)
	if !ok {
		httputils.ReturnError(w, http.StatusBadRequest, "Unsupported resource type")
		return
	}

	resourceID, err := parseResourceID(r)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid resource ID")
		return
	}

	userID, err := parseUserID(r)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var request schemas.UpdateCollaborator
	err = httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		httputils.HandleValidationError(w, err)
		return
	}

	db := httputils.GetDatabase(r)
	user := httputils.GetCurrentUser(r)

	err = ac.accessControlService.UpdateCollaborator(db, user, resourceType, resourceID, userID, request.Permission)
	if err != nil {
		httputils.HandleServiceError(w, err, db, ac.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Collaborator permission updated successfully"))
}

// RemoveCollaborator godoc
//
//	@Tags			access-control
//	@Summary		Remove a collaborator from a resource
//	@Description	Remove a user's collaborator access to a resource. Only users with manage permission can remove collaborators. Cannot remove the creator.
//	@Produce		json
//	@Param			resource_type	path		string	true	"Resource type (contests|tasks)"
//	@Param			resource_id		path		int		true	"Resource ID"
//	@Param			user_id			path		int		true	"User ID"
//	@Failure		400				{object}	httputils.APIError
//	@Failure		403				{object}	httputils.APIError
//	@Failure		404				{object}	httputils.APIError
//	@Failure		500				{object}	httputils.APIError
//	@Success		200				{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/access-control/resources/{resource_type}/{resource_id}/collaborators/{user_id} [delete]
func (ac *accessControlRoute) RemoveCollaborator(w http.ResponseWriter, r *http.Request) {
	// Generic handler retained
	if r.Method != http.MethodDelete {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	resourceType, ok := parseResourceType(r)
	if !ok {
		httputils.ReturnError(w, http.StatusBadRequest, "Unsupported resource type")
		return
	}

	resourceID, err := parseResourceID(r)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid resource ID")
		return
	}

	userID, err := parseUserID(r)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	db := httputils.GetDatabase(r)
	user := httputils.GetCurrentUser(r)

	err = ac.accessControlService.RemoveCollaborator(db, user, resourceType, resourceID, userID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, ac.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Collaborator removed successfully"))
}

// GetAssignableUsers godoc
//
//	@Tags			access-control
//	@Summary		Get assignable users for a resource
//	@Description	List users (teachers) who can be granted access for the given resource. Only users with manage permission can view assignable users. Returned users do not currently have any access entry for the resource.
//	@Produce		json
//	@Param			resource_type	path		string	true	"Resource type (contests|tasks)"
//	@Param			resource_id		path		int		true	"Resource ID"
//	@Param			limit			query		int		false	"Pagination limit"
//	@Param			offset			query		int		false	"Pagination offset"
//	@Param			sort			query		string	false	"Sort fields, e.g. 'id:asc,created_at:desc'"
//	@Failure		400				{object}	httputils.APIError
//	@Failure		403				{object}	httputils.APIError
//	@Failure		404				{object}	httputils.APIError
//	@Failure		500				{object}	httputils.APIError
//	@Success		200				{object}	httputils.APIResponse[schemas.PaginatedResult[[]schemas.User]]
//	@Router			/access-control/resources/{resource_type}/{resource_id}/assignable [get]
func (ac *accessControlRoute) GetAssignableUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	resourceType, ok := parseResourceType(r)
	if !ok {
		httputils.ReturnError(w, http.StatusBadRequest, "Unsupported resource type")
		return
	}

	resourceID, err := parseResourceID(r)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid resource ID")
		return
	}

	// Parse pagination query params
	q := r.URL.Query()
	queryParams, err := httputils.GetQueryParams(&q)
	if err != nil {
		httputils.HandleValidationError(w, err)
		return
	}
	pagination := httputils.ExtractPaginationParams(queryParams)

	db := httputils.GetDatabase(r)
	user := httputils.GetCurrentUser(r)

	resp, err := ac.accessControlService.GetAssignableUsers(db, user, resourceType, resourceID, pagination)
	if err != nil {
		httputils.HandleServiceError(w, err, db, ac.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, resp)
}

// Task Collaborators

func NewAccessControlRoute(accessControlService service.AccessControlService) AccessControlRoute {
	route := &accessControlRoute{
		accessControlService: accessControlService,
		logger:               utils.NewNamedLogger("access_control_route"),
	}
	if err := utils.ValidateStruct(*route); err != nil {
		panic(err)
	}
	return route
}

func RegisterAccessControlRoutes(mux *mux.Router, route AccessControlRoute) {
	// Configurable set of resource types; extend this slice to add more resources.
	resourceTypes := []string{"contests", "tasks"}
	rtPattern := strings.Join(resourceTypes, "|")

	// Generic collaborators collection route
	mux.HandleFunc("/resources/{resource_type:"+rtPattern+"}/{resource_id}/collaborators", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetCollaborators(w, r)
		case http.MethodPost:
			route.AddCollaborator(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// Generic single collaborator route
	mux.HandleFunc("/resources/{resource_type:"+rtPattern+"}/{resource_id}/collaborators/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			route.UpdateCollaborator(w, r)
		case http.MethodDelete:
			route.RemoveCollaborator(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// Assignable users route
	mux.HandleFunc("/resources/{resource_type:"+rtPattern+"}/{resource_id}/assignable", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetAssignableUsers(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
}
