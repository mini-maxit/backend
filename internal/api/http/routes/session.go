package routes

import (
	"encoding/json"
	"net/http"

	"github.com/mini-maxit/backend/internal/api/http/utils"
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

	session, err := sr.sessionService.CreateSession(request.UserId)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Failed to create session. "+err.Error())
		return
	}
	utils.ReturnSuccess(w, http.StatusOK, session)
}

func (sr *SessionRouteImpl) ValidateSession(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.Header.Get("session")
	if sessionToken == "" {
		utils.ReturnError(w, http.StatusUnauthorized, "Session token is empty")
		return
	}
	valid, err := sr.sessionService.ValidateSession(sessionToken)
	if err != nil {
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

	if valid {
		utils.ReturnSuccess(w, http.StatusOK, "Session is valid")
		return
	} else {
		utils.ReturnError(w, http.StatusUnauthorized, "Session was invalid, without any error. If you see this, please report it to the developers.")
		return
	}

}

func (sr *SessionRouteImpl) InvalidateSession(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.Header.Get("session")
	if sessionToken == "" {
		utils.ReturnError(w, http.StatusUnauthorized, "Session token is empty")
		return
	}
	err := sr.sessionService.InvalidateSession(sessionToken)
	if err != nil {
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
