package routes

import (
	"log"
	"net/http"
	"slices"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
)

type UserRoute interface {
	GetAllUsers(w http.ResponseWriter, r *http.Request)
	GetMe(w http.ResponseWriter, r *http.Request)
	GetUserByID(w http.ResponseWriter, r *http.Request)
	EditUser(w http.ResponseWriter, r *http.Request)
	CreateUsers(w http.ResponseWriter, r *http.Request)
	ChangePassword(w http.ResponseWriter, r *http.Request)
	ResetPassword(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)
	RestoreUser(w http.ResponseWriter, r *http.Request)
}

type UserRouteImpl struct {
	userService service.UserService
	logger      *zap.SugaredLogger
}

// GetAllUsers godoc
//
//	@Tags			user
//	@Summary		Get all users
//	@Description	Get all users
//	@Produce		json
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Param			search	query		string	false	"Search users by name, surname or email"
//	@Param			role	query		string	false	"Filter by role (student, teacher, admin)"
//	@Param			deleted	query		bool	false	"List only soft-deleted users"
//	@Success		200		{object}	httputils.APIResponse[schemas.PaginatedResult[[]schemas.User]]
//	@Failure		400		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/users [get]
func (u *UserRouteImpl) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := httputils.GetDatabase(r)
	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	paginationParams := httputils.ExtractPaginationParams(queryParams)
	search, _ := queryParams["search"].(string)
	role, _ := queryParams["role"].(string)
	if role != "" && !slices.Contains([]types.UserRole{types.UserRoleStudent, types.UserRoleTeacher, types.UserRoleAdmin}, types.UserRole(role)) {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid role")
		return
	}
	deleted := r.URL.Query().Get("deleted") == "true"
	result, err := u.userService.GetAll(db, paginationParams, search, role, deleted)
	if err != nil {
		httputils.HandleServiceError(w, err, db, u.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, result)
}

// GetUserByID godoc
//
//	@Tags			user
//	@Summary		Get user by ID
//	@Description	Get user by ID
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	httputils.APIResponse[schemas.User]
//	@Failure		400	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Router			/users/{id} [get]
func (u *UserRouteImpl) GetUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userIDStr := httputils.GetPathValue(r, "id")

	if userIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "UserID cannot be empty")
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid userID: "+err.Error())
		return
	}

	db := httputils.GetDatabase(r)
	user, err := u.userService.Get(db, userID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, u.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, user)
}

// EditUser godoc
//
//	@Tags			user
//	@Summary		Edit user
//	@Description	Edit user
//	@Produce		json
//	@Param			id		path		int					true	"User ID"
//	@Param			request	body		schemas.UserEdit	true	"User Edit Request"
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Failure		400		{object}	httputils.ValidationErrorResponse
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/users/{id} [patch].
func (u *UserRouteImpl) EditUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userIDStr := httputils.GetPathValue(r, "id")

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid userID")
		return
	}

	var request schemas.UserEdit
	err = httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		httputils.HandleValidationError(w, err)
		return
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	err = u.userService.Edit(db, *currentUser, userID, &request)
	if err != nil {
		httputils.HandleServiceError(w, err, db, u.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Update successful"))
}

// ChangePassword godoc
//
//	@Tags			user
//	@Summary		Change user password
//	@Description	Change user password
//	@Produce		json
//	@Param			id		path		int							true	"User ID"
//	@Param			request	body		schemas.UserChangePassword	true	"User Change Password Request"
//	@Success		200		{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Failure		400		{object}	httputils.ValidationErrorResponse
//	@Failure		403		{object}	httputils.APIError
//	@Failure		404		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/users/{id}/password [patch].
func (u *UserRouteImpl) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userIDStr := httputils.GetPathValue(r, "id")
	if userIDStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "UserID cannot be empty")
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid userID")
		return
	}

	request := &schemas.UserChangePassword{}
	err = httputils.ShouldBindJSON(r.Body, request)
	if err != nil {
		httputils.HandleValidationError(w, err)
		return
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	err = u.userService.ChangePassword(db, *currentUser, userID, request)
	if err != nil {
		httputils.HandleServiceError(w, err, db, u.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Password changed successfully"))
}

// ResetPassword godoc
//
//	@Tags			user
//	@Summary		Reset user password (admin)
//	@Description	Generates a new random password for the user and returns it once
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	httputils.APIResponse[schemas.UserResetPasswordResponse]
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Router			/users/{id}/password/reset [post]
func (u *UserRouteImpl) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, err := strconv.ParseInt(httputils.GetPathValue(r, "id"), 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid userID")
		return
	}

	db := httputils.GetDatabase(r)
	currentUser := httputils.GetCurrentUser(r)

	password, err := u.userService.ResetPassword(db, *currentUser, userID)
	if err != nil {
		httputils.HandleServiceError(w, err, db, u.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, schemas.UserResetPasswordResponse{NewPassword: password})
}

// DeleteUser godoc
//
//	@Tags			user
//	@Summary		Soft-delete user (admin)
//	@Description	Marks the user as deleted; the user can no longer log in
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Router			/users/{id} [delete]
func (u *UserRouteImpl) DeleteUser(w http.ResponseWriter, r *http.Request) {
	u.setDeleted(w, r, u.userService.Delete, "User deleted successfully")
}

// RestoreUser godoc
//
//	@Tags			user
//	@Summary		Restore soft-deleted user (admin)
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Failure		400	{object}	httputils.APIError
//	@Failure		403	{object}	httputils.APIError
//	@Failure		404	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Router			/users/{id}/restore [post]
func (u *UserRouteImpl) RestoreUser(w http.ResponseWriter, r *http.Request) {
	u.setDeleted(w, r, u.userService.Restore, "User restored successfully")
}

func (u *UserRouteImpl) setDeleted(
	w http.ResponseWriter,
	r *http.Request,
	action func(database.Database, schemas.User, int64) error,
	message string,
) {
	userID, err := strconv.ParseInt(httputils.GetPathValue(r, "id"), 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid userID")
		return
	}
	db := httputils.GetDatabase(r)
	if err := action(db, *httputils.GetCurrentUser(r), userID); err != nil {
		httputils.HandleServiceError(w, err, db, u.logger)
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse(message))
}

func (u *UserRouteImpl) CreateUsers(w http.ResponseWriter, _ *http.Request) {
	// this funcion allows admin to ctreate new users with their email and given role
	// the users will be created with a default password and will be required to change it on first login

	httputils.ReturnError(w, http.StatusNotImplemented, "Not implemented")
}

// GetMe godoc
//
//	@Tags			user
//	@Summary		Get current user
//	@Description	Get current user
//	@Produce		json
//	@Success		200	{object}	httputils.APIResponse[schemas.User]
//	@Failure		405	{object}	httputils.APIError
//	@Router			/users/me [get]
func (u *UserRouteImpl) GetMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	currentUser := httputils.GetCurrentUser(r)

	httputils.ReturnSuccess(w, http.StatusOK, currentUser)
}

func NewUserRoute(userService service.UserService) UserRoute {
	route := &UserRouteImpl{
		userService: userService,
		logger:      utils.NewNamedLogger("users"),
	}
	err := utils.ValidateStruct(*route)
	if err != nil {
		log.Panicf("UserRoute struct is not valid: %s", err.Error())
	}
	return route
}

func RegisterUserRoutes(mux *mux.Router, route UserRoute) {
	mux.HandleFunc("/users", route.GetAllUsers)
	mux.HandleFunc("/users/me", route.GetMe)
	mux.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetUserByID(w, r)
		case http.MethodPatch:
			route.EditUser(w, r)
		case http.MethodDelete:
			route.DeleteUser(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	mux.HandleFunc("/users/{id}/password", route.ChangePassword)
	mux.HandleFunc("/users/{id}/password/reset", route.ResetPassword)
	mux.HandleFunc("/users/{id}/restore", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		route.RestoreUser(w, r)
	})
}
