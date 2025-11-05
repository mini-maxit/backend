package routes_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/api/http/routes"
	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/schemas"
	myerrors "github.com/mini-maxit/backend/package/errors"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetOngoingContests(t *testing.T) {
	ctrl := gomock.NewController(t)
	cs := mock_service.NewMockContestService(ctrl)
	ss := mock_service.NewMockSubmissionService(ctrl)
	gs := mock_service.NewMockGroupService(ctrl)
	ts := mock_service.NewMockTaskService(ctrl)
	qs := mock_service.NewMockQueueService(ctrl)

	route := routes.NewStudentRoute(cs, ss, gs, ts, qs)
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
					ID:               1,
					Name:             "Ongoing Contest",
					Description:      "Test Description",
					CreatedBy:        1,
					StartAt:          &now,
					EndAt:            nil,
					CreatedAt:        now,
					UpdatedAt:        now,
					ParticipantCount: 5,
					TaskCount:        3,
				},
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
