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

type Contest struct {
	BaseContest
	ParticipantCount int64               `json:"participantCount"`
	TaskCount        int64               `json:"taskCount"`
	Status           types.ContestStatus `json:"status"` // "upcoming", "ongoing", "past"
}

type AvailableContest struct {
	Contest
	RegistrationStatus string `json:"registrationStatus"` // "registered", "canRegister", "awaitingApproval", "registrationClosed"
}

type CreatedContest struct {
	BaseContest
	CreatedAt          time.Time `json:"createdAt"`
	IsRegistrationOpen *bool     `json:"isRegistrationOpen"`
	IsSubmissionOpen   *bool     `json:"isSubmissionOpen"`
	IsVisible          *bool     `json:"isVisible"`
}

type ManagedContest struct {
	CreatedContest
	PermissionType types.Permission `json:"permissionType"` // "view", "edit", "manage", "owner"
}

type CreateContest struct {
	Name               string     `json:"name" validate:"required,gte=3,lte=100"`
	Description        string     `json:"description" validate:"required"`
	StartAt            time.Time  `json:"startAt" validate:"required"`
	EndAt              *time.Time `json:"endAt,omitempty"`
	IsRegistrationOpen *bool      `json:"isRegistrationOpen"`
	IsSubmissionOpen   *bool      `json:"isSubmissionOpen"`
	IsVisible          *bool      `json:"isVisible"`
}

type EditContest struct {
	Name               *string    `json:"name,omitempty" validate:"omitempty,gte=3,lte=100"`
	Description        *string    `json:"description,omitempty"`
	StartAt            *time.Time `json:"startAt,omitempty"`
	EndAt              *time.Time `json:"endAt,omitempty"`
	IsRegistrationOpen *bool      `json:"isRegistrationOpen,omitempty"`
	IsSubmissionOpen   *bool      `json:"isSubmissionOpen,omitempty"`
	IsVisible          *bool      `json:"isVisible,omitempty"`
}

type ContestWithStats struct {
	Contest
	SolvedTaskCount int64 `json:"solvedTaskCount"`
}

type UpcomingContest struct {
	Contest
}

type PastContestWithStats struct {
	Contest
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

type AddTaskToContest struct {
	TaskID  int64      `json:"taskId" validate:"required"`
	StartAt *time.Time `json:"startAt,omitempty"`
	EndAt   *time.Time `json:"endAt,omitempty"`
}

type RegistrationRequest struct {
	ID        int64                           `json:"id"`
	ContestID int64                           `json:"contestId"`
	UserID    int64                           `json:"userId"`
	User      User                            `json:"user"`
	Status    types.RegistrationRequestStatus `json:"status"`
	CreatedAt time.Time                       `json:"createdAt"`
}

type ContestTask struct {
	Task             TaskInfo   `json:"task"`
	CreatorName      string     `json:"creatorName"`
	StartAt          time.Time  `json:"startAt"`
	EndAt            *time.Time `json:"endAt"`
	IsSubmissionOpen bool       `json:"isSubmissionOpen"`
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
