package routes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/api/http/routes"
	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	myerrors "github.com/mini-maxit/backend/package/errors"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestGetAllUsers(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	us := mock_service.NewMockUserService(ctrl)
	route := routes.NewUserRoute(us)
	db := &testutils.MockDatabase{}

	t.Run("Accept only GET", func(t *testing.T) {
		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			handler := testutils.MockDatabaseMiddleware(http.HandlerFunc(route.GetAllUsers), db)
			server := httptest.NewServer(handler)
			defer server.Close()

			req, err := http.NewRequest(method, server.URL, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Transaction was not started by middleware", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.QueryParamsKey, map[string]any{})
			ctx = context.WithValue(ctx, httputils.DatabaseKey, db)
			route.GetAllUsers(w, r.WithContext(ctx))
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		db.Invalidate()
		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		db.Validate()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		bodyString := string(bodyBytes)

		assert.Contains(t, bodyString, "Database connection error")
	})

	t.Run("Internal server error", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.QueryParamsKey, map[string]any{})
			route.GetAllUsers(w, r.WithContext(ctx))
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		us.EXPECT().GetAll(gomock.Any(), gomock.Any()).Return(schemas.PaginatedResult[[]schemas.User]{}, gorm.ErrInvalidDB).Times(1)

		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		bodyString := string(bodyBytes)

		assert.Contains(t, bodyString, "User service temporarily unavailable")
	})

	t.Run("Success", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.QueryParamsKey, map[string]any{})
			route.GetAllUsers(w, r.WithContext(ctx))
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		expectedUsers := []schemas.User{
			{ID: 1, Name: "User1", Email: "user1@email.com", Role: types.UserRoleStudent},
			{ID: 2, Name: "User2", Email: "user2@email.com", Role: types.UserRoleAdmin},
		}

		paginatedResult := schemas.NewPaginatedResult(expectedUsers, 0, 0, int64(len(expectedUsers)))
		us.EXPECT().GetAll(gomock.Any(), gomock.Any()).Return(paginatedResult, nil).Times(1)

		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		response := &httputils.APIResponse[schemas.PaginatedResult[[]schemas.User]]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Equal(t, expectedUsers, response.Data.Items)
	})

	t.Run("Success with empty list", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.QueryParamsKey, map[string]any{})
			route.GetAllUsers(w, r.WithContext(ctx))
		})
		server := httptest.NewServer(handler)
		defer server.Close()

		emptyResult := schemas.NewPaginatedResult([]schemas.User{}, 0, 0, 0)
		us.EXPECT().GetAll(gomock.Any(), gomock.Any()).Return(emptyResult, nil).Times(1)

		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		response := &httputils.APIResponse[schemas.PaginatedResult[[]schemas.User]]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Empty(t, response.Data.Items)
	})
}

func TestGetUserByID(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	us := mock_service.NewMockUserService(ctrl)
	route := routes.NewUserRoute(us)
	db := &testutils.MockDatabase{}

	t.Run("Accept only GET", func(t *testing.T) {
		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			handler := testutils.MockDatabaseMiddleware(http.HandlerFunc(route.GetUserByID), db)
			server := httptest.NewServer(handler)
			defer server.Close()

			req, err := http.NewRequest(method, server.URL, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Empty user ID", func(t *testing.T) {
		handler := testutils.MockDatabaseMiddleware(http.HandlerFunc(route.GetUserByID), db)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "UserID cannot be empty")
	})

	t.Run("Invalid user ID", func(t *testing.T) {
		handler := testutils.MockDatabaseMiddleware(http.HandlerFunc(route.GetUserByID), db)
		req := httptest.NewRequest(http.MethodGet, "/abc", nil)
		req = SetPathValue(req, "id", "abc")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid userID")
	})

	t.Run("Transaction was not started by middleware", func(t *testing.T) {
		handler := testutils.MockDatabaseMiddleware(http.HandlerFunc(route.GetUserByID), db)
		req := httptest.NewRequest(http.MethodGet, "/1", nil)
		req = SetPathValue(req, "id", "1")
		w := httptest.NewRecorder()

		db.Invalidate()
		handler.ServeHTTP(w, req)
		db.Validate()

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Database connection error")
	})

	t.Run("User not found", func(t *testing.T) {
		handler := testutils.MockDatabaseMiddleware(http.HandlerFunc(route.GetUserByID), db)
		req := httptest.NewRequest(http.MethodGet, "/999", nil)
		req = SetPathValue(req, "id", "999")
		w := httptest.NewRecorder()

		us.EXPECT().Get(gomock.Any(), int64(999)).Return(nil, myerrors.ErrUserNotFound).Times(1)

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "User not found")
	})

	t.Run("Internal server error", func(t *testing.T) {
		handler := testutils.MockDatabaseMiddleware(http.HandlerFunc(route.GetUserByID), db)
		req := httptest.NewRequest(http.MethodGet, "/1", nil)
		req = SetPathValue(req, "id", "1")
		w := httptest.NewRecorder()

		us.EXPECT().Get(gomock.Any(), int64(1)).Return(nil, gorm.ErrInvalidDB).Times(1)

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "User service temporarily unavailable")
	})

	t.Run("Success", func(t *testing.T) {
		handler := testutils.MockDatabaseMiddleware(http.HandlerFunc(route.GetUserByID), db)
		req := httptest.NewRequest(http.MethodGet, "/1", nil)
		req = SetPathValue(req, "id", "1")
		w := httptest.NewRecorder()

		expectedUser := &schemas.User{
			ID:      1,
			Name:    "Test",
			Surname: "User",
			Email:   "test@email.com",
			Role:    types.UserRoleStudent,
		}

		us.EXPECT().Get(gomock.Any(), int64(1)).Return(expectedUser, nil).Times(1)

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		response := &httputils.APIResponse[schemas.User]{}
		err := json.Unmarshal(w.Body.Bytes(), response)
		require.NoError(t, err)
		assert.Equal(t, expectedUser, &response.Data)
	})
}

func TestEditUser(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	us := mock_service.NewMockUserService(ctrl)
	route := routes.NewUserRoute(us)
	db := &testutils.MockDatabase{}

	currentUser := schemas.User{
		ID:      1,
		Name:    "Current",
		Surname: "User",
		Email:   "current@email.com",
		Role:    types.UserRoleAdmin,
	}

	t.Run("Accept only PATCH", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

		for _, method := range methods {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
				ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
				route.EditUser(w, r.WithContext(ctx))
			})
			server := httptest.NewServer(handler)
			defer server.Close()

			req, err := http.NewRequest(method, server.URL, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Invalid user ID", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.EditUser(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserEdit{Name: stringPtr("NewName")}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/abc", bytes.NewBuffer(jsonBody))
		req = SetPathValue(req, "id", "abc")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid userID")
	})

	t.Run("Invalid request body", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.EditUser(w, r.WithContext(ctx))
		})
		req := httptest.NewRequest(http.MethodPatch, "/1", bytes.NewBufferString("invalid json"))
		req = SetPathValue(req, "id", "1")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("Transaction was not started by middleware", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.EditUser(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserEdit{Name: stringPtr("NewName")}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/1", bytes.NewBuffer(jsonBody))
		req = SetPathValue(req, "id", "1")
		w := httptest.NewRecorder()

		db.Invalidate()
		handler.ServeHTTP(w, req)
		db.Validate()

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Database connection error")
	})

	t.Run("User not found", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.EditUser(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserEdit{Name: stringPtr("NewName")}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/999", bytes.NewBuffer(jsonBody))
		req = SetPathValue(req, "id", "999")
		w := httptest.NewRecorder()

		us.EXPECT().Edit(gomock.Any(), currentUser, int64(999), gomock.Any()).Return(myerrors.ErrUserNotFound).Times(1)

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "User not found")
	})

	t.Run("Not authorized", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.EditUser(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserEdit{Name: stringPtr("NewName")}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/2", bytes.NewBuffer(jsonBody))
		req = SetPathValue(req, "id", "2")
		w := httptest.NewRecorder()

		us.EXPECT().Edit(gomock.Any(), currentUser, int64(2), gomock.Any()).Return(myerrors.ErrForbidden).Times(1)

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "You are not authorized to edit this user")
	})

	t.Run("Not allowed", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.EditUser(w, r.WithContext(ctx))
		})
		role := types.UserRoleAdmin
		reqBody := schemas.UserEdit{Role: &role}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/2", bytes.NewBuffer(jsonBody))
		req = SetPathValue(req, "id", "2")
		w := httptest.NewRecorder()

		us.EXPECT().Edit(gomock.Any(), currentUser, int64(2), gomock.Any()).Return(myerrors.ErrNotAllowed).Times(1)

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "You are not allowed to change user role")
	})

	t.Run("Internal server error", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.EditUser(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserEdit{Name: stringPtr("NewName")}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/1", bytes.NewBuffer(jsonBody))
		req = SetPathValue(req, "id", "1")
		w := httptest.NewRecorder()

		us.EXPECT().Edit(gomock.Any(), currentUser, int64(1), gomock.Any()).Return(gorm.ErrInvalidDB).Times(1)

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "User edit service temporarily unavailable")
	})

	t.Run("Success", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.EditUser(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserEdit{Name: stringPtr("NewName")}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/1", bytes.NewBuffer(jsonBody))
		req = SetPathValue(req, "id", "1")
		w := httptest.NewRecorder()

		us.EXPECT().Edit(gomock.Any(), currentUser, int64(1), gomock.Any()).Return(nil).Times(1)

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := &httputils.APIResponse[httputils.MessageResponse]{}
		err := json.Unmarshal(w.Body.Bytes(), resp)
		require.NoError(t, err)
		assert.Equal(t, *httputils.NewMessageResponse("Update successful"), resp.Data)
	})
}

func TestChangePassword(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	us := mock_service.NewMockUserService(ctrl)
	route := routes.NewUserRoute(us)
	db := &testutils.MockDatabase{}

	currentUser := schemas.User{
		ID:      1,
		Name:    "Current",
		Surname: "User",
		Email:   "current@email.com",
		Role:    types.UserRoleAdmin,
	}

	t.Run("Accept only PATCH", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

		for _, method := range methods {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
				ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
				route.ChangePassword(w, r.WithContext(ctx))
			})
			server := httptest.NewServer(handler)
			defer server.Close()

			req, err := http.NewRequest(method, server.URL, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Empty user ID", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.ChangePassword(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserChangePassword{
			OldPassword:        "OldPass123!",
			NewPassword:        "NewPass123!",
			NewPasswordConfirm: "NewPass123!",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/password", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "UserID cannot be empty")
	})

	t.Run("Invalid user ID", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.ChangePassword(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserChangePassword{
			OldPassword:        "OldPass123!",
			NewPassword:        "NewPass123!",
			NewPasswordConfirm: "NewPass123!",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/abc/password", bytes.NewBuffer(jsonBody))
		req = SetPathValue(req, "id", "abc")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid userID")
	})

	t.Run("Invalid request body", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.ChangePassword(w, r.WithContext(ctx))
		})
		req := httptest.NewRequest(http.MethodPatch, "/1/password", bytes.NewBufferString("invalid json"))
		req = SetPathValue(req, "id", "1")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("Transaction was not started by middleware", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.ChangePassword(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserChangePassword{
			OldPassword:        "OldPass123!",
			NewPassword:        "NewPass123!",
			NewPasswordConfirm: "NewPass123!",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/1/password", bytes.NewBuffer(jsonBody))
		req = SetPathValue(req, "id", "1")
		w := httptest.NewRecorder()

		db.Invalidate()
		handler.ServeHTTP(w, req)
		db.Validate()

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Database connection error")
	})

	t.Run("User not found", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.ChangePassword(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserChangePassword{
			OldPassword:        "OldPass123!",
			NewPassword:        "NewPass123!",
			NewPasswordConfirm: "NewPass123!",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/999/password", bytes.NewBuffer(jsonBody))
		req = SetPathValue(req, "id", "999")
		w := httptest.NewRecorder()

		us.EXPECT().ChangePassword(gomock.Any(), currentUser, int64(999), gomock.Any()).Return(myerrors.ErrUserNotFound).Times(1)

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "User not found")
	})

	t.Run("Not authorized", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.ChangePassword(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserChangePassword{
			OldPassword:        "OldPass123!",
			NewPassword:        "NewPass123!",
			NewPasswordConfirm: "NewPass123!",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/2/password", bytes.NewBuffer(jsonBody))
		req = SetPathValue(req, "id", "2")
		w := httptest.NewRecorder()

		us.EXPECT().ChangePassword(gomock.Any(), currentUser, int64(2), gomock.Any()).Return(myerrors.ErrForbidden).Times(1)

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "You are not authorized to edit this user")
	})

	t.Run("Not allowed", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.ChangePassword(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserChangePassword{
			OldPassword:        "OldPass123!",
			NewPassword:        "NewPass123!",
			NewPasswordConfirm: "NewPass123!",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/2/password", bytes.NewBuffer(jsonBody))
		req = SetPathValue(req, "id", "2")
		w := httptest.NewRecorder()

		us.EXPECT().ChangePassword(gomock.Any(), currentUser, int64(2), gomock.Any()).Return(myerrors.ErrNotAllowed).Times(1)

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "You are not allowed to change user role")
	})

	t.Run("Invalid credentials", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.ChangePassword(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserChangePassword{
			OldPassword:        "WrongOldPass123!",
			NewPassword:        "NewPass123!",
			NewPasswordConfirm: "NewPass123!",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/1/password", bytes.NewBuffer(jsonBody))
		req = SetPathValue(req, "id", "1")
		w := httptest.NewRecorder()

		us.EXPECT().ChangePassword(gomock.Any(), currentUser, int64(1), gomock.Any()).Return(myerrors.ErrInvalidCredentials).Times(1)

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid old password")
	})

	t.Run("Invalid data - passwords don't match", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.ChangePassword(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserChangePassword{
			OldPassword:        "OldPass123!",
			NewPassword:        "NewPass123!",
			NewPasswordConfirm: "NewPass123!",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/1/password", bytes.NewBuffer(jsonBody))
		req = SetPathValue(req, "id", "1")
		w := httptest.NewRecorder()

		us.EXPECT().ChangePassword(gomock.Any(), currentUser, int64(1), gomock.Any()).Return(myerrors.ErrInvalidData).Times(1)

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "New password and confirm password do not match")
	})

	t.Run("Internal server error", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.ChangePassword(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserChangePassword{
			OldPassword:        "OldPass123!",
			NewPassword:        "NewPass123!",
			NewPasswordConfirm: "NewPass123!",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/1/password", bytes.NewBuffer(jsonBody))
		req = SetPathValue(req, "id", "1")
		w := httptest.NewRecorder()

		us.EXPECT().ChangePassword(gomock.Any(), currentUser, int64(1), gomock.Any()).Return(gorm.ErrInvalidDB).Times(1)

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Password change service temporarily unavailable")
	})

	t.Run("Success", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.ChangePassword(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserChangePassword{
			OldPassword:        "OldPass123!",
			NewPassword:        "NewPass123!",
			NewPasswordConfirm: "NewPass123!",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/1/password", bytes.NewBuffer(jsonBody))
		req = SetPathValue(req, "id", "1")
		w := httptest.NewRecorder()

		us.EXPECT().ChangePassword(gomock.Any(), currentUser, int64(1), gomock.Any()).Return(nil).Times(1)

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Password changed successfully")
	})
}

// Helper function
func TestGetMe(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	us := mock_service.NewMockUserService(ctrl)
	route := routes.NewUserRoute(us)

	currentUser := schemas.User{
		ID:       1,
		Name:     "John",
		Surname:  "Doe",
		Email:    "john.doe@example.com",
		Username: "johndoe",
		Role:     types.UserRoleStudent,
	}

	t.Run("Accept only GET", func(t *testing.T) {
		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), httputils.UserKey, currentUser)
				route.GetMe(w, r.WithContext(ctx))
			})

			req := httptest.NewRequest(method, "/user/me", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

			var response httputils.APIError
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "Method not allowed", response.Data.Message)
		}
	})

	t.Run("Success - Return current user", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.UserKey, currentUser)
			route.GetMe(w, r.WithContext(ctx))
		})

		req := httptest.NewRequest(http.MethodGet, "/user/me", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response httputils.APIResponse[schemas.User]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Ok)
		assert.Equal(t, currentUser, response.Data)
	})

	t.Run("Success - Different user role", func(t *testing.T) {
		adminUser := schemas.User{
			ID:       2,
			Name:     "Admin",
			Surname:  "User",
			Email:    "admin@example.com",
			Username: "admin",
			Role:     types.UserRoleAdmin,
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.UserKey, adminUser)
			route.GetMe(w, r.WithContext(ctx))
		})

		req := httptest.NewRequest(http.MethodGet, "/user/me", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response httputils.APIResponse[schemas.User]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Ok)
		assert.Equal(t, adminUser, response.Data)
		assert.Equal(t, types.UserRoleAdmin, response.Data.Role)
	})

	t.Run("Error - User context missing", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Don't set user context at all
			route.GetMe(w, r)
		})

		req := httptest.NewRequest(http.MethodGet, "/user/me", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response httputils.APIError
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Could not retrieve user. Verify that you are logged in.", response.Data.Message)
	})

	t.Run("Error - User context wrong type", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set user context with wrong type (string instead of schemas.User)
			ctx := context.WithValue(r.Context(), httputils.UserKey, "invalid_user_data")
			route.GetMe(w, r.WithContext(ctx))
		})

		req := httptest.NewRequest(http.MethodGet, "/user/me", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response httputils.APIError
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Could not retrieve user. Verify that you are logged in.", response.Data.Message)
	})
}

func stringPtr(s string) *string {
	return &s
}

func SetPathValue(r *http.Request, name, value string) *http.Request {
	r = mux.SetURLVars(r, map[string]string{name: value})
	return r
}
