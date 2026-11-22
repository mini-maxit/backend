package service_test

import (
	"testing"
	"time"

	"github.com/mini-maxit/backend/internal/testutils"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/errors"
	mock_repository "github.com/mini-maxit/backend/package/repository/mocks"
	"github.com/mini-maxit/backend/package/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// 14:35 UTC, so the 24h window starts at 15:00 the previous day and ends with the 14:00 bucket.
var dashboardNow = time.Date(2026, 11, 21, 14, 35, 0, 0, time.UTC)

func newDashboardService(t *testing.T) (service.DashboardService, *mock_repository.MockDashboardRepository) {
	ctrl := gomock.NewController(t)
	repo := mock_repository.NewMockDashboardRepository(ctrl)
	return service.NewDashboardService(repo, func() time.Time { return dashboardNow }), repo
}

func code(c types.SubmissionResultCode) *types.SubmissionResultCode { return &c }

func TestDashboardService_GetAdminStats(t *testing.T) {
	db := &testutils.MockDatabase{}
	windowStart := time.Date(2026, 11, 20, 15, 0, 0, 0, time.UTC)

	t.Run("non-admin is forbidden", func(t *testing.T) {
		ds, _ := newDashboardService(t)
		for _, role := range []types.UserRole{types.UserRoleStudent, types.UserRoleTeacher} {
			stats, err := ds.GetAdminStats(db, &schemas.User{ID: 1, Role: role})
			require.ErrorIs(t, err, errors.ErrForbidden)
			assert.Nil(t, stats)
		}
	})

	t.Run("aggregates counts, hours and outcomes", func(t *testing.T) {
		ds, repo := newDashboardService(t)

		repo.EXPECT().CountUsersByRole(db).Return([]models.RoleCount{
			{Role: types.UserRoleStudent, Count: 44},
			{Role: types.UserRoleTeacher, Count: 2},
			{Role: types.UserRoleAdmin, Count: 8},
		}, nil)
		repo.EXPECT().CountContests(db, dashboardNow).Return(int64(19), int64(2), nil)
		repo.EXPECT().CountTasks(db).Return(int64(42), nil)
		repo.EXPECT().CountSubmissions(db, windowStart).Return(int64(1615), int64(9), nil)
		repo.EXPECT().HourlySubmissions(db, windowStart).Return([]models.HourlySubmissionCount{
			{Hour: windowStart, Count: 3},
			{Hour: time.Date(2026, 11, 21, 14, 0, 0, 0, time.UTC), Count: 6},
		}, nil)
		repo.EXPECT().SubmissionOutcomes(db, windowStart).Return([]models.SubmissionOutcomeCount{
			{Status: types.SubmissionStatusEvaluated, Code: code(types.SubmissionResultCodeSuccess), Count: 4},
			{Status: types.SubmissionStatusEvaluated, Code: code(types.SubmissionResultCodeTestFailed), Count: 2},
			{Status: types.SubmissionStatusEvaluated, Code: code(types.SubmissionResultCodeCompilationError), Count: 1},
			{Status: types.SubmissionStatusLost, Count: 1},
			{Status: types.SubmissionStatusReceived, Count: 1},
			{Status: types.SubmissionStatusSentForEvaluation, Count: 1},
		}, nil)

		stats, err := ds.GetAdminStats(db, &schemas.User{ID: 1, Role: types.UserRoleAdmin})
		require.NoError(t, err)

		assert.Equal(t, schemas.DashboardUserCounts{Students: 44, Teachers: 2, Admins: 8}, stats.Users)
		assert.Equal(t, schemas.DashboardContestCounts{Total: 19, Ongoing: 2}, stats.Contests)
		assert.Equal(t, int64(42), stats.Tasks)
		assert.Equal(t, int64(1615), stats.Submissions.Total)
		assert.Equal(t, int64(9), stats.Submissions.Last24h)

		require.Len(t, stats.Submissions.Hourly, 24)
		assert.Equal(t, windowStart, stats.Submissions.Hourly[0].Hour)
		assert.Equal(t, int64(3), stats.Submissions.Hourly[0].Count)
		assert.Equal(t, int64(0), stats.Submissions.Hourly[12].Count)
		assert.Equal(t, time.Date(2026, 11, 21, 14, 0, 0, 0, time.UTC), stats.Submissions.Hourly[23].Hour)
		assert.Equal(t, int64(6), stats.Submissions.Hourly[23].Count)

		assert.Equal(t, schemas.SubmissionOutcomes{
			Passed:           4,
			Failed:           3, // test failure + lost
			CompilationError: 1,
			Pending:          2,
		}, stats.Submissions.Outcomes)
	})

	t.Run("repository error is returned", func(t *testing.T) {
		ds, repo := newDashboardService(t)
		repo.EXPECT().CountUsersByRole(db).Return(nil, assert.AnError)

		stats, err := ds.GetAdminStats(db, &schemas.User{ID: 1, Role: types.UserRoleAdmin})
		require.ErrorIs(t, err, assert.AnError)
		assert.Nil(t, stats)
	})
}

func TestDashboardService_GetContestActivity(t *testing.T) {
	db := &testutils.MockDatabase{}
	weekAgo := dashboardNow.Add(-7 * 24 * time.Hour)
	dayAgo := dashboardNow.Add(-24 * time.Hour)
	last := dashboardNow.Add(-time.Hour)

	t.Run("student is forbidden", func(t *testing.T) {
		ds, _ := newDashboardService(t)
		activity, err := ds.GetContestActivity(db, &schemas.User{ID: 5, Role: types.UserRoleStudent})
		require.ErrorIs(t, err, errors.ErrForbidden)
		assert.Nil(t, activity)
	})

	t.Run("teacher sees only managed contests", func(t *testing.T) {
		ds, repo := newDashboardService(t)
		teacherID := int64(7)
		repo.EXPECT().ContestActivity(db, &teacherID, dashboardNow, weekAgo, dayAgo).Return([]models.ContestActivity{
			{ContestID: 3, NewSubmissions: 5, LastSubmittedAt: &last},
			{ContestID: 4},
		}, nil)

		activity, err := ds.GetContestActivity(db, &schemas.User{ID: teacherID, Role: types.UserRoleTeacher})
		require.NoError(t, err)
		assert.Equal(t, []schemas.ContestActivity{
			{ContestID: 3, NewSubmissions: 5, LastSubmittedAt: &last},
			{ContestID: 4},
		}, activity)
	})

	t.Run("admin sees all contests", func(t *testing.T) {
		ds, repo := newDashboardService(t)
		repo.EXPECT().ContestActivity(db, (*int64)(nil), dashboardNow, weekAgo, dayAgo).Return(nil, nil)

		activity, err := ds.GetContestActivity(db, &schemas.User{ID: 1, Role: types.UserRoleAdmin})
		require.NoError(t, err)
		assert.Empty(t, activity)
	})
}
