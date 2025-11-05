package routes

import (
	"errors"
	"log"
	"net/http"

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

type UserRoute interface {
	GetMe(w http.ResponseWriter, r *http.Request)
	ChangeMyPassword(w http.ResponseWriter, r *http.Request)
}

type UserRouteImpl struct {
	userService service.UserService
	logger      *zap.SugaredLogger
}

// ChangeMyPassword godoc
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
func (u *UserRouteImpl) ChangeMyPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	request := &schemas.UserChangePassword{}
	err := httputils.ShouldBindJSON(r.Body, request)
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

	currentUser := r.Context().Value(httputils.UserKey).(schemas.User)

	err = u.userService.ChangePassword(tx, currentUser, currentUser.ID, request)
	if err != nil {
		db.Rollback()

		// Define mapping of myerrors to HTTP status codes and messages
		errorResponses := map[error]struct {
			code    int
			message string
		}{
			myerrors.ErrNotAllowed:         {http.StatusForbidden, "You are not allowed to change user role"},
			myerrors.ErrUserNotFound:       {http.StatusNotFound, "User not found"},
			myerrors.ErrNotAuthorized:      {http.StatusForbidden, "You are not authorized to edit this user"},
			myerrors.ErrInvalidCredentials: {http.StatusBadRequest, "Invalid old password"},
			myerrors.ErrInvalidData:        {http.StatusBadRequest, "New password and confirm password do not match"},
		}

		// Special handling for validation errors
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			httputils.ReturnError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Check if err exists in the errorResponses map
		if resp, exists := errorResponses[err]; exists {
			httputils.ReturnError(w, resp.code, resp.message)
			return
		}

		// Default case
		u.logger.Errorw("Failed to change password", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Password change service temporarily unavailable")
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Password changed successfully"))
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
	mux.HandleFunc("/me", route.GetMe)
	mux.HandleFunc("/me/password", route.ChangeMyPassword)
}
