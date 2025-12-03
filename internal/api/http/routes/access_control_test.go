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

func createAccessControlHandler(route routes.AccessControlRoute, db *testutils.MockDatabase, currentUser schemas.User, handler http.HandlerFunc) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
handler(w, r.WithContext(ctx))
})
}

// Contest Collaborator Tests

func TestAddContestCollaborator(t *testing.T) {
ctrl, acs, route, db, currentUser := setupAccessControlTest(t)
defer ctrl.Finish()

t.Run("Accept only POST", func(t *testing.T) {
methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

for _, method := range methods {
handler := createAccessControlHandler(route, db, currentUser, route.AddContestCollaborator)
req := httptest.NewRequest(method, "/contests/1/collaborators", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
})

t.Run("Empty contest ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.AddContestCollaborator)
reqBody := schemas.AddCollaborator{UserID: 2, Permission: types.PermissionEdit}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPost, "/contests//collaborators", bytes.NewBuffer(jsonBody))
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "Contest ID cannot be empty")
})

t.Run("Invalid contest ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.AddContestCollaborator)
reqBody := schemas.AddCollaborator{UserID: 2, Permission: types.PermissionEdit}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPost, "/contests/abc/collaborators", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"resource_id": "abc"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "Invalid contest ID")
})

t.Run("Invalid request body", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.AddContestCollaborator)
req := httptest.NewRequest(http.MethodPost, "/contests/1/collaborators", bytes.NewBufferString("invalid json"))
req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
})

t.Run("Not authorized", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.AddContestCollaborator)
reqBody := schemas.AddCollaborator{UserID: 2, Permission: types.PermissionEdit}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPost, "/contests/1/collaborators", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
w := httptest.NewRecorder()

acs.EXPECT().AddCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2), types.PermissionEdit).Return(errors.ErrForbidden).Times(1)

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusForbidden, w.Code)
})

t.Run("User not found", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.AddContestCollaborator)
reqBody := schemas.AddCollaborator{UserID: 999, Permission: types.PermissionEdit}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPost, "/contests/1/collaborators", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
w := httptest.NewRecorder()

acs.EXPECT().AddCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(999), types.PermissionEdit).Return(errors.ErrNotFound).Times(1)

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusNotFound, w.Code)
})

t.Run("Access already exists", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.AddContestCollaborator)
reqBody := schemas.AddCollaborator{UserID: 2, Permission: types.PermissionEdit}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPost, "/contests/1/collaborators", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
w := httptest.NewRecorder()

acs.EXPECT().AddCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2), types.PermissionEdit).Return(errors.ErrAccessAlreadyExists).Times(1)

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusConflict, w.Code)
})

t.Run("Internal server error", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.AddContestCollaborator)
reqBody := schemas.AddCollaborator{UserID: 2, Permission: types.PermissionEdit}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPost, "/contests/1/collaborators", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
w := httptest.NewRecorder()

acs.EXPECT().AddCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2), types.PermissionEdit).Return(gorm.ErrInvalidDB).Times(1)

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusInternalServerError, w.Code)
})

t.Run("Success", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.AddContestCollaborator)
reqBody := schemas.AddCollaborator{UserID: 2, Permission: types.PermissionEdit}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPost, "/contests/1/collaborators", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
w := httptest.NewRecorder()

acs.EXPECT().AddCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2), types.PermissionEdit).Return(nil).Times(1)

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusOK, w.Code)
assert.Contains(t, w.Body.String(), "Collaborator added successfully")
})
}

func TestGetContestCollaborators(t *testing.T) {
ctrl, acs, route, db, currentUser := setupAccessControlTest(t)
defer ctrl.Finish()

t.Run("Accept only GET", func(t *testing.T) {
methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

for _, method := range methods {
handler := createAccessControlHandler(route, db, currentUser, route.GetContestCollaborators)
req := httptest.NewRequest(method, "/contests/1/collaborators", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
})

t.Run("Empty contest ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.GetContestCollaborators)
req := httptest.NewRequest(http.MethodGet, "/contests//collaborators", nil)
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "Contest ID cannot be empty")
})

t.Run("Invalid contest ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.GetContestCollaborators)
req := httptest.NewRequest(http.MethodGet, "/contests/abc/collaborators", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "abc"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "Invalid contest ID")
})

t.Run("Not authorized", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.GetContestCollaborators)
req := httptest.NewRequest(http.MethodGet, "/contests/1/collaborators", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
w := httptest.NewRecorder()

acs.EXPECT().GetCollaborators(gomock.Any(), gomock.Any(), gomock.Any(), int64(1)).Return(nil, errors.ErrForbidden).Times(1)

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusForbidden, w.Code)
})

t.Run("Internal server error", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.GetContestCollaborators)
req := httptest.NewRequest(http.MethodGet, "/contests/1/collaborators", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
w := httptest.NewRecorder()

acs.EXPECT().GetCollaborators(gomock.Any(), gomock.Any(), gomock.Any(), int64(1)).Return(nil, gorm.ErrInvalidDB).Times(1)

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusInternalServerError, w.Code)
})

t.Run("Success", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.GetContestCollaborators)
req := httptest.NewRequest(http.MethodGet, "/contests/1/collaborators", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
w := httptest.NewRecorder()

expectedCollaborators := []schemas.Collaborator{
{UserID: 1, UserName: "User1", UserEmail: "user1@email.com", Permission: types.PermissionOwner, AddedAt: "2024-01-01T00:00:00Z"},
{UserID: 2, UserName: "User2", UserEmail: "user2@email.com", Permission: types.PermissionEdit, AddedAt: "2024-01-02T00:00:00Z"},
}

acs.EXPECT().GetCollaborators(gomock.Any(), gomock.Any(), gomock.Any(), int64(1)).Return(expectedCollaborators, nil).Times(1)

handler.ServeHTTP(w, req)

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

t.Run("Accept only PUT", func(t *testing.T) {
methods := []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPatch}

for _, method := range methods {
handler := createAccessControlHandler(route, db, currentUser, route.UpdateContestCollaborator)
req := httptest.NewRequest(method, "/contests/1/collaborators/2", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "1", "user_id": "2"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
})

t.Run("Empty contest ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.UpdateContestCollaborator)
reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPut, "/contests//collaborators/2", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"user_id": "2"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "Contest ID cannot be empty")
})

t.Run("Invalid contest ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.UpdateContestCollaborator)
reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPut, "/contests/abc/collaborators/2", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"resource_id": "abc", "user_id": "2"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "Invalid contest ID")
})

t.Run("Empty user ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.UpdateContestCollaborator)
reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPut, "/contests/1/collaborators/", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "User ID cannot be empty")
})

t.Run("Invalid user ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.UpdateContestCollaborator)
reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPut, "/contests/1/collaborators/abc", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"resource_id": "1", "user_id": "abc"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "Invalid user ID")
})

t.Run("Invalid request body", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.UpdateContestCollaborator)
req := httptest.NewRequest(http.MethodPut, "/contests/1/collaborators/2", bytes.NewBufferString("invalid json"))
req = mux.SetURLVars(req, map[string]string{"resource_id": "1", "user_id": "2"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
})

t.Run("Not authorized", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.UpdateContestCollaborator)
reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPut, "/contests/1/collaborators/2", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"resource_id": "1", "user_id": "2"})
w := httptest.NewRecorder()

acs.EXPECT().UpdateCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2), types.PermissionManage).Return(errors.ErrForbidden).Times(1)

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusForbidden, w.Code)
})

t.Run("Collaborator not found", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.UpdateContestCollaborator)
reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPut, "/contests/1/collaborators/999", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"resource_id": "1", "user_id": "999"})
w := httptest.NewRecorder()

acs.EXPECT().UpdateCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(999), types.PermissionManage).Return(errors.ErrNotFound).Times(1)

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusNotFound, w.Code)
})

t.Run("Internal server error", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.UpdateContestCollaborator)
reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPut, "/contests/1/collaborators/2", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"resource_id": "1", "user_id": "2"})
w := httptest.NewRecorder()

acs.EXPECT().UpdateCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2), types.PermissionManage).Return(gorm.ErrInvalidDB).Times(1)

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusInternalServerError, w.Code)
})

t.Run("Success", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.UpdateContestCollaborator)
reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPut, "/contests/1/collaborators/2", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"resource_id": "1", "user_id": "2"})
w := httptest.NewRecorder()

acs.EXPECT().UpdateCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2), types.PermissionManage).Return(nil).Times(1)

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusOK, w.Code)
assert.Contains(t, w.Body.String(), "Collaborator permission updated successfully")
})
}

func TestRemoveContestCollaborator(t *testing.T) {
ctrl, acs, route, db, currentUser := setupAccessControlTest(t)
defer ctrl.Finish()

t.Run("Accept only DELETE", func(t *testing.T) {
methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch}

for _, method := range methods {
handler := createAccessControlHandler(route, db, currentUser, route.RemoveContestCollaborator)
req := httptest.NewRequest(method, "/contests/1/collaborators/2", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "1", "user_id": "2"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
})

t.Run("Empty contest ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.RemoveContestCollaborator)
req := httptest.NewRequest(http.MethodDelete, "/contests//collaborators/2", nil)
req = mux.SetURLVars(req, map[string]string{"user_id": "2"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "Contest ID cannot be empty")
})

t.Run("Invalid contest ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.RemoveContestCollaborator)
req := httptest.NewRequest(http.MethodDelete, "/contests/abc/collaborators/2", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "abc", "user_id": "2"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "Invalid contest ID")
})

t.Run("Empty user ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.RemoveContestCollaborator)
req := httptest.NewRequest(http.MethodDelete, "/contests/1/collaborators/", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "User ID cannot be empty")
})

t.Run("Invalid user ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.RemoveContestCollaborator)
req := httptest.NewRequest(http.MethodDelete, "/contests/1/collaborators/abc", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "1", "user_id": "abc"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "Invalid user ID")
})

t.Run("Not authorized", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.RemoveContestCollaborator)
req := httptest.NewRequest(http.MethodDelete, "/contests/1/collaborators/2", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "1", "user_id": "2"})
w := httptest.NewRecorder()

acs.EXPECT().RemoveCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(errors.ErrForbidden).Times(1)

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusForbidden, w.Code)
})

t.Run("Collaborator not found", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.RemoveContestCollaborator)
req := httptest.NewRequest(http.MethodDelete, "/contests/1/collaborators/999", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "1", "user_id": "999"})
w := httptest.NewRecorder()

acs.EXPECT().RemoveCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(999)).Return(errors.ErrNotFound).Times(1)

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusNotFound, w.Code)
})

t.Run("Internal server error", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.RemoveContestCollaborator)
req := httptest.NewRequest(http.MethodDelete, "/contests/1/collaborators/2", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "1", "user_id": "2"})
w := httptest.NewRecorder()

acs.EXPECT().RemoveCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(gorm.ErrInvalidDB).Times(1)

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusInternalServerError, w.Code)
})

t.Run("Success", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.RemoveContestCollaborator)
req := httptest.NewRequest(http.MethodDelete, "/contests/1/collaborators/2", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "1", "user_id": "2"})
w := httptest.NewRecorder()

acs.EXPECT().RemoveCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(nil).Times(1)

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusOK, w.Code)
assert.Contains(t, w.Body.String(), "Collaborator removed successfully")
})
}

// Task Collaborator Tests

func TestAddTaskCollaborator(t *testing.T) {
ctrl, acs, route, db, currentUser := setupAccessControlTest(t)
defer ctrl.Finish()

t.Run("Accept only POST", func(t *testing.T) {
methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

for _, method := range methods {
handler := createAccessControlHandler(route, db, currentUser, route.AddTaskCollaborator)
req := httptest.NewRequest(method, "/tasks/1/collaborators", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
})

t.Run("Empty task ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.AddTaskCollaborator)
reqBody := schemas.AddCollaborator{UserID: 2, Permission: types.PermissionEdit}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPost, "/tasks//collaborators", bytes.NewBuffer(jsonBody))
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "Task ID cannot be empty")
})

t.Run("Invalid task ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.AddTaskCollaborator)
reqBody := schemas.AddCollaborator{UserID: 2, Permission: types.PermissionEdit}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPost, "/tasks/abc/collaborators", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"resource_id": "abc"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "Invalid task ID")
})

t.Run("Invalid request body", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.AddTaskCollaborator)
req := httptest.NewRequest(http.MethodPost, "/tasks/1/collaborators", bytes.NewBufferString("invalid json"))
req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
})

t.Run("Not authorized", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.AddTaskCollaborator)
reqBody := schemas.AddCollaborator{UserID: 2, Permission: types.PermissionEdit}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPost, "/tasks/1/collaborators", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
w := httptest.NewRecorder()

acs.EXPECT().AddCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2), types.PermissionEdit).Return(errors.ErrForbidden).Times(1)

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusForbidden, w.Code)
})

t.Run("Success", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.AddTaskCollaborator)
reqBody := schemas.AddCollaborator{UserID: 2, Permission: types.PermissionEdit}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPost, "/tasks/1/collaborators", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
w := httptest.NewRecorder()

acs.EXPECT().AddCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2), types.PermissionEdit).Return(nil).Times(1)

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusOK, w.Code)
assert.Contains(t, w.Body.String(), "Collaborator added successfully")
})
}

func TestGetTaskCollaborators(t *testing.T) {
ctrl, acs, route, db, currentUser := setupAccessControlTest(t)
defer ctrl.Finish()

t.Run("Accept only GET", func(t *testing.T) {
methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

for _, method := range methods {
handler := createAccessControlHandler(route, db, currentUser, route.GetTaskCollaborators)
req := httptest.NewRequest(method, "/tasks/1/collaborators", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
})

t.Run("Empty task ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.GetTaskCollaborators)
req := httptest.NewRequest(http.MethodGet, "/tasks//collaborators", nil)
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "Task ID cannot be empty")
})

t.Run("Invalid task ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.GetTaskCollaborators)
req := httptest.NewRequest(http.MethodGet, "/tasks/abc/collaborators", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "abc"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "Invalid task ID")
})

t.Run("Success", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.GetTaskCollaborators)
req := httptest.NewRequest(http.MethodGet, "/tasks/1/collaborators", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
w := httptest.NewRecorder()

expectedCollaborators := []schemas.Collaborator{
{UserID: 1, UserName: "User1", UserEmail: "user1@email.com", Permission: types.PermissionOwner, AddedAt: "2024-01-01T00:00:00Z"},
}

acs.EXPECT().GetCollaborators(gomock.Any(), gomock.Any(), gomock.Any(), int64(1)).Return(expectedCollaborators, nil).Times(1)

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusOK, w.Code)
})
}

func TestUpdateTaskCollaborator(t *testing.T) {
ctrl, acs, route, db, currentUser := setupAccessControlTest(t)
defer ctrl.Finish()

t.Run("Accept only PUT", func(t *testing.T) {
methods := []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPatch}

for _, method := range methods {
handler := createAccessControlHandler(route, db, currentUser, route.UpdateTaskCollaborator)
req := httptest.NewRequest(method, "/tasks/1/collaborators/2", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "1", "user_id": "2"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
})

t.Run("Empty task ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.UpdateTaskCollaborator)
reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPut, "/tasks//collaborators/2", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"user_id": "2"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "Task ID cannot be empty")
})

t.Run("Invalid task ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.UpdateTaskCollaborator)
reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPut, "/tasks/abc/collaborators/2", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"resource_id": "abc", "user_id": "2"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "Invalid task ID")
})

t.Run("Empty user ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.UpdateTaskCollaborator)
reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPut, "/tasks/1/collaborators/", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "User ID cannot be empty")
})

t.Run("Invalid user ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.UpdateTaskCollaborator)
reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPut, "/tasks/1/collaborators/abc", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"resource_id": "1", "user_id": "abc"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "Invalid user ID")
})

t.Run("Success", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.UpdateTaskCollaborator)
reqBody := schemas.UpdateCollaborator{Permission: types.PermissionManage}
jsonBody, _ := json.Marshal(reqBody)
req := httptest.NewRequest(http.MethodPut, "/tasks/1/collaborators/2", bytes.NewBuffer(jsonBody))
req = mux.SetURLVars(req, map[string]string{"resource_id": "1", "user_id": "2"})
w := httptest.NewRecorder()

acs.EXPECT().UpdateCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2), types.PermissionManage).Return(nil).Times(1)

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusOK, w.Code)
assert.Contains(t, w.Body.String(), "Collaborator permission updated successfully")
})
}

func TestRemoveTaskCollaborator(t *testing.T) {
ctrl, acs, route, db, currentUser := setupAccessControlTest(t)
defer ctrl.Finish()

t.Run("Accept only DELETE", func(t *testing.T) {
methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch}

for _, method := range methods {
handler := createAccessControlHandler(route, db, currentUser, route.RemoveTaskCollaborator)
req := httptest.NewRequest(method, "/tasks/1/collaborators/2", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "1", "user_id": "2"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
})

t.Run("Empty task ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.RemoveTaskCollaborator)
req := httptest.NewRequest(http.MethodDelete, "/tasks//collaborators/2", nil)
req = mux.SetURLVars(req, map[string]string{"user_id": "2"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "Task ID cannot be empty")
})

t.Run("Invalid task ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.RemoveTaskCollaborator)
req := httptest.NewRequest(http.MethodDelete, "/tasks/abc/collaborators/2", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "abc", "user_id": "2"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "Invalid task ID")
})

t.Run("Empty user ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.RemoveTaskCollaborator)
req := httptest.NewRequest(http.MethodDelete, "/tasks/1/collaborators/", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "1"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "User ID cannot be empty")
})

t.Run("Invalid user ID", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.RemoveTaskCollaborator)
req := httptest.NewRequest(http.MethodDelete, "/tasks/1/collaborators/abc", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "1", "user_id": "abc"})
w := httptest.NewRecorder()

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusBadRequest, w.Code)
assert.Contains(t, w.Body.String(), "Invalid user ID")
})

t.Run("Success", func(t *testing.T) {
handler := createAccessControlHandler(route, db, currentUser, route.RemoveTaskCollaborator)
req := httptest.NewRequest(http.MethodDelete, "/tasks/1/collaborators/2", nil)
req = mux.SetURLVars(req, map[string]string{"resource_id": "1", "user_id": "2"})
w := httptest.NewRecorder()

acs.EXPECT().RemoveCollaborator(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(nil).Times(1)

handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusOK, w.Code)
assert.Contains(t, w.Body.String(), "Collaborator removed successfully")
})
}
