package routes_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/api/http/middleware"
	"github.com/mini-maxit/backend/internal/api/http/routes"
	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func newRegistrationServer(t *testing.T) (*httptest.Server, *mock_service.MockContestService) {
	ctrl := gomock.NewController(t)
	cs := mock_service.NewMockContestService(ctrl)
	ss := mock_service.NewMockSubmissionService(ctrl)
	db := &testutils.MockDatabase{}

	router := mux.NewRouter()
	routes.RegisterContestRoutes(router, routes.NewContestRoute(cs, ss))
	routes.RegisterContestsManagementRoute(router.PathPrefix("/contests-management").Subrouter(), routes.NewContestsManagementRoute(cs, ss))
	handler := middleware.QueryParamsMiddleware(httputils.MockDatabaseMiddleware(router, db))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := schemas.User{ID: 1, Role: testAdmin, Email: testExampleMail}
		handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), httputils.UserKey, user)))
	}))
	t.Cleanup(server.Close)
	return server, cs
}

func doRequest(t *testing.T, method, url string) int {
	t.Helper()
	req, err := http.NewRequest(method, url, nil)
	assert.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()
	return resp.StatusCode
}

func TestGetPublicContests(t *testing.T) {
	t.Run("routes /contests/public before /contests/{id} and passes status", func(t *testing.T) {
		server, cs := newRegistrationServer(t)
		cs.EXPECT().GetPublicContests(gomock.Any(), gomock.Any(), types.ContestStatusUpcoming, gomock.Any()).
			Return(schemas.PaginatedResult[[]schemas.PublicContest]{}, nil)

		assert.Equal(t, http.StatusOK, doRequest(t, http.MethodGet, server.URL+"/contests/public?status=upcoming"))
	})

	t.Run("invalid status", func(t *testing.T) {
		server, _ := newRegistrationServer(t)
		assert.Equal(t, http.StatusBadRequest, doRequest(t, http.MethodGet, server.URL+"/contests/public?status=bogus"))
	})
}

func TestRegisterForContest(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
	}{
		{"success", nil, http.StatusOK},
		{"not found", errors.ErrNotFound, http.StatusNotFound},
		{"already registered", errors.ErrAlreadyRegistered, http.StatusConflict},
		{"contest ended", errors.ErrContestEnded, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server, cs := newRegistrationServer(t)
			cs.EXPECT().RegisterForContest(gomock.Any(), gomock.Any(), int64(7)).Return(tc.err)
			assert.Equal(t, tc.status, doRequest(t, http.MethodPost, server.URL+"/contests/7/register"))
		})
	}

	t.Run("only POST", func(t *testing.T) {
		server, _ := newRegistrationServer(t)
		assert.Equal(t, http.StatusMethodNotAllowed, doRequest(t, http.MethodGet, server.URL+"/contests/7/register"))
	})
}

func TestRegistrationRequestsManagement(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		server, cs := newRegistrationServer(t)
		cs.EXPECT().GetRegistrationRequests(gomock.Any(), gomock.Any(), int64(7)).Return([]schemas.RegistrationRequest{}, nil)
		assert.Equal(t, http.StatusOK, doRequest(t, http.MethodGet, server.URL+"/contests-management/contests/7/registration-requests"))
	})

	t.Run("approve", func(t *testing.T) {
		server, cs := newRegistrationServer(t)
		cs.EXPECT().ApproveRegistrationRequest(gomock.Any(), gomock.Any(), int64(7), int64(3)).Return(nil)
		assert.Equal(t, http.StatusOK, doRequest(t, http.MethodPost, server.URL+"/contests-management/contests/7/registration-requests/3/approve"))
	})

	t.Run("reject without pending request", func(t *testing.T) {
		server, cs := newRegistrationServer(t)
		cs.EXPECT().RejectRegistrationRequest(gomock.Any(), gomock.Any(), int64(7), int64(3)).Return(errors.ErrNoPendingRegistration)
		assert.Equal(t, http.StatusNotFound, doRequest(t, http.MethodPost, server.URL+"/contests-management/contests/7/registration-requests/3/reject"))
	})

	t.Run("invalid user id", func(t *testing.T) {
		server, _ := newRegistrationServer(t)
		assert.Equal(t, http.StatusBadRequest, doRequest(t, http.MethodPost, server.URL+"/contests-management/contests/7/registration-requests/x/approve"))
	})
}
