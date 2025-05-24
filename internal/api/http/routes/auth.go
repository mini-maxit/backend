package routes

import (
	"errors"
	"log"
	"net/http"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	myerrors "github.com/mini-maxit/backend/package/errors"
	"github.com/mini-maxit/backend/package/service"
	"github.com/mini-maxit/backend/package/utils"
)

type AuthRoute interface {
	Login(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
	RefreshToken(w http.ResponseWriter, r *http.Request)
}

type AuthRouteImpl struct {
	userService service.UserService
	authService service.AuthService
}

// Login godoc
//
//	@Tags			auth
//	@Summary		Login a user
//	@Description	Logs in a user with email and password, returns JWT tokens
//	@Accept			json
//	@Produce		json
//	@Param			request	body		schemas.UserLoginRequest	true	"User Login Request"
//	@Failure		400		{object}	httputils.ApiError
//	@Failure		401		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Success		200		{object}	httputils.ApiResponse[schemas.JWTTokens]
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

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	tokens, err := ar.authService.Login(tx, request)
	if err != nil {
		db.Rollback()
		switch {
		case errors.Is(err, myerrors.ErrUserNotFound):
			httputils.ReturnError(w, http.StatusUnauthorized, "User not found. This email is not registered.")
		case errors.Is(err, myerrors.ErrInvalidCredentials):
			httputils.ReturnError(w, http.StatusUnauthorized, "Invalid credentials.")
		default:
			httputils.ReturnError(w, http.StatusInternalServerError, "Failed to login. "+err.Error())
		}
		return
	}

	err = db.Commit()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to commit transaction. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, tokens)
}

// Register godoc
//
//	@Tags			auth
//	@Summary		Register a user
//	@Description	Registers a user with name, surname, email, username and password, returns JWT tokens
//	@Accept			json
//	@Produce		json
//	@Param			request	body		schemas.UserRegisterRequest	true	"User Register Request"
//	@Failure		400		{object}	httputils.ApiError
//	@Failure		405		{object}	httputils.ApiError
//	@Failure		409		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Success		201		{object}	httputils.ApiResponse[schemas.JWTTokens]
//	@Router			/register [post]
func (ar *AuthRouteImpl) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.UserRegisterRequest
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

	tokens, err := ar.authService.Register(tx, request)
	switch err {
	case nil:
		break
	case myerrors.ErrUserAlreadyExists:
		db.Rollback()
		switch {
		case errors.Is(err, myerrors.ErrUserAlreadyExists):
			httputils.ReturnError(w, http.StatusConflict, err.Error())
		default:
			httputils.ReturnError(w, http.StatusInternalServerError, "Failed to register. "+err.Error())
		}
		return
	}

	err = db.Commit()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to commit transaction. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusCreated, tokens)
}

// RefreshToken godoc
//
//	@Tags			auth
//	@Summary		Refresh JWT tokens
//	@Description	Refreshes JWT tokens using a valid refresh token
//	@Accept			json
//	@Produce		json
//	@Param			request	body		schemas.RefreshTokenRequest	true	"Refresh Token Request"
//	@Failure		400		{object}	httputils.ApiError
//	@Failure		401		{object}	httputils.ApiError
//	@Failure		405		{object}	httputils.ApiError
//	@Failure		500		{object}	httputils.ApiError
//	@Success		200		{object}	httputils.ApiResponse[schemas.JWTTokens]
//	@Router			/auth/refresh [post]
func (ar *AuthRouteImpl) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.RefreshTokenRequest
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

	tokens, err := ar.authService.RefreshTokens(tx, request)
	if err != nil {
		db.Rollback()
		if err == service.ErrInvalidToken || err == service.ErrTokenExpired || err == service.ErrInvalidTokenType {
			httputils.ReturnError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
			return
		}
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to refresh tokens. "+err.Error())
		return
	}

	err = db.Commit()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to commit transaction. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, tokens)
}

func NewAuthRoute(userService service.UserService, authService service.AuthService) AuthRoute {
	route := &AuthRouteImpl{
		userService: userService,
		authService: authService,
	}
	err := utils.ValidateStruct(*route)
	if err != nil {
		log.Panicf("AuthRoute struct is not valid: %s", err.Error())
	}
	return route
}
