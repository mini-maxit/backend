package routes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/api/http/routes"
	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/errors"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestDeleteTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ts := mock_service.NewMockTaskService(ctrl)
	route := routes.NewTasksManagementRoute(ts)
	db := &testutils.MockDatabase{}

	router := mux.NewRouter()
	router.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		route.DeleteTask(w, r)
	})

	handler := httputils.MockDatabaseMiddleware(router, db)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "teacher",
			Email: "test@example.com",
		}
		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	t.Run("Accept only DELETE", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL+"/1", nil)
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

	t.Run("Invalid task ID", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, server.URL+"/invalid", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Task not found", func(t *testing.T) {
		ts.EXPECT().Delete(gomock.Any(), gomock.Any(), int64(999)).Return(errors.ErrNotFound)

		req, err := http.NewRequest(http.MethodDelete, server.URL+"/999", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Not authorized", func(t *testing.T) {
		ts.EXPECT().Delete(gomock.Any(), gomock.Any(), int64(1)).Return(errors.ErrForbidden)

		req, err := http.NewRequest(http.MethodDelete, server.URL+"/1", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Internal server error", func(t *testing.T) {
		ts.EXPECT().Delete(gomock.Any(), gomock.Any(), int64(1)).Return(gorm.ErrInvalidDB)

		req, err := http.NewRequest(http.MethodDelete, server.URL+"/1", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		ts.EXPECT().Delete(gomock.Any(), gomock.Any(), int64(1)).Return(nil)

		req, err := http.NewRequest(http.MethodDelete, server.URL+"/1", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response httputils.APIResponse[httputils.MessageResponse]
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		assert.True(t, response.Ok)
		assert.Equal(t, "Task deleted successfully", response.Data.Message)
	})
}

func TestEditTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ts := mock_service.NewMockTaskService(ctrl)
	route := routes.NewTasksManagementRoute(ts)
	db := &testutils.MockDatabase{}

	router := mux.NewRouter()
	router.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		route.EditTask(w, r)
	})

	handler := httputils.MockDatabaseMiddleware(router, db)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "teacher",
			Email: "test@example.com",
		}
		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	t.Run("Accept only PATCH", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL+"/1", nil)
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

	t.Run("Invalid task ID", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPatch, server.URL+"/invalid", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Task not found", func(t *testing.T) {
		ts.EXPECT().Edit(gomock.Any(), gomock.Any(), int64(999), gomock.Any()).Return(errors.ErrNotFound)

		req, err := http.NewRequest(http.MethodPatch, server.URL+"/999", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "multipart/form-data")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Not authorized", func(t *testing.T) {
		ts.EXPECT().Edit(gomock.Any(), gomock.Any(), int64(1), gomock.Any()).Return(errors.ErrForbidden)

		req, err := http.NewRequest(http.MethodPatch, server.URL+"/1", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Internal server error", func(t *testing.T) {
		ts.EXPECT().Edit(gomock.Any(), gomock.Any(), int64(1), gomock.Any()).Return(gorm.ErrInvalidDB)

		req, err := http.NewRequest(http.MethodPatch, server.URL+"/1", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("Success without archive", func(t *testing.T) {
		ts.EXPECT().Edit(gomock.Any(), gomock.Any(), int64(1), gomock.Any()).Return(nil)

		// Create a multipart form request without the archive field
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("title", "Updated Task Title")
		writer.Close()

		req, err := http.NewRequest(http.MethodPatch, server.URL+"/1", body)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response httputils.APIResponse[httputils.MessageResponse]
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		assert.True(t, response.Ok)
		assert.Equal(t, "Task updated successfully", response.Data.Message)
	})
}

func TestGetAllCreatedTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ts := mock_service.NewMockTaskService(ctrl)
	route := routes.NewTasksManagementRoute(ts)
	db := &testutils.MockDatabase{}

	handler := httputils.MockDatabaseMiddleware(http.HandlerFunc(route.GetAllCreatedTasks), db)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				t.Logf("Recovered from panic: %v", rec)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		mockUser := schemas.User{
			ID:    1,
			Role:  "teacher",
			Email: "test@example.com",
		}
		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		ctx = context.WithValue(ctx, httputils.QueryParamsKey, map[string]any{
			"limit":  10,
			"offset": 0,
			"sort":   "",
		})
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	t.Run("Accept only GET", func(t *testing.T) {
		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
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

	t.Run("Not authorized - wrong role", func(t *testing.T) {
		ts.EXPECT().GetAllCreated(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			schemas.PaginatedResult[[]schemas.Task]{},
			errors.ErrForbidden,
		)

		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Internal server error", func(t *testing.T) {
		ts.EXPECT().GetAllCreated(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			schemas.PaginatedResult[[]schemas.Task]{},
			gorm.ErrInvalidDB,
		)

		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("Success with empty list", func(t *testing.T) {
		result := schemas.NewPaginatedResult([]schemas.Task{}, 0, 10, 0)
		ts.EXPECT().GetAllCreated(gomock.Any(), gomock.Any(), gomock.Any()).Return(result, nil)

		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response httputils.APIResponse[schemas.PaginatedResult[[]schemas.Task]]
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		assert.True(t, response.Ok)
		assert.Empty(t, response.Data.Items)
	})

	t.Run("Success with tasks", func(t *testing.T) {
		now := time.Now()
		tasks := []schemas.Task{
			{
				ID:        1,
				Title:     "Test Task 1",
				CreatedBy: 1,
				CreatedAt: now,
				UpdatedAt: now,
				IsVisible: true,
			},
			{
				ID:        2,
				Title:     "Test Task 2",
				CreatedBy: 1,
				CreatedAt: now,
				UpdatedAt: now,
				IsVisible: false,
			},
		}
		result := schemas.NewPaginatedResult(tasks, 0, 10, 2)
		ts.EXPECT().GetAllCreated(gomock.Any(), gomock.Any(), gomock.Any()).Return(result, nil)

		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response httputils.APIResponse[schemas.PaginatedResult[[]schemas.Task]]
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		assert.True(t, response.Ok)
		assert.Len(t, response.Data.Items, 2)
		assert.Equal(t, "Test Task 1", response.Data.Items[0].Title)
		assert.Equal(t, "Test Task 2", response.Data.Items[1].Title)
	})
}

func TestGetLimits(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ts := mock_service.NewMockTaskService(ctrl)
	route := routes.NewTasksManagementRoute(ts)
	db := &testutils.MockDatabase{}

	router := mux.NewRouter()
	router.HandleFunc("/{id}/limits", func(w http.ResponseWriter, r *http.Request) {
		route.GetLimits(w, r)
	})

	handler := httputils.MockDatabaseMiddleware(router, db)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "teacher",
			Email: "test@example.com",
		}
		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	t.Run("Accept only GET", func(t *testing.T) {
		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL+"/1/limits", nil)
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

	t.Run("Invalid task ID", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/invalid/limits")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Task not found", func(t *testing.T) {
		ts.EXPECT().GetLimits(gomock.Any(), gomock.Any(), int64(999)).Return(nil, errors.ErrNotFound)

		resp, err := http.Get(server.URL + "/999/limits")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Internal server error", func(t *testing.T) {
		ts.EXPECT().GetLimits(gomock.Any(), gomock.Any(), int64(1)).Return(nil, gorm.ErrInvalidDB)

		resp, err := http.Get(server.URL + "/1/limits")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("Success with empty limits", func(t *testing.T) {
		ts.EXPECT().GetLimits(gomock.Any(), gomock.Any(), int64(1)).Return([]schemas.TestCase{}, nil)

		resp, err := http.Get(server.URL + "/1/limits")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response httputils.APIResponse[[]schemas.TestCase]
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		assert.True(t, response.Ok)
		assert.Empty(t, response.Data)
	})

	t.Run("Success with limits", func(t *testing.T) {
		limits := []schemas.TestCase{
			{
				Order:       1,
				TimeLimit:   1000,
				MemoryLimit: 256,
			},
			{
				Order:       2,
				TimeLimit:   2000,
				MemoryLimit: 512,
			},
		}
		ts.EXPECT().GetLimits(gomock.Any(), gomock.Any(), int64(1)).Return(limits, nil)

		resp, err := http.Get(server.URL + "/1/limits")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response httputils.APIResponse[[]schemas.TestCase]
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		assert.True(t, response.Ok)
		assert.Len(t, response.Data, 2)
		assert.Equal(t, 1, response.Data[0].Order)
		assert.Equal(t, int64(1000), response.Data[0].TimeLimit)
		assert.Equal(t, int64(256), response.Data[0].MemoryLimit)
	})
}

func TestPutLimits(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ts := mock_service.NewMockTaskService(ctrl)
	route := routes.NewTasksManagementRoute(ts)
	db := &testutils.MockDatabase{}

	router := mux.NewRouter()
	router.HandleFunc("/{id}/limits", func(w http.ResponseWriter, r *http.Request) {
		route.PutLimits(w, r)
	})

	handler := httputils.MockDatabaseMiddleware(router, db)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "teacher",
			Email: "test@example.com",
		}
		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	t.Run("Accept only PUT", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL+"/1/limits", nil)
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

	t.Run("Invalid task ID", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPut, server.URL+"/invalid/limits", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Invalid request body", func(t *testing.T) {
		invalidBodies := []string{
			`{invalid json}`,
			`{"limits": "not an array"}`,
		}

		for _, body := range invalidBodies {
			req, err := http.NewRequest(http.MethodPut, server.URL+"/1/limits", bytes.NewBufferString(body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("Task not found", func(t *testing.T) {
		body := schemas.PutTestCaseLimitsRequest{
			Limits: []schemas.PutTestCase{
				{Order: 1, TimeLimit: 1000, MemoryLimit: 256},
			},
		}
		jsonBody, _ := json.Marshal(body)

		ts.EXPECT().PutLimits(gomock.Any(), gomock.Any(), int64(999), gomock.Any()).Return(errors.ErrNotFound)

		req, err := http.NewRequest(http.MethodPut, server.URL+"/999/limits", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Not authorized", func(t *testing.T) {
		body := schemas.PutTestCaseLimitsRequest{
			Limits: []schemas.PutTestCase{
				{Order: 1, TimeLimit: 1000, MemoryLimit: 256},
			},
		}
		jsonBody, _ := json.Marshal(body)

		ts.EXPECT().PutLimits(gomock.Any(), gomock.Any(), int64(1), gomock.Any()).Return(errors.ErrForbidden)

		req, err := http.NewRequest(http.MethodPut, server.URL+"/1/limits", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Internal server error", func(t *testing.T) {
		body := schemas.PutTestCaseLimitsRequest{
			Limits: []schemas.PutTestCase{
				{Order: 1, TimeLimit: 1000, MemoryLimit: 256},
			},
		}
		jsonBody, _ := json.Marshal(body)

		ts.EXPECT().PutLimits(gomock.Any(), gomock.Any(), int64(1), gomock.Any()).Return(gorm.ErrInvalidDB)

		req, err := http.NewRequest(http.MethodPut, server.URL+"/1/limits", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		body := schemas.PutTestCaseLimitsRequest{
			Limits: []schemas.PutTestCase{
				{Order: 1, TimeLimit: 1000, MemoryLimit: 256},
				{Order: 2, TimeLimit: 2000, MemoryLimit: 512},
			},
		}
		jsonBody, _ := json.Marshal(body)

		ts.EXPECT().PutLimits(gomock.Any(), gomock.Any(), int64(1), gomock.Any()).Return(nil)

		req, err := http.NewRequest(http.MethodPut, server.URL+"/1/limits", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		var response httputils.APIResponse[httputils.MessageResponse]
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		assert.True(t, response.Ok)
		assert.Equal(t, "Task limits updated successfully", response.Data.Message)
	})
}
