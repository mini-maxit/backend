package schemas

import "time"

type EditTask struct {
	Title     *string `json:"title,omitempty"`
	IsVisible *bool   `json:"isVisible,omitempty"`
}

type Task struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	CreatedBy int64     `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	IsVisible bool      `json:"isVisible"`
}

// Struct to embed basic task info
type TaskInfo struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type TaskDetailed struct {
	ID             int64     `json:"id"`
	Title          string    `json:"title"`
	DescriptionURL string    `json:"descriptionUrl"`
	CreatedBy      int64     `json:"createdBy"`
	CreatedByName  string    `json:"createdByName"`
	CreatedAt      time.Time `json:"createdAt"`
}

type TaskCreateResponse struct {
	ID int64 `json:"id"`
}

type TaskWithContestStats struct {
	ID              int64           `json:"id"`
	Title           string          `json:"title"`
	CreatedBy       int64           `json:"createdBy"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
	AttemptsSummary AttemptsSummary `json:"attemptsSummary,omitempty"`
}

type AttemptsSummary struct {
	BestScore    float64 `json:"bestScore"`    // Best score achieved by the user (percentage of passed tests)
	AttemptCount int     `json:"attemptCount"` // Number of submission attempts by the user
}

type TaskWithAttempts struct {
	Task
	AttemptsSummary AttemptsSummary `json:"attemptsSummary,omitempty"`
}

type ContestWithTasks struct {
	ContestID   int64              `json:"contestId"`
	ContestName string             `json:"contestName"`
	StartAt     time.Time          `json:"startAt"`
	EndAt       *time.Time         `json:"endAt"`
	Tasks       []TaskWithAttempts `json:"tasks"`
}

type MyTasksResponse struct {
	Contests []ContestWithTasks `json:"contests"`
}
