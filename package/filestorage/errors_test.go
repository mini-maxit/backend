//nolint:testpackage // to access unexported identifiers for testing
package filestorage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Export internal types for testing
// This file is only included in test builds

// NewDecompressor creates a new decompressor for testing.
func NewDecompressor() Decompressor {
	return &decompressor{}
}

// RemoveDirectory is exported for testing.
func RemoveDirectory(path string) error {
	return removeDirectory(path)
}

// TestableFileStorageService exposes internal methods for testing.
type TestableFileStorageService struct {
	*fileStorageService
}

// NewTestableFileStorageService creates a FileStorageService with injected dependencies for testing.
func NewTestableFileStorageService(decompressor Decompressor, validator ArchiveValidator, bucketName string) *TestableFileStorageService {
	return &TestableFileStorageService{
		fileStorageService: &fileStorageService{
			decompressor: decompressor,
			validator:    validator,
			bucketName:   bucketName,
		},
	}
}

// NormalizeFolderPath exposes the normalizeFolderPath method for testing.
func (f *TestableFileStorageService) NormalizeFolderPath(folderPath string) (string, error) {
	return f.normalizeFolderPath(folderPath)
}

// SetDecompressor allows injecting a custom decompressor (mock/fake) for tests.
func (f *TestableFileStorageService) SetDecompressor(d Decompressor) {
	f.decompressor = d
}

// SetValidator allows injecting a custom archive validator (mock/fake) for tests.
func (f *TestableFileStorageService) SetValidator(v ArchiveValidator) {
	f.validator = v
}

// SetBucketName allows overriding the bucket name for tests.
func (f *TestableFileStorageService) SetBucketName(name string) {
	f.bucketName = name
}

func TestValidationError(t *testing.T) {
	t.Run("Error returns formatted string without cause", func(t *testing.T) {
		err := &ValidationError{
			RuleName: "test-rule",
			Message:  "test message",
			Cause:    nil,
			Context:  map[string]any{"key": "value"},
		}

		expected := "validation error [test-rule]: test message"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("Error returns formatted string with cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := &ValidationError{
			RuleName: "test-rule",
			Message:  "test message",
			Cause:    cause,
			Context:  map[string]any{"key": "value"},
		}

		expected := "validation error [test-rule]: test message: underlying error"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("Unwrap returns cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := &ValidationError{
			RuleName: "test-rule",
			Message:  "test message",
			Cause:    cause,
		}

		unwrapped := err.Unwrap()
		assert.Equal(t, cause, unwrapped)
	})

	t.Run("Unwrap returns nil when no cause", func(t *testing.T) {
		err := &ValidationError{
			RuleName: "test-rule",
			Message:  "test message",
			Cause:    nil,
		}

		unwrapped := err.Unwrap()
		assert.NoError(t, unwrapped)
	})

	t.Run("Error is compatible with errors.Is", func(t *testing.T) {
		cause := errors.New("specific error")
		err := &ValidationError{
			RuleName: "test-rule",
			Message:  "test message",
			Cause:    cause,
		}

		require.ErrorIs(t, err, cause)
	})

	t.Run("Error is compatible with errors.As", func(t *testing.T) {
		err := &ValidationError{
			RuleName: "test-rule",
			Message:  "test message",
			Cause:    nil,
		}

		var validationErr *ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "test-rule", validationErr.RuleName)
	})
}

func TestDecompressionError(t *testing.T) {
	t.Run("Error returns formatted string without cause", func(t *testing.T) {
		err := &DecompressionError{
			ArchivePath: "/path/to/archive.zip",
			Message:     "failed to decompress",
			Cause:       nil,
			Context:     map[string]any{"key": "value"},
		}

		expected := "decompression error: failed to decompress"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("Error returns formatted string with cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := &DecompressionError{
			ArchivePath: "/path/to/archive.zip",
			Message:     "failed to decompress",
			Cause:       cause,
			Context:     map[string]any{"key": "value"},
		}

		expected := "decompression error: failed to decompress: underlying error"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("Unwrap returns cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := &DecompressionError{
			ArchivePath: "/path/to/archive.zip",
			Message:     "failed to decompress",
			Cause:       cause,
		}

		unwrapped := err.Unwrap()
		assert.Equal(t, cause, unwrapped)
	})

	t.Run("Unwrap returns nil when no cause", func(t *testing.T) {
		err := &DecompressionError{
			ArchivePath: "/path/to/archive.zip",
			Message:     "failed to decompress",
			Cause:       nil,
		}

		unwrapped := err.Unwrap()
		assert.NoError(t, unwrapped)
	})

	t.Run("Error is compatible with errors.Is", func(t *testing.T) {
		cause := errors.New("specific error")
		err := &DecompressionError{
			ArchivePath: "/path/to/archive.zip",
			Message:     "failed to decompress",
			Cause:       cause,
		}

		require.ErrorIs(t, err, cause)
	})

	t.Run("Error is compatible with errors.As", func(t *testing.T) {
		err := &DecompressionError{
			ArchivePath: "/path/to/archive.zip",
			Message:     "failed to decompress",
			Cause:       nil,
		}

		var decompressionErr *DecompressionError
		require.ErrorAs(t, err, &decompressionErr)
		assert.Equal(t, "/path/to/archive.zip", decompressionErr.ArchivePath)
	})
}
