package routes

import (
	"encoding/json"
	"net/http"

	"github.com/mini-maxit/backend/internal/api/http/middleware"
	"github.com/mini-maxit/backend/internal/api/http/utils"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/service"
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
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request struct {
		UserId int64 `json:"user_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}
	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	session, err := sr.sessionService.CreateSession(tx, request.UserId)
	if err != nil {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, "Failed to create session. "+err.Error())
		return
	}
	utils.ReturnSuccess(w, http.StatusOK, session)
}

func (sr *SessionRouteImpl) ValidateSession(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.Header.Get("Session")
	if sessionToken == "" {
		utils.ReturnError(w, http.StatusUnauthorized, "Session token is empty")
		return
	}
	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	validateSession, err := sr.sessionService.ValidateSession(tx, sessionToken)
	if err != nil {
		db.Rollback()
		if err == service.ErrSessionNotFound {
			utils.ReturnError(w, http.StatusUnauthorized, "Session not found")
			return
		}
		if err == service.ErrSessionExpired {
			utils.ReturnError(w, http.StatusUnauthorized, "Session expired")
			return
		}
		if err == service.ErrSessionUserNotFound {
			utils.ReturnError(w, http.StatusUnauthorized, "User associated with session not found")
			return
		}
		utils.ReturnError(w, http.StatusInternalServerError, "Failed to validate session. "+err.Error())
		return
	}

	if validateSession.Valid {
		utils.ReturnSuccess(w, http.StatusOK, validateSession)
		return
	} else {
		db.Rollback()
		utils.ReturnError(w, http.StatusUnauthorized, "Session was invalid, without any error. If you see this, please report it to the developers.")
		return
	}

}

func (sr *SessionRouteImpl) InvalidateSession(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.Header.Get("Session")
	if sessionToken == "" {
		utils.ReturnError(w, http.StatusUnauthorized, "Session token is empty")
		return
	}
	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	err = sr.sessionService.InvalidateSession(tx, sessionToken)
	if err != nil {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, "Failed to invalidate session. "+err.Error())
		return
	}
	utils.ReturnSuccess(w, http.StatusOK, "Session invalidated")
}

func NewSessionRoute(sessionService service.SessionService) SessionRoute {
	return &SessionRouteImpl{
		sessionService: sessionService,
	}
}
