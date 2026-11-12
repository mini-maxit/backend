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
	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/errors"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	testContestName        = "Test Contest"
	testContestDescription = "Test Description"
)

func TestGetContestParticipants(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cs := mock_service.NewMockContestService(ctrl)
	ss := mock_service.NewMockSubmissionService(ctrl)
	route := routes.NewContestsManagementRoute(cs, ss)
	db := &testutils.MockDatabase{}

	mux := mux.NewRouter()
	mux.HandleFunc("/{id}/participants", func(w http.ResponseWriter, r *http.Request) {
		route.GetContestParticipants(w, r)
	})

	handler := httputils.MockDatabaseMiddleware(mux, db)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  testAdmin,
			Email: testExampleMail,
		}

		query := r.URL.Query()
		queryParams, _ := httputils.GetQueryParams(&query)

		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		ctx = context.WithValue(ctx, httputils.QueryParamsKey, queryParams)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	t.Run("Accept only GET", func(t *testing.T) {
		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL+"/1/participants", nil)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Returns participants with access sources", func(t *testing.T) {
		contestID := int64(1)
		expected := schemas.PaginatedResult[[]schemas.ContestParticipantAccess]{
			Items: []schemas.ContestParticipantAccess{
				{
					User: schemas.User{
						ID:      1,
						Name:    "Alice",
						Surname: "A",
						Email:   "alice@example.com",
						Role:    testAdmin,
					},
					Direct: true,
					Groups: []schemas.GroupRef{
						{ID: 7, Name: "Class 4-B"},
					},
				},
				{
					User: schemas.User{
						ID:      2,
						Name:    "Bob",
						Surname: "B",
						Email:   "bob@example.com",
						Role:    testStudent,
					},
					Direct: false,
					Groups: []schemas.GroupRef{
						{ID: 9, Name: "Robotics Club"},
					},
				},
			},
			Pagination: schemas.PaginationMetadata{
				CurrentPage: 1,
				PageSize:    20,
				TotalItems:  2,
				TotalPages:  1,
			},
		}

		cs.EXPECT().GetContestParticipants(gomock.Any(), gomock.Any(), contestID, gomock.Any()).Return(expected, nil).Times(1)

		resp, err := http.Get(server.URL + "/1/participants")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		response := &httputils.APIResponse[schemas.PaginatedResult[[]schemas.ContestParticipantAccess]]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.True(t, response.Ok)
		assert.Equal(t, expected.Items, response.Data.Items)
		assert.Equal(t, expected.Pagination.TotalItems, response.Data.Pagination.TotalItems)
	})
}

func TestGetContestGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cs := mock_service.NewMockContestService(ctrl)
	ss := mock_service.NewMockSubmissionService(ctrl)
	route := routes.NewContestsManagementRoute(cs, ss)
	db := &testutils.MockDatabase{}

	mux := mux.NewRouter()
	mux.HandleFunc("/{id}/groups", func(w http.ResponseWriter, r *http.Request) {
		route.GetContestGroups(w, r)
	})

	handler := httputils.MockDatabaseMiddleware(mux, db)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  testAdmin,
			Email: testExampleMail,
		}

		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	t.Run("Accept only GET", func(t *testing.T) {
		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL+"/1/groups", nil)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Returns groups with members", func(t *testing.T) {
		contestID := int64(1)
		expected := []schemas.GroupWithMembers{
			{
				Group: schemas.Group{
					ID:   7,
					Name: "Group Seven",
				},
				Members: []schemas.UserInfo{
					{ID: 1, Name: "Alice", Surname: "A"},
					{ID: 2, Name: "Bob", Surname: "B"},
				},
			},
		}

		cs.EXPECT().GetContestGroups(gomock.Any(), gomock.Any(), contestID).Return(expected, nil).Times(1)

		resp, err := http.Get(server.URL + "/1/groups")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		response := &httputils.APIResponse[[]schemas.GroupWithMembers]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.True(t, response.Ok)
		require.Len(t, response.Data, 1)
		assert.Equal(t, expected[0].ID, response.Data[0].ID)
		assert.Equal(t, expected[0].Name, response.Data[0].Name)
		assert.Equal(t, expected[0].Members, response.Data[0].Members)
	})
}

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
			Role:  testAdmin,
			Email: testExampleMail,
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
			Name:        testContestName,
			Description: testContestDescription,
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
			Name:        testContestName,
			Description: testContestDescription,
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
			Role:  testAdmin,
			Email: testExampleMail,
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

		isSubmissionOpen := true
		contest := &schemas.CreatedContest{
			ContestDetailed: schemas.ContestDetailed{
				BaseContest: schemas.BaseContest{
					ID:          1,
					Name:        "Updated Contest",
					CreatedBy:   1,
					Description: testContestDescription,
				},
				IsSubmissionOpen: isSubmissionOpen,
			},
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

	t.Run("Null endAt decodes as explicit clear", func(t *testing.T) {
		// Raw JSON with explicit null for endAt
		rawBody := `{"name":"Updated Contest","endAt":null}`

		cs.EXPECT().Edit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(db database.Database, u *schemas.User, contestID int64, editInfo *schemas.EditContest) (*schemas.CreatedContest, error) {
				require.NotNil(t, editInfo.EndAt, "EndAt must be present (Set=true)")
				assert.True(t, editInfo.EndAt.Set, "EndAt.Set must be true for explicit null")
				assert.Nil(t, editInfo.EndAt.Value, "EndAt.Value must be nil for explicit null")
				assert.False(t, editInfo.StartAt.Set, "StartAt absent")
				return &schemas.CreatedContest{}, nil
			})

		req, err := http.NewRequest(http.MethodPut, server.URL+"/1", bytes.NewBufferString(rawBody))
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
			Role:  testAdmin,
			Email: testExampleMail,
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

func TestRemoveTaskFromContest(t *testing.T) {
	ctrl := gomock.NewController(t)
	cs := mock_service.NewMockContestService(ctrl)
	ss := mock_service.NewMockSubmissionService(ctrl)
	route := routes.NewContestsManagementRoute(cs, ss)
	db := &testutils.MockDatabase{}

	router := mux.NewRouter()
	router.HandleFunc("/{id}/tasks", func(w http.ResponseWriter, r *http.Request) {
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
			Role:  testAdmin,
			Email: testExampleMail,
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, httputils.UserKey, mockUser)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	t.Run("Accept only DELETE", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL+"/1/tasks", nil)
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
		req, err := http.NewRequest(http.MethodDelete, server.URL+"/invalid/tasks", nil)
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

	t.Run("Empty body", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, server.URL+"/1/tasks", nil)
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

		body := bytes.NewBufferString(`{"taskIds":[2]}`)
		req, err := http.NewRequest(http.MethodDelete, server.URL+"/1/tasks", body)
		req.Header.Set("Content-Type", "application/json")
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

		body := bytes.NewBufferString(`{"taskIds":[2]}`)
		req, err := http.NewRequest(http.MethodDelete, server.URL+"/1/tasks", body)
		req.Header.Set("Content-Type", "application/json")
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

		body := bytes.NewBufferString(`{"taskIds":[2]}`)
		req, err := http.NewRequest(http.MethodDelete, server.URL+"/1/tasks", body)
		req.Header.Set("Content-Type", "application/json")
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

		body := bytes.NewBufferString(`{"taskIds":[2]}`)
		req, err := http.NewRequest(http.MethodDelete, server.URL+"/1/tasks", body)
		req.Header.Set("Content-Type", "application/json")
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
		assert.Equal(t, "Tasks removed from contest successfully", response.Data.Message)
	})
}

func TestGetContestSubmissions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cs := mock_service.NewMockContestService(ctrl)
	ss := mock_service.NewMockSubmissionService(ctrl)
	route := routes.NewContestsManagementRoute(cs, ss)
	db := &testutils.MockDatabase{}

	mux := mux.NewRouter()
	mux.HandleFunc("/{id}/submissions", route.GetContestSubmissions)
	handler := httputils.MockDatabaseMiddleware(mux, db)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		queryParams, _ := httputils.GetQueryParams(&query)
		ctx := context.WithValue(r.Context(), httputils.UserKey, schemas.User{ID: 1, Role: testAdmin})
		ctx = context.WithValue(ctx, httputils.QueryParamsKey, queryParams)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	empty := &schemas.PaginatedResult[[]schemas.Submission]{}

	t.Run("Combines taskId with search filters", func(t *testing.T) {
		taskID := int64(7)
		ss.EXPECT().
			GetAllForContest(gomock.Any(), int64(3), &taskID, gomock.Any(), gomock.Any(), "jan", "sort").
			Return(empty, nil).Times(1)

		resp, err := http.Get(server.URL + "/3/submissions?taskId=7&userSearch=jan&taskSearch=sort")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("No taskId passes nil", func(t *testing.T) {
		ss.EXPECT().
			GetAllForContest(gomock.Any(), int64(3), (*int64)(nil), gomock.Any(), gomock.Any(), "", "").
			Return(empty, nil).Times(1)

		resp, err := http.Get(server.URL + "/3/submissions")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Invalid taskId", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/3/submissions?taskId=abc")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
