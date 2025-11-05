package routes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/api/http/routes"
	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/schemas"
	myerrors "github.com/mini-maxit/backend/package/errors"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type teacherRouteTest struct {
	Route             routes.TeacherRoute
	ContestSevice     *mock_service.MockContestService
	TaskService       *mock_service.MockTaskService
	GroupService      *mock_service.MockGroupService
	SubmissionService *mock_service.MockSubmissionService
	Mux               *mux.Router
}

func setupTeacherRoute(t *testing.T) (*teacherRouteTest, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	contestService := mock_service.NewMockContestService(ctrl)
	taskService := mock_service.NewMockTaskService(ctrl)
	groupService := mock_service.NewMockGroupService(ctrl)
	submissionService := mock_service.NewMockSubmissionService(ctrl)

	route := routes.NewTeacherRoute(contestService, taskService, groupService, submissionService)
	mux := mux.NewRouter()
	routes.RegisterTeacherRoutes(mux, route)

	return &teacherRouteTest{
		Route:             route,
		ContestSevice:     contestService,
		TaskService:       taskService,
		GroupService:      groupService,
		SubmissionService: submissionService,
		Mux:               mux,
	}, ctrl
}

func TestEditContest(t *testing.T) {
	setup, ctrl := setupTeacherRoute(t)
	defer ctrl.Finish()

	path := "/contests/1"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		db := &testutils.MockDatabase{}
		ctx = context.WithValue(ctx, httputils.DatabaseKey, db)
		setup.Mux.ServeHTTP(w, r.WithContext(ctx))
	}))

	doRequest := func(t *testing.T, method string, url string, body io.Reader) *http.Response {
		req, err := http.NewRequest(method, url, body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		return resp
	}

	defer server.Close()
	newName := "Updated Contest"

	t.Run("Invalid request body", func(t *testing.T) {
		invalidBodies := []string{
			`{invalid json}`,
		}

		for _, body := range invalidBodies {
			resp := doRequest(t, http.MethodPut, server.URL+path, bytes.NewBufferString(body))
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("Contest not found", func(t *testing.T) {
		body := schemas.EditContest{
			Name: &newName,
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		setup.ContestSevice.EXPECT().Edit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, myerrors.ErrNotFound)

		resp := doRequest(t, http.MethodPut, server.URL+path, bytes.NewBuffer(jsonBody))
		defer resp.Body.Close()
		responseBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, string(responseBody))
	})

	t.Run("Not authorized", func(t *testing.T) {
		body := schemas.EditContest{
			Name: &newName,
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		setup.ContestSevice.EXPECT().Edit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, myerrors.ErrNotAuthorized)

		resp := doRequest(t, http.MethodPut, server.URL+path, bytes.NewBuffer(jsonBody))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		body := schemas.EditContest{
			Name: &newName,
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		contest := &schemas.Contest{
			ID:          1,
			Name:        "Updated Contest",
			Description: "Test Description",
			CreatedBy:   1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		setup.ContestSevice.EXPECT().Edit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(contest, nil)

		resp := doRequest(t, http.MethodPut, server.URL+path, bytes.NewBuffer(jsonBody))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestDeleteContest(t *testing.T) {
	setup, ctrl := setupTeacherRoute(t)
	defer ctrl.Finish()

	path := "/contests/1"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		db := &testutils.MockDatabase{}
		ctx = context.WithValue(ctx, httputils.DatabaseKey, db)
		setup.Mux.ServeHTTP(w, r.WithContext(ctx))
	}))

	doRequest := func(t *testing.T, method string, url string) *http.Response {
		req, err := http.NewRequest(method, url, nil)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		return resp
	}
	defer server.Close()

	t.Run("Contest not found", func(t *testing.T) {
		setup.ContestSevice.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(myerrors.ErrNotFound)

		resp := doRequest(t, http.MethodDelete, server.URL+path)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Not authorized", func(t *testing.T) {
		setup.ContestSevice.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(myerrors.ErrNotAuthorized)

		resp := doRequest(t, http.MethodDelete, server.URL+path)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		setup.ContestSevice.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		resp := doRequest(t, http.MethodDelete, server.URL+path)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestCreateContest(t *testing.T) {
	setup, ctrl := setupTeacherRoute(t)
	defer ctrl.Finish()
	db := &testutils.MockDatabase{}
	path := "/contests"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, httputils.UserKey, mockUser)
		ctx = context.WithValue(ctx, httputils.DatabaseKey, db)
		setup.Mux.ServeHTTP(w, r.WithContext(ctx))
	}))
	doRequest := func(t *testing.T, url string, body io.Reader) *http.Response {
		resp, err := http.Post(url, "application/json", body)
		require.NoError(t, err)
		return resp
	}
	defer server.Close()

	t.Run("Invalid request body", func(t *testing.T) {
		invalidBodies := []string{
			`{"name": "Test Contest", "description": "Test", "extra": "field"}`,
			`{"name": "Test Contest", "description": "Test"}{"extra": "json"}`,
		}

		for _, body := range invalidBodies {
			t.Logf("Testing with body: %s", body)
			resp := doRequest(t, server.URL+path, strings.NewReader(body))
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}
			bodyString := string(bodyBytes)

			assert.Contains(t, bodyString, "Could not validate request data")
		}
	})

	t.Run("Transaction was not started by middleware", func(t *testing.T) {
		body := schemas.CreateContest{
			Name:        "Test Contest",
			Description: "Test Description",
		}
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		db.Invalidate()
		resp := doRequest(t, server.URL+path, bytes.NewBuffer(jsonBody))
		db.Validate()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		bodyString := string(bodyBytes)

		assert.Contains(t, bodyString, "Database connection error")
	})

	t.Run("Not authorized", func(t *testing.T) {
		body := schemas.CreateContest{
			Name:        "Test Contest",
			Description: "Test Description",
		}
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)

		setup.ContestSevice.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), myerrors.ErrNotAuthorized)

		resp := doRequest(t, server.URL+path, bytes.NewBuffer(jsonBody))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		body := schemas.CreateContest{
			Name:        "Test Contest",
			Description: "Test Description",
		}
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)

		setup.ContestSevice.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)

		resp := doRequest(t, server.URL+path, bytes.NewBuffer(jsonBody))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestApproveRegistrationRequest(t *testing.T) {
	setup, ctrl := setupTeacherRoute(t)
	defer ctrl.Finish()
	db := &testutils.MockDatabase{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		ctx = context.WithValue(ctx, httputils.DatabaseKey, db)
		setup.Mux.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()
	doRequest := func(t *testing.T, url string) *http.Response {
		resp, err := http.Post(url, "application/json", nil)
		require.NoError(t, err)
		return resp
	}

	t.Run("Invalid contest ID", func(t *testing.T) {
		resp := doRequest(t, server.URL+"/contests/invalid/registration-requests/2/approve")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Invalid user ID", func(t *testing.T) {
		resp := doRequest(t, server.URL+"/contests/1/registration-requests/invalid/approve")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Contest not found", func(t *testing.T) {
		setup.ContestSevice.EXPECT().ApproveRegistrationRequest(gomock.Any(), gomock.Any(), int64(999), int64(2)).Return(myerrors.ErrNotFound)

		resp := doRequest(t, server.URL+"/contests/999/registration-requests/2/approve")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("User not found", func(t *testing.T) {
		setup.ContestSevice.EXPECT().ApproveRegistrationRequest(gomock.Any(), gomock.Any(), int64(1), int64(999)).Return(myerrors.ErrNotFound)

		resp := doRequest(t, server.URL+"/contests/1/registration-requests/999/approve")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Not authorized", func(t *testing.T) {
		setup.ContestSevice.EXPECT().ApproveRegistrationRequest(gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(myerrors.ErrNotAuthorized)

		resp := doRequest(t, server.URL+"/contests/1/registration-requests/2/approve")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("No pending registration", func(t *testing.T) {
		setup.ContestSevice.EXPECT().ApproveRegistrationRequest(gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(myerrors.ErrNoPendingRegistration)

		resp := doRequest(t, server.URL+"/contests/1/registration-requests/2/approve")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("User already participant", func(t *testing.T) {
		setup.ContestSevice.EXPECT().ApproveRegistrationRequest(gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(myerrors.ErrAlreadyParticipant)

		resp := doRequest(t, server.URL+"/contests/1/registration-requests/2/approve")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		setup.ContestSevice.EXPECT().ApproveRegistrationRequest(gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(nil)

		resp := doRequest(t, server.URL+"/contests/1/registration-requests/2/approve")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response httputils.APIResponse[httputils.MessageResponse]
		err := json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.True(t, response.Ok)
		assert.Equal(t, "Registration request approved successfully", response.Data.Message)
	})
}

func TestRejectRegistrationRequest(t *testing.T) {
	setup, ctrl := setupTeacherRoute(t)
	defer ctrl.Finish()
	db := &testutils.MockDatabase{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		ctx = context.WithValue(ctx, httputils.DatabaseKey, db)
		setup.Mux.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()
	doRequest := func(t *testing.T, url string) *http.Response {
		resp, err := http.Post(url, "application/json", nil)
		require.NoError(t, err)
		return resp
	}

	t.Run("Invalid contest ID", func(t *testing.T) {
		resp := doRequest(t, server.URL+"/contests/invalid/registration-requests/2/reject")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Invalid user ID", func(t *testing.T) {
		resp := doRequest(t, server.URL+"/contests/1/registration-requests/invalid/reject")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Contest not found", func(t *testing.T) {
		setup.ContestSevice.EXPECT().RejectRegistrationRequest(gomock.Any(), gomock.Any(), int64(999), int64(2)).Return(myerrors.ErrNotFound)

		resp := doRequest(t, server.URL+"/contests/999/registration-requests/2/reject")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Not authorized", func(t *testing.T) {
		setup.ContestSevice.EXPECT().RejectRegistrationRequest(gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(myerrors.ErrNotAuthorized)

		resp := doRequest(t, server.URL+"/contests/1/registration-requests/2/reject")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("No pending registration", func(t *testing.T) {
		setup.ContestSevice.EXPECT().RejectRegistrationRequest(gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(myerrors.ErrNoPendingRegistration)

		resp := doRequest(t, server.URL+"/contests/1/registration-requests/2/reject")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		setup.ContestSevice.EXPECT().RejectRegistrationRequest(gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(nil)

		resp := doRequest(t, server.URL+"/contests/1/registration-requests/2/reject")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response httputils.APIResponse[httputils.MessageResponse]
		err := json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.True(t, response.Ok)
		assert.Equal(t, "Registration request rejected successfully", response.Data.Message)
	})
}

func TestAddTaskToContest(t *testing.T) {
	setup, ctrl := setupTeacherRoute(t)
	defer ctrl.Finish()
	db := &testutils.MockDatabase{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		ctx = context.WithValue(ctx, httputils.DatabaseKey, db)
		setup.Mux.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	doRequest := func(t *testing.T, url string, body io.Reader) *http.Response {
		resp, err := http.Post(url, "application/json", body)
		require.NoError(t, err)
		return resp
	}

	body := `{"taskId": 1}`
	t.Run("Invalid contest ID", func(t *testing.T) {
		resp := doRequest(t, server.URL+"/contests/invalid/tasks", strings.NewReader(body))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Invalid request body", func(t *testing.T) {
		resp := doRequest(t, server.URL+"/contests/1/tasks", strings.NewReader(`{invalid json}`))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Contest not found", func(t *testing.T) {
		setup.ContestSevice.EXPECT().AddTaskToContest(gomock.Any(), gomock.Any(), int64(999), gomock.Any()).Return(myerrors.ErrNotFound)

		resp := doRequest(t, server.URL+"/contests/999/tasks", strings.NewReader(body))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Not authorized", func(t *testing.T) {
		setup.ContestSevice.EXPECT().AddTaskToContest(gomock.Any(), gomock.Any(), int64(1), gomock.Any()).Return(myerrors.ErrNotAuthorized)

		resp := doRequest(t, server.URL+"/contests/1/tasks", strings.NewReader(body))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		setup.ContestSevice.EXPECT().AddTaskToContest(gomock.Any(), gomock.Any(), int64(1), gomock.Any()).Return(nil)

		resp := doRequest(t, server.URL+"/contests/1/tasks", strings.NewReader(body))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response httputils.APIResponse[httputils.MessageResponse]
		err := json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.True(t, response.Ok)
	})
}

func TestDeleteTask(t *testing.T) {
	setup, ctrl := setupTeacherRoute(t)
	defer ctrl.Finish()
	db := &testutils.MockDatabase{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		ctx = context.WithValue(ctx, httputils.DatabaseKey, db)
		setup.Mux.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	doRequest := func(t *testing.T, method string, url string) *http.Response {
		req, err := http.NewRequest(method, url, nil)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		return resp
	}

	t.Run("Invalid task ID", func(t *testing.T) {
		resp := doRequest(t, http.MethodDelete, server.URL+"/tasks/invalid")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Task not found", func(t *testing.T) {
		setup.TaskService.EXPECT().Delete(gomock.Any(), gomock.Any(), int64(999)).Return(myerrors.ErrTaskNotFound)

		resp := doRequest(t, http.MethodDelete, server.URL+"/tasks/999")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Not authorized", func(t *testing.T) {
		setup.TaskService.EXPECT().Delete(gomock.Any(), gomock.Any(), int64(1)).Return(myerrors.ErrNotAuthorized)

		resp := doRequest(t, http.MethodDelete, server.URL+"/tasks/1")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		setup.TaskService.EXPECT().Delete(gomock.Any(), gomock.Any(), int64(1)).Return(nil)

		resp := doRequest(t, http.MethodDelete, server.URL+"/tasks/1")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response httputils.APIResponse[httputils.MessageResponse]
		err := json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.True(t, response.Ok)
		assert.Equal(t, "Task deleted successfully", response.Data.Message)
	})
}

func TestAssignTaskToUsers(t *testing.T) {
	setup, ctrl := setupTeacherRoute(t)
	defer ctrl.Finish()
	db := &testutils.MockDatabase{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		ctx = context.WithValue(ctx, httputils.DatabaseKey, db)
		setup.Mux.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	doRequest := func(t *testing.T, url string, body io.Reader) *http.Response {
		resp, err := http.Post(url, "application/json", body)
		require.NoError(t, err)
		return resp
	}

	t.Run("Invalid task ID", func(t *testing.T) {
		resp := doRequest(t, server.URL+"/tasks/invalid/assign/users", strings.NewReader(`{"userIds": [1, 2]}`))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Invalid request body", func(t *testing.T) {
		resp := doRequest(t, server.URL+"/tasks/1/assign/users", strings.NewReader(`{invalid json}`))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	body := `{"userIds": [1, 2]}`
	t.Run("Task not found", func(t *testing.T) {
		userIDs := []int64{1, 2}
		setup.TaskService.EXPECT().AssignToUsers(gomock.Any(), gomock.Any(), int64(999), userIDs).Return(myerrors.ErrTaskNotFound)

		resp := doRequest(t, server.URL+"/tasks/999/assign/users", strings.NewReader(body))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Not authorized", func(t *testing.T) {
		userIDs := []int64{1, 2}
		setup.TaskService.EXPECT().AssignToUsers(gomock.Any(), gomock.Any(), int64(1), userIDs).Return(myerrors.ErrNotAuthorized)

		resp := doRequest(t, server.URL+"/tasks/1/assign/users", strings.NewReader(body))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		userIDs := []int64{1, 2}
		setup.TaskService.EXPECT().AssignToUsers(gomock.Any(), gomock.Any(), int64(1), userIDs).Return(nil)

		resp := doRequest(t, server.URL+"/tasks/1/assign/users", strings.NewReader(body))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response httputils.APIResponse[httputils.MessageResponse]
		err := json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.True(t, response.Ok)
		assert.Equal(t, "Task assigned successfully", response.Data.Message)
	})
}

func TestAssignTaskToGroups(t *testing.T) {
	setup, ctrl := setupTeacherRoute(t)
	defer ctrl.Finish()
	db := &testutils.MockDatabase{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		ctx = context.WithValue(ctx, httputils.DatabaseKey, db)
		setup.Mux.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	doRequest := func(t *testing.T, url string, body io.Reader) *http.Response {
		resp, err := http.Post(url, "application/json", body)
		require.NoError(t, err)
		return resp
	}

	t.Run("Invalid task ID", func(t *testing.T) {
		resp := doRequest(t, server.URL+"/tasks/invalid/assign/groups", strings.NewReader(`{"groupIds": [1, 2]}`))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Invalid request body", func(t *testing.T) {
		resp := doRequest(t, server.URL+"/tasks/1/assign/groups", strings.NewReader(`{invalid json}`))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	body := `{"groupIds": [1, 2]}`
	t.Run("Task not found", func(t *testing.T) {
		groupIDs := []int64{1, 2}
		setup.TaskService.EXPECT().AssignToGroups(gomock.Any(), gomock.Any(), int64(999), groupIDs).Return(myerrors.ErrTaskNotFound)

		resp := doRequest(t, server.URL+"/tasks/999/assign/groups", strings.NewReader(body))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Not authorized", func(t *testing.T) {
		groupIDs := []int64{1, 2}
		setup.TaskService.EXPECT().AssignToGroups(gomock.Any(), gomock.Any(), int64(1), groupIDs).Return(myerrors.ErrNotAuthorized)

		resp := doRequest(t, server.URL+"/tasks/1/assign/groups", strings.NewReader(body))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		groupIDs := []int64{1, 2}
		setup.TaskService.EXPECT().AssignToGroups(gomock.Any(), gomock.Any(), int64(1), groupIDs).Return(nil)

		resp := doRequest(t, server.URL+"/tasks/1/assign/groups", strings.NewReader(body))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response httputils.APIResponse[httputils.MessageResponse]
		err := json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.True(t, response.Ok)
		assert.Equal(t, "Task assigned successfully", response.Data.Message)
	})
}

func TestUnAssignTaskFromUsers(t *testing.T) {
	setup, ctrl := setupTeacherRoute(t)
	defer ctrl.Finish()
	db := &testutils.MockDatabase{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		ctx = context.WithValue(ctx, httputils.DatabaseKey, db)
		setup.Mux.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	doRequest := func(t *testing.T, url string, body io.Reader) *http.Response {
		resp, err := http.Post(url, "application/json", body)
		require.NoError(t, err)
		return resp
	}

	t.Run("Invalid task ID", func(t *testing.T) {
		resp := doRequest(t, server.URL+"/tasks/invalid/unassign/users", strings.NewReader(`{"userIds": [1, 2]}`))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Invalid request body", func(t *testing.T) {
		resp := doRequest(t, server.URL+"/tasks/1/unassign/users", strings.NewReader(`{invalid json}`))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	body := `{"userIds": [1, 2]}`
	t.Run("Task not found", func(t *testing.T) {
		userIDs := []int64{1, 2}
		setup.TaskService.EXPECT().UnassignFromUsers(gomock.Any(), gomock.Any(), int64(999), userIDs).Return(myerrors.ErrTaskNotFound)

		resp := doRequest(t, server.URL+"/tasks/999/unassign/users", strings.NewReader(body))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Not authorized", func(t *testing.T) {
		userIDs := []int64{1, 2}
		setup.TaskService.EXPECT().UnassignFromUsers(gomock.Any(), gomock.Any(), int64(1), userIDs).Return(myerrors.ErrNotAuthorized)

		resp := doRequest(t, server.URL+"/tasks/1/unassign/users", strings.NewReader(body))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		userIDs := []int64{1, 2}
		setup.TaskService.EXPECT().UnassignFromUsers(gomock.Any(), gomock.Any(), int64(1), userIDs).Return(nil)

		resp := doRequest(t, server.URL+"/tasks/1/unassign/users", strings.NewReader(body))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response httputils.APIResponse[httputils.MessageResponse]
		err := json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.True(t, response.Ok)
		assert.Equal(t, "Task unassigned successfully", response.Data.Message)
	})
}

func TestUnAssignTaskFromGroups(t *testing.T) {
	setup, ctrl := setupTeacherRoute(t)
	defer ctrl.Finish()
	db := &testutils.MockDatabase{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		ctx = context.WithValue(ctx, httputils.DatabaseKey, db)
		setup.Mux.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	doRequest := func(t *testing.T, url string, body io.Reader) *http.Response {
		resp, err := http.Post(url, "application/json", body)
		require.NoError(t, err)
		return resp
	}

	t.Run("Invalid task ID", func(t *testing.T) {
		resp := doRequest(t, server.URL+"/tasks/invalid/unassign/groups", strings.NewReader(`{"groupIds": [1, 2]}`))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Invalid request body", func(t *testing.T) {
		resp := doRequest(t, server.URL+"/tasks/1/unassign/groups", strings.NewReader(`{invalid json}`))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	body := `{"groupIds": [1, 2]}`
	t.Run("Task not found", func(t *testing.T) {
		groupIDs := []int64{1, 2}
		setup.TaskService.EXPECT().UnassignFromGroups(gomock.Any(), gomock.Any(), int64(999), groupIDs).Return(myerrors.ErrTaskNotFound)

		resp := doRequest(t, server.URL+"/tasks/999/unassign/groups", strings.NewReader(body))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Not authorized", func(t *testing.T) {
		groupIDs := []int64{1, 2}
		setup.TaskService.EXPECT().UnassignFromGroups(gomock.Any(), gomock.Any(), int64(1), groupIDs).Return(myerrors.ErrNotAuthorized)

		resp := doRequest(t, server.URL+"/tasks/1/unassign/groups", strings.NewReader(body))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		groupIDs := []int64{1, 2}
		setup.TaskService.EXPECT().UnassignFromGroups(gomock.Any(), gomock.Any(), int64(1), groupIDs).Return(nil)

		resp := doRequest(t, server.URL+"/tasks/1/unassign/groups", strings.NewReader(body))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response httputils.APIResponse[httputils.MessageResponse]
		err := json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.True(t, response.Ok)
		assert.Equal(t, "Task unassigned successfully", response.Data.Message)
	})
}

func TestGetTaskLimits(t *testing.T) {
	setup, ctrl := setupTeacherRoute(t)
	defer ctrl.Finish()
	db := &testutils.MockDatabase{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		ctx = context.WithValue(ctx, httputils.DatabaseKey, db)
		setup.Mux.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	doRequest := func(t *testing.T, url string) *http.Response {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		return resp
	}

	t.Run("Invalid task ID", func(t *testing.T) {
		resp := doRequest(t, server.URL+"/tasks/invalid/limits")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Task not found", func(t *testing.T) {
		setup.TaskService.EXPECT().GetLimits(gomock.Any(), gomock.Any(), int64(999)).Return(nil, myerrors.ErrNotFound)

		resp := doRequest(t, server.URL+"/tasks/999/limits")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		limits := []schemas.TestCase{
			{
				ID:          1,
				TaskID:      1,
				Order:       1,
				TimeLimit:   1000,
				MemoryLimit: 256000,
			},
		}
		setup.TaskService.EXPECT().GetLimits(gomock.Any(), gomock.Any(), int64(1)).Return(limits, nil)

		resp := doRequest(t, server.URL+"/tasks/1/limits")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response httputils.APIResponse[[]schemas.TestCase]
		err := json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.True(t, response.Ok)
		assert.Len(t, response.Data, 1)
	})
}

func TestPutTaskLimits(t *testing.T) {
	setup, ctrl := setupTeacherRoute(t)
	defer ctrl.Finish()
	db := &testutils.MockDatabase{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		ctx = context.WithValue(ctx, httputils.DatabaseKey, db)
		setup.Mux.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	doRequest := func(t *testing.T, url string, body io.Reader) *http.Response {
		req, err := http.NewRequest(http.MethodPut, url, body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		return resp
	}
	body := `{"limits": [{"order": 1, "timeLimit": 1000, "memoryLimit": 256000}]}`

	t.Run("Invalid task ID", func(t *testing.T) {
		resp := doRequest(t, server.URL+"/tasks/invalid/limits", strings.NewReader(body))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Invalid request body", func(t *testing.T) {
		resp := doRequest(t, server.URL+"/tasks/1/limits", strings.NewReader(`{invalid json}`))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Task not found", func(t *testing.T) {
		setup.TaskService.EXPECT().PutLimits(gomock.Any(), gomock.Any(), int64(999), gomock.Any()).Return(myerrors.ErrNotFound)

		resp := doRequest(t, server.URL+"/tasks/999/limits", strings.NewReader(body))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Not authorized", func(t *testing.T) {
		setup.TaskService.EXPECT().PutLimits(gomock.Any(), gomock.Any(), int64(1), gomock.Any()).Return(myerrors.ErrNotAuthorized)

		resp := doRequest(t, server.URL+"/tasks/1/limits", strings.NewReader(body))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		setup.TaskService.EXPECT().PutLimits(gomock.Any(), gomock.Any(), int64(1), gomock.Any()).Return(nil)

		resp := doRequest(t, server.URL+"/tasks/1/limits", strings.NewReader(body))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response httputils.APIResponse[httputils.MessageResponse]
		err := json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.True(t, response.Ok)
		assert.Equal(t, "Task limits updated successfully", response.Data.Message)
	})
}
