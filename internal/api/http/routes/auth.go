package routes

import (
	"encoding/json"
	"net/http"

	"github.com/mini-maxit/backend/internal/api/http/middleware"
	"github.com/mini-maxit/backend/internal/api/http/utils"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
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
//	@Failure		500		{object}	utils.ApiResponse[schemas.UserLoginErrorResponse]
//	@Success		200		{object}	utils.ApiResponse[schemas.UserLoginSuccessResponse]
//	@Router			/login [post]
func (ar *AuthRouteImpl) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.UserLoginRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
	}

	session, err := ar.authService.Login(tx, request)
	if err != nil {
		db.Rollback()
		if err == service.ErrInvalidCredentials {
			utils.ReturnError(w, http.StatusUnauthorized, "Invalid credentials. Verify your email and password and try again.")
			return
		}
		utils.ReturnError(w, http.StatusInternalServerError, "Failed to login. "+err.Error())
		return
	}
	utils.ReturnSuccess(w, http.StatusOK, session)
}

// Register godoc
//
//	@Tags			auth
//	@Summary		Register a user
//	@Description	Registers a user with name, surname, email, username and password
//	@Accept			json
//	@Produce		json
//	@Param			request	body		schemas.UserRegisterRequest	true	"User Register Request"
//	@Failure		500		{object}	utils.ApiResponse[schemas.UserRegisterErrorResponse]
//	@Success		201		{object}	utils.ApiResponse[schemas.UserRegisterSuccessResponse]
//	@Router			/register [post]
func (ar *AuthRouteImpl) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.UserRegisterRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}
	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
	}

	session, err := ar.authService.Register(tx, request)
	switch err {
	case nil:
		break
	case service.ErrUserAlreadyExists:
		db.Rollback()
		utils.ReturnError(w, http.StatusBadRequest, err.Error())
		return
	default:
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, "Failed to register. "+err.Error())
		return
	}

	utils.ReturnSuccess(w, http.StatusCreated, session)
}

func NewAuthRoute(userService service.UserService, authService service.AuthService) AuthRoute {
	return &AuthRouteImpl{
		userService: userService,
		authService: authService,
	}
}
