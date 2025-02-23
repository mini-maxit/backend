package routes

import (
	"encoding/json"
	"net/http"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/api/http/middleware"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/service"
)

type AuthRoute interface {
	Login(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
}

type AuthRouteImpl struct {
	userService service.UserService
	authService service.AuthService
}

// Login godoc
//
//	@Tags			auth
//	@Summary		Login a user
//	@Description	Logs in a user with email and password
//	@Accept			json
//	@Produce		json
//	@Param			request	body		schemas.UserLoginRequest	true	"User Login Request"
//	@Failure		400		{object}	httputils.ApiError
//	@Failure		401		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Success		200		{object}	httputils.ApiResponse[schemas.Session]
//	@Router			/login [post]
func (ar *AuthRouteImpl) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.UserLoginRequest
	err := httputils.ShouldBindJSON(r.Body, &request)
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

	session, err := ar.authService.Login(tx, request)
	if err != nil {
		db.Rollback()
		if err == errors.ErrUserNotFound {
			httputils.ReturnError(w, http.StatusUnauthorized, "User not found. This email is not registerd.")
			return
		}
		if err == errors.ErrInvalidCredentials {
			httputils.ReturnError(w, http.StatusUnauthorized, "Invalid credentials. Verify your email and password and try again.")
			return
		}
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to login. "+err.Error())
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, session)
}

// Register godoc
//
//	@Tags			auth
//	@Summary		Register a user
//	@Description	Registers a user with name, surname, email, username and password
//	@Accept			json
//	@Produce		json
//	@Param			request	body		schemas.UserRegisterRequest	true	"User Register Request"
//	@Failure		400		{object}	httputils.ApiError
//	@Failure		405		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Success		201		{object}	httputils.ApiResponse[schemas.Session]
//	@Router			/register [post]
func (ar *AuthRouteImpl) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.UserRegisterRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}
	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
	}

	session, err := ar.authService.Register(tx, request)
	switch err {
	case nil:
		break
	case errors.ErrUserAlreadyExists:
		db.Rollback()
		httputils.ReturnError(w, http.StatusBadRequest, err.Error())
		return
	default:
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to register. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusCreated, session)
}

func NewAuthRoute(userService service.UserService, authService service.AuthService) AuthRoute {
	return &AuthRouteImpl{
		userService: userService,
		authService: authService,
	}
}
