package routes_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/backend/internal/api/http/routes"
	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	mock_service "github.com/mini-maxit/backend/package/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func dashboardHandler(t *testing.T, pick func(routes.DashboardRoute) http.HandlerFunc, user schemas.User) (http.Handler, *mock_service.MockDashboardService) {
	ctrl := gomock.NewController(t)
	ds := mock_service.NewMockDashboardService(ctrl)
	route := routes.NewDashboardRoute(ds)
	withUser := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pick(route).ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), httputils.UserKey, user)))
	})
	return httputils.MockDatabaseMiddleware(withUser, &testutils.MockDatabase{}), ds
}

func TestDashboardRoute_GetAdminStats(t *testing.T) {
	admin := schemas.User{ID: 1, Role: types.UserRoleAdmin}
	pick := func(r routes.DashboardRoute) http.HandlerFunc { return r.GetAdminStats }

	t.Run("returns stats", func(t *testing.T) {
		handler, ds := dashboardHandler(t, pick, admin)
		ds.EXPECT().GetAdminStats(gomock.Any(), gomock.Any()).
			Return(&schemas.AdminDashboardStats{Tasks: 42, Users: schemas.DashboardUserCounts{Students: 3}}, nil)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin", nil))

		require.Equal(t, http.StatusOK, w.Code)
		var body httputils.APIResponse[schemas.AdminDashboardStats]
		require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
		assert.Equal(t, int64(42), body.Data.Tasks)
		assert.Equal(t, int64(3), body.Data.Users.Students)
	})

	t.Run("forbidden", func(t *testing.T) {
		handler, ds := dashboardHandler(t, pick, schemas.User{ID: 2, Role: types.UserRoleStudent})
		ds.EXPECT().GetAdminStats(gomock.Any(), gomock.Any()).Return(nil, errors.ErrForbidden)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin", nil))
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("wrong method", func(t *testing.T) {
		handler, _ := dashboardHandler(t, pick, admin)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/admin", nil))
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestDashboardRoute_GetContestActivity(t *testing.T) {
	teacher := schemas.User{ID: 7, Role: types.UserRoleTeacher}
	pick := func(r routes.DashboardRoute) http.HandlerFunc { return r.GetContestActivity }

	t.Run("returns activity", func(t *testing.T) {
		handler, ds := dashboardHandler(t, pick, teacher)
		ds.EXPECT().GetContestActivity(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ any, u *schemas.User) ([]schemas.ContestActivity, error) {
				assert.Equal(t, int64(7), u.ID)
				return []schemas.ContestActivity{{ContestID: 3, NewSubmissions: 5}}, nil
			})

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/teacher/contest-activity", nil))

		require.Equal(t, http.StatusOK, w.Code)
		var body httputils.APIResponse[[]schemas.ContestActivity]
		require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
		assert.Equal(t, []schemas.ContestActivity{{ContestID: 3, NewSubmissions: 5}}, body.Data)
	})

	t.Run("service error", func(t *testing.T) {
		handler, ds := dashboardHandler(t, pick, teacher)
		ds.EXPECT().GetContestActivity(gomock.Any(), gomock.Any()).Return(nil, assert.AnError)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/teacher/contest-activity", nil))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
