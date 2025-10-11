package routes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/api/http/routes"
	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/schemas"
	myerrors "github.com/mini-maxit/backend/package/errors"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestGetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ss := mock_service.NewMockSubmissionService(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	qs := mock_service.NewMockQueueService(ctrl)
	route := routes.NewSubmissionRoutes(ss, "http://filestorage", qs, ts)
	db := &testutils.MockDatabase{}
	handler := testutils.MockDatabaseMiddleware(http.HandlerFunc(route.GetAll), db)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, httputils.UserKey, mockUser)
		ctx = context.WithValue(ctx, httputils.QueryParamsKey, map[string]interface{}{})
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	t.Run("Accept only GET", func(t *testing.T) {
		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL, nil)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Transaction was not started by middleware", func(t *testing.T) {
		db.Invalidate()
		resp, err := http.Get(server.URL)
		require.NoError(t, err)
		db.Validate()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Transaction was not started by middleware")
	})

	t.Run("Internal server error", func(t *testing.T) {
		ss.EXPECT().GetAll(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, gorm.ErrInvalidDB).Times(1)

		resp, err := http.Get(server.URL)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Internal Server Error")
	})

	t.Run("Success with empty list", func(t *testing.T) {
		ss.EXPECT().GetAll(gomock.Any(), gomock.Any(), gomock.Any()).Return([]schemas.Submission{}, nil).Times(1)

		resp, err := http.Get(server.URL)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		response := &httputils.APIResponse[[]schemas.Submission]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Empty(t, response.Data)
	})

	t.Run("Success with submissions", func(t *testing.T) {
		submissions := []schemas.Submission{
			{
				ID:     1,
				TaskID: 1,
				UserID: 1,
				Status: "completed",
			},
			{
				ID:     2,
				TaskID: 2,
				UserID: 1,
				Status: "pending",
			},
		}
		ss.EXPECT().GetAll(gomock.Any(), gomock.Any(), gomock.Any()).Return(submissions, nil).Times(1)

		resp, err := http.Get(server.URL)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		response := &httputils.APIResponse[[]schemas.Submission]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Equal(t, submissions, response.Data)
	})
}

func TestGetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ss := mock_service.NewMockSubmissionService(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	qs := mock_service.NewMockQueueService(ctrl)
	route := routes.NewSubmissionRoutes(ss, "http://filestorage", qs, ts)
	db := &testutils.MockDatabase{}

	mux := http.NewServeMux()
	mux.HandleFunc("/{id}", route.GetByID)
	handler := testutils.MockDatabaseMiddleware(mux, db)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	t.Run("Accept only GET", func(t *testing.T) {
		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL+"/1", nil)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Invalid submission ID", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/invalid")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Invalid submission id")
	})

	t.Run("Transaction was not started by middleware", func(t *testing.T) {
		db.Invalidate()
		resp, err := http.Get(server.URL + "/1")
		require.NoError(t, err)
		db.Validate()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Transaction was not started by middleware")
	})

	t.Run("Internal server error", func(t *testing.T) {
		ss.EXPECT().Get(gomock.Any(), int64(1), gomock.Any()).Return(schemas.Submission{}, gorm.ErrInvalidDB).Times(1)

		resp, err := http.Get(server.URL + "/1")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Internal Server Error")
	})

	t.Run("Success", func(t *testing.T) {
		submission := schemas.Submission{
			ID:     1,
			TaskID: 1,
			UserID: 1,
			Status: "completed",
		}
		ss.EXPECT().Get(gomock.Any(), int64(1), gomock.Any()).Return(submission, nil).Times(1)

		resp, err := http.Get(server.URL + "/1")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		response := &httputils.APIResponse[schemas.Submission]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Equal(t, submission, response.Data)
	})
}

func TestGetAllForUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ss := mock_service.NewMockSubmissionService(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	qs := mock_service.NewMockQueueService(ctrl)
	route := routes.NewSubmissionRoutes(ss, "http://filestorage", qs, ts)
	db := &testutils.MockDatabase{}

	mux := http.NewServeMux()
	mux.HandleFunc("/user/{id}", route.GetAllForUser)
	handler := testutils.MockDatabaseMiddleware(mux, db)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, httputils.UserKey, mockUser)
		ctx = context.WithValue(ctx, httputils.QueryParamsKey, map[string]interface{}{})
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	t.Run("Accept only GET", func(t *testing.T) {
		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL+"/user/1", nil)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Invalid user ID", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/user/invalid")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Invalid user id")
	})

	t.Run("Permission denied", func(t *testing.T) {
		ss.EXPECT().GetAllForUser(gomock.Any(), int64(2), gomock.Any(), gomock.Any()).Return(nil, myerrors.ErrPermissionDenied).Times(1)

		resp, err := http.Get(server.URL + "/user/2")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Permission denied")
	})

	t.Run("Internal server error", func(t *testing.T) {
		ss.EXPECT().GetAllForUser(gomock.Any(), int64(1), gomock.Any(), gomock.Any()).Return(nil, gorm.ErrInvalidDB).Times(1)

		resp, err := http.Get(server.URL + "/user/1")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Internal Server Error")
	})

	t.Run("Success", func(t *testing.T) {
		submissions := []schemas.Submission{
			{ID: 1, UserID: 1, Status: "completed"},
			{ID: 2, UserID: 1, Status: "pending"},
		}
		ss.EXPECT().GetAllForUser(gomock.Any(), int64(1), gomock.Any(), gomock.Any()).Return(submissions, nil).Times(1)

		resp, err := http.Get(server.URL + "/user/1")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		response := &httputils.APIResponse[[]schemas.Submission]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Equal(t, submissions, response.Data)
	})
}

func TestGetAllForUserShort(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ss := mock_service.NewMockSubmissionService(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	qs := mock_service.NewMockQueueService(ctrl)
	route := routes.NewSubmissionRoutes(ss, "http://filestorage", qs, ts)
	db := &testutils.MockDatabase{}

	mux := http.NewServeMux()
	mux.HandleFunc("/user/{id}/short", route.GetAllForUserShort)
	handler := testutils.MockDatabaseMiddleware(mux, db)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, httputils.UserKey, mockUser)
		ctx = context.WithValue(ctx, httputils.QueryParamsKey, map[string]interface{}{})
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	t.Run("Accept only GET", func(t *testing.T) {
		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL+"/user/1/short", nil)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Invalid user ID", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/user/invalid/short")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Invalid user id")
	})

	t.Run("Permission denied", func(t *testing.T) {
		ss.EXPECT().GetAllForUserShort(gomock.Any(), int64(2), gomock.Any(), gomock.Any()).Return(nil, myerrors.ErrPermissionDenied).Times(1)

		resp, err := http.Get(server.URL + "/user/2/short")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Permission denied")
	})

	t.Run("Success", func(t *testing.T) {
		submissions := []schemas.SubmissionShort{
			{ID: 1, UserID: 1, TaskID: 1, Passed: true, HowManyPassed: 10},
			{ID: 2, UserID: 1, TaskID: 2, Passed: false, HowManyPassed: 5},
		}
		ss.EXPECT().GetAllForUserShort(gomock.Any(), int64(1), gomock.Any(), gomock.Any()).Return(submissions, nil).Times(1)

		resp, err := http.Get(server.URL + "/user/1/short")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		response := &httputils.APIResponse[[]schemas.SubmissionShort]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Equal(t, submissions, response.Data)
	})
}

func TestGetAllForGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ss := mock_service.NewMockSubmissionService(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	qs := mock_service.NewMockQueueService(ctrl)
	route := routes.NewSubmissionRoutes(ss, "http://filestorage", qs, ts)
	db := &testutils.MockDatabase{}

	mux := http.NewServeMux()
	mux.HandleFunc("/group/{id}", route.GetAllForGroup)
	handler := testutils.MockDatabaseMiddleware(mux, db)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, httputils.UserKey, mockUser)
		ctx = context.WithValue(ctx, httputils.QueryParamsKey, map[string]interface{}{})
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	t.Run("Accept only GET", func(t *testing.T) {
		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL+"/group/1", nil)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Invalid group ID", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/group/invalid")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Invalid group id")
	})

	t.Run("Internal server error", func(t *testing.T) {
		ss.EXPECT().GetAllForGroup(gomock.Any(), int64(1), gomock.Any(), gomock.Any()).Return(nil, gorm.ErrInvalidDB).Times(1)

		resp, err := http.Get(server.URL + "/group/1")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Internal Server Error")
	})

	t.Run("Success", func(t *testing.T) {
		submissions := []schemas.Submission{
			{ID: 1, TaskID: 1, Status: "completed"},
			{ID: 2, TaskID: 1, Status: "pending"},
		}
		ss.EXPECT().GetAllForGroup(gomock.Any(), int64(1), gomock.Any(), gomock.Any()).Return(submissions, nil).Times(1)

		resp, err := http.Get(server.URL + "/group/1")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		response := &httputils.APIResponse[[]schemas.Submission]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Equal(t, submissions, response.Data)
	})
}

func TestGetAllForTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ss := mock_service.NewMockSubmissionService(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	qs := mock_service.NewMockQueueService(ctrl)
	route := routes.NewSubmissionRoutes(ss, "http://filestorage", qs, ts)
	db := &testutils.MockDatabase{}

	mux := http.NewServeMux()
	mux.HandleFunc("/task/{id}", route.GetAllForTask)
	handler := testutils.MockDatabaseMiddleware(mux, db)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, httputils.UserKey, mockUser)
		ctx = context.WithValue(ctx, httputils.QueryParamsKey, map[string]interface{}{})
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	t.Run("Accept only GET", func(t *testing.T) {
		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL+"/task/1", nil)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Invalid task ID", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/task/invalid")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Invalid task id")
	})

	t.Run("Internal server error", func(t *testing.T) {
		ss.EXPECT().GetAllForTask(gomock.Any(), int64(1), gomock.Any(), gomock.Any()).Return(nil, gorm.ErrInvalidDB).Times(1)

		resp, err := http.Get(server.URL + "/task/1")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Internal Server Error")
	})

	t.Run("Success", func(t *testing.T) {
		submissions := []schemas.Submission{
			{ID: 1, TaskID: 1, UserID: 1, Status: "completed"},
			{ID: 2, TaskID: 1, UserID: 2, Status: "pending"},
		}
		ss.EXPECT().GetAllForTask(gomock.Any(), int64(1), gomock.Any(), gomock.Any()).Return(submissions, nil).Times(1)

		resp, err := http.Get(server.URL + "/task/1")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		response := &httputils.APIResponse[[]schemas.Submission]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Equal(t, submissions, response.Data)
	})
}

func TestGetAvailableLanguages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ss := mock_service.NewMockSubmissionService(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	qs := mock_service.NewMockQueueService(ctrl)
	route := routes.NewSubmissionRoutes(ss, "http://filestorage", qs, ts)
	db := &testutils.MockDatabase{}
	handler := testutils.MockDatabaseMiddleware(http.HandlerFunc(route.GetAvailableLanguages), db)

	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("Transaction was not started by middleware", func(t *testing.T) {
		db.Invalidate()
		resp, err := http.Get(server.URL)
		require.NoError(t, err)
		db.Validate()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Transaction was not started by middleware")
	})

	t.Run("Internal server error", func(t *testing.T) {
		ss.EXPECT().GetAvailableLanguages(gomock.Any()).Return(nil, gorm.ErrInvalidDB).Times(1)

		resp, err := http.Get(server.URL)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Internal Server Error")
	})

	t.Run("Success", func(t *testing.T) {
		languages := []schemas.LanguageConfig{
			{ID: 1, Type: "Go", Version: "1.21", FileExtension: ".go"},
			{ID: 2, Type: "Python", Version: "3.11", FileExtension: ".py"},
		}
		ss.EXPECT().GetAvailableLanguages(gomock.Any()).Return(languages, nil).Times(1)

		resp, err := http.Get(server.URL)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		response := &httputils.APIResponse[[]schemas.LanguageConfig]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Equal(t, languages, response.Data)
	})
}

func TestSubmitSolution(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ss := mock_service.NewMockSubmissionService(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	qs := mock_service.NewMockQueueService(ctrl)
	route := routes.NewSubmissionRoutes(ss, "http://filestorage", qs, ts)
	db := &testutils.MockDatabase{}
	handler := testutils.MockDatabaseMiddleware(http.HandlerFunc(route.SubmitSolution), db)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "student",
			Email: "test@example.com",
		}
		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	t.Run("Accept only POST", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL, nil)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Missing task ID", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req, err := http.NewRequest(http.MethodPost, server.URL, body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Task ID is required")
	})

	t.Run("Invalid task ID", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.WriteField("taskID", "invalid")
		writer.Close()

		req, err := http.NewRequest(http.MethodPost, server.URL, body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Invalid task ID")
	})

	t.Run("Missing solution file", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.WriteField("taskID", "1")
		writer.Close()

		req, err := http.NewRequest(http.MethodPost, server.URL, body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "No solution file found")
	})

	t.Run("Missing language ID", func(t *testing.T) {
		// Create a temporary file for testing
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "solution.go")
		err := os.WriteFile(tmpFile, []byte("package main\nfunc main() {}"), 0644)
		require.NoError(t, err)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.WriteField("taskID", "1")

		file, err := os.Open(tmpFile)
		require.NoError(t, err)
		defer file.Close()

		part, err := writer.CreateFormFile("solution", filepath.Base(tmpFile))
		require.NoError(t, err)
		_, err = io.Copy(part, file)
		require.NoError(t, err)
		writer.Close()

		req, err := http.NewRequest(http.MethodPost, server.URL, body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Language ID is required")
	})

	t.Run("Invalid language ID", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "solution.go")
		err := os.WriteFile(tmpFile, []byte("package main\nfunc main() {}"), 0644)
		require.NoError(t, err)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.WriteField("taskID", "1")
		writer.WriteField("languageID", "invalid")

		file, err := os.Open(tmpFile)
		require.NoError(t, err)
		defer file.Close()

		part, err := writer.CreateFormFile("solution", filepath.Base(tmpFile))
		require.NoError(t, err)
		_, err = io.Copy(part, file)
		require.NoError(t, err)
		writer.Close()

		req, err := http.NewRequest(http.MethodPost, server.URL, body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Invalid language ID")
	})

	t.Run("Transaction was not started by middleware", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "solution.go")
		err := os.WriteFile(tmpFile, []byte("package main\nfunc main() {}"), 0644)
		require.NoError(t, err)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.WriteField("taskID", "1")
		writer.WriteField("languageID", "1")

		file, err := os.Open(tmpFile)
		require.NoError(t, err)
		defer file.Close()

		part, err := writer.CreateFormFile("solution", filepath.Base(tmpFile))
		require.NoError(t, err)
		_, err = io.Copy(part, file)
		require.NoError(t, err)
		writer.Close()

		db.Invalidate()
		req, err := http.NewRequest(http.MethodPost, server.URL, body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		db.Validate()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Transaction was not started by middleware")
	})

	t.Run("Task not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "solution.go")
		err := os.WriteFile(tmpFile, []byte("package main\nfunc main() {}"), 0644)
		require.NoError(t, err)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.WriteField("taskID", "999")
		writer.WriteField("languageID", "1")

		file, err := os.Open(tmpFile)
		require.NoError(t, err)
		defer file.Close()

		part, err := writer.CreateFormFile("solution", filepath.Base(tmpFile))
		require.NoError(t, err)
		_, err = io.Copy(part, file)
		require.NoError(t, err)
		writer.Close()

		ss.EXPECT().Submit(gomock.Any(), gomock.Any(), int64(999), int64(1), gomock.Any()).Return(int64(0), myerrors.ErrNotFound).Times(1)

		req, err := http.NewRequest(http.MethodPost, server.URL, body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Task not found")
	})

	t.Run("Success", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "solution.go")
		err := os.WriteFile(tmpFile, []byte("package main\nfunc main() {}"), 0644)
		require.NoError(t, err)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.WriteField("taskID", "1")
		writer.WriteField("languageID", "1")

		file, err := os.Open(tmpFile)
		require.NoError(t, err)
		defer file.Close()

		part, err := writer.CreateFormFile("solution", filepath.Base(tmpFile))
		require.NoError(t, err)
		_, err = io.Copy(part, file)
		require.NoError(t, err)
		writer.Close()

		ss.EXPECT().Submit(gomock.Any(), gomock.Any(), int64(1), int64(1), gomock.Any()).Return(int64(123), nil).Times(1)

		req, err := http.NewRequest(http.MethodPost, server.URL, body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		response := &httputils.APIResponse[map[string]int64]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Equal(t, int64(123), response.Data["submissionId"])
	})
}
