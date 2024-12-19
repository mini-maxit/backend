package routes

import (
	"encoding/json"
	"net/http"

	"github.com/mini-maxit/backend/internal/api/http/utils"
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

func (ar *AuthRouteImpl) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ReturnError(w, http.StatusMethodNotAllowed, utils.CodeMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.UserLoginRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, utils.CodeBadRequest, "Invalid request body. "+err.Error())
		return
	}

	_, err = ar.userService.GetUserByEmail(request.Email)
	if err != nil {
		utils.ReturnError(w, http.StatusUnauthorized, utils.CodeUnauthorized, "Given email does not exist. Verify your email and try again.")
		return
	}

	session, err := ar.authService.Login(request.Email, request.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			utils.ReturnError(w, http.StatusUnauthorized, utils.CodeUnauthorized, "Invalid credentials. Verify your email and password and try again.")
			return
		}
		utils.ReturnError(w, http.StatusInternalServerError, utils.CodeInternalServerError, "Failed to login. "+err.Error())
		return
	}
	utils.ReturnSuccess(w, http.StatusOK, session)
}

func (ar *AuthRouteImpl) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ReturnError(w, http.StatusMethodNotAllowed, utils.CodeMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.UserRegisterRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, utils.CodeBadRequest, "Invalid request body. "+err.Error())
		return
	}

	session, err := ar.authService.Register(request)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, utils.CodeInternalServerError, "Failed to register. "+err.Error())
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
