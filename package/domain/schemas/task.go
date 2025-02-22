package schemas

import "time"

type EditTask struct {
	Title string `json:"title"`
}

type Task struct {
	Id        int64     `json:"id"`
	Title     string    `json:"title"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

type TaskDetailed struct {
	Id             int64     `json:"id"`
	Title          string    `json:"title"`
	DescriptionURL string    `json:"description_url"`
	CreatedBy      int64     `json:"created_by"`
	CreatedByName  string    `json:"created_by_name"`
	CreatedAt      time.Time `json:"created_at"`
}

type TaskCreateResponse struct {
	Id int64 `json:"id"`
}
