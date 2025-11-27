package filestorage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mini-maxit/backend/package/filestorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test directory structure
func setupTestDirectory(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()

	// Create input directory with valid files
	inputDir := filepath.Join(tempDir, "input")
	require.NoError(t, os.MkdirAll(inputDir, 0755))

	// Create output directory with valid files
	outputDir := filepath.Join(tempDir, "output")
	require.NoError(t, os.MkdirAll(outputDir, 0755))

	// Create description.pdf
	descPath := filepath.Join(tempDir, "description.pdf")
	require.NoError(t, os.WriteFile(descPath, []byte("test description"), 0644))

	// Create input files
	require.NoError(t, os.WriteFile(filepath.Join(inputDir, "1.txt"), []byte("input 1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(inputDir, "2.txt"), []byte("input 2"), 0644))

	// Create output files
	require.NoError(t, os.WriteFile(filepath.Join(outputDir, "1.txt"), []byte("output 1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(outputDir, "2.txt"), []byte("output 2"), 0644))

	return tempDir
}

func TestNewArchiveValidator(t *testing.T) {
	t.Run("Create new archive validator", func(t *testing.T) {
		validator := filestorage.NewArchiveValidator()
		assert.NotNil(t, validator)
	})
}

func TestArchiveValidatorAddRule(t *testing.T) {
	t.Run("Add single rule", func(t *testing.T) {
		validator := filestorage.NewArchiveValidator()
		rule := &filestorage.NonEmptyArchiveRule{}
		validator.AddRule(rule)
		// Validator should not panic and rule should be added
		assert.NotNil(t, validator)
	})

	t.Run("Add multiple rules", func(t *testing.T) {
		validator := filestorage.NewArchiveValidator()
		validator.AddRule(&filestorage.NonEmptyArchiveRule{})
		validator.AddRule(&filestorage.RequiredEntriesRule{RequiredEntries: []string{"input", "output"}})
		assert.NotNil(t, validator)
	})
}

func TestArchiveValidatorValidate(t *testing.T) {
	t.Run("Validate with passing rules", func(t *testing.T) {
		tempDir := setupTestDirectory(t)

		validator := filestorage.NewArchiveValidator()
		validator.AddRule(&filestorage.NonEmptyArchiveRule{})

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  tempDir,
		}

		err := validator.Validate(ctx)
		require.NoError(t, err)
	})

	t.Run("Validate with failing rule", func(t *testing.T) {
		tempDir := t.TempDir()
		// Create an empty directory

		validator := filestorage.NewArchiveValidator()
		validator.AddRule(&filestorage.NonEmptyArchiveRule{})

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  tempDir,
		}

		err := validator.Validate(ctx)
		require.Error(t, err)

		var validationErr *filestorage.ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "non-empty-archive", validationErr.RuleName)
	})

	t.Run("Validate with no rules", func(t *testing.T) {
		tempDir := t.TempDir()

		validator := filestorage.NewArchiveValidator()

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  tempDir,
		}

		err := validator.Validate(ctx)
		require.NoError(t, err)
	})

	t.Run("Stop at first failing rule", func(t *testing.T) {
		tempDir := t.TempDir()
		// Empty directory will fail non-empty rule

		validator := filestorage.NewArchiveValidator()
		validator.AddRule(&filestorage.NonEmptyArchiveRule{})
		validator.AddRule(&filestorage.RequiredEntriesRule{RequiredEntries: []string{"input", "output"}})

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  tempDir,
		}

		err := validator.Validate(ctx)
		require.Error(t, err)

		var validationErr *filestorage.ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "non-empty-archive", validationErr.RuleName)
	})
}

func TestNonEmptyArchiveRule(t *testing.T) {
	rule := &filestorage.NonEmptyArchiveRule{}

	t.Run("Name returns correct value", func(t *testing.T) {
		assert.Equal(t, "non-empty-archive", rule.Name())
	})

	t.Run("Validate success with non-empty archive", func(t *testing.T) {
		tempDir := setupTestDirectory(t)

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  tempDir,
		}

		err := rule.Validate(ctx)
		require.NoError(t, err)
	})

	t.Run("Validate fails with empty archive", func(t *testing.T) {
		tempDir := t.TempDir()

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  tempDir,
		}

		err := rule.Validate(ctx)
		require.Error(t, err)

		var validationErr *filestorage.ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "non-empty-archive", validationErr.RuleName)
		assert.Contains(t, validationErr.Message, "empty")
	})

	t.Run("Validate fails with invalid path", func(t *testing.T) {
		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  "/nonexistent/path",
		}

		err := rule.Validate(ctx)
		require.Error(t, err)

		var validationErr *filestorage.ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "non-empty-archive", validationErr.RuleName)
		assert.Contains(t, validationErr.Message, "failed to read")
	})
}

func TestRequiredEntriesRule(t *testing.T) {
	t.Run("Name returns correct value", func(t *testing.T) {
		rule := &filestorage.RequiredEntriesRule{RequiredEntries: []string{"input", "output"}}
		assert.Equal(t, "required-entries", rule.Name())
	})

	t.Run("Validate success with all required entries", func(t *testing.T) {
		tempDir := setupTestDirectory(t)

		rule := &filestorage.RequiredEntriesRule{
			RequiredEntries: []string{"input", "output", "description.pdf"},
		}

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  tempDir,
		}

		err := rule.Validate(ctx)
		require.NoError(t, err)
	})

	t.Run("Validate fails with missing entry", func(t *testing.T) {
		tempDir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "input"), 0755))

		rule := &filestorage.RequiredEntriesRule{
			RequiredEntries: []string{"input", "output"},
		}

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  tempDir,
		}

		err := rule.Validate(ctx)
		require.Error(t, err)

		var validationErr *filestorage.ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "required-entries", validationErr.RuleName)
	})

	t.Run("Validate fails with extra entries", func(t *testing.T) {
		tempDir := setupTestDirectory(t)
		// Add an extra file
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "extra.txt"), []byte("extra"), 0644))

		rule := &filestorage.RequiredEntriesRule{
			RequiredEntries: []string{"input", "output", "description.pdf"},
		}

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  tempDir,
		}

		err := rule.Validate(ctx)
		require.Error(t, err)

		var validationErr *filestorage.ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "required-entries", validationErr.RuleName)
		assert.Contains(t, validationErr.Message, "exactly")
	})

	t.Run("Validate fails with invalid path", func(t *testing.T) {
		rule := &filestorage.RequiredEntriesRule{RequiredEntries: []string{"input", "output"}}

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  "/nonexistent/path",
		}

		err := rule.Validate(ctx)
		require.Error(t, err)

		var validationErr *filestorage.ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "required-entries", validationErr.RuleName)
		assert.Contains(t, validationErr.Message, "failed to read")
	})
}

func TestDirectoryFilesRule(t *testing.T) {
	t.Run("Name returns correct value with directory name", func(t *testing.T) {
		rule := &filestorage.DirectoryFilesRule{
			Config: filestorage.DirectoryConfig{
				Name:               "input",
				AcceptedExtensions: []string{".txt"},
			},
		}
		assert.Equal(t, "directory-files-input", rule.Name())
	})

	t.Run("Validate success with valid files", func(t *testing.T) {
		tempDir := setupTestDirectory(t)

		rule := &filestorage.DirectoryFilesRule{
			Config: filestorage.DirectoryConfig{
				Name:               "input",
				AcceptedExtensions: []string{".txt"},
				RequireSequential:  true,
			},
		}

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  tempDir,
		}

		err := rule.Validate(ctx)
		require.NoError(t, err)
	})

	t.Run("Validate fails with subdirectory in directory", func(t *testing.T) {
		tempDir := t.TempDir()
		inputDir := filepath.Join(tempDir, "input")
		require.NoError(t, os.MkdirAll(inputDir, 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(inputDir, "subdir"), 0755))

		rule := &filestorage.DirectoryFilesRule{
			Config: filestorage.DirectoryConfig{
				Name:               "input",
				AcceptedExtensions: []string{".txt"},
			},
		}

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  tempDir,
		}

		err := rule.Validate(ctx)
		require.Error(t, err)

		var validationErr *filestorage.ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Contains(t, validationErr.Message, "subdirectories")
	})

	t.Run("Validate fails with wrong extension", func(t *testing.T) {
		tempDir := t.TempDir()
		inputDir := filepath.Join(tempDir, "input")
		require.NoError(t, os.MkdirAll(inputDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(inputDir, "1.json"), []byte("test"), 0644))

		rule := &filestorage.DirectoryFilesRule{
			Config: filestorage.DirectoryConfig{
				Name:               "input",
				AcceptedExtensions: []string{".txt", ".in"},
			},
		}

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  tempDir,
		}

		err := rule.Validate(ctx)
		require.Error(t, err)

		var validationErr *filestorage.ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Contains(t, validationErr.Message, "extension")
	})

	t.Run("Validate fails with non-sequential naming", func(t *testing.T) {
		tempDir := t.TempDir()
		inputDir := filepath.Join(tempDir, "input")
		require.NoError(t, os.MkdirAll(inputDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(inputDir, "1.txt"), []byte("test"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(inputDir, "3.txt"), []byte("test"), 0644)) // Skipped 2

		rule := &filestorage.DirectoryFilesRule{
			Config: filestorage.DirectoryConfig{
				Name:               "input",
				AcceptedExtensions: []string{".txt"},
				RequireSequential:  true,
			},
		}

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  tempDir,
		}

		err := rule.Validate(ctx)
		require.Error(t, err)

		var validationErr *filestorage.ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Contains(t, validationErr.Message, "named")
	})

	t.Run("Validate fails with empty file", func(t *testing.T) {
		tempDir := t.TempDir()
		inputDir := filepath.Join(tempDir, "input")
		require.NoError(t, os.MkdirAll(inputDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(inputDir, "1.txt"), []byte{}, 0644)) // Empty file

		rule := &filestorage.DirectoryFilesRule{
			Config: filestorage.DirectoryConfig{
				Name:               "input",
				AcceptedExtensions: []string{".txt"},
				RequireSequential:  true,
			},
		}

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  tempDir,
		}

		err := rule.Validate(ctx)
		require.Error(t, err)

		var validationErr *filestorage.ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Contains(t, validationErr.Message, "empty")
	})

	t.Run("Validate fails with nonexistent directory", func(t *testing.T) {
		tempDir := t.TempDir()

		rule := &filestorage.DirectoryFilesRule{
			Config: filestorage.DirectoryConfig{
				Name:               "input",
				AcceptedExtensions: []string{".txt"},
			},
		}

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  tempDir,
		}

		err := rule.Validate(ctx)
		require.Error(t, err)

		var validationErr *filestorage.ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Contains(t, validationErr.Message, "failed to read")
	})

	t.Run("Validate success without sequential requirement", func(t *testing.T) {
		tempDir := t.TempDir()
		inputDir := filepath.Join(tempDir, "input")
		require.NoError(t, os.MkdirAll(inputDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(inputDir, "test1.txt"), []byte("test"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(inputDir, "test2.txt"), []byte("test"), 0644))

		rule := &filestorage.DirectoryFilesRule{
			Config: filestorage.DirectoryConfig{
				Name:               "input",
				AcceptedExtensions: []string{".txt"},
				RequireSequential:  false,
			},
		}

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  tempDir,
		}

		err := rule.Validate(ctx)
		require.NoError(t, err)
	})
}

func TestInputOutputMatchRule(t *testing.T) {
	rule := &filestorage.InputOutputMatchRule{}

	t.Run("Name returns correct value", func(t *testing.T) {
		assert.Equal(t, "input-output-match", rule.Name())
	})

	t.Run("Validate success with matching counts", func(t *testing.T) {
		tempDir := setupTestDirectory(t)

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  tempDir,
		}

		err := rule.Validate(ctx)
		require.NoError(t, err)
	})

	t.Run("Validate fails with mismatched counts", func(t *testing.T) {
		tempDir := t.TempDir()
		inputDir := filepath.Join(tempDir, "input")
		outputDir := filepath.Join(tempDir, "output")
		require.NoError(t, os.MkdirAll(inputDir, 0755))
		require.NoError(t, os.MkdirAll(outputDir, 0755))

		// Create 2 input files
		require.NoError(t, os.WriteFile(filepath.Join(inputDir, "1.txt"), []byte("input 1"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(inputDir, "2.txt"), []byte("input 2"), 0644))

		// Create 1 output file
		require.NoError(t, os.WriteFile(filepath.Join(outputDir, "1.txt"), []byte("output 1"), 0644))

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  tempDir,
		}

		err := rule.Validate(ctx)
		require.Error(t, err)

		var validationErr *filestorage.ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "input-output-match", validationErr.RuleName)
		assert.Contains(t, validationErr.Message, "same")
	})

	t.Run("Validate fails with empty directories", func(t *testing.T) {
		tempDir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "input"), 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "output"), 0755))

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  tempDir,
		}

		err := rule.Validate(ctx)
		require.Error(t, err)

		var validationErr *filestorage.ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Contains(t, validationErr.Message, "non-zero")
	})

	t.Run("Validate fails with missing input directory", func(t *testing.T) {
		tempDir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "output"), 0755))

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  tempDir,
		}

		err := rule.Validate(ctx)
		require.Error(t, err)

		var validationErr *filestorage.ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Contains(t, validationErr.Message, "input")
	})

	t.Run("Validate fails with missing output directory", func(t *testing.T) {
		tempDir := t.TempDir()
		inputDir := filepath.Join(tempDir, "input")
		require.NoError(t, os.MkdirAll(inputDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(inputDir, "1.txt"), []byte("input 1"), 0644))

		ctx := filestorage.ValidationContext{
			ArchivePath: "test.zip",
			FolderPath:  tempDir,
		}

		err := rule.Validate(ctx)
		require.Error(t, err)

		var validationErr *filestorage.ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Contains(t, validationErr.Message, "output")
	})
}
