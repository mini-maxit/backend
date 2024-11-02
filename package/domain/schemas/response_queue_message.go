package schemas

type ResponseMessage struct {
	MessageId      string `json:"message_id"`
	TaskId         int64  `json:"task_id"`
	UserId         int64  `json:"user_id"`
	UserSolutionId int64  `json:"user_solution_id"`
	Result         Result `json:"result"`
}

type Result struct {
	Success     bool         `json:"Success"`
	StatusCode  int64        `json:"StatusCode"`
	Code        string       `json:"Code"`
	Message     string       `json:"Message"`
	TestResults []TestResult `json:"TestResults"`
}

type TestResult struct {
	InputFile    string `json:"InputFile"`
	ExpectedFile string `json:"ExpectedFile"`
	ActualFile   string `json:"ActualFile"`
	Passed       bool   `json:"Passed"`
	ErrorMessage string `json:"ErrorMessage"`
	Order        int64  `json:"Order"`
}
