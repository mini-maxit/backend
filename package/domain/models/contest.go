package models

import (
	"time"

	"github.com/mini-maxit/backend/package/domain/types"
)

type Contest struct {
	ID                 int64      `gorm:"primaryKey"`
	Name               string     `gorm:"not null;unique"`
	Description        string     `gorm:"type:text;null"`
	CreatedBy          int64      `gorm:"foreignKey:UserID;not null"`
	StartAt            *time.Time `gorm:"null"`
	EndAt              *time.Time `gorm:"null"`
	IsRegistrationOpen *bool      `gorm:"default:true;not null"`  // if false, only admins can add participants
	IsSubmissionOpen   *bool      `gorm:"default:false;not null"` // if true, contest is active and participants can submit solutions
	IsVisible          *bool      `gorm:"default:false;not null"` // if true, contest is visible to all users

	BaseModel

	Creator           User    `gorm:"foreignKey:CreatedBy; references:ID"`
	Tasks             []Task  `gorm:"many2many:contest_tasks;"`
	Participants      []User  `gorm:"many2many:contest_participants;"`
	ParticipantGroups []Group `gorm:"many2many:contest_participant_groups;"`
}

type ContestTask struct {
	ContestID        int64      `gorm:"primaryKey"`
	TaskID           int64      `gorm:"primaryKey"`
	StartAt          time.Time  `gorm:"not null"`
	EndAt            *time.Time `gorm:"null"`
	IsSubmissionOpen bool       `gorm:"default:true;not null"`

	Contest Contest `gorm:"foreignKey:ContestID;references:ID"`
	Task    Task    `gorm:"foreignKey:TaskID;references:ID"`
}

type ContestParticipant struct {
	ContestID int64 `gorm:"primaryKey"`
	UserID    int64 `gorm:"primaryKey"`

	Contest Contest `gorm:"foreignKey:ContestID;references:ID"`
	User    User    `gorm:"foreignKey:UserID;references:ID"`
}

type ContestParticipantGroup struct {
	ContestID int64 `gorm:"primaryKey"`
	GroupID   int64 `gorm:"primaryKey"`

	Contest Contest `gorm:"foreignKey:ContestID;references:ID"`
	Group   Group   `gorm:"foreignKey:GroupID;references:ID"`
}

type ContestRegistrationRequests struct {
	ID        int64                           `gorm:"primaryKey;autoIncrement"`
	ContestID int64                           `gorm:"not null"`
	UserID    int64                           `gorm:"not null"`
	Status    types.RegistrationRequestStatus `gorm:"type:registration_request_status;not null"`
	BaseModel

	Contest Contest `gorm:"foreignKey:ContestID;references:ID"`
	User    User    `gorm:"foreignKey:UserID;references:ID"`
}

// ContestWithStats extends Contest with computed fields for efficient querying
type ContestWithStats struct {
	Contest
	ParticipantCount int64 `gorm:"column:participant_count"`
	TaskCount        int64 `gorm:"column:task_count"`
	IsParticipant    bool  `gorm:"column:is_participant"`
	HasPendingReg    bool  `gorm:"column:has_pending_reg"`
}

type ParticipantContestStats struct {
	Contest
	ParticipantCount int64 `gorm:"column:participant_count"`
	TaskCount        int64 `gorm:"column:task_count"`
	SolvedCount      int64 `gorm:"column:solved_count"`
}
