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
	&Group{},
	&UserGroup{},
	&Submission{},
	&SubmissionResult{},
	&TestResult{},
	&Contest{},
	&ContestParticipant{},
	&ContestParticipantGroup{},
	&ContestTask{},
	&ContestRegistrationRequests{},
	&ContestCollaborator{},
	&TaskCollaborator{},
	&File{},
	&TestCase{},
	&LanguageConfig{},
	&QueueMessage{},
}
