package filestorage

import (
	"fmt"
)

// ValidationError represents validation-related errors
type ValidationError struct {
	RuleName string
	Message  string
	Cause    error
	Context  map[string]any
}

func (e *ValidationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("validation error [%s]: %s: %v", e.RuleName, e.Message, e.Cause)
	}
	return fmt.Sprintf("validation error [%s]: %s", e.RuleName, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Cause
}

// DecompressionError represents decompression-related errors
type DecompressionError struct {
	ArchivePath string
	Message     string
	Cause       error
	Context     map[string]any
}

func (e *DecompressionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("decompression error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("decompression error: %s", e.Message)
}

func (e *DecompressionError) Unwrap() error {
	return e.Cause
}
