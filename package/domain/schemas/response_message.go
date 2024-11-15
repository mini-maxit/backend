package schemas

type ResponseMessage struct {
	MessageId string `json:"message_id"`
	Result    Result `json:"result"`
}

type Result struct {
	Success     bool         `json:"Success"`
	StatusCode  int64        `json:"StatusCode"`
	Code        string       `json:"Code"`
	Message     string       `json:"Message"`
	TestResults []TestResult `json:"TestResults"`
}

type TestResult struct {
	Order        int64  `json:"Order"`
	Passed       bool   `json:"Passed"`
	ErrorMessage string `json:"ErrorMessage"`
}
