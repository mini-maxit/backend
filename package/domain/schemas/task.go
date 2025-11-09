package schemas

import "time"

type EditTask struct {
	Title *string `json:"title,omitempty"`
}

type Task struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	CreatedBy int64     `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TaskDetailed struct {
	ID             int64     `json:"id"`
	Title          string    `json:"title"`
	DescriptionURL string    `json:"descriptionUrl"`
	CreatedBy      int64     `json:"createdBy"`
	CreatedByName  string    `json:"createdByName"`
	CreatedAt      time.Time `json:"createdAt"`
	GroupIDs       []int64   `json:"groupIds"`
}

type TaskCreateResponse struct {
	ID int64 `json:"id"`
}

type TaskWithContestStats struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	CreatedBy    int64     `json:"createdBy"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	BestScore    *float64  `json:"bestScore"`    // Best score achieved by the user (percentage of passed tests)
	AttemptCount int       `json:"attemptCount"` // Number of submission attempts by the user
}

type TaskWithAttempts struct {
	Task
	BestScore    *float64 `json:"bestScore,omitempty"`    // Best score achieved by the user (percentage of passed tests)
	AttemptCount int      `json:"attemptCount,omitempty"` // Number of submission attempts by the user
}

type ContestWithTasks struct {
	ContestID   int64              `json:"contestId"`
	ContestName string             `json:"contestName"`
	StartAt     *time.Time         `json:"startAt"`
	EndAt       *time.Time         `json:"endAt"`
	Tasks       []TaskWithAttempts `json:"tasks"`
}

type MyTasksResponse struct {
	Contests        []ContestWithTasks `json:"contests"`
	NonContestTasks []TaskWithAttempts `json:"nonContestTasks"`
}
