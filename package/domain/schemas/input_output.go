package schemas

type TestCase struct {
	ID          int64 `json:"-"`
	TaskID      int64 `json:"-"`
	Order       int   `json:"order"`
	TimeLimit   int64 `json:"timeLimit"`
	MemoryLimit int64 `json:"memoryLimit"`
}

type PutTestCaseLimitsRequest struct {
	Limits []PutTestCase `json:"limits"`
}

type PutTestCase struct {
	Order       int   `json:"order" validate:"gt=0"`
	TimeLimit   int64 `json:"timeLimit" validate:"gt=0"`
	MemoryLimit int64 `json:"memoryLimit" validate:"gt=0"`
}
