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
	"gorm.io/gorm"
)

// GroupMatcher is a custom matcher for schemas.Group.
type GroupMatcher struct {
	Expected *schemas.Group
}

func (gm *GroupMatcher) Matches(x interface{}) bool {
	group, ok := x.(*schemas.Group)
	if !ok {
		return false
	}
	return group.Name == gm.Expected.Name && group.CreatedBy == gm.Expected.CreatedBy
}

func (gm *GroupMatcher) String() string {
	return "matches expected schemas.Group"
}

func TestCreateGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	gs := mock_service.NewMockGroupService(ctrl)
	route := routes.NewGroupRoute(gs)
	db := &testutils.MockDatabase{}
	handler := testutils.MockDatabaseMiddleware(http.HandlerFunc(route.CreateGroup), db)
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
			`{"name": "Test Group", "extra": "field"}`, // extra field
			`{"name": "Test Group"}{"extra": "json"}`,  // multiple JSON objects
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

			assert.Contains(t, bodyString, "Could not validate request data")
		}
	})

	t.Run("Transaction was not started by middleware", func(t *testing.T) {
		body := schemas.CreateGroup{
			Name: "Test Group",
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

		assert.Contains(t, bodyString, "Database connection error")
	})

	t.Run("Not authorized", func(t *testing.T) {
		body := schemas.CreateGroup{
			Name: "Test Group",
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		expectedGroup := &schemas.Group{
			Name:      "Test Group",
			CreatedBy: 1, // Match the mock user ID
		}

		gs.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
			func(tx *gorm.DB, user schemas.User, group *schemas.Group) (int64, error) {
				assert.Equal(t, expectedGroup.Name, group.Name)
				assert.Equal(t, expectedGroup.CreatedBy, group.CreatedBy)
				return 0, myerrors.ErrNotAuthorized
			}).Times(1)

		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		bodyString := string(bodyBytes)

		assert.Contains(t, bodyString, "Group creation failed")
	})

	t.Run("Internal server error", func(t *testing.T) {
		body := schemas.CreateGroup{
			Name: "Test Group",
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		gs.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), gorm.ErrInvalidDB).Times(1)

		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonBody))
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

		assert.Contains(t, bodyString, "Internal Server Error")
	})

	t.Run("Success", func(t *testing.T) {
		body := schemas.CreateGroup{
			Name: "Test Group",
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		gs.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
			func(tx *gorm.DB, user schemas.User, group *schemas.Group) (int64, error) {
				assert.Equal(t, "Test Group", group.Name)
				assert.Equal(t, int64(1), group.CreatedBy)
				return 1, nil
			}).Times(1)

		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		response := &httputils.APIResponse[httputils.IDResponse]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Equal(t, *httputils.NewIDResponse(int64(1)), response.Data)
	})
}

func TestGetGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	gs := mock_service.NewMockGroupService(ctrl)
	route := routes.NewGroupRoute(gs)
	db := &testutils.MockDatabase{}

	mux := mux.NewRouter()

	mux.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		route.GetGroup(w, r)
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
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Invalid group ID", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/invalid")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Invalid group ID")
	})

	t.Run("Group not found", func(t *testing.T) {
		gs.EXPECT().Get(gomock.Any(), gomock.Any(), int64(999)).Return(nil, myerrors.ErrGroupNotFound).Times(1)

		resp, err := http.Get(server.URL + "/999")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Group retrieval failed")
	})

	t.Run("Database error", func(t *testing.T) {
		gs.EXPECT().Get(gomock.Any(), gomock.Any(), int64(1)).Return(nil, gorm.ErrInvalidDB).Times(1)

		resp, err := http.Get(server.URL + "/1")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Internal Server Error")
	})

	t.Run("Not authorized", func(t *testing.T) {
		gs.EXPECT().Get(gomock.Any(), gomock.Any(), int64(1)).Return(nil, myerrors.ErrNotAuthorized).Times(1)

		resp, err := http.Get(server.URL + "/1")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Group retrieval failed")
	})

	t.Run("Success", func(t *testing.T) {
		group := schemas.GroupDetailed{
			ID:        1,
			Name:      "Test Group",
			CreatedBy: 1,
		}
		gs.EXPECT().Get(gomock.Any(), gomock.Any(), int64(1)).Return(&group, nil).Times(1)

		resp, err := http.Get(server.URL + "/1")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		response := &httputils.APIResponse[schemas.GroupDetailed]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Equal(t, group, response.Data)
	})
}

func TestGetAllGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	gs := mock_service.NewMockGroupService(ctrl)
	route := routes.NewGroupRoute(gs)
	db := &testutils.MockDatabase{}

	mux := mux.NewRouter()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		route.GetAllGroup(w, r)
	})

	handler := testutils.MockDatabaseMiddleware(mux, db)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUser := schemas.User{
			ID:    1,
			Role:  "admin",
			Email: "test@example.com",
		}

		ctx := context.WithValue(r.Context(), httputils.UserKey, mockUser)
		ctx = context.WithValue(ctx, httputils.QueryParamsKey, map[string]interface{}{
			"limit":  "2",
			"offset": "0",
		})
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	t.Run("Accept only GET", func(t *testing.T) {
		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL+"/", nil)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Database error", func(t *testing.T) {
		gs.EXPECT().GetAll(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, gorm.ErrInvalidDB).Times(1)

		resp, err := http.Get(server.URL + "/")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Internal Server Error")
	})
	t.Run("Not authorized", func(t *testing.T) {
		gs.EXPECT().GetAll(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, myerrors.ErrNotAuthorized).Times(1)

		resp, err := http.Get(server.URL + "/")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Group listing failed")
	})

	t.Run("No groups found", func(t *testing.T) {
		gs.EXPECT().GetAll(gomock.Any(), gomock.Any(), gomock.Any()).Return([]schemas.Group{}, nil).Times(1)

		resp, err := http.Get(server.URL + "/")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		response := &httputils.APIResponse[[]schemas.Group]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Empty(t, response.Data)
	})

	t.Run("Limit and offset query params", func(t *testing.T) {
		gs.EXPECT().GetAll(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
			func(tx *gorm.DB, user schemas.User, queryParams map[string]interface{}) ([]schemas.Group, error) {
				assert.Equal(t, "2", queryParams["limit"])
				assert.Equal(t, "0", queryParams["offset"])
				return []schemas.Group{}, nil
			}).Times(1)
		resp, err := http.Get(server.URL + "/?limit=2&offset=0")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		response := &httputils.APIResponse[[]schemas.Group]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Empty(t, response.Data)
	})

	t.Run("Success", func(t *testing.T) {
		groups := []schemas.Group{
			{ID: 1, Name: "Group 1"},
			{ID: 2, Name: "Group 2"},
		}
		gs.EXPECT().GetAll(gomock.Any(), gomock.Any(), gomock.Any()).Return(groups, nil).Times(1)

		resp, err := http.Get(server.URL + "/")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		response := &httputils.APIResponse[[]schemas.Group]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Equal(t, groups, response.Data)
	})
}

func TestEditGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	gs := mock_service.NewMockGroupService(ctrl)
	route := routes.NewGroupRoute(gs)
	db := &testutils.MockDatabase{}

	mux := mux.NewRouter()
	mux.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		route.EditGroup(w, r)
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

	t.Run("Accept only PUT", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL+"/1", nil)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Invalid request body", func(t *testing.T) {
		resp, err := http.NewRequest(http.MethodPut, server.URL+"/1", bytes.NewBufferString(`{invalid_json}`))
		require.NoError(t, err)

		resp.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(resp)
		require.NoError(t, err)
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		bodyBytes, _ := io.ReadAll(res.Body)
		assert.Contains(t, string(bodyBytes), "Could not validate request data")
	})

	t.Run("Transaction not started", func(t *testing.T) {
		db.Invalidate()

		updatedName := "Updated Name"
		body := schemas.EditGroup{Name: &updatedName}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest(http.MethodPut, server.URL+"/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		db.Validate()
		defer res.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
		bodyBytes, _ := io.ReadAll(res.Body)
		assert.Contains(t, string(bodyBytes), "Database connection error")
	})

	t.Run("Invalid group ID in path", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPut, server.URL+"/invalid", nil)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Not authorized", func(t *testing.T) {
		updatedName := "Updated Name"
		body := schemas.EditGroup{Name: &updatedName}
		jsonBody, _ := json.Marshal(body)

		gs.EXPECT().Edit(gomock.Any(), gomock.Any(), int64(1), &body).
			Return(nil, myerrors.ErrNotAuthorized).Times(1)

		req, _ := http.NewRequest(http.MethodPut, server.URL+"/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Group edit failed")
	})

	t.Run("Internal server error", func(t *testing.T) {
		name := "Test Group"
		body := schemas.EditGroup{Name: &name}
		jsonBody, _ := json.Marshal(body)

		gs.EXPECT().Edit(gomock.Any(), gomock.Any(), int64(1), &body).
			Return(nil, gorm.ErrInvalidDB).Times(1)

		req, _ := http.NewRequest(http.MethodPut, server.URL+"/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Internal Server Error")
	})

	t.Run("Success", func(t *testing.T) {
		updatedName := "Updated Group"
		body := schemas.EditGroup{Name: &updatedName}
		updated := schemas.Group{
			ID:   1,
			Name: "Updated Group",
		}

		gs.EXPECT().Edit(gomock.Any(), gomock.Any(), int64(1), &body).Return(&updated, nil).Times(1)

		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest(http.MethodPut, server.URL+"/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		response := &httputils.APIResponse[schemas.Group]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Equal(t, updated, response.Data)
	})
}

func TestAddUsersToGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	gs := mock_service.NewMockGroupService(ctrl)
	route := routes.NewGroupRoute(gs)
	db := &testutils.MockDatabase{}

	mux := mux.NewRouter()
	mux.HandleFunc("/{id}/users", func(w http.ResponseWriter, r *http.Request) {
		route.AddUsersToGroup(w, r)
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

	t.Run("Accept only POST", func(t *testing.T) {
		methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL+"/1/users", nil)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Invalid request body", func(t *testing.T) {
		resp, err := http.NewRequest(http.MethodPost, server.URL+"/1/users", bytes.NewBufferString(`{invalid_json}`))
		require.NoError(t, err)

		resp.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(resp)
		require.NoError(t, err)
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		bodyBytes, _ := io.ReadAll(res.Body)
		assert.Contains(t, string(bodyBytes), "Could not validate request data")
	})

	t.Run("Transaction not started", func(t *testing.T) {
		db.Invalidate()

		body := schemas.UserIDs{
			UserIDs: []int64{1, 2, 3},
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/1/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		db.Validate()
		defer res.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
		bodyBytes, _ := io.ReadAll(res.Body)
		assert.Contains(t, string(bodyBytes), "Database connection error")
	})

	t.Run("Invalid group ID in path", func(t *testing.T) {
		body := schemas.UserIDs{
			UserIDs: []int64{1, 2, 3},
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/invalid/users", bytes.NewBuffer(jsonBody))
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Not authorized", func(t *testing.T) {
		body := schemas.UserIDs{
			UserIDs: []int64{1, 2, 3},
		}
		jsonBody, _ := json.Marshal(body)

		gs.EXPECT().AddUsers(gomock.Any(), gomock.Any(), int64(1), body.UserIDs).
			Return(myerrors.ErrNotAuthorized).Times(1)

		req, _ := http.NewRequest(http.MethodPost, server.URL+"/1/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "User addition to group failed")
	})

	t.Run("Invalid user IDs", func(t *testing.T) {
		body := schemas.UserIDs{
			UserIDs: []int64{1, 2, 3},
		}
		jsonBody, _ := json.Marshal(body)

		gs.EXPECT().AddUsers(gomock.Any(), gomock.Any(), int64(1), body.UserIDs).
			Return(gorm.ErrRecordNotFound).Times(1)

		req, _ := http.NewRequest(http.MethodPost, server.URL+"/1/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "User addition to group failed")
	})
	t.Run("Group not found", func(t *testing.T) {
		body := schemas.UserIDs{
			UserIDs: []int64{1, 2, 3},
		}
		jsonBody, _ := json.Marshal(body)

		gs.EXPECT().AddUsers(gomock.Any(), gomock.Any(), int64(1), body.UserIDs).
			Return(myerrors.ErrGroupNotFound).Times(1)

		req, _ := http.NewRequest(http.MethodPost, server.URL+"/1/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "User addition to group failed")
	})

	t.Run("Internal server error", func(t *testing.T) {
		body := schemas.UserIDs{
			UserIDs: []int64{1, 2, 3},
		}
		jsonBody, _ := json.Marshal(body)

		gs.EXPECT().AddUsers(gomock.Any(), gomock.Any(), int64(1), body.UserIDs).
			Return(gorm.ErrInvalidDB).Times(1)

		req, _ := http.NewRequest(http.MethodPost, server.URL+"/1/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Internal Server Error")
	})

	t.Run("Success", func(t *testing.T) {
		body := schemas.UserIDs{
			UserIDs: []int64{1, 2, 3},
		}
		jsonBody, _ := json.Marshal(body)

		gs.EXPECT().AddUsers(gomock.Any(), gomock.Any(), int64(1), body.UserIDs).
			Return(nil).Times(1)

		req, _ := http.NewRequest(http.MethodPost, server.URL+"/1/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		response := &httputils.APIResponse[any]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Contains(t, string(bodyBytes), "Users added")
	})
}

func TestDeleteUsersFromGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	gs := mock_service.NewMockGroupService(ctrl)
	route := routes.NewGroupRoute(gs)
	db := &testutils.MockDatabase{}

	mux := mux.NewRouter()
	mux.HandleFunc("/{id}/users", func(w http.ResponseWriter, r *http.Request) {
		route.DeleteUsersFromGroup(w, r)
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
			req, err := http.NewRequest(method, server.URL+"/1/users", nil)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Invalid request body", func(t *testing.T) {
		resp, err := http.NewRequest(http.MethodDelete, server.URL+"/1/users", bytes.NewBufferString(`{invalid_json}`))
		require.NoError(t, err)

		resp.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(resp)
		require.NoError(t, err)
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		bodyBytes, _ := io.ReadAll(res.Body)
		assert.Contains(t, string(bodyBytes), "Could not validate request data")
	})

	t.Run("Transaction not started", func(t *testing.T) {
		db.Invalidate()

		body := schemas.UserIDs{
			UserIDs: []int64{1, 2, 3},
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest(http.MethodDelete, server.URL+"/1/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		db.Validate()
		defer res.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
		bodyBytes, _ := io.ReadAll(res.Body)
		assert.Contains(t, string(bodyBytes), "Database connection error")
	})

	t.Run("Invalid group ID in path", func(t *testing.T) {
		body := schemas.UserIDs{
			UserIDs: []int64{1, 2, 3},
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest(http.MethodDelete, server.URL+"/invalid/users", bytes.NewBuffer(jsonBody))
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Not authorized", func(t *testing.T) {
		body := schemas.UserIDs{
			UserIDs: []int64{1, 2, 3},
		}
		jsonBody, _ := json.Marshal(body)

		gs.EXPECT().DeleteUsers(gomock.Any(), gomock.Any(), int64(1), body.UserIDs).
			Return(myerrors.ErrNotAuthorized).Times(1)

		req, _ := http.NewRequest(http.MethodDelete, server.URL+"/1/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "User deletion from group failed")
	})

	t.Run("Invalid user IDs", func(t *testing.T) {
		body := schemas.UserIDs{
			UserIDs: []int64{1, 2, 3},
		}
		jsonBody, _ := json.Marshal(body)

		gs.EXPECT().DeleteUsers(gomock.Any(), gomock.Any(), int64(1), body.UserIDs).
			Return(gorm.ErrRecordNotFound).Times(1)

		req, _ := http.NewRequest(http.MethodDelete, server.URL+"/1/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "User deletion from group failed")
	})

	t.Run("Group not found", func(t *testing.T) {
		body := schemas.UserIDs{
			UserIDs: []int64{1, 2, 3},
		}
		jsonBody, _ := json.Marshal(body)

		gs.EXPECT().DeleteUsers(gomock.Any(), gomock.Any(), int64(1), body.UserIDs).
			Return(myerrors.ErrGroupNotFound).Times(1)

		req, _ := http.NewRequest(http.MethodDelete, server.URL+"/1/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "User deletion from group failed")
	})

	t.Run("Internal server error", func(t *testing.T) {
		body := schemas.UserIDs{
			UserIDs: []int64{1, 2, 3},
		}
		jsonBody, _ := json.Marshal(body)

		gs.EXPECT().DeleteUsers(gomock.Any(), gomock.Any(), int64(1), body.UserIDs).
			Return(gorm.ErrInvalidDB).Times(1)

		req, _ := http.NewRequest(http.MethodDelete, server.URL+"/1/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Internal Server Error")
	})

	t.Run("Success", func(t *testing.T) {
		body := schemas.UserIDs{
			UserIDs: []int64{1, 2, 3},
		}
		jsonBody, _ := json.Marshal(body)

		gs.EXPECT().DeleteUsers(gomock.Any(), gomock.Any(), int64(1), body.UserIDs).
			Return(nil).Times(1)

		req, _ := http.NewRequest(http.MethodDelete, server.URL+"/1/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		response := &httputils.APIResponse[any]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Contains(t, string(bodyBytes), "Users deleted")
	})
}

func TestGetGroupUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	gs := mock_service.NewMockGroupService(ctrl)
	route := routes.NewGroupRoute(gs)
	db := &testutils.MockDatabase{}

	mux := mux.NewRouter()
	mux.HandleFunc("/{id}/users", func(w http.ResponseWriter, r *http.Request) {
		route.GetGroupUsers(w, r)
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
			req, err := http.NewRequest(method, server.URL+"/1/users", nil)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Invalid group ID", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/invalid/users")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Invalid group ID")
	})

	t.Run("Group not found", func(t *testing.T) {
		gs.EXPECT().GetUsers(gomock.Any(), gomock.Any(), int64(999)).Return(nil, myerrors.ErrGroupNotFound).Times(1)

		resp, err := http.Get(server.URL + "/999/users")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Group users retrieval failed")
	})

	t.Run("Database error", func(t *testing.T) {
		gs.EXPECT().GetUsers(gomock.Any(), gomock.Any(), int64(1)).Return(nil, gorm.ErrInvalidDB).Times(1)

		resp, err := http.Get(server.URL + "/1/users")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Internal Server Error")
	})

	t.Run("Not authorized", func(t *testing.T) {
		gs.EXPECT().GetUsers(gomock.Any(), gomock.Any(), int64(1)).Return(nil, myerrors.ErrNotAuthorized).Times(1)

		resp, err := http.Get(server.URL + "/1/users")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Group users retrieval failed")
	})

	t.Run("No users found", func(t *testing.T) {
		gs.EXPECT().GetUsers(gomock.Any(), gomock.Any(), int64(1)).Return([]schemas.User{}, nil).Times(1)

		resp, err := http.Get(server.URL + "/1/users")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		response := &httputils.APIResponse[[]schemas.User]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Empty(t, response.Data)
	})

	t.Run("Success", func(t *testing.T) {
		users := []schemas.User{
			{ID: 1, Email: "test@example.com"},
			{ID: 2, Email: "test@example.com"},
		}
		gs.EXPECT().GetUsers(gomock.Any(), gomock.Any(), int64(1)).Return(users, nil).Times(1)
		resp, err := http.Get(server.URL + "/1/users")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		response := &httputils.APIResponse[[]schemas.User]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Equal(t, users, response.Data)
	})
}

func TestGetGroupTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	gs := mock_service.NewMockGroupService(ctrl)
	route := routes.NewGroupRoute(gs)
	db := &testutils.MockDatabase{}

	mux := mux.NewRouter()
	mux.HandleFunc("/{id}/tasks", func(w http.ResponseWriter, r *http.Request) {
		route.GetGroupTasks(w, r)
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
			req, err := http.NewRequest(method, server.URL+"/1/tasks", nil)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Invalid group ID", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/invalid/tasks")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Invalid group ID")
	})

	t.Run("Group not found", func(t *testing.T) {
		gs.EXPECT().GetTasks(gomock.Any(), gomock.Any(), int64(999)).Return(nil, myerrors.ErrGroupNotFound).Times(1)

		resp, err := http.Get(server.URL + "/999/tasks")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Group tasks retrieval failed")
	})

	t.Run("Database error", func(t *testing.T) {
		gs.EXPECT().GetTasks(gomock.Any(), gomock.Any(), int64(1)).Return(nil, gorm.ErrInvalidDB).Times(1)

		resp, err := http.Get(server.URL + "/1/tasks")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Internal Server Error")
	})

	t.Run("Not authorized", func(t *testing.T) {
		gs.EXPECT().GetTasks(gomock.Any(), gomock.Any(), int64(1)).Return(nil, myerrors.ErrNotAuthorized).Times(1)

		resp, err := http.Get(server.URL + "/1/tasks")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Group tasks retrieval failed")
	})

	t.Run("No tasks found", func(t *testing.T) {
		gs.EXPECT().GetTasks(gomock.Any(), gomock.Any(), int64(1)).Return([]schemas.Task{}, nil).Times(1)

		resp, err := http.Get(server.URL + "/1/tasks")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		response := &httputils.APIResponse[[]schemas.Task]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Empty(t, response.Data)
	})

	t.Run("Success", func(t *testing.T) {
		tasks := []schemas.Task{
			{ID: 1, Title: "Task 1"},
			{ID: 2, Title: "Task 2"},
		}
		gs.EXPECT().GetTasks(gomock.Any(), gomock.Any(), int64(1)).Return(tasks, nil).Times(1)
		resp, err := http.Get(server.URL + "/1/tasks")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		response := &httputils.APIResponse[[]schemas.Task]{}
		err = json.Unmarshal(bodyBytes, response)
		require.NoError(t, err)
		assert.Equal(t, tasks, response.Data)
	})
}
