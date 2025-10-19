package types

type SubmissionStatus string

const (
	SubmissionStatusReceived          SubmissionStatus = "received"
	SubmissionStatusSentForEvaluation SubmissionStatus = "sent for evaluation"
	SubmissionStatusEvaluated         SubmissionStatus = "evaluated"
	SubmissionStatusLost              SubmissionStatus = "lost"
)

type SubmissionResultCode int

const (
	SubmissionResultCodeUnknown SubmissionResultCode = iota // used internally when creating empty submissionResult
	// This statuses are returned by worker
	SubmissionResultCodeSuccess
	SubmissionResultCodeTestFailed
	SubmissionResultCodeCompilationError
	SubmissionResultCodeInitializationError
	SubmissionResultCodeInternlError
	// backend only
	SubmissionResultCodeInvalid // used when received status is not one of the above
)

func (s SubmissionResultCode) IsValid() bool {
	return s >= SubmissionResultCodeSuccess && s <= SubmissionResultCodeInternlError
}

func (s SubmissionResultCode) String() string {
	return [...]string{
		"unknown",
		"success",
		"test_failed",
		"compilation_error",
		"initialization_error",
		"internal_error",
		"invalid",
	}[s]
}

type TestResultStatusCode int

const (
	TestResultStatusCodeOK TestResultStatusCode = iota + 1
	TestResultStatusCodeOutputDifference
	TestResultStatusCodeTimeLimit
	TestResultStatusCodeMemoryLimit
	TestResultStatusCodeRuntimeError
	// backend only
	TestResultStatusCodeNotExecuted // used when creating empty test result
	TestResultStatusCodeInvalid     // used when received status is not one of the above
)

func (t TestResultStatusCode) String() string {
	return [...]string{
		"ok",
		"output_difference",
		"time_limit_exceeded",
		"memory_limit_exceeded",
		"runtime_error",
		"not_executed",
		"invalid",
	}[t-1]
}

func (t TestResultStatusCode) IsValid() bool {
	return t >= TestResultStatusCodeOK && t < TestResultStatusCodeNotExecuted
}
