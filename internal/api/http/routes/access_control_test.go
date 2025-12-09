package routes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/api/http/routes"
	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func setupAccessControlTest(t *testing.T) (*gomock.Controller, *mock_service.MockAccessControlService, routes.AccessControlRoute, *testutils.MockDatabase, schemas.User) {
	ctrl := gomock.NewController(t)
	acs := mock_service.NewMockAccessControlService(ctrl)
	route := routes.NewAccessControlRoute(acs)
	db := &testutils.MockDatabase{}
	currentUser := schemas.User{
		ID:      1,
		Name:    "Test",
		Surname: "User",
		Email:   "test@email.com",
		Role:    types.UserRoleAdmin,
	}
	return ctrl, acs, route, db, currentUser
}

func createAccessControlRouter(db *testutils.MockDatabase, currentUser schemas.User, route routes.AccessControlRoute) *mux.Router {
	r := mux.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
	routes.RegisterAccessControlRoutes(r, route)
	return r
}

// Contest Collaborator Tests

func TestAddContestCollaborator(t *testing.T) {
	ctrl, acs, route, db, currentUser := setupAccessControlTest(t)
	defer ctrl.Finish()

	t.Run("Invalid contest ID", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		reqBody := schemas.AddCollaborator{UserID: 2, Permission: types.PermissionEdit}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/resources/contests/abc/collaborators", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid resource ID")
	})

	t.Run("Invalid request body", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodPost, "/resources/contests/1/collaborators", bytes.NewBufferString("invalid json"))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Not authorized", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		reqBody := schemas.AddCollaborator{UserID: 2, Permission: types.PermissionEdit}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/resources/contests/1/collaborators", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		acs.EXPECT().AddCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2), types.PermissionEdit).Return(errors.ErrForbidden).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("User not found", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		reqBody := schemas.AddCollaborator{UserID: 999, Permission: types.PermissionEdit}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/resources/contests/1/collaborators", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		acs.EXPECT().AddCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(999), types.PermissionEdit).Return(errors.ErrNotFound).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Access already exists", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		reqBody := schemas.AddCollaborator{UserID: 2, Permission: types.PermissionEdit}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/resources/contests/1/collaborators", bytes.NewBuffer(jsonBody))
		req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
		w := httptest.NewRecorder()

		acs.EXPECT().AddCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2), types.PermissionEdit).Return(errors.ErrAccessAlreadyExists).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Internal server error", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		reqBody := schemas.AddCollaborator{UserID: 2, Permission: types.PermissionEdit}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/resources/contests/1/collaborators", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		acs.EXPECT().AddCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2), types.PermissionEdit).Return(gorm.ErrInvalidDB).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		reqBody := schemas.AddCollaborator{UserID: 2, Permission: types.PermissionEdit}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/resources/contests/1/collaborators", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		acs.EXPECT().AddCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2), types.PermissionEdit).Return(nil).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Collaborator added successfully")
	})
}

func TestGetContestCollaborators(t *testing.T) {
	ctrl, acs, route, db, currentUser := setupAccessControlTest(t)
	defer ctrl.Finish()

	t.Run("Invalid contest ID", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodGet, "/resources/contests/abc/collaborators", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid resource ID")
	})

	t.Run("Contest not found", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodGet, "/resources/contests/999/collaborators", nil)
		w := httptest.NewRecorder()

		acs.EXPECT().GetCollaborators(gomock.Any(), gomock.Any(), gomock.Any(), int64(999)).Return(nil, errors.ErrNotFound).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Not authorized", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodGet, "/resources/contests/1/collaborators", nil)
		w := httptest.NewRecorder()

		acs.EXPECT().GetCollaborators(gomock.Any(), gomock.Any(), gomock.Any(), int64(1)).Return(nil, errors.ErrForbidden).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Internal server error", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodGet, "/resources/contests/1/collaborators", nil)
		w := httptest.NewRecorder()

		acs.EXPECT().GetCollaborators(gomock.Any(), gomock.Any(), gomock.Any(), int64(1)).Return(nil, gorm.ErrInvalidDB).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodGet, "/resources/contests/1/collaborators", nil)
		w := httptest.NewRecorder()

		expectedCollaborators := []schemas.Collaborator{
			{UserID: 1, UserName: "User1", UserEmail: "user1@email.com", Permission: types.PermissionOwner, AddedAt: "2024-01-01T00:00:00Z"},
			{UserID: 2, UserName: "User2", UserEmail: "user2@email.com", Permission: types.PermissionEdit, AddedAt: "2024-01-02T00:00:00Z"},
		}

		acs.EXPECT().GetCollaborators(gomock.Any(), gomock.Any(), gomock.Any(), int64(1)).Return(expectedCollaborators, nil).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		response := &httputils.APIResponse[[]schemas.Collaborator]{}
		err := json.Unmarshal(w.Body.Bytes(), response)
		require.NoError(t, err)
		assert.Equal(t, expectedCollaborators, response.Data)
	})
}

func TestUpdateContestCollaborator(t *testing.T) {
	ctrl, acs, route, db, currentUser := setupAccessControlTest(t)
	defer ctrl.Finish()

	t.Run("Invalid contest ID", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/resources/contests/abc/collaborators/2", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid resource ID")
	})
	t.Run("Invalid user ID", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/resources/contests/1/collaborators/abc", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid user ID")
	})

	t.Run("Invalid request body", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodPut, "/resources/contests/1/collaborators/2", bytes.NewBufferString("invalid json"))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Not authorized", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/resources/contests/1/collaborators/2", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		acs.EXPECT().UpdateCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2), types.PermissionManage).Return(errors.ErrForbidden).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Collaborator not found", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/resources/contests/1/collaborators/999", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		acs.EXPECT().UpdateCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(999), types.PermissionManage).Return(errors.ErrNotFound).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Internal server error", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/resources/contests/1/collaborators/2", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		acs.EXPECT().UpdateCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2), types.PermissionManage).Return(gorm.ErrInvalidDB).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/resources/contests/1/collaborators/2", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		acs.EXPECT().UpdateCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2), types.PermissionManage).Return(nil).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Collaborator permission updated successfully")
	})
}

func TestRemoveContestCollaborator(t *testing.T) {
	ctrl, acs, route, db, currentUser := setupAccessControlTest(t)
	defer ctrl.Finish()

	t.Run("Invalid contest ID", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodDelete, "/resources/contests/abc/collaborators/2", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid resource ID")
	})

	t.Run("Invalid user ID", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodDelete, "/resources/contests/1/collaborators/abc", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid user ID")
	})

	t.Run("Not authorized", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodDelete, "/resources/contests/1/collaborators/2", nil)
		w := httptest.NewRecorder()

		acs.EXPECT().RemoveCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(errors.ErrForbidden).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Collaborator not found", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodDelete, "/resources/contests/1/collaborators/999", nil)
		w := httptest.NewRecorder()

		acs.EXPECT().RemoveCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(999)).Return(errors.ErrNotFound).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Internal server error", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodDelete, "/resources/contests/1/collaborators/2", nil)
		w := httptest.NewRecorder()

		acs.EXPECT().RemoveCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(gorm.ErrInvalidDB).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodDelete, "/resources/contests/1/collaborators/2", nil)
		w := httptest.NewRecorder()

		acs.EXPECT().RemoveCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(nil).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Collaborator removed successfully")
	})
}

// Task Collaborator Tests

func TestAddTaskCollaborator(t *testing.T) {
	ctrl, acs, route, db, currentUser := setupAccessControlTest(t)
	defer ctrl.Finish()

	t.Run("Invalid task ID", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		reqBody := schemas.AddCollaborator{UserID: 2, Permission: types.PermissionEdit}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/resources/tasks/abc/collaborators", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid resource ID")
	})

	t.Run("Invalid request body", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodPost, "/resources/tasks/1/collaborators", bytes.NewBufferString("invalid json"))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Not authorized", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		reqBody := schemas.AddCollaborator{UserID: 2, Permission: types.PermissionEdit}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/resources/tasks/1/collaborators", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		acs.EXPECT().AddCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2), types.PermissionEdit).Return(errors.ErrForbidden).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		reqBody := schemas.AddCollaborator{UserID: 2, Permission: types.PermissionEdit}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/resources/tasks/1/collaborators", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		acs.EXPECT().AddCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2), types.PermissionEdit).Return(nil).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Collaborator added successfully")
	})
}

func TestGetTaskCollaborators(t *testing.T) {
	ctrl, acs, route, db, currentUser := setupAccessControlTest(t)
	defer ctrl.Finish()

	t.Run("Invalid task ID", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodGet, "/resources/tasks/abc/collaborators", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid resource ID")
	})

	t.Run("Task not found", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodGet, "/resources/tasks/999/collaborators", nil)
		w := httptest.NewRecorder()

		acs.EXPECT().GetCollaborators(gomock.Any(), gomock.Any(), gomock.Any(), int64(999)).Return(nil, errors.ErrNotFound).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodGet, "/resources/tasks/1/collaborators", nil)
		w := httptest.NewRecorder()

		expectedCollaborators := []schemas.Collaborator{
			{UserID: 1, UserName: "User1", UserEmail: "user1@email.com", Permission: types.PermissionOwner, AddedAt: "2024-01-01T00:00:00Z"},
		}

		acs.EXPECT().GetCollaborators(gomock.Any(), gomock.Any(), gomock.Any(), int64(1)).Return(expectedCollaborators, nil).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Error is handled", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodGet, "/resources/tasks/1/collaborators", nil)
		w := httptest.NewRecorder()

		acs.EXPECT().GetCollaborators(gomock.Any(), gomock.Any(), gomock.Any(), int64(1)).Return(nil, errors.ErrInvalidData).Times(1)

		router.ServeHTTP(w, req)

		assert.Condition(t, func() bool {
			return w.Result().StatusCode >= 400 && w.Result().StatusCode < 600
		})
	})
}

func TestUpdateTaskCollaborator(t *testing.T) {
	ctrl, acs, route, db, currentUser := setupAccessControlTest(t)
	defer ctrl.Finish()

	t.Run("Invalid task ID", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/resources/tasks/abc/collaborators/2", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid resource ID")
	})

	t.Run("Invalid user ID", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/resources/tasks/1/collaborators/abc", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid user ID")
	})

	t.Run("Invalid request body", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		reqBody := []byte("invalid json")
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/resources/tasks/1/collaborators/1", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), httputils.InvalidRequestBodyMessage)
	})

	t.Run("Service error is handled", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/resources/tasks/1/collaborators/1", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		acs.EXPECT().UpdateCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(1), types.PermissionManage).Return(errors.ErrInvalidData).Times(1)

		router.ServeHTTP(w, req)

		assert.Condition(t, func() bool {
			return w.Result().StatusCode >= 400 && w.Result().StatusCode < 600
		})
	})

	t.Run("Success", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/resources/tasks/1/collaborators/2", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		acs.EXPECT().UpdateCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2), types.PermissionManage).Return(nil).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Collaborator permission updated successfully")
	})
}

func TestRemoveTaskCollaborator(t *testing.T) {
	ctrl, acs, route, db, currentUser := setupAccessControlTest(t)
	defer ctrl.Finish()

	t.Run("Invalid task ID", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodDelete, "/resources/tasks/abc/collaborators/2", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid resource ID")
	})

	t.Run("Invalid user ID", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodDelete, "/resources/tasks/1/collaborators/abc", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid user ID")
	})
	t.Run("Service error is handled", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodDelete, "/resources/tasks/1/collaborators/2", nil)
		w := httptest.NewRecorder()

		acs.EXPECT().RemoveCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(errors.ErrInvalidData).Times(1)

		router.ServeHTTP(w, req)

		assert.Condition(t, func() bool {
			return w.Result().StatusCode >= 400 && w.Result().StatusCode < 600
		})
	})
	t.Run("Success", func(t *testing.T) {
		router := createAccessControlRouter(db, currentUser, route)
		req := httptest.NewRequest(http.MethodDelete, "/resources/tasks/1/collaborators/2", nil)
		w := httptest.NewRecorder()

		acs.EXPECT().RemoveCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(nil).Times(1)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Collaborator removed successfully")
	})
}
