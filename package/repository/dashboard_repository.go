package repository

import (
	"fmt"
	"time"

	"github.com/mini-maxit/backend/internal/database"
	"github.com/mini-maxit/backend/package/domain/models"
	"github.com/mini-maxit/backend/package/domain/types"
)

// DashboardRepository holds the aggregate queries behind the dashboard pages.
type DashboardRepository interface {
	// CountUsersByRole counts non-deleted users per role.
	CountUsersByRole(db database.Database) ([]models.RoleCount, error)
	// CountContests returns the number of contests and how many are running at `now`.
	CountContests(db database.Database, now time.Time) (total, ongoing int64, err error)
	// CountTasks counts non-deleted tasks.
	CountTasks(db database.Database) (int64, error)
	// CountSubmissions returns the number of submissions and how many were made at or after `since`.
	CountSubmissions(db database.Database, since time.Time) (total, recent int64, err error)
	// HourlySubmissions counts submissions made at or after `since`, grouped by hour. Empty hours are omitted.
	HourlySubmissions(db database.Database, since time.Time) ([]models.HourlySubmissionCount, error)
	// SubmissionOutcomes counts submissions made at or after `since`, grouped by status and result code.
	SubmissionOutcomes(db database.Database, since time.Time) ([]models.SubmissionOutcomeCount, error)
	// ContestActivity returns, for contests that are running or ended after `endedAfter`,
	// how many submissions were made at or after `since` and when the latest one was made.
	// With managerID set, only contests that user has an access-control entry for are included.
	ContestActivity(db database.Database, managerID *int64, now, endedAfter, since time.Time) ([]models.ContestActivity, error)
}

type dashboardRepository struct{}

func NewDashboardRepository() DashboardRepository {
	return &dashboardRepository{}
}

func (dr *dashboardRepository) CountUsersByRole(db database.Database) ([]models.RoleCount, error) {
	var rows []models.RoleCount
	err := db.GetInstance().Model(&models.User{}).
		Select("role, COUNT(*) AS count").
		Group("role").
		Scan(&rows).Error
	return rows, err
}

func (dr *dashboardRepository) CountContests(db database.Database, now time.Time) (int64, int64, error) {
	var row struct {
		Total   int64
		Ongoing int64
	}
	err := db.GetInstance().Model(&models.Contest{}).
		Select(`COUNT(*) AS total,
			COUNT(*) FILTER (WHERE start_at <= ? AND (end_at IS NULL OR end_at > ?)) AS ongoing`, now, now).
		Scan(&row).Error
	return row.Total, row.Ongoing, err
}

func (dr *dashboardRepository) CountTasks(db database.Database) (int64, error) {
	var count int64
	err := db.GetInstance().Model(&models.Task{}).Count(&count).Error
	return count, err
}

func (dr *dashboardRepository) CountSubmissions(db database.Database, since time.Time) (int64, int64, error) {
	var row struct {
		Total  int64
		Recent int64
	}
	err := db.GetInstance().Model(&models.Submission{}).
		Select("COUNT(*) AS total, COUNT(*) FILTER (WHERE submitted_at >= ?) AS recent", since).
		Scan(&row).Error
	return row.Total, row.Recent, err
}

func (dr *dashboardRepository) HourlySubmissions(db database.Database, since time.Time) ([]models.HourlySubmissionCount, error) {
	var rows []models.HourlySubmissionCount
	err := db.GetInstance().Model(&models.Submission{}).
		Select("date_trunc('hour', submitted_at) AS hour, COUNT(*) AS count").
		Where("submitted_at >= ?", since).
		Group("hour").
		Order("hour").
		Scan(&rows).Error
	return rows, err
}

func (dr *dashboardRepository) SubmissionOutcomes(db database.Database, since time.Time) ([]models.SubmissionOutcomeCount, error) {
	tx := db.GetInstance()
	var rows []models.SubmissionOutcomeCount
	err := tx.Model(&models.Submission{}).
		Select("submissions.status AS status, sr.code AS code, COUNT(*) AS count").
		Joins(fmt.Sprintf("LEFT JOIN %s sr ON sr.submission_id = submissions.id",
			database.ResolveTableName(tx, &models.SubmissionResult{}))).
		Where("submissions.submitted_at >= ?", since).
		Group("submissions.status, sr.code").
		Scan(&rows).Error
	return rows, err
}

func (dr *dashboardRepository) ContestActivity(
	db database.Database,
	managerID *int64,
	now, endedAfter, since time.Time,
) ([]models.ContestActivity, error) {
	tx := db.GetInstance()
	query := tx.Model(&models.Contest{}).
		Select(`contests.id AS contest_id,
			COUNT(s.id) FILTER (WHERE s.submitted_at >= ?) AS new_submissions,
			MAX(s.submitted_at) AS last_submitted_at`, since).
		Joins(fmt.Sprintf("LEFT JOIN %s s ON s.contest_id = contests.id",
			database.ResolveTableName(tx, &models.Submission{}))).
		Where("contests.start_at <= ? AND (contests.end_at IS NULL OR contests.end_at > ?)", now, endedAfter).
		Group("contests.id")

	if managerID != nil {
		query = query.Joins(
			fmt.Sprintf("JOIN %s ac ON ac.resource_type = ? AND ac.resource_id = contests.id AND ac.user_id = ?",
				database.ResolveTableName(tx, &models.AccessControl{})),
			types.ResourceTypeContest, *managerID,
		)
	}

	var rows []models.ContestActivity
	err := query.Scan(&rows).Error
	return rows, err
}
