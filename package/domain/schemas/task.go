package schemas

import "time"

type EditTask struct {
	Title *string `json:"title,omitempty"`
}

type Task struct {
	Id        int64     `json:"id"`
	Title     string    `json:"title"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TaskDetailed struct {
	Id             int64     `json:"id"`
	Title          string    `json:"title"`
	DescriptionURL string    `json:"description_url"`
	CreatedBy      int64     `json:"created_by"`
	CreatedByName  string    `json:"created_by_name"`
	CreatedAt      time.Time `json:"created_at"`
	GroupIds       []int64   `json:"group_ids"`
}

type TaskCreateResponse struct {
	Id int64 `json:"id"`
}
