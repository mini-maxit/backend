package schemas

import (
	"time"

	"github.com/mini-maxit/backend/package/domain/types"
)

type Submission struct {
	ID          int64                  `json:"id"`
	TaskID      int64                  `json:"taskId"`
	UserID      int64                  `json:"userId"`
	ContestID   *int64                 `json:"contestId,omitempty"`
	Order       int                    `json:"order"`
	LanguageID  int64                  `json:"languageId"`
	Status      types.SubmissionStatus `json:"status"`
	SubmittedAt time.Time              `json:"submittedAt"`
	CheckedAt   time.Time              `json:"checkedAt"`
	Language    LanguageConfig         `json:"language"`
	Task        Task                   `json:"task"`
	User        User                   `json:"user"`
	Result      *SubmissionResult      `json:"result"`
}

type SubmissionShort struct {
	ID            int64 `json:"id"`
	TaskID        int64 `json:"taskId"`
	UserID        int64 `json:"userId"`
	Passed        bool  `json:"passed"`
	HowManyPassed int64 `json:"howManyPassed"`
}

type SubmissionResult struct {
	ID           int64        `json:"id"`
	SubmissionID int64        `json:"submissionId"`
	Code         string       `json:"code"`
	Message      string       `json:"message"`
	CreatedAt    time.Time    `json:"createdAt"`
	TestResults  []TestResult `json:"testResults"`
}

type TestResult struct {
	ID                 int64  `json:"id"`
	SubmissionResultID int64  `json:"submissionResultId"`
	TestCaseID         int64  `json:"testCaseId"`
	Passed             bool   `json:"passed"`
	Code               string `json:"code"`
	ErrorMessage       string `json:"errorMessage"`
}

// ContestTaskStats contains aggregated statistics for a task in a contest
type ContestTaskStats struct {
	TaskID               int64   `json:"taskId"`
	TaskTitle            string  `json:"taskTitle"`
	TotalParticipants    int64   `json:"totalParticipants"`
	SubmittedCount       int64   `json:"submittedCount"`
	FullySolvedCount     int64   `json:"fullySolvedCount"`
	PartiallySolvedCount int64   `json:"partiallySolvedCount"`
	SuccessRate          float64 `json:"successRate"`
	AverageScore         float64 `json:"averageScore"`
}

// TaskUserStats contains statistics for a user on a specific task
type TaskUserStats struct {
	UserID           int64                  `json:"userId"`
	Username         string                 `json:"username"`
	SubmissionCount  int                    `json:"submissionCount"`
	BestScore        float64                `json:"bestScore"`
	LatestStatus     types.SubmissionStatus `json:"latestStatus"`
	LatestResultCode string                 `json:"latestResultCode"`
}

// UserContestStats contains overall performance statistics for a user in a contest
type UserContestStats struct {
	UserID               int64                 `json:"userId"`
	Username             string                `json:"username"`
	TasksAttempted       int                   `json:"tasksAttempted"`
	TasksSolved          int                   `json:"tasksSolved"`
	TasksPartiallySolved int                   `json:"tasksPartiallySolved"`
	TaskBreakdown        []UserTaskPerformance `json:"taskBreakdown"`
}

// UserTaskPerformance contains performance details for a user on a specific task
type UserTaskPerformance struct {
	TaskID       int64   `json:"taskId"`
	TaskTitle    string  `json:"taskTitle"`
	BestScore    float64 `json:"bestScore"`
	AttemptCount int     `json:"attemptCount"`
	IsSolved     bool    `json:"isSolved"`
}
