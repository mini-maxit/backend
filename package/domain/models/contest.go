package models

import "time"

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
	ContestID        int64     `gorm:"primaryKey"`
	TaskID           int64     `gorm:"primaryKey"`
	StartAt          time.Time `gorm:"null"` // null means immediately available
	EndAt            time.Time `gorm:"null"` // null means available until the end of the contest
	IsSubmissionOpen *bool     `gorm:"default:true;not null"`
}

type ContestParticipant struct {
	ContestID int64 `gorm:"primaryKey"`
	UserID    int64 `gorm:"primaryKey"`
}

type ContestParticipantGroup struct {
	ContestID int64 `gorm:"primaryKey"`
	GroupID   int64 `gorm:"primaryKey"`
}
