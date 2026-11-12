package schemas

import (
	"time"

	"github.com/mini-maxit/backend/package/domain/types"
)

type BaseContest struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatedBy   int64      `json:"createdBy"`
	StartAt     time.Time  `json:"startAt"`
	EndAt       *time.Time `json:"endAt"`
}

type ContestDetailed struct {
	BaseContest
	CreatorName      string              `json:"creatorName"`
	ParticipantCount int64               `json:"participantCount"`
	TaskCount        int64               `json:"taskCount"`
	Status           types.ContestStatus `json:"status"` // "upcoming", "ongoing", "past"
	IsSubmissionOpen bool                `json:"isSubmissionOpen"`
}

type CreatedContest struct {
	ContestDetailed
	CreatedAt time.Time `json:"createdAt"`
}

type ManagedContest struct {
	CreatedContest
	PermissionType types.Permission `json:"permissionType"` // "view", "edit", "manage", "owner"
}

type CreateContest struct {
	Name             string       `json:"name" validate:"required,gte=3,lte=100"`
	Description      string       `json:"description" validate:"required"`
	StartAt          time.Time    `json:"startAt" validate:"required"`
	EndAt            OptionalTime `json:"endAt"`
	IsSubmissionOpen bool         `json:"isSubmissionOpen"`
}

type EditContest struct {
	Name             *string      `json:"name,omitempty" validate:"omitempty,gte=3,lte=100"`
	Description      *string      `json:"description,omitempty"`
	StartAt          OptionalTime `json:"startAt"`
	EndAt            OptionalTime `json:"endAt"`
	IsSubmissionOpen *bool        `json:"isSubmissionOpen,omitempty"`
}

type ContestWithStats struct {
	ContestDetailed
	SolvedTaskCount int64 `json:"solvedTaskCount"`
}

type UpcomingContest struct {
	ContestDetailed
}

type PastContestWithStats struct {
	ContestDetailed
	SolvedTaskPercentage float64 `json:"solvedPercentage"`
	Score                int64   `json:"score"`
	MaximumScore         int64   `json:"maximumScore"`
	Rank                 int64   `json:"rank"`
}

type UserContestsWithStats struct {
	Ongoing  []ContestWithStats     `json:"ongoing"`
	Past     []PastContestWithStats `json:"past"`
	Upcoming []ContestWithStats     `json:"upcoming"`
}

// AddTaskToContest limits override the task's per-test-case limits for this contest; absent or null keeps the task defaults.
type AddTaskToContest struct {
	TaskID        int64         `json:"taskId" validate:"required"`
	StartAt       OptionalTime  `json:"startAt"`
	EndAt         OptionalTime  `json:"endAt"`
	TimeLimitMs   OptionalInt64 `json:"timeLimitMs"`
	MemoryLimitKB OptionalInt64 `json:"memoryLimitKb"`
}

// UpdateTaskInContest limits: absent leaves the current override, null clears it (back to task defaults), a number sets it.
type UpdateTaskInContest struct {
	StartAt       OptionalTime  `json:"startAt"`
	EndAt         OptionalTime  `json:"endAt"`
	TimeLimitMs   OptionalInt64 `json:"timeLimitMs"`
	MemoryLimitKB OptionalInt64 `json:"memoryLimitKb"`
}

// ReorderContestTasks lists every task of the contest in the desired order.
type ReorderContestTasks struct {
	TaskIDs []int64 `json:"taskIds" validate:"required"`
}

type ContestTask struct {
	Task             TaskInfo   `json:"task"`
	CreatorName      string     `json:"creatorName"`
	StartAt          time.Time  `json:"startAt"`
	EndAt            *time.Time `json:"endAt"`
	IsSubmissionOpen bool       `json:"isSubmissionOpen"`
	Position         int64      `json:"position"`
	TimeLimitMs      *int64     `json:"timeLimitMs"`   // contest override, null = task default
	MemoryLimitKB    *int64     `json:"memoryLimitKb"` // contest override, null = task default
}

type TaskResult struct {
	Task             TaskInfo `json:"task"`
	SubmissionCount  int      `json:"submissionCount"`
	BestScore        float64  `json:"bestScore"`
	BestSubmissionID *int64   `json:"bestSubmissionId"`
}

type ContestResults struct {
	Contest     BaseContest  `json:"contest"`
	TaskResults []TaskResult `json:"taskResults"`
}

type GroupRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ContestParticipantAccess struct {
	User
	Direct bool       `json:"direct"`
	Groups []GroupRef `json:"groups"`
}
