package types

// ResourceType represents the type of resource in access control.
type ResourceType string

const (
	// ResourceTypeContest represents a contest resource.
	ResourceTypeContest ResourceType = "contest"
	// ResourceTypeTask represents a task resource.
	ResourceTypeTask ResourceType = "task"
	// ResourceTypeGroup represents a group resource.
	ResourceTypeGroup ResourceType = "group"
)
