package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/api/http/middleware"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/service"
)

type UserRoute interface {
	GetAllUsers(w http.ResponseWriter, r *http.Request)
	GetUserById(w http.ResponseWriter, r *http.Request)
	GetUserByEmail(w http.ResponseWriter, r *http.Request)
	EditUser(w http.ResponseWriter, r *http.Request)
	CreateUsers(w http.ResponseWriter, r *http.Request)
	ChangePassword(w http.ResponseWriter, r *http.Request)
}

type UserRouteImpl struct {
	userService service.UserService
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
//	@Success		200		{object}	httputils.ApiResponse[schemas.User]
//	@Failure		405		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Router			/user/ [get]
func (u *UserRouteImpl) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error connecting to database. %s", err.Error()))
		return
	}

	queryParams := r.Context().Value(middleware.QueryParamsKey).(map[string]interface{})
	users, err := u.userService.GetAllUsers(tx, queryParams)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error getting users. %s", err.Error()))
		return
	}

	if users == nil {
		users = []schemas.User{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, users)
}

// GetUserById godoc
//
//	@Tags			user
//	@Summary		Get user by ID
//	@Description	Get user by ID
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	httputils.ApiResponse[schemas.User]
//	@Failure		400	{object}	httputils.ApiError
//	@Failure		404	{object}	httputils.ApiError
//	@Failure		405	{object}	httputils.ApiError
//	@Failure		500	{object}	httputils.ApiError
//	@Router			/user/{id} [get]
func (u *UserRouteImpl) GetUserById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userIdStr := r.PathValue("id")

	if userIdStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "UserId cannot be empty")
		return
	}

	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, fmt.Sprintf("Invalid userId: %s", err.Error()))
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error connecting to database. %s", err.Error()))
		return
	}

	user, err := u.userService.GetUserById(tx, userId)
	if err != nil {
		db.Rollback()
		if err == errors.ErrUserNotFound {
			httputils.ReturnError(w, http.StatusNotFound, "User not found")
			return
		}
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error fetching user: %s", err.Error()))
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, user)
}

// GetUserByEmail godoc
//
//	@Tags			user
//	@Summary		Get user by email
//	@Description	Get user by email
//	@Produce		json
//	@Param			email	query		string	true	"User email"
//	@Success		200		{object}	httputils.ApiResponse[schemas.User]
//	@Failure		400		{object}	httputils.ApiError
//	@Failure		404		{object}	httputils.ApiError
//	@Failure		405		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Router			/user/email [get]
func (u *UserRouteImpl) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	queryParams := r.Context().Value(middleware.QueryParamsKey).(map[string]interface{})
	email := queryParams["email"].(string)
	if email == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Email query cannot be empty")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error connecting to database. %s", err.Error()))
		return
	}

	user, err := u.userService.GetUserByEmail(tx, email)
	if err != nil {
		db.Rollback()
		if err == errors.ErrUserNotFound {
			httputils.ReturnError(w, http.StatusNotFound, "User not found")
			return
		}
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error getting user. %s", err.Error()))
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, user)
}

// EditUser godoc
//
// @Tags			user
// @Summary		Edit user
// @Description	Edit user
// @Accept			json
// @Produce		json
// @Param			id		path		int					true	"User ID"
// @Param			body	body		schemas.UserEdit	true	"User edit object"
// @Success		200		{object}	httputils.ApiResponse[string]
// @Failure		400		{object}	httputils.ApiError
// @Failure		403		{object}	httputils.ApiError
// @Failure		404		{object}	httputils.ApiError
// @Failure		405		{object}	httputils.ApiError
// @Failure		500		{object}	httputils.ApiError
// @Router			/user/{id} [patch]
func (u *UserRouteImpl) EditUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.UserEdit

	userIdStr := r.PathValue("id")

	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid userId")
		return
	}

	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error connecting to database. %s", err.Error()))
		return
	}

	currentUser := r.Context().Value(middleware.UserKey).(schemas.User)

	err = u.userService.EditUser(tx, currentUser, userId, &request)
	if err != nil {
		db.Rollback()
		if err == errors.ErrNotAllowed {
			httputils.ReturnError(w, http.StatusForbidden, "You are not allowed to change user role")
			return
		}
		if err == errors.ErrUserNotFound {
			httputils.ReturnError(w, http.StatusNotFound, "User not found")
			return
		}
		if err == errors.ErrNotAuthorized {
			httputils.ReturnError(w, http.StatusForbidden, "You are not authorized to edit this user")
			return
		}
		httputils.ReturnError(w, http.StatusInternalServerError, "Error ocured during editing. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, "Update successfull")
}

// ChangePassword godoc
//
// @Tags			user
// @Summary		Change user password
// @Description	Change user password
// @Accept			json
// @Produce		json
// @Param			id		path		int							true	"User ID"
// @Param			body	body		schemas.UserChangePassword	true	"User change password object"
// @Success		200		{object}	httputils.ApiResponse[string]
// @Failure		400		{object}	httputils.ApiError
// @Failure		403		{object}	httputils.ApiError
// @Failure		404		{object}	httputils.ApiError
// @Failure		405		{object}	httputils.ApiError
// @Failure		500		{object}	httputils.ApiError
// @Router			/user/{id}/password [patch]
func (u *UserRouteImpl) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userIdStr := r.PathValue("id")
	if userIdStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "UserId cannot be empty")
		return
	}
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid userId")
		return
	}

	request := &schemas.UserChangePassword{}
	err = json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error connecting to database. %s", err.Error()))
		return
	}

	currentUser := r.Context().Value(middleware.UserKey).(schemas.User)

	err = u.userService.ChangePassword(tx, currentUser, userId, request)
	if err != nil {
		db.Rollback()
		if reflect.TypeOf(err) == reflect.TypeOf(validator.ValidationErrors{}) {
			httputils.ReturnError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err == errors.ErrNotAllowed {
			httputils.ReturnError(w, http.StatusForbidden, "You are not allowed to change user role")
			return
		}
		if err == errors.ErrUserNotFound {
			httputils.ReturnError(w, http.StatusNotFound, "User not found")
			return
		}
		if err == errors.ErrNotAuthorized {
			httputils.ReturnError(w, http.StatusForbidden, "You are not authorized to edit this user")
			return
		}
		if err == errors.ErrInvalidCredentials {
			httputils.ReturnError(w, http.StatusBadRequest, "Invalid old password")
			return
		}
		if err == errors.ErrInvalidData {
			httputils.ReturnError(w, http.StatusBadRequest, "New password and confirm password do not match")
			return
		}
		httputils.ReturnError(w, http.StatusInternalServerError, "Error ocured during editing. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, "Password changed successfully")
}

func (u *UserRouteImpl) CreateUsers(w http.ResponseWriter, r *http.Request) {
	// this funcion allows admin to ctreate new users with their email and given role
	// the users will be created with a default password and will be required to change it on first login

	httputils.ReturnError(w, http.StatusNotImplemented, "Not implemented")
}

func NewUserRoute(userService service.UserService) UserRoute {
	return &UserRouteImpl{userService: userService}
}

func RegisterUserRoutes(mux *http.ServeMux, route UserRoute) {
	mux.HandleFunc("/", route.GetAllUsers)
	mux.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			route.GetUserById(w, r)
		} else if r.Method == http.MethodPatch {
			route.EditUser(w, r)
		}
	})
	mux.HandleFunc("/email", route.GetUserByEmail)
	mux.HandleFunc("/{id}/password", route.ChangePassword)
}
