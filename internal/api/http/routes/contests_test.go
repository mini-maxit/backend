package routes_test

import (
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
	"github.com/mini-maxit/backend/package/errors"
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

func TestGetContest(t *testing.T) {
	ctrl := gomock.NewController(t)
	cs := mock_service.NewMockContestService(ctrl)
	ss := mock_service.NewMockSubmissionService(ctrl)
	route := routes.NewContestRoute(cs, ss)
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
		cs.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.ErrNotFound)

		resp, err := http.Get(server.URL + "/1")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Not authorized", func(t *testing.T) {
		cs.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.ErrForbidden)

		resp, err := http.Get(server.URL + "/1")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		contest := &schemas.Contest{
			BaseContest: schemas.BaseContest{
				ID:          1,
				Name:        "Test Contest",
				Description: "Test Description",
				CreatedBy:   1,
			},
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

func TestRegisterForContest(t *testing.T) {
	ctrl := gomock.NewController(t)
	cs := mock_service.NewMockContestService(ctrl)
	ss := mock_service.NewMockSubmissionService(ctrl)
	route := routes.NewContestRoute(cs, ss)
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
		cs.EXPECT().RegisterForContest(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.ErrNotFound)

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
		cs.EXPECT().RegisterForContest(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.ErrForbidden)

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
		cs.EXPECT().RegisterForContest(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.ErrContestRegistrationClosed)

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
		cs.EXPECT().RegisterForContest(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.ErrContestEnded)

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
		cs.EXPECT().RegisterForContest(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.ErrAlreadyRegistered)

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
	ss := mock_service.NewMockSubmissionService(ctrl)
	route := routes.NewContestRoute(cs, ss)
	db := &testutils.MockDatabase{}
	handler := testutils.MockDatabaseMiddleware(http.HandlerFunc(route.GetContests), db)
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
			"status": "ongoing",
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
		contests := []schemas.AvailableContest{
			{
				Contest: schemas.Contest{
					BaseContest: schemas.BaseContest{
						ID:          1,
						Name:        "Ongoing Contest",
						Description: "Test Description",
						CreatedBy:   1,
						StartAt:     now,
						EndAt:       nil,
					},
					ParticipantCount: 5,
					TaskCount:        3,
				},
				RegistrationStatus: "registered",
			},
		}

		paginatedResult := schemas.NewPaginatedResult(contests, 0, 10, int64(len(contests)))
		cs.EXPECT().GetOngoingContests(gomock.Any(), gomock.Any(), gomock.Any()).Return(paginatedResult, nil)

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

		var response httputils.APIResponse[schemas.PaginatedResult[[]schemas.AvailableContest]]
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		assert.True(t, response.Ok)
		assert.Len(t, response.Data.Items, 1)
		assert.Equal(t, "Ongoing Contest", response.Data.Items[0].Name)
		assert.Equal(t, int64(5), response.Data.Items[0].ParticipantCount)
		assert.Equal(t, int64(3), response.Data.Items[0].TaskCount)
		assert.Equal(t, 1, response.Data.Pagination.CurrentPage)
		assert.Equal(t, int64(1), int64(response.Data.Pagination.TotalItems))
	})

	t.Run("Internal server error", func(t *testing.T) {
		cs.EXPECT().GetOngoingContests(gomock.Any(), gomock.Any(), gomock.Any()).Return(schemas.PaginatedResult[[]schemas.AvailableContest]{}, errors.ErrNotFound)

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

func TestGetMyContestResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	cs := mock_service.NewMockContestService(ctrl)
	ss := mock_service.NewMockSubmissionService(ctrl)
	route := routes.NewContestRoute(cs, ss)
	db := &testutils.MockDatabase{}

	mux := mux.NewRouter()

	mux.HandleFunc("/{id}/results/my", func(w http.ResponseWriter, r *http.Request) {
		route.GetMyContestResults(w, r)
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

	t.Run("Accept only GET", func(t *testing.T) {
		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL+"/1/results/my", nil)
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
		cs.EXPECT().GetMyContestResults(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.ErrNotFound)

		resp, err := http.Get(server.URL + "/999/results/my")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Not authorized", func(t *testing.T) {
		cs.EXPECT().GetMyContestResults(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.ErrForbidden)

		resp, err := http.Get(server.URL + "/1/results/my")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Success", func(t *testing.T) {
		submissionID := int64(10)
		results := &schemas.ContestResults{
			Contest: schemas.BaseContest{
				ID: int64(1),
			},
			TaskResults: []schemas.TaskResult{
				{
					Task: schemas.TaskInfo{
						ID: int64(1),
					},
					SubmissionCount:  5,
					BestScore:        80.5,
					BestSubmissionID: &submissionID,
				},
				{
					Task: schemas.TaskInfo{
						ID: int64(12),
					},
					SubmissionCount:  3,
					BestScore:        100.0,
					BestSubmissionID: &submissionID,
				},
			},
		}

		cs.EXPECT().GetMyContestResults(gomock.Any(), gomock.Any(), int64(1)).Return(results, nil)

		resp, err := http.Get(server.URL + "/1/results/my")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		var response httputils.APIResponse[schemas.ContestResults]
		err = json.Unmarshal(body, &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		assert.True(t, response.Ok)
		assert.Equal(t, int64(1), response.Data.Contest.ID)
		assert.Len(t, response.Data.TaskResults, 2)
		assert.Equal(t, int64(1), response.Data.TaskResults[0].Task.ID)
		assert.Equal(t, 5, response.Data.TaskResults[0].SubmissionCount)
		assert.InDelta(t, 80.5, response.Data.TaskResults[0].BestScore, 0.001)
		assert.NotNil(t, response.Data.TaskResults[0].BestSubmissionID)
		assert.Equal(t, int64(10), *response.Data.TaskResults[0].BestSubmissionID)
	})
}

func TestGetContestDetails(t *testing.T) {
	ctrl := gomock.NewController(t)
	cs := mock_service.NewMockContestService(ctrl)
	ss := mock_service.NewMockSubmissionService(ctrl)
	route := routes.NewContestRoute(cs, ss)
	db := &testutils.MockDatabase{}
	router := mux.NewRouter()
	router.HandleFunc("/{id}/details", func(w http.ResponseWriter, r *http.Request) {
		route.GetContestDetails(w, r)
	})
	handler := testutils.MockDatabaseMiddleware(router, db)
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
	t.Run("Accept only GET", func(t *testing.T) {
		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
		for _, method := range methods {
			req, err := http.NewRequest(method, server.URL+"/1/details", nil)
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
		cs.EXPECT().GetDetails(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.ErrNotFound)
		resp, err := http.Get(server.URL + "/1/details")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
	t.Run("Not authorized - not a participant", func(t *testing.T) {
		cs.EXPECT().GetDetails(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.ErrForbidden)
		resp, err := http.Get(server.URL + "/1/details")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
	t.Run("Success", func(t *testing.T) {
		isRegOpen := true
		isSubOpen := true
		contest := &schemas.ContestDetailed{
			Contest: schemas.Contest{
				BaseContest: schemas.BaseContest{
					ID:          1,
					Name:        "Test Contest",
					Description: "Test Description",
					CreatedBy:   1,
				},
				ParticipantCount: 5,
				TaskCount:        3,
			},
			CreatorName:        "Test Creator",
			IsRegistrationOpen: &isRegOpen,
			IsSubmissionOpen:   &isSubOpen,
		}
		cs.EXPECT().GetDetails(gomock.Any(), gomock.Any(), gomock.Any()).Return(contest, nil)
		resp, err := http.Get(server.URL + "/1/details")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		var response httputils.APIResponse[schemas.ContestDetailed]
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		assert.True(t, response.Ok)
		assert.Equal(t, "Test Contest", response.Data.Name)
		assert.Equal(t, "Test Creator", response.Data.CreatorName)
		assert.Equal(t, int64(5), response.Data.ParticipantCount)
	})
}
func TestGetContestTasksFiltered(t *testing.T) {
	ctrl := gomock.NewController(t)
	cs := mock_service.NewMockContestService(ctrl)
	ss := mock_service.NewMockSubmissionService(ctrl)
	route := routes.NewContestRoute(cs, ss)
	db := &testutils.MockDatabase{}
	router := mux.NewRouter()
	router.HandleFunc("/{id}/tasks", func(w http.ResponseWriter, r *http.Request) {
		route.GetContestTasksFiltered(w, r)
	})
	handler := testutils.MockDatabaseMiddleware(router, db)
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
	t.Run("Accept only GET", func(t *testing.T) {
		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
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
	t.Run("Contest not found", func(t *testing.T) {
		cs.EXPECT().GetVisibleTasksForContest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.ErrNotFound)
		resp, err := http.Get(server.URL + "/1/tasks")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
	t.Run("Not authorized - not a participant", func(t *testing.T) {
		cs.EXPECT().GetVisibleTasksForContest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.ErrForbidden)
		resp, err := http.Get(server.URL + "/1/tasks")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
	t.Run("Invalid status parameter", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/1/tasks?status=invalid")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
	t.Run("Success without status filter", func(t *testing.T) {
		now := time.Now()
		tasks := []schemas.ContestTask{
			{
				Task: schemas.TaskInfo{
					ID:    1,
					Title: "Test Task 1",
				},
				CreatorName:      "Test Creator",
				StartAt:          now.Add(-time.Hour),
				EndAt:            nil,
				IsSubmissionOpen: true,
				IsVisible:        true,
			},
		}
		cs.EXPECT().GetVisibleTasksForContest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(tasks, nil)
		resp, err := http.Get(server.URL + "/1/tasks")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		var response httputils.APIResponse[[]schemas.ContestTask]
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		assert.True(t, response.Ok)
		assert.Len(t, response.Data, 1)
		assert.Equal(t, "Test Task 1", response.Data[0].Task.Title)
		assert.True(t, response.Data[0].IsVisible)
	})
	t.Run("Success with status filter", func(t *testing.T) {
		now := time.Now()
		tasks := []schemas.ContestTask{
			{
				Task: schemas.TaskInfo{
					ID:    1,
					Title: "Ongoing Task",
				},
				CreatorName:      "Test Creator",
				StartAt:          now.Add(-time.Hour),
				EndAt:            nil,
				IsSubmissionOpen: true,
				IsVisible:        true,
			},
		}
		cs.EXPECT().GetVisibleTasksForContest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(tasks, nil)
		resp, err := http.Get(server.URL + "/1/tasks?status=ongoing")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
