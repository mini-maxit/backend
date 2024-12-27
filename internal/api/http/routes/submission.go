package routes

import (
	"net/http"
	"strconv"

	"github.com/mini-maxit/backend/internal/api/http/middleware"
	"github.com/mini-maxit/backend/internal/api/http/utils"
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/service"
)

type SubmissionRoutes interface {
	GetAll(w http.ResponseWriter, r *http.Request)
	GetById(w http.ResponseWriter, r *http.Request)
	GetAllForUser(w http.ResponseWriter, r *http.Request) 
	GetAllForGroup(w http.ResponseWriter, r *http.Request)
	GetAllForTask(w http.ResponseWriter, r *http.Request)
	GetAllForTaskAndUser(w http.ResponseWriter, r *http.Request)
	GetAllForTaskAndGroup(w http.ResponseWriter, r *http.Request)
}

type SumbissionImpl struct {
	submissionService service.SubmissionService
}


func (s *SumbissionImpl) GetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.UserSession)

	qurry := r.URL.Query()
	limitStr := qurry.Get("limit")
	offsetStr := qurry.Get("offset")

	limit, offset, err := utils.GetLimitAndOffset(limitStr, offsetStr)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid limit or offset. "+err.Error())
		return
	}

	submissions, err := s.submissionService.GetAll(tx, limit, offset, current_user)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Failed to get submissions. "+err.Error())
		return
	}

	if submissions == nil {
		submissions = []schemas.Submission{}
	}

	utils.ReturnSuccess(w, http.StatusOK, submissions)
}

func (s *SumbissionImpl) GetById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
 
	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	submissionIdStr := r.PathValue("id")
	submissionId, err := strconv.ParseInt(submissionIdStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid submission id. "+err.Error())
		return
	}

	user := r.Context().Value(middleware.UserKey).(schemas.UserSession)

	submission, err := s.submissionService.GetById(tx, submissionId, user)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Failed to get submission. "+err.Error())
		return
	}

	utils.ReturnSuccess(w, http.StatusOK, submission)
}

func (s *SumbissionImpl) GetAllForUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userIdStr := r.PathValue("id")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid user id. "+err.Error())
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.UserSession)

	qurry := r.URL.Query()
	limitStr := qurry.Get("limit")
	offsetStr := qurry.Get("offset")
	limit, offset, err := utils.GetLimitAndOffset(limitStr, offsetStr)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid limit or offset. "+err.Error())
		return
	}

	submissions, err := s.submissionService.GetAllForUser(tx, userId, limit, offset, current_user)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Failed to get submissions. "+err.Error())
		return
	}

	if submissions == nil {
		submissions = []schemas.Submission{}
	}

	utils.ReturnSuccess(w, http.StatusOK, submissions)
}

func (s *SumbissionImpl) GetAllForGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ReturnError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	groupIdStr := r.PathValue("id")
	groupId, err := strconv.ParseInt(groupIdStr, 10, 64)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid group id. "+err.Error())
		return
	}

	db := r.Context().Value(middleware.DatabaseKey).(database.Database)
	tx, err := db.Connect()
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Transaction was not started by middleware. "+err.Error())
		return
	}

	current_user := r.Context().Value(middleware.UserKey).(schemas.UserSession)

	qurry := r.URL.Query()
	limitStr := qurry.Get("limit")
	offsetStr := qurry.Get("offset")
	limit, offset, err := utils.GetLimitAndOffset(limitStr, offsetStr)
	if err != nil {
		utils.ReturnError(w, http.StatusBadRequest, "Invalid limit or offset. "+err.Error())
		return
	}

	submissions, err := s.submissionService.GetAllForGroup(tx, groupId, limit, offset, current_user)
	if err != nil {
		utils.ReturnError(w, http.StatusInternalServerError, "Failed to get submissions. "+err.Error())
		return
	}

	if submissions == nil {
		submissions = []schemas.Submission{}
	}

	utils.ReturnSuccess(w, http.StatusOK, submissions)
}

func (s *SumbissionImpl) GetAllForTask(w http.ResponseWriter, r *http.Request) {
	utils.ReturnError(w, http.StatusNotImplemented, "Not implemented")
}

func (s *SumbissionImpl) GetAllForTaskAndUser(w http.ResponseWriter, r *http.Request) {
	utils.ReturnError(w, http.StatusNotImplemented, "Not implemented")
}

func (s *SumbissionImpl) GetAllForTaskAndGroup(w http.ResponseWriter, r *http.Request) {
	utils.ReturnError(w, http.StatusNotImplemented, "Not implemented")
}

func NewSubmissionRoutes(submissionService service.SubmissionService) SubmissionRoutes {
	return &SumbissionImpl{
		submissionService: submissionService,
	}
}