package schemas

type QueueResponseMessage struct {
	MessageId string      `json:"message_id"`
	Result    QueueResult `json:"result"`
}

type QueueResult struct {
	Success     bool              `json:"Success"`
	StatusCode  int64             `json:"StatusCode"`
	Code        string            `json:"Code"`
	Message     string            `json:"Message"`
	TestResults []QueueTestResult `json:"TestResults"`
}

type QueueTestResult struct {
	Order        int64  `json:"Order"`
	Passed       bool   `json:"Passed"`
	ErrorMessage string `json:"ErrorMessage"`
}
