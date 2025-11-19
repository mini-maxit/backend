package types

type ContestStatus string

const (
	ContestStatusUpcoming ContestStatus = "upcoming"
	ContestStatusOngoing  ContestStatus = "ongoing"
	ContestStatusPast     ContestStatus = "past"
)
