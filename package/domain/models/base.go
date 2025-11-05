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

// GetSchemaTableName returns a table name with schema prefix
// Example: GetSchemaTableName("users") returns "maxit.users"
func GetSchemaTableName(tableName string) string {
	// Import the schema name from database package to avoid circular imports
	// For now, we'll hardcode it here, but it could be made configurable
	return "maxit." + tableName
}
