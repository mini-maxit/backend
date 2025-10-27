package models

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	CreatedAt time.Time      `gorm:"autoCreateTime;default:current_timestamp"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;default:current_timestamp"`
	DeletedAt gorm.DeletedAt `gorm:"index;default:null"`
}

// AllModels contains all GORM model types for schema operations
// This is used by both the schema plugin and database loaders
var AllModels = []interface{}{
	&User{},
	&Task{},
	&TaskUser{},
	&Group{},
	&UserGroup{},
	&TaskGroup{},
	&Submission{},
	&SubmissionResult{},
	&TestResult{},
	&Contest{},
	&ContestParticipant{},
	&ContestParticipantGroup{},
	&ContestTask{},
	&ContestRegistrationRequests{},
	&File{},
	&TestCase{},
	&LanguageConfig{},
	&QueueMessage{},
}

// TableNames provides easy access to table names without schema prefix
// Use these with the schema plugin for automatic schema prefixing
var TableNames = struct {
	Users                       string
	Tasks                       string
	TaskUsers                   string
	Groups                      string
	UserGroups                  string
	TaskGroups                  string
	Submissions                 string
	SubmissionResults           string
	TestResults                 string
	Contests                    string
	ContestParticipants         string
	ContestParticipantGroups    string
	ContestTasks                string
	ContestPendingRegistrations string
	Files                       string
	TestCases                   string
	LanguageConfigs             string
	QueueMessages               string
}{
	Users:                       "users",
	Tasks:                       "tasks",
	TaskUsers:                   "task_users",
	Groups:                      "groups",
	UserGroups:                  "user_groups",
	TaskGroups:                  "task_groups",
	Submissions:                 "submissions",
	SubmissionResults:           "submission_results",
	TestResults:                 "test_results",
	Contests:                    "contests",
	ContestParticipants:         "contest_participants",
	ContestParticipantGroups:    "contest_participant_groups",
	ContestTasks:                "contest_tasks",
	ContestPendingRegistrations: "contest_pending_registrations",
	Files:                       "files",
	TestCases:                   "test_cases",
	LanguageConfigs:             "language_configs",
	QueueMessages:               "queue_messages",
}

// GetSchemaTableName returns a table name with schema prefix
// Example: GetSchemaTableName("users") returns "maxit.users"
func GetSchemaTableName(tableName string) string {
	// Import the schema name from database package to avoid circular imports
	// For now, we'll hardcode it here, but it could be made configurable
	return "maxit." + tableName
}

// SchemaTableNames provides full schema-prefixed table names
// Use these when you need explicit schema-prefixed names
var SchemaTableNames = struct {
	Users                       string
	Tasks                       string
	TaskUsers                   string
	Groups                      string
	UserGroups                  string
	TaskGroups                  string
	Submissions                 string
	SubmissionResults           string
	TestResults                 string
	Contests                    string
	ContestParticipants         string
	ContestParticipantGroups    string
	ContestTasks                string
	ContestPendingRegistrations string
	Files                       string
	TestCases                   string
	LanguageConfigs             string
	QueueMessages               string
}{
	Users:                       GetSchemaTableName(TableNames.Users),
	Tasks:                       GetSchemaTableName(TableNames.Tasks),
	TaskUsers:                   GetSchemaTableName(TableNames.TaskUsers),
	Groups:                      GetSchemaTableName(TableNames.Groups),
	UserGroups:                  GetSchemaTableName(TableNames.UserGroups),
	TaskGroups:                  GetSchemaTableName(TableNames.TaskGroups),
	Submissions:                 GetSchemaTableName(TableNames.Submissions),
	SubmissionResults:           GetSchemaTableName(TableNames.SubmissionResults),
	TestResults:                 GetSchemaTableName(TableNames.TestResults),
	Contests:                    GetSchemaTableName(TableNames.Contests),
	ContestParticipants:         GetSchemaTableName(TableNames.ContestParticipants),
	ContestParticipantGroups:    GetSchemaTableName(TableNames.ContestParticipantGroups),
	ContestTasks:                GetSchemaTableName(TableNames.ContestTasks),
	ContestPendingRegistrations: GetSchemaTableName(TableNames.ContestPendingRegistrations),
	Files:                       GetSchemaTableName(TableNames.Files),
	TestCases:                   GetSchemaTableName(TableNames.TestCases),
	LanguageConfigs:             GetSchemaTableName(TableNames.LanguageConfigs),
	QueueMessages:               GetSchemaTableName(TableNames.QueueMessages),
}
