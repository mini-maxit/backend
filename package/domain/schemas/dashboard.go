package schemas

import "time"

// AdminDashboardStats summarises platform usage for the admin dashboard.
type AdminDashboardStats struct {
	Users       DashboardUserCounts       `json:"users"`
	Contests    DashboardContestCounts    `json:"contests"`
	Tasks       int64                     `json:"tasks"`
	Submissions DashboardSubmissionCounts `json:"submissions"`
}

type DashboardUserCounts struct {
	Students int64 `json:"students"`
	Teachers int64 `json:"teachers"`
	Admins   int64 `json:"admins"`
}

type DashboardContestCounts struct {
	Total   int64 `json:"total"`
	Ongoing int64 `json:"ongoing"`
}

type DashboardSubmissionCounts struct {
	Total   int64 `json:"total"`
	Last24h int64 `json:"last24h"`
	// Hourly holds exactly 24 consecutive hours ending with the current one, oldest first.
	Hourly []HourlySubmissions `json:"hourly"`
	// Outcomes covers the same 24 hours as Hourly.
	Outcomes SubmissionOutcomes `json:"outcomes"`
}

type HourlySubmissions struct {
	Hour  time.Time `json:"hour"`
	Count int64     `json:"count"`
}

type SubmissionOutcomes struct {
	Passed           int64 `json:"passed"`
	Failed           int64 `json:"failed"`
	CompilationError int64 `json:"compilationError"`
	Pending          int64 `json:"pending"`
}

// ContestActivity is the recent submission activity of one contest the caller manages.
type ContestActivity struct {
	ContestID       int64      `json:"contestId"`
	NewSubmissions  int64      `json:"newSubmissions"`
	LastSubmittedAt *time.Time `json:"lastSubmittedAt"`
}
