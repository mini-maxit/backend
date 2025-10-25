package types_test

import (
	"testing"

	"github.com/mini-maxit/backend/package/domain/types"
)

func TestSubmissionResultCodeString(t *testing.T) {
	tests := []struct {
		name     string
		code     types.SubmissionResultCode
		expected string
	}{
		{"Unknown", types.SubmissionResultCodeUnknown, "unknown"},
		{"Success", types.SubmissionResultCodeSuccess, "success"},
		{"TestFailed", types.SubmissionResultCodeTestFailed, "test_failed"},
		{"CompilationError", types.SubmissionResultCodeCompilationError, "compilation_error"},
		{"InitializationError", types.SubmissionResultCodeInitializationError, "initialization_error"},
		{"InternalError", types.SubmissionResultCodeInternlError, "internal_error"},
		{"Invalid", types.SubmissionResultCodeInvalid, "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("String() method panicked for %s: %v", tt.name, r)
				}
			}()

			result := tt.code.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestTestResultStatusCodeString(t *testing.T) {
	tests := []struct {
		name     string
		code     types.TestResultStatusCode
		expected string
	}{
		{"OK", types.TestResultStatusCodeOK, "ok"},
		{"OutputDifference", types.TestResultStatusCodeOutputDifference, "output_difference"},
		{"TimeLimit", types.TestResultStatusCodeTimeLimit, "time_limit_exceeded"},
		{"MemoryLimit", types.TestResultStatusCodeMemoryLimit, "memory_limit_exceeded"},
		{"RuntimeError", types.TestResultStatusCodeRuntimeError, "runtime_error"},
		{"NotExecuted", types.TestResultStatusCodeNotExecuted, "not_executed"},
		{"Invalid", types.TestResultStatusCodeInvalid, "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("String() method panicked for %s: %v", tt.name, r)
				}
			}()

			result := tt.code.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
