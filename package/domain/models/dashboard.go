package models

import (
	"time"

	"github.com/mini-maxit/backend/package/domain/types"
)

// RoleCount is a scan target for users grouped by role.
type RoleCount struct {
	Role  types.UserRole
	Count int64
}

// HourlySubmissionCount is a scan target for submissions grouped by hour.
type HourlySubmissionCount struct {
	Hour  time.Time
	Count int64
}

// SubmissionOutcomeCount is a scan target for submissions grouped by status and result code.
// Code is nil when the submission has no result yet.
type SubmissionOutcomeCount struct {
	Status types.SubmissionStatus
	Code   *types.SubmissionResultCode
	Count  int64
}

// ContestActivity is a scan target for recent submission activity in one contest.
type ContestActivity struct {
	ContestID       int64
	NewSubmissions  int64
	LastSubmittedAt *time.Time
}
