package schemas

import "time"

type EditTask struct {
	Title *string `json:"title,omitempty"`
}

type Task struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	CreatedBy int64     `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TaskDetailed struct {
	ID             int64     `json:"id"`
	Title          string    `json:"title"`
	DescriptionURL string    `json:"descriptionUrl"`
	CreatedBy      int64     `json:"createdBy"`
	CreatedByName  string    `json:"createdByName"`
	CreatedAt      time.Time `json:"createdAt"`
	GroupIDs       []int64   `json:"groupIds"`
}

type TaskCreateResponse struct {
	ID int64 `json:"id"`
}
