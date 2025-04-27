package schemas

type InputOutput struct {
	ID          int64 `json:"-"`
	TaskID      int64 `json:"-"`
	Order       int   `json:"order"`
	TimeLimit   int64 `json:"time_limit"`
	MemoryLimit int64 `json:"memory_limit"`
}

type PutInputOutputRequest struct {
	Limits []PutInputOutput `json:"limits"`
}

type PutInputOutput struct {
	Order       int   `json:"order" validate:"gt=0"`
	TimeLimit   int64 `json:"time_limit" validate:"gt=0"`
	MemoryLimit int64 `json:"memory_limit" validate:"gt=0"`
}
