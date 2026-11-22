package service

import (
	"time"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/schemas"
	"github.com/mini-maxit/backend/package/domain/types"
	"github.com/mini-maxit/backend/package/repository"
	"github.com/mini-maxit/backend/package/utils"
)

const (
	dashboardWindow      = 24 * time.Hour
	dashboardEndedWithin = 7 * 24 * time.Hour
)

type DashboardService interface {
	// GetAdminStats returns platform usage figures. Admin only.
	GetAdminStats(db database.Database, currentUser *schemas.User) (*schemas.AdminDashboardStats, error)
	// GetContestActivity returns submission activity for contests the user manages that are running
	// or ended in the last 7 days. Teachers see contests they collaborate on, admins see all.
	GetContestActivity(db database.Database, currentUser *schemas.User) ([]schemas.ContestActivity, error)
}

type dashboardService struct {
	repo repository.DashboardRepository
	now  func() time.Time
}

// NewDashboardService takes the clock as a parameter so tests can pin the current time.
func NewDashboardService(repo repository.DashboardRepository, now func() time.Time) DashboardService {
	return &dashboardService{repo: repo, now: now}
}

func (ds *dashboardService) GetAdminStats(db database.Database, currentUser *schemas.User) (*schemas.AdminDashboardStats, error) {
	if err := utils.ValidateRoleAccess(currentUser.Role, []types.UserRole{types.UserRoleAdmin}); err != nil {
		return nil, err
	}

	now := ds.now().UTC()
	// The window is the 24 whole hours ending with the current one, so the first bucket is complete.
	windowStart := now.Truncate(time.Hour).Add(time.Hour - dashboardWindow)

	stats := &schemas.AdminDashboardStats{}

	roles, err := ds.repo.CountUsersByRole(db)
	if err != nil {
		return nil, err
	}
	for _, r := range roles {
		switch r.Role {
		case types.UserRoleStudent:
			stats.Users.Students = r.Count
		case types.UserRoleTeacher:
			stats.Users.Teachers = r.Count
		case types.UserRoleAdmin:
			stats.Users.Admins = r.Count
		}
	}

	if stats.Contests.Total, stats.Contests.Ongoing, err = ds.repo.CountContests(db, now); err != nil {
		return nil, err
	}
	if stats.Tasks, err = ds.repo.CountTasks(db); err != nil {
		return nil, err
	}
	if stats.Submissions.Total, stats.Submissions.Last24h, err = ds.repo.CountSubmissions(db, windowStart); err != nil {
		return nil, err
	}

	hourly, err := ds.repo.HourlySubmissions(db, windowStart)
	if err != nil {
		return nil, err
	}
	stats.Submissions.Hourly = fillHours(hourly, windowStart)

	outcomes, err := ds.repo.SubmissionOutcomes(db, windowStart)
	if err != nil {
		return nil, err
	}
	stats.Submissions.Outcomes = groupOutcomes(outcomes)

	return stats, nil
}

func (ds *dashboardService) GetContestActivity(db database.Database, currentUser *schemas.User) ([]schemas.ContestActivity, error) {
	if err := utils.ValidateRoleAccess(
		currentUser.Role,
		[]types.UserRole{types.UserRoleTeacher, types.UserRoleAdmin},
	); err != nil {
		return nil, err
	}

	var managerID *int64
	if currentUser.Role != types.UserRoleAdmin {
		managerID = &currentUser.ID
	}

	now := ds.now().UTC()
	rows, err := ds.repo.ContestActivity(db, managerID, now, now.Add(-dashboardEndedWithin), now.Add(-dashboardWindow))
	if err != nil {
		return nil, err
	}

	result := make([]schemas.ContestActivity, len(rows))
	for i, r := range rows {
		result[i] = schemas.ContestActivity{
			ContestID:       r.ContestID,
			NewSubmissions:  r.NewSubmissions,
			LastSubmittedAt: r.LastSubmittedAt,
		}
	}
	return result, nil
}

// fillHours returns 24 hourly buckets starting at windowStart, with zero for hours the query omitted.
func fillHours(rows []models.HourlySubmissionCount, windowStart time.Time) []schemas.HourlySubmissions {
	buckets := make([]schemas.HourlySubmissions, int(dashboardWindow/time.Hour))
	for i := range buckets {
		buckets[i].Hour = windowStart.Add(time.Duration(i) * time.Hour)
	}
	for _, r := range rows {
		i := int(r.Hour.UTC().Sub(windowStart) / time.Hour)
		if i >= 0 && i < len(buckets) {
			buckets[i].Count += r.Count
		}
	}
	return buckets
}

// groupOutcomes folds status/result-code pairs into the four outcomes the dashboard shows.
func groupOutcomes(rows []models.SubmissionOutcomeCount) schemas.SubmissionOutcomes {
	var o schemas.SubmissionOutcomes
	for _, r := range rows {
		switch {
		case r.Status == types.SubmissionStatusReceived || r.Status == types.SubmissionStatusSentForEvaluation:
			o.Pending += r.Count
		case r.Status == types.SubmissionStatusEvaluated && r.Code != nil && *r.Code == types.SubmissionResultCodeSuccess:
			o.Passed += r.Count
		case r.Status == types.SubmissionStatusEvaluated && r.Code != nil && *r.Code == types.SubmissionResultCodeCompilationError:
			o.CompilationError += r.Count
		default:
			// Lost submissions and every other result code count as failed.
			o.Failed += r.Count
		}
	}
	return o
}
