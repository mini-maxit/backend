package models

// UserContestStatsModel represents aggregated statistics for a single user in a contest.
// It is intended to be populated directly from a SQL query using gorm.DB.Raw(...).Scan(&[]UserContestStatsModel).
// The TaskBreakdownJson field holds a JSON array ( produced by json_agg(...) ) that can be unmarshaled
// into []UserTaskPerformanceModel at the service layer.
type UserContestStatsModel struct {
	User                 User  `gorm:"embedded;embeddedPrefix:user_"`
	TasksAttempted       int64 `gorm:"column:tasks_attempted"`
	TasksSolved          int64 `gorm:"column:tasks_solved"`
	TasksPartiallySolved int64 `gorm:"column:tasks_partially_solved"`
	// Raw JSON produced by the SQL query (json_agg of per-task objects).
	// Service layer can unmarshal this into []UserTaskPerformanceModel.
	TaskBreakdownJson string `gorm:"column:task_breakdown"`
}

// UserTaskPerformanceModel represents performance details for a user on a single task inside a contest.
// This struct matches the JSON objects constructed in the repository SQL (json_build_object(...)).
type UserTaskPerformanceModel struct {
	TaskID       int64   `json:"taskId" gorm:"column:task_id"`
	TaskTitle    string  `json:"taskTitle" gorm:"column:task_title"`
	BestScore    float64 `json:"bestScore" gorm:"column:best_score"`
	AttemptCount int     `json:"attemptCount" gorm:"column:attempt_count"`
	IsSolved     bool    `json:"isSolved" gorm:"column:is_solved"`
}

// TaskUserStatsModel represents per-user statistics for a specific task within a contest.
// Used by GetUserStatsForContestTask repository method.
type TaskUserStatsModel struct {
	User             User    `gorm:"embedded;embeddedPrefix:user_"`
	SubmissionCount  int64   `gorm:"column:submission_count"`
	BestScore        float64 `gorm:"column:best_score"`
	BestSubmissionID int64   `gorm:"column:best_submission_id"`
}

// ContestTaskStatsModel contains aggregated statistics for each task in a contest.
// Used by GetTaskStatsForContest repository method.
type ContestTaskStatsModel struct {
	Task                 Task    `gorm:"embedded;embeddedPrefix:task_"`
	TotalParticipants    int64   `gorm:"column:total_participants"`
	SubmittedCount       int64   `gorm:"column:submitted_count"`
	FullySolvedCount     int64   `gorm:"column:fully_solved_count"`
	PartiallySolvedCount int64   `gorm:"column:partially_solved_count"`
	AverageScore         float64 `gorm:"column:average_score"`
}

// UserContestStatsFull is the structured form (no raw JSON) used by the two-query repository approach.
type UserContestStatsFull struct {
	User                 User
	TasksAttempted       int64
	TasksSolved          int64
	TasksPartiallySolved int64
	TaskBreakdown        []UserTaskPerformanceModel
}

// UserContestSummaryRow maps the first (summary) query result per user.
type UserContestSummaryRow struct {
	UserID               int64  `gorm:"column:user_id"`
	UserUsername         string `gorm:"column:user_username"`
	UserName             string `gorm:"column:user_name"`
	UserSurname          string `gorm:"column:user_surname"`
	TasksAttempted       int64  `gorm:"column:tasks_attempted"`
	TasksSolved          int64  `gorm:"column:tasks_solved"`
	TasksPartiallySolved int64  `gorm:"column:tasks_partially_solved"`
}

// UserTaskPerformanceRow maps the second query rows (one per (user, task)).
type UserTaskPerformanceRow struct {
	UserID       int64   `gorm:"column:user_id"`
	TaskID       int64   `gorm:"column:task_id"`
	TaskTitle    string  `gorm:"column:task_title"`
	BestScore    float64 `gorm:"column:best_score"`
	AttemptCount int64   `gorm:"column:attempt_count"`
	IsSolved     bool    `gorm:"column:is_solved"`
}
