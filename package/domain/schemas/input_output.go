package schemas

type InputOutput struct {
	ID          int64   `json:"-"`
	TaskID      int64   `json:"task_id"`
	Order       int     `json:"order"`
	TimeLimit   float64 `json:"time_limit"`
	MemoryLimit float64 `json:"memory_limit"`
}
