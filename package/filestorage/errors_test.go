package filestorage_test

import (
	"errors"
	"testing"

	"github.com/mini-maxit/backend/package/filestorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationError(t *testing.T) {
	t.Run("Error returns formatted string without cause", func(t *testing.T) {
		err := &filestorage.ValidationError{
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
		err := &filestorage.ValidationError{
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
		err := &filestorage.ValidationError{
			RuleName: "test-rule",
			Message:  "test message",
			Cause:    cause,
		}

		unwrapped := err.Unwrap()
		assert.Equal(t, cause, unwrapped)
	})

	t.Run("Unwrap returns nil when no cause", func(t *testing.T) {
		err := &filestorage.ValidationError{
			RuleName: "test-rule",
			Message:  "test message",
			Cause:    nil,
		}

		unwrapped := err.Unwrap()
		assert.NoError(t, unwrapped)
	})

	t.Run("Error is compatible with errors.Is", func(t *testing.T) {
		cause := errors.New("specific error")
		err := &filestorage.ValidationError{
			RuleName: "test-rule",
			Message:  "test message",
			Cause:    cause,
		}

		require.ErrorIs(t, err, cause)
	})

	t.Run("Error is compatible with errors.As", func(t *testing.T) {
		err := &filestorage.ValidationError{
			RuleName: "test-rule",
			Message:  "test message",
			Cause:    nil,
		}

		var validationErr *filestorage.ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "test-rule", validationErr.RuleName)
	})
}

func TestDecompressionError(t *testing.T) {
	t.Run("Error returns formatted string without cause", func(t *testing.T) {
		err := &filestorage.DecompressionError{
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
		err := &filestorage.DecompressionError{
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
		err := &filestorage.DecompressionError{
			ArchivePath: "/path/to/archive.zip",
			Message:     "failed to decompress",
			Cause:       cause,
		}

		unwrapped := err.Unwrap()
		assert.Equal(t, cause, unwrapped)
	})

	t.Run("Unwrap returns nil when no cause", func(t *testing.T) {
		err := &filestorage.DecompressionError{
			ArchivePath: "/path/to/archive.zip",
			Message:     "failed to decompress",
			Cause:       nil,
		}

		unwrapped := err.Unwrap()
		assert.NoError(t, unwrapped)
	})

	t.Run("Error is compatible with errors.Is", func(t *testing.T) {
		cause := errors.New("specific error")
		err := &filestorage.DecompressionError{
			ArchivePath: "/path/to/archive.zip",
			Message:     "failed to decompress",
			Cause:       cause,
		}

		require.ErrorIs(t, err, cause)
	})

	t.Run("Error is compatible with errors.As", func(t *testing.T) {
		err := &filestorage.DecompressionError{
			ArchivePath: "/path/to/archive.zip",
			Message:     "failed to decompress",
			Cause:       nil,
		}

		var decompressionErr *filestorage.DecompressionError
		require.ErrorAs(t, err, &decompressionErr)
		assert.Equal(t, "/path/to/archive.zip", decompressionErr.ArchivePath)
	})
}
