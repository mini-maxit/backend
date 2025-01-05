package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mini-maxit/backend/internal/api/http/middleware"
	"github.com/mini-maxit/backend/internal/api/http/utils"
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
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	query := r.URL.Query()
	utils.SetDefaultQueryParams(&query, utils.UserDefaultSortOrder)

	queryParams := map[string][]string{}
	for key, value := range query {
		queryParams[key] = value
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error connecting to database. %s", err.Error()))
		return
	}

	users, err := u.userService.GetAllUsers(tx, queryParams)
	if err != nil {
		db.Rollback()
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error getting users. %s", err.Error()))
		return
	}

	if users == nil {
		users = []schemas.User{}
	}

	utils.ReturnSuccess(w, http.StatusOK, users)
}

func (u *UserRouteImpl) GetUserById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userIdStr := r.PathValue("id")

	if userIdStr == "" {
		utils.ReturnError(w, http.StatusBadRequest, "UserId cannot be empty")
		return
	}

	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, fmt.Sprintf("Invalid userId: %s", err.Error()))
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error connecting to database. %s", err.Error()))
		return
	}

	user, err := u.userService.GetUserById(tx, userId)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error fetching user: %s", err.Error()))
		return
	}

	utils.ReturnSuccess(w, http.StatusOK, user)
}

func (u *UserRouteImpl) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	query := r.URL.Query()
	email := query.Get("email")
	if email == "" {
		utils.ReturnError(w, http.StatusBadRequest, "Email cannot be empty")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error connecting to database. %s", err.Error()))
		return
	}

	user, err := u.userService.GetUserByEmail(tx, email)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error getting user. %s", err.Error()))
		return
	}

	utils.ReturnSuccess(w, http.StatusOK, user)
}

func (u *UserRouteImpl) EditUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request schemas.UserEdit

	userIdStr := r.PathValue("id")

	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid userId")
		return
	}

	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid request body. "+err.Error())
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, fmt.Sprintf("Error connecting to database. %s", err.Error()))
		return
	}

	err = u.userService.EditUser(tx, userId, &request)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Error ocured during editing. "+err.Error())
		return
	}

	utils.ReturnSuccess(w, http.StatusOK, "Update successfull")
}

func (u *UserRouteImpl) CreateUsers(w http.ResponseWriter, r *http.Request) {
	// this funcion allows admin to ctreate new users with their email and given role
	// the users will be created with a default password and will be required to change it on first login

	utils.ReturnError(w, http.StatusNotImplemented, "Not implemented")
}

func NewUserRoute(userService service.UserService) UserRoute {
	return &UserRouteImpl{userService: userService}
}
