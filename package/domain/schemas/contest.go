package schemas

import "time"

type Contest struct {
	ID                 int64      `json:"id"`
	Name               string     `json:"name"`
	Description        string     `json:"description"`
	CreatedBy          int64      `json:"createdBy"`
	StartAt            *time.Time `json:"startAt"`
	EndAt              *time.Time `json:"endAt"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	ParticipantCount   int64      `json:"participantCount"`
	TaskCount          int64      `json:"taskCount"`
	RegistrationStatus string     `json:"registrationStatus"` // "registered", "canRegister", "awaitingApproval", "registrationClosed"
}

type CreateContest struct {
	Name               string     `json:"name" validate:"required,gte=3,lte=100"`
	Description        string     `json:"description,omitempty"`
	StartAt            *time.Time `json:"startAt,omitempty"`
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
