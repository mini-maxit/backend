package schemas

type SubmitResponse struct {
	Message          string `json:"message"`
	SubmissionNumber int64  `json:"submissionNumber"`
}
