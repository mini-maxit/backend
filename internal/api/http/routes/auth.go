package routes

import (
	"errors"
	"log"
	"net/http"
	"time"

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

// AuthResponse represents the response for auth endpoints (excludes refresh token for security)
type AuthResponse struct {
	AccessToken string    `json:"accessToken"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

// newAuthResponse creates an AuthResponse from JWTTokens, excluding the refresh token
func newAuthResponse(tokens *schemas.JWTTokens) AuthResponse {
	return AuthResponse{
		AccessToken: tokens.AccessToken,
		ExpiresAt:   tokens.ExpiresAt,
	}
}

// setRefreshTokenCookie sets the refresh token as an httpOnly cookie
func setRefreshTokenCookie(w http.ResponseWriter, path, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     path,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})
}

type AuthRoute interface {
	Login(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
	RefreshToken(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
}

type AuthRouteImpl struct {
	refreshTokenPath string
	userService      service.UserService
	authService      service.AuthService
	logger           *zap.SugaredLogger
}

// Login godoc
//
//	@Tags			auth
//	@Summary		Login a user
//	@Description	Logs in a user with email and password, returns JWT tokens
//	@Accept			json
//	@Produce		json
//	@Param			request	body		schemas.UserLoginRequest	true	"User Login Request"
//	@Failure		400		{object}	httputils.ValidationErrorResponse
//	@Failure		401		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[AuthResponse]
//	@Router			/auth/login [post]
func (ar *AuthRouteImpl) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.UserLoginRequest
	err := httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			httputils.ReturnValidationError(w, valErrs)
			return
		}
		httputils.ReturnError(w, http.StatusBadRequest, "Could not validate request data.")
		return
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		ar.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
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
			ar.logger.Errorw("Login failed", "error", err)
			httputils.ReturnError(w, http.StatusInternalServerError, "Authentication service temporarily unavailable")
		}
		return
	}

	setRefreshTokenCookie(w, ar.refreshTokenPath, tokens.RefreshToken)

	authResponse := newAuthResponse(tokens)

	httputils.ReturnSuccess(w, http.StatusOK, authResponse)
}

// Register godoc
//
//	@Tags			auth
//	@Summary		Register a user
//	@Description	Registers a user with name, surname, email, username and password, returns JWT tokens
//	@Accept			json
//	@Produce		json
//	@Param			request	body		schemas.UserRegisterRequest	true	"User Register Request"
//	@Failure		400		{object}	httputils.ValidationErrorResponse
//	@Failure		405		{object}	httputils.APIError
//	@Failure		409		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		201		{object}	httputils.APIResponse[AuthResponse]
//	@Router			/auth/register [post]
func (ar *AuthRouteImpl) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.UserRegisterRequest
	err := httputils.ShouldBindJSON(r.Body, &request)
	if err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			httputils.ReturnValidationError(w, valErrs)
			return
		}
		httputils.ReturnError(w, http.StatusBadRequest, "Could not validate request data.")
		return
	}
	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		ar.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	tokens, err := ar.authService.Register(tx, request)
	switch {
	case err == nil:
		break
	case errors.Is(err, myerrors.ErrUserAlreadyExists):
		db.Rollback()
		httputils.ReturnError(w, http.StatusConflict, err.Error())
		return
	default:
		db.Rollback()
		ar.logger.Errorw("Registration failed", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Registration service temporarily unavailable")
		return
	}

	setRefreshTokenCookie(w, ar.refreshTokenPath, tokens.RefreshToken)

	authResponse := newAuthResponse(tokens)

	httputils.ReturnSuccess(w, http.StatusCreated, authResponse)
}

// RefreshToken godoc
//
//	@Tags			auth
//	@Summary		Refresh JWT tokens
//	@Description	Refreshes JWT tokens using a valid refresh token from cookie
//	@Produce		json
//	@Failure		400	{object}	httputils.APIError
//	@Failure		401	{object}	httputils.APIError
//	@Failure		405	{object}	httputils.APIError
//	@Failure		500	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[AuthResponse]
//	@Router			/auth/refresh [post]
func (ar *AuthRouteImpl) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	refreshTokenCookie, err := r.Cookie("refresh_token")
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Refresh token cookie not found")
		return
	}

	request := schemas.RefreshTokenRequest{
		RefreshToken: refreshTokenCookie.Value,
	}

	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		ar.logger.Errorw("Failed to begin database transaction", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Database connection error")
		return
	}

	tokens, err := ar.authService.RefreshTokens(tx, request)
	if err != nil {
		db.Rollback()
		if errors.Is(err, myerrors.ErrInvalidToken) ||
			errors.Is(err, myerrors.ErrTokenExpired) ||
			errors.Is(err, myerrors.ErrInvalidTokenType) {
			httputils.ReturnError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
			return
		}
		ar.logger.Errorw("Token refresh failed", "error", err)
		httputils.ReturnError(w, http.StatusInternalServerError, "Token refresh service temporarily unavailable")
		return
	}

	setRefreshTokenCookie(w, ar.refreshTokenPath, tokens.RefreshToken)

	authResponse := newAuthResponse(tokens)

	httputils.ReturnSuccess(w, http.StatusOK, authResponse)
}

// Logout godoc
//
//	@Tags			auth
//	@Summary		Logout a user
//	@Description	Logs out a user by clearing the refresh token cookie
//	@Produce		json
//	@Failure		405	{object}	httputils.APIError
//	@Success		200	{object}	httputils.APIResponse[httputils.MessageResponse]
//	@Router			/auth/logout [post]
func (ar *AuthRouteImpl) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Clear the refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Path:     ar.refreshTokenPath,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
	})

	httputils.ReturnSuccess(w, http.StatusOK, httputils.NewMessageResponse("Logged out successfully"))
}

func NewAuthRoute(userService service.UserService, authService service.AuthService, refreshTokenPath string) AuthRoute {
	route := &AuthRouteImpl{
		refreshTokenPath: refreshTokenPath,
		userService:      userService,
		authService:      authService,
		logger:           utils.NewNamedLogger("auth"),
	}
	err := utils.ValidateStruct(*route)
	if err != nil {
		log.Panicf("AuthRoute struct is not valid: %s", err.Error())
	}
	return route
}
func RegisterAuthRoute(mux *mux.Router, route AuthRoute) {
	mux.HandleFunc("/login", route.Login)
	mux.HandleFunc("/register", route.Register)
	mux.HandleFunc("/refresh", route.RefreshToken)
	mux.HandleFunc("/logout", route.Logout)
}
