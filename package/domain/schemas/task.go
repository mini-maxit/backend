package schemas

type UpdateTask struct {
	Title string `json:"title"`
}

type Task struct {
	Title     string `json:"title"`
	CreatedBy int64  `json:"created_by"`
}
