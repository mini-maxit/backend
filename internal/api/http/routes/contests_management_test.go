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
	"github.com/mini-maxit/backend/package/errors"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCreateContest(t *testing.T) {
	ctrl := gomock.NewController(t)
	cs := mock_service.NewMockContestService(ctrl)
	ss := mock_service.NewMockSubmissionService(ctrl)
	route := routes.NewContestsManagementRoute(cs, ss)
	db := &testutils.MockDatabase{}
	handler := httputils.MockDatabaseMiddleware(http.HandlerFunc(route.CreateContest), db)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				t.Logf("Recovered from panic: %v", rec)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		// Mock user and add to context
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, httputils.UserKey, mockUser)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	t.Run("Accept only POST", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

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

	t.Run("Invalid request body", func(t *testing.T) {
		// Test malformed JSON that triggers non-validation errors
		invalidBodies := []string{
			`{"name": "Test Contest", "description": "Test", "extra": "field"}`, // extra field
			`{"name": "Test Contest", "description": "Test"}{"extra": "json"}`,  // multiple JSON objects
		}

		for _, body := range invalidBodies {
			t.Logf("Testing with body: %s", body)

			resp, err := http.Post(server.URL, "application/json", strings.NewReader(body))
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}
			bodyString := string(bodyBytes)

			assert.Contains(t, bodyString, "Invalid request body")
		}
	})

	t.Run("Not authorized", func(t *testing.T) {
		body := schemas.CreateContest{
			Name:        "Test Contest",
			Description: "Test Description",
			StartAt:     time.Now().Add(1 * time.Hour),
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		cs.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), errors.ErrForbidden)

		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		body := schemas.CreateContest{
			Name:        "Test Contest",
			Description: "Test Description",
			StartAt:     time.Now().Add(1 * time.Hour),
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		cs.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)

		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestEditContest(t *testing.T) {
	ctrl := gomock.NewController(t)
	cs := mock_service.NewMockContestService(ctrl)
	ss := mock_service.NewMockSubmissionService(ctrl)
	route := routes.NewContestsManagementRoute(cs, ss)
	db := &testutils.MockDatabase{}

	mux := mux.NewRouter()

	mux.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		route.EditContest(w, r)
	})

	handler := httputils.MockDatabaseMiddleware(mux, db)

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
	newName := "Updated Contest"

	t.Run("Accept only PUT", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPatch}

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

	t.Run("Invalid request body", func(t *testing.T) {
		invalidBodies := []string{
			`{invalid json}`,
		}

		for _, body := range invalidBodies {
			req, err := http.NewRequest(http.MethodPut, server.URL+"/1", bytes.NewBufferString(body))
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

	t.Run("Contest not found", func(t *testing.T) {
		body := schemas.EditContest{
			Name: &newName,
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		cs.EXPECT().Edit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.ErrNotFound)

		req, err := http.NewRequest(http.MethodPut, server.URL+"/1", bytes.NewBuffer(jsonBody))
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
		body := schemas.EditContest{
			Name: &newName,
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		cs.EXPECT().Edit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.ErrForbidden)

		req, err := http.NewRequest(http.MethodPut, server.URL+"/1", bytes.NewBuffer(jsonBody))
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

	t.Run("Success", func(t *testing.T) {
		body := schemas.EditContest{
			Name: &newName,
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		isVisible := true
		isRegistrationOpen := true
		isSubmissionOpen := true
		contest := &schemas.CreatedContest{
			ContestDetailed: schemas.ContestDetailed{
				BaseContest: schemas.BaseContest{
					ID:          1,
					Name:        "Updated Contest",
					CreatedBy:   1,
					Description: "Test Description",
				},
				IsSubmissionOpen: isSubmissionOpen,
			},
			IsVisible:          isVisible,
			IsRegistrationOpen: isRegistrationOpen,
		}

		cs.EXPECT().Edit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(contest, nil)

		req, err := http.NewRequest(http.MethodPut, server.URL+"/1", bytes.NewBuffer(jsonBody))
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
	})
}

func TestDeleteContest(t *testing.T) {
	ctrl := gomock.NewController(t)
	cs := mock_service.NewMockContestService(ctrl)
	ss := mock_service.NewMockSubmissionService(ctrl)
	route := routes.NewContestsManagementRoute(cs, ss)
	db := &testutils.MockDatabase{}

	mux := mux.NewRouter()

	mux.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		route.DeleteContest(w, r)
	})

	handler := httputils.MockDatabaseMiddleware(mux, db)

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

	t.Run("Contest not found", func(t *testing.T) {
		cs.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.ErrNotFound)

		req, err := http.NewRequest(http.MethodDelete, server.URL+"/1", nil)
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
		cs.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.ErrForbidden)

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

	t.Run("Success", func(t *testing.T) {
		cs.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

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
	})
}

func TestGetRegistrationRequests(t *testing.T) {
	ctrl := gomock.NewController(t)
	cs := mock_service.NewMockContestService(ctrl)
	ss := mock_service.NewMockSubmissionService(ctrl)
	route := routes.NewContestsManagementRoute(cs, ss)
	db := &testutils.MockDatabase{}

	mux := mux.NewRouter()

	mux.HandleFunc("/{id}/registration-requests", func(w http.ResponseWriter, r *http.Request) {
		route.GetRegistrationRequests(w, r)
	})

	handler := httputils.MockDatabaseMiddleware(mux, db)

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
			req, err := http.NewRequest(method, server.URL+"/1/registration-requests", nil)
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

	t.Run("Contest not found", func(t *testing.T) {
		cs.EXPECT().GetRegistrationRequests(gomock.Any(), gomock.Any(), int64(1), gomock.Any()).Return(nil, errors.ErrNotFound)

		resp, err := http.Get(server.URL + "/1/registration-requests")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Not authorized", func(t *testing.T) {
		cs.EXPECT().GetRegistrationRequests(gomock.Any(), gomock.Any(), int64(1), gomock.Any()).Return(nil, errors.ErrForbidden)

		resp, err := http.Get(server.URL + "/1/registration-requests")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		mockRequests := []schemas.RegistrationRequest{
			{
				ID:        1,
				ContestID: 1,
				UserID:    2,
				User: schemas.User{
					ID:       2,
					Name:     "John",
					Surname:  "Doe",
					Email:    "john@example.com",
					Username: "johndoe",
					Role:     "student",
				},
				CreatedAt: time.Now(),
			},
		}

		cs.EXPECT().GetRegistrationRequests(gomock.Any(), gomock.Any(), int64(1), gomock.Any()).Return(mockRequests, nil)

		resp, err := http.Get(server.URL + "/1/registration-requests")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response httputils.APIResponse[[]schemas.RegistrationRequest]
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		assert.True(t, response.Ok)
		assert.Len(t, response.Data, 1)
		assert.Equal(t, int64(1), response.Data[0].ID)
		assert.Equal(t, "John", response.Data[0].User.Name)
	})
}

func TestApproveRegistrationRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	cs := mock_service.NewMockContestService(ctrl)
	ss := mock_service.NewMockSubmissionService(ctrl)
	route := routes.NewContestsManagementRoute(cs, ss)
	db := &testutils.MockDatabase{}

	mux := mux.NewRouter()

	mux.HandleFunc("/{id}/registration-requests/{user_id}/approve", func(w http.ResponseWriter, r *http.Request) {
		route.ApproveRegistrationRequest(w, r)
	})

	handler := httputils.MockDatabaseMiddleware(mux, db)

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

	t.Run("Accept only POST", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL+"/1/registration-requests/2/approve", nil)
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

	t.Run("Invalid contest ID", func(t *testing.T) {
		resp, err := http.Post(server.URL+"/invalid/registration-requests/2/approve", "application/json", nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Invalid user ID", func(t *testing.T) {
		resp, err := http.Post(server.URL+"/1/registration-requests/invalid/approve", "application/json", nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Contest not found", func(t *testing.T) {
		cs.EXPECT().ApproveRegistrationRequest(gomock.Any(), gomock.Any(), int64(999), int64(2)).Return(errors.ErrNotFound)

		resp, err := http.Post(server.URL+"/999/registration-requests/2/approve", "application/json", nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("User not found", func(t *testing.T) {
		cs.EXPECT().ApproveRegistrationRequest(gomock.Any(), gomock.Any(), int64(1), int64(999)).Return(errors.ErrNotFound)

		resp, err := http.Post(server.URL+"/1/registration-requests/999/approve", "application/json", nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Not authorized", func(t *testing.T) {
		cs.EXPECT().ApproveRegistrationRequest(gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(errors.ErrForbidden)

		resp, err := http.Post(server.URL+"/1/registration-requests/2/approve", "application/json", nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("No pending registration", func(t *testing.T) {
		cs.EXPECT().ApproveRegistrationRequest(gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(errors.ErrNoPendingRegistration)

		resp, err := http.Post(server.URL+"/1/registration-requests/2/approve", "application/json", nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("User already participant", func(t *testing.T) {
		cs.EXPECT().ApproveRegistrationRequest(gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(errors.ErrAlreadyParticipant)

		resp, err := http.Post(server.URL+"/1/registration-requests/2/approve", "application/json", nil)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		cs.EXPECT().ApproveRegistrationRequest(gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(nil)

		resp, err := http.Post(server.URL+"/1/registration-requests/2/approve", "application/json", nil)
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
		assert.Equal(t, "Registration request approved successfully", response.Data.Message)
	})
}

func TestRemoveTaskFromContest(t *testing.T) {
	ctrl := gomock.NewController(t)
	cs := mock_service.NewMockContestService(ctrl)
	ss := mock_service.NewMockSubmissionService(ctrl)
	route := routes.NewContestsManagementRoute(cs, ss)
	db := &testutils.MockDatabase{}

	router := mux.NewRouter()
	router.HandleFunc("/{id}/tasks/{taskId}", func(w http.ResponseWriter, r *http.Request) {
		route.RemoveTaskFromContest(w, r)
	})

	handler := httputils.MockDatabaseMiddleware(router, db)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				t.Logf("Recovered from panic: %v", rec)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		// Mock user and add to context
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, httputils.UserKey, mockUser)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	t.Run("Accept only DELETE", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL+"/1/tasks/2", nil)
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

	t.Run("Invalid contest ID", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, server.URL+"/invalid/tasks/2", nil)
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

	t.Run("Invalid task ID", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, server.URL+"/1/tasks/invalid", nil)
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

	t.Run("Task not in contest", func(t *testing.T) {
		cs.EXPECT().RemoveTaskFromContest(gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(errors.ErrTaskNotInContest)

		req, err := http.NewRequest(http.MethodDelete, server.URL+"/1/tasks/2", nil)
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
		cs.EXPECT().RemoveTaskFromContest(gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(errors.ErrForbidden)

		req, err := http.NewRequest(http.MethodDelete, server.URL+"/1/tasks/2", nil)
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

	t.Run("Contest not found", func(t *testing.T) {
		cs.EXPECT().RemoveTaskFromContest(gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(errors.ErrNotFound)

		req, err := http.NewRequest(http.MethodDelete, server.URL+"/1/tasks/2", nil)
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

	t.Run("Success", func(t *testing.T) {
		cs.EXPECT().RemoveTaskFromContest(gomock.Any(), gomock.Any(), int64(1), int64(2)).Return(nil)

		req, err := http.NewRequest(http.MethodDelete, server.URL+"/1/tasks/2", nil)
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
		assert.Equal(t, "Task removed from contest successfully", response.Data.Message)
	})
}
