package routes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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
	"go.uber.org/mock/gomock"
)

// ContestMatcher is a custom matcher for schemas.CreateContest.
type ContestMatcher struct {
	Expected *schemas.CreateContest
}

func (cm *ContestMatcher) Matches(x interface{}) bool {
	contest, ok := x.(*schemas.CreateContest)
	if !ok {
		return false
	}
	return contest.Name == cm.Expected.Name
}

func (cm *ContestMatcher) String() string {
	return "matches expected schemas.CreateContest"
}

func TestCreateContest(t *testing.T) {
	ctrl := gomock.NewController(t)
	cs := mock_service.NewMockContestService(ctrl)
	route := routes.NewContestRoute(cs)
	db := &testutils.MockDatabase{}
	handler := testutils.MockDatabaseMiddleware(http.HandlerFunc(route.CreateContest), db)
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
		invalidBodies := []any{
			struct {
				Name string `json:"name"`
			}{
				Name: "",
			},
			struct {
				Name  string `json:"name"`
				Extra string `json:"extra"`
			}{
				Name:  "Contest Name",
				Extra: "Invalid Field",
			},
		}

		for _, body := range invalidBodies {
			t.Logf("Testing with body: %v", body)
			jsonBody, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonBody))
			if err != nil {
				t.Logf("Body: %s", string(jsonBody))
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

	t.Run("Transaction was not started by middleware", func(t *testing.T) {
		body := schemas.CreateContest{
			Name:        "Test Contest",
			Description: "Test Description",
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		db.Invalidate()
		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonBody))
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

		assert.Contains(t, bodyString, "Transaction was not started by middleware")
	})

	t.Run("Not authorized", func(t *testing.T) {
		body := schemas.CreateContest{
			Name:        "Test Contest",
			Description: "Test Description",
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		cs.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), myerrors.ErrNotAuthorized)

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

func TestGetContest(t *testing.T) {
	ctrl := gomock.NewController(t)
	cs := mock_service.NewMockContestService(ctrl)
	route := routes.NewContestRoute(cs)
	db := &testutils.MockDatabase{}

	mux := mux.NewRouter()

	mux.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		route.GetContest(w, r)
	})

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
		cs.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, myerrors.ErrNotFound)

		resp, err := http.Get(server.URL + "/1")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Not authorized", func(t *testing.T) {
		cs.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, myerrors.ErrNotAuthorized)

		resp, err := http.Get(server.URL + "/1")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		contest := &schemas.Contest{
			ID:          1,
			Name:        "Test Contest",
			Description: "Test Description",
			CreatedBy:   1,
		}

		cs.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(contest, nil)

		resp, err := http.Get(server.URL + "/1")
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
	route := routes.NewContestRoute(cs)
	db := &testutils.MockDatabase{}

	mux := mux.NewRouter()

	mux.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		route.EditContest(w, r)
	})

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

		cs.EXPECT().Edit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, myerrors.ErrNotFound)

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

		cs.EXPECT().Edit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, myerrors.ErrNotAuthorized)

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

		contest := &schemas.Contest{
			ID:          1,
			Name:        "Updated Contest",
			Description: "Test Description",
			CreatedBy:   1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
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

func TestGetAllContests(t *testing.T) {
	ctrl := gomock.NewController(t)
	cs := mock_service.NewMockContestService(ctrl)
	route := routes.NewContestRoute(cs)
	db := &testutils.MockDatabase{}
	handler := testutils.MockDatabaseMiddleware(http.HandlerFunc(route.GetAllContests), db)
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
		ctx = context.WithValue(ctx, httputils.QueryParamsKey, map[string]any{
			"limit":  10,
			"offset": 0,
			"sort":   "created_at:desc",
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

	t.Run("Success with empty list", func(t *testing.T) {
		cs.EXPECT().GetAll(gomock.Any(), gomock.Any(), gomock.Any()).Return([]schemas.Contest{}, nil)

		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Success with contests", func(t *testing.T) {
		contests := []schemas.Contest{
			{
				ID:          1,
				Name:        "Contest 1",
				Description: "Description 1",
				CreatedBy:   1,
			},
			{
				ID:          2,
				Name:        "Contest 2",
				Description: "Description 2",
				CreatedBy:   1,
			},
		}

		cs.EXPECT().GetAll(gomock.Any(), gomock.Any(), gomock.Any()).Return(contests, nil)

		resp, err := http.Get(server.URL)
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
	route := routes.NewContestRoute(cs)
	db := &testutils.MockDatabase{}

	mux := mux.NewRouter()

	mux.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		route.DeleteContest(w, r)
	})

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
		cs.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(myerrors.ErrNotFound)

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
		cs.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(myerrors.ErrNotAuthorized)

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

func TestRegisterForContest(t *testing.T) {
	ctrl := gomock.NewController(t)
	cs := mock_service.NewMockContestService(ctrl)
	route := routes.NewContestRoute(cs)
	db := &testutils.MockDatabase{}

	mux := mux.NewRouter()

	mux.HandleFunc("/{id}/register", func(w http.ResponseWriter, r *http.Request) {
		route.RegisterForContest(w, r)
	})

	handler := testutils.MockDatabaseMiddleware(mux, db)

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
			req, err := http.NewRequest(method, server.URL+"/1/register", nil)
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
		cs.EXPECT().RegisterForContest(gomock.Any(), gomock.Any(), gomock.Any()).Return(myerrors.ErrNotFound)

		req, err := http.NewRequest(http.MethodPost, server.URL+"/1/register", nil)
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

	t.Run("Not authorized - contest not visible", func(t *testing.T) {
		cs.EXPECT().RegisterForContest(gomock.Any(), gomock.Any(), gomock.Any()).Return(myerrors.ErrNotAuthorized)

		req, err := http.NewRequest(http.MethodPost, server.URL+"/1/register", nil)
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

	t.Run("Registration closed", func(t *testing.T) {
		cs.EXPECT().RegisterForContest(gomock.Any(), gomock.Any(), gomock.Any()).Return(myerrors.ErrContestRegistrationClosed)

		req, err := http.NewRequest(http.MethodPost, server.URL+"/1/register", nil)
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

	t.Run("Contest ended", func(t *testing.T) {
		cs.EXPECT().RegisterForContest(gomock.Any(), gomock.Any(), gomock.Any()).Return(myerrors.ErrContestEnded)

		req, err := http.NewRequest(http.MethodPost, server.URL+"/1/register", nil)
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

	t.Run("Already registered", func(t *testing.T) {
		cs.EXPECT().RegisterForContest(gomock.Any(), gomock.Any(), gomock.Any()).Return(myerrors.ErrAlreadyRegistered)

		req, err := http.NewRequest(http.MethodPost, server.URL+"/1/register", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		cs.EXPECT().RegisterForContest(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		req, err := http.NewRequest(http.MethodPost, server.URL+"/1/register", nil)
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

func TestGetOngoingContests(t *testing.T) {
	ctrl := gomock.NewController(t)
	cs := mock_service.NewMockContestService(ctrl)
	route := routes.NewContestRoute(cs)
	db := &testutils.MockDatabase{}
	handler := testutils.MockDatabaseMiddleware(http.HandlerFunc(route.GetOngoingContests), db)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				t.Logf("Recovered from panic: %v", rec)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		ctx := context.WithValue(r.Context(), httputils.QueryParamsKey, map[string]any{
			"limit":  10,
			"offset": 0,
			"sort":   "",
		})
		ctx = context.WithValue(ctx, httputils.UserKey, schemas.User{ID: 1})
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	t.Run("Accept only GET", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, server.URL, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("Success with ongoing contests", func(t *testing.T) {
		now := time.Now()
		contests := []schemas.Contest{
			{
				ID:                 1,
				Name:               "Ongoing Contest",
				Description:        "Test Description",
				CreatedBy:          1,
				StartAt:            &now,
				EndAt:              nil,
				CreatedAt:          now,
				UpdatedAt:          now,
				ParticipantCount:   5,
				TaskCount:          3,
				RegistrationStatus: "registered",
			},
		}

		cs.EXPECT().GetOngoingContests(gomock.Any(), gomock.Any(), gomock.Any()).Return(contests, nil)

		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
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

		var response httputils.APIResponse[[]schemas.Contest]
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		assert.True(t, response.Ok)
		assert.Len(t, response.Data, 1)
		assert.Equal(t, "Ongoing Contest", response.Data[0].Name)
		assert.Equal(t, int64(5), response.Data[0].ParticipantCount)
		assert.Equal(t, int64(3), response.Data[0].TaskCount)
	})

	t.Run("Internal server error", func(t *testing.T) {
		cs.EXPECT().GetOngoingContests(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, myerrors.ErrNotFound)

		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
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
}
