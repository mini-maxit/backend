package routes

import (
	"encoding/json"
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

type SessionRoute interface {
	CreateSession(w http.ResponseWriter, r *http.Request)
	ValidateSession(w http.ResponseWriter, r *http.Request)
	InvalidateSession(w http.ResponseWriter, r *http.Request)
}

type SessionRouteImpl struct {
	sessionService service.SessionService
}

func (sr *SessionRouteImpl) CreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request struct {
		UserID int64 `json:"user_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
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

	session, err := sr.sessionService.Create(tx, request.UserID)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to create session. "+err.Error())
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, session)
}

// ValidateSession godoc
//
//	@Tags			session
//	@Summary		Validate a session
//	@Description	Validates a session token
//	@Produce		json
//	@Param			Session	header		string	true	"Session Token"
//	@Failure		400		{object}	httputils.APIError
//	@Failure		401		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[[]schemas.Task]
//	@Router			/session/validate [get]
func (sr *SessionRouteImpl) ValidateSession(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.Header.Get("Session")
	if sessionToken == "" {
		httputils.ReturnError(w, http.StatusUnauthorized, "Session token is empty")
		return
	}
	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	var validateSession schemas.ValidateSessionResponse
	validateSession, err = sr.sessionService.Validate(tx, sessionToken)
	if err != nil {
		db.Rollback()
		switch {
		case errors.Is(err, myerrors.ErrSessionNotFound):
			httputils.ReturnError(w, http.StatusUnauthorized, "Session not found")
			return
		case errors.Is(err, myerrors.ErrSessionExpired):
			httputils.ReturnError(w, http.StatusUnauthorized, "Session expired")
			return
		case errors.Is(err, myerrors.ErrSessionUserNotFound):
			httputils.ReturnError(w, http.StatusUnauthorized, "User associated with session not found")
			return
		}
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to validate session. "+err.Error())
		return
	}

	if validateSession.Valid {
		httputils.ReturnSuccess(w, http.StatusOK, validateSession)
		return
	}
	db.Rollback()
	httputils.ReturnError(w,
		http.StatusUnauthorized,
		"Session was invalid, without any error. If you see this, please report it to the developers.",
	)
}

// InvalidateSession godoc
//
//	@Tags			session
//	@Summary		Invalidate a session
//	@Description	Invalidates a session token
//	@Produce		json
//	@Param			Session	header		string	true	"Session Token"
//	@Failure		400		{object}	httputils.APIError
//	@Failure		401		{object}	httputils.APIError
//	@Failure		405		{object}	httputils.APIError
//	@Failure		500		{object}	httputils.APIError
//	@Success		200		{object}	httputils.APIResponse[string]
//	@Router			/session/invalidate [post]
func (sr *SessionRouteImpl) InvalidateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	sessionToken := r.Header.Get("Session")
	if sessionToken == "" {
		httputils.ReturnError(w, http.StatusUnauthorized, "Session token is empty")
		return
	}
	db := r.Context().Value(httputils.DatabaseKey).(database.Database)
	tx, err := db.BeginTransaction()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	err = sr.sessionService.Invalidate(tx, sessionToken)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, "Failed to invalidate session. "+err.Error())
		return
	}
	httputils.ReturnSuccess(w, http.StatusOK, "Session invalidated")
}

func NewSessionRoute(sessionService service.SessionService) SessionRoute {
	route := &SessionRouteImpl{
		sessionService: sessionService,
	}

	err := utils.ValidateStruct(*route)
	if err != nil {
		log.Panicf("SessionRoute struct is not valid: %s", err.Error())
	}

	return route
}
