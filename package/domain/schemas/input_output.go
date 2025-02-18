package schemas

type InputOutput struct {
	Id          int64   `json:"-"`
	TaskId      int64   `json:"task_id"`
	Order       int     `json:"order"`
	TimeLimit   float64 `json:"time_limit"`
	MemoryLimit float64 `json:"memory_limit"`
}
