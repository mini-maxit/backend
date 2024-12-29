package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/api/http/middleware"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/service"
)

type UserRoute interface {
	GetAllUsers(w http.ResponseWriter, r *http.Request)
	GetUserById(w http.ResponseWriter, r *http.Request)
	GetUserByEmail(w http.ResponseWriter, r *http.Request)
	EditUser(w http.ResponseWriter, r *http.Request)
	CreateUsers(w http.ResponseWriter, r *http.Request)
}

type UserRouteImpl struct {
	userService service.UserService
}

func (u *UserRouteImpl) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	query := r.URL.Query()

	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	if limitStr == "" {
		limitStr = httputils.DefaultPaginationLimitStr
	}

	if offsetStr == "" {
		offsetStr = httputils.DefaultPaginationOffsetStr
	}

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid limit")
		return
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid offset")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error connecting to database. %s", err.Error()))
		return
	}

	users, err := u.userService.GetAllUsers(tx, limit, offset)
	if err != nil {
		db.Rollback()
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error getting users. %s", err.Error()))
		return
	}

	if users == nil {
		users = []schemas.User{}
	}

	httputils.ReturnSuccess(w, http.StatusOK, users)
}

func (u *UserRouteImpl) GetUserById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userIdStr := r.PathValue("id")

	if userIdStr == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "UserId cannot be empty")
		return
	}

	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, fmt.Sprintf("Invalid userId: %s", err.Error()))
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error connecting to database. %s", err.Error()))
		return
	}

	user, err := u.userService.GetUserById(tx, userId)
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error fetching user: %s", err.Error()))
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, user)
}

func (u *UserRouteImpl) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	query := r.URL.Query()
	email := query.Get("email")
	if email == "" {
		httputils.ReturnError(w, http.StatusBadRequest, "Email cannot be empty")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error connecting to database. %s", err.Error()))
		return
	}

	user, err := u.userService.GetUserByEmail(tx, email)
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error getting user. %s", err.Error()))
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, user)
}

func (u *UserRouteImpl) EditUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		httputils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.UserEdit

	userIdStr := r.PathValue("id")

	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid userId")
		return
	}

	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		httputils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error connecting to database. %s", err.Error()))
		return
	}

	err = u.userService.EditUser(tx, userId, &request)
	if err != nil {
		httputils.ReturnError(w, http.StatusInternalServerError, "Error ocured during editing. "+err.Error())
		return
	}

	httputils.ReturnSuccess(w, http.StatusOK, "Update successfull")
}

func (u *UserRouteImpl) CreateUsers(w http.ResponseWriter, r *http.Request) {
	// this funcion allows admin to ctreate new users with their email and given role
	// the users will be created with a default password and will be required to change it on first login

	httputils.ReturnError(w, http.StatusNotImplemented, "Not implemented")
}

func NewUserRoute(userService service.UserService) UserRoute {
	return &UserRouteImpl{userService: userService}
}
