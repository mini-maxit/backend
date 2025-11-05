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
	myerrors "github.com/mini-maxit/backend/package/errors"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

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
				route.ChangeMyPassword(w, r.WithContext(ctx))
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

	t.Run("Invalid request body", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.ChangeMyPassword(w, r.WithContext(ctx))
		})
		req := httptest.NewRequest(http.MethodPatch, "/me/password", bytes.NewBufferString("invalid json"))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("Transaction was not started by middleware", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.ChangeMyPassword(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserChangePassword{
			OldPassword:        "OldPass123!",
			NewPassword:        "NewPass123!",
			NewPasswordConfirm: "NewPass123!",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/me/password", bytes.NewBuffer(jsonBody))
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
			route.ChangeMyPassword(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserChangePassword{
			OldPassword:        "OldPass123!",
			NewPassword:        "NewPass123!",
			NewPasswordConfirm: "NewPass123!",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/me/password", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		us.EXPECT().ChangePassword(gomock.Any(), currentUser, currentUser.ID, gomock.Any()).Return(myerrors.ErrUserNotFound).Times(1)

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "User not found")
	})

	t.Run("Invalid credentials", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), httputils.DatabaseKey, db)
			ctx = context.WithValue(ctx, httputils.UserKey, currentUser)
			route.ChangeMyPassword(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserChangePassword{
			OldPassword:        "WrongOldPass123!",
			NewPassword:        "NewPass123!",
			NewPasswordConfirm: "NewPass123!",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/me/password", bytes.NewBuffer(jsonBody))
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
			route.ChangeMyPassword(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserChangePassword{
			OldPassword:        "OldPass123!",
			NewPassword:        "NewPass123!",
			NewPasswordConfirm: "NewPass123!",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/me/password", bytes.NewBuffer(jsonBody))
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
			route.ChangeMyPassword(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserChangePassword{
			OldPassword:        "OldPass123!",
			NewPassword:        "NewPass123!",
			NewPasswordConfirm: "NewPass123!",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/me/password", bytes.NewBuffer(jsonBody))
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
			route.ChangeMyPassword(w, r.WithContext(ctx))
		})
		reqBody := schemas.UserChangePassword{
			OldPassword:        "OldPass123!",
			NewPassword:        "NewPass123!",
			NewPasswordConfirm: "NewPass123!",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/me/password", bytes.NewBuffer(jsonBody))
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
