package routes

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
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
//	@Success		200		{object}	httputils.APIResponse[schemas.PaginatedResult[[]schemas.User]]
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Router			/users/ [get]
func (u *UserRouteImpl) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		u.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	queryParams := r.Context().Value(httputils.QueryParamsKey).(map[string]any)
	paginationParams := httputils.ExtractPaginationParams(queryParams)
	result, err := u.userService.GetAll(tx, paginationParams)
	if err != nil {
		db.Rollback()
		u.logger.Errorw("Failed to get users", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "User service temporarily unavailable")
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

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		u.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	user, err := u.userService.Get(tx, userID)
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

	var request schemas.UserEdit

	userIDStr := httputils.GetPathValue(r, "id")

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid userID")
		return
	}

	err = httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			httputils.ReturnValidationError(w, valErrs)
			return
		}
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		u.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := httputils.GetCurrentUser(r)

	err = u.userService.Edit(tx, *currentUser, userID, &request)
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
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			httputils.ReturnValidationError(w, valErrs)
			return
		}
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		u.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	currentUser := httputils.GetCurrentUser(r)

	err = u.userService.ChangePassword(tx, *currentUser, userID, request)
	if err != nil {
		httputils.HandleServiceError(w, err, db, u.logger)
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Password changed successfully"))
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

	userVal := r.Context().Value(httputils.UserKey)
	currentUser, ok := userVal.(schemas.User)
	if !ok {
		httputils.ReturnError(w, http.StatusInternalServerError, "Could not retrieve user. Verify that you are logged in.")
		return
	}
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
	mux.HandleFunc("/", route.GetAllUsers)
	mux.HandleFunc("/me", route.GetMe)
	mux.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			route.GetUserByID(w, r)
		case http.MethodPatch:
			route.EditUser(w, r)
		default:
			httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	mux.HandleFunc("/{id}/password", route.ChangePassword)
}
