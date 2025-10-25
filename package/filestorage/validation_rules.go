package filestorage

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/mini-maxit/backend/package/utils"
)

var logger = utils.NewNamedLogger("filestorageValidator")

// ValidationContext provides context for validation operations
type ValidationContext struct {
	ArchivePath string
	FolderPath  string
}

// ValidationRule represents a single validation rule
type ValidationRule interface {
	Validate(ctx ValidationContext) error
	Name() string
}

// ArchiveValidator orchestrates multiple validation rules
type ArchiveValidator interface {
	AddRule(rule ValidationRule)
	Validate(ctx ValidationContext) error
}

type archiveValidator struct {
	rules []ValidationRule
}

func NewArchiveValidator() ArchiveValidator {
	return &archiveValidator{
		rules: make([]ValidationRule, 0),
	}
}

func (v *archiveValidator) AddRule(rule ValidationRule) {
	v.rules = append(v.rules, rule)
}

func (v *archiveValidator) Validate(ctx ValidationContext) error {
	for _, rule := range v.rules {
		if err := rule.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

// NonEmptyArchiveRule validates that the archive is not empty
type NonEmptyArchiveRule struct{}

func (r *NonEmptyArchiveRule) Name() string {
	return "non-empty-archive"
}

func (r *NonEmptyArchiveRule) Validate(ctx ValidationContext) error {
	entries, err := os.ReadDir(ctx.FolderPath)
	if err != nil {
		logger.Error("Failed to read decompressed archive directory.", "error", err)
		return &ValidationError{
			RuleName: r.Name(),
			Message:  "failed to read decompressed archive directory",
			Cause:    err,
			Context: map[string]interface{}{
				"folder_path": ctx.FolderPath,
			},
		}
	}
	if len(entries) == 0 {
		return &ValidationError{
			RuleName: r.Name(),
			Message:  "archive is empty or does not contain any files",
			Cause:    nil,
			Context: map[string]interface{}{
				"folder_path": ctx.FolderPath,
			},
		}
	}
	return nil
}

// RequiredEntriesRule validates that required entries are present
type RequiredEntriesRule struct {
	RequiredEntries []string
}

func (r *RequiredEntriesRule) Name() string {
	return "required-entries"
}

func (r *RequiredEntriesRule) Validate(ctx ValidationContext) error {
	entries, err := os.ReadDir(ctx.FolderPath)
	if err != nil {
		return &ValidationError{
			RuleName: r.Name(),
			Message:  "failed to read directory",
			Cause:    err,
			Context: map[string]interface{}{
				"folder_path": ctx.FolderPath,
			},
		}
	}

	entryNames := make([]string, len(entries))
	for i, entry := range entries {
		entryNames[i] = entry.Name()
	}

	if len(entries) != len(r.RequiredEntries) {
		context := map[string]interface{}{
			"required_entries": r.RequiredEntries,
			"found_entries":    entryNames,
			"expected_count":   len(r.RequiredEntries),
			"actual_count":     len(entries),
		}
		return &ValidationError{
			RuleName: r.Name(),
			Message: fmt.Sprintf("archive should contain exactly %d entries: %v, found: %v",
				len(r.RequiredEntries), r.RequiredEntries, entryNames),
			Cause:   nil,
			Context: context,
		}
	}

	for _, required := range r.RequiredEntries {
		if !slices.Contains(entryNames, required) {
			context := map[string]interface{}{
				"missing_entry":    required,
				"found_entries":    entryNames,
				"required_entries": r.RequiredEntries,
			}
			return &ValidationError{
				RuleName: r.Name(),
				Message:  fmt.Sprintf("archive is missing '%s' file, found entries: %v", required, entryNames),
				Cause:    nil,
				Context:  context,
			}
		}
	}

	return nil
}

// DirectoryConfig represents configuration for directory validation
type DirectoryConfig struct {
	Name               string
	AcceptedExtensions []string
	RequireSequential  bool
}

// DirectoryFilesRule validates files in a directory according to specific rules
type DirectoryFilesRule struct {
	Config DirectoryConfig
}

func (r *DirectoryFilesRule) Name() string {
	return fmt.Sprintf("directory-files-%s", r.Config.Name)
}

func (r *DirectoryFilesRule) Validate(ctx ValidationContext) error {
	dirPath := filepath.Join(ctx.FolderPath, r.Config.Name)
	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		logger.Error("Failed to read directory in the archive.", "directory", r.Config.Name, "error", err)
		return &ValidationError{
			RuleName: r.Name(),
			Message:  fmt.Sprintf("failed to read %s directory in the archive", r.Config.Name),
			Cause:    err,
			Context: map[string]interface{}{
				"directory_path": dirPath,
				"directory_name": r.Config.Name,
			},
		}
	}

	for i, file := range dirEntries {
		if file.IsDir() {
			context := map[string]interface{}{
				"directory_name": r.Config.Name,
				"subdirectory":   file.Name(),
			}
			return &ValidationError{
				RuleName: r.Name(),
				Message:  fmt.Sprintf("%s directory should not contain subdirectories, found: %s", r.Config.Name, file.Name()),
				Cause:    nil,
				Context:  context,
			}
		}

		if err := r.validateFileExtension(file.Name()); err != nil {
			return err
		}

		if r.Config.RequireSequential {
			if err := r.validateSequentialNaming(file.Name(), i+1); err != nil {
				return err
			}
		}

		if err := r.validateFileNotEmpty(dirPath, file.Name()); err != nil {
			return err
		}
	}

	return nil
}

func (r *DirectoryFilesRule) validateFileExtension(fileName string) error {
	ext := filepath.Ext(fileName)
	if !slices.Contains(r.Config.AcceptedExtensions, ext) {
		context := map[string]interface{}{
			"directory_name":      r.Config.Name,
			"file_name":           fileName,
			"file_extension":      ext,
			"accepted_extensions": r.Config.AcceptedExtensions,
		}
		return &ValidationError{
			RuleName: r.Name(),
			Message: fmt.Sprintf("%s files should have one of these extensions %v, found: %s",
				r.Config.Name, r.Config.AcceptedExtensions, fileName),
			Cause:   nil,
			Context: context,
		}
	}
	return nil
}

func (r *DirectoryFilesRule) validateSequentialNaming(fileName string, expectedNumber int) error {
	ext := filepath.Ext(fileName)
	base := fileName[:len(fileName)-len(ext)]
	if base != strconv.Itoa(expectedNumber) {
		context := map[string]interface{}{
			"directory_name":   r.Config.Name,
			"file_name":        fileName,
			"expected_number":  expectedNumber,
			"actual_base_name": base,
			"expected_pattern": fmt.Sprintf("%d%s", expectedNumber, r.Config.AcceptedExtensions[0]),
		}
		return &ValidationError{
			RuleName: r.Name(),
			Message: fmt.Sprintf("%s files should be named as 1%s, 2%s, etc., found: %s",
				r.Config.Name, r.Config.AcceptedExtensions[0], r.Config.AcceptedExtensions[0], fileName),
			Cause:   nil,
			Context: context,
		}
	}
	return nil
}

func (r *DirectoryFilesRule) validateFileNotEmpty(dirPath, fileName string) error {
	filePath := filepath.Join(dirPath, fileName)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		logger.Error("Failed to get file info for file in the archive.", "directory", r.Config.Name, "file", fileName, "error", err)
		return &ValidationError{
			RuleName: r.Name(),
			Message:  fmt.Sprintf("failed to get file info for %s file in the archive", r.Config.Name),
			Cause:    err,
			Context: map[string]interface{}{
				"file_path":      filePath,
				"directory_name": r.Config.Name,
				"file_name":      fileName,
			},
		}
	}
	if fileInfo.Size() == 0 {
		context := map[string]interface{}{
			"directory_name": r.Config.Name,
			"file_name":      fileName,
			"file_size":      fileInfo.Size(),
		}
		return &ValidationError{
			RuleName: r.Name(),
			Message:  fmt.Sprintf("%s files should not be empty, found: %s. size: %d", r.Config.Name, fileName, fileInfo.Size()),
			Cause:    nil,
			Context:  context,
		}
	}
	return nil
}

// InputOutputMatchRule validates that input and output directories have matching test counts
type InputOutputMatchRule struct{}

func (r *InputOutputMatchRule) Name() string {
	return "input-output-match"
}

func (r *InputOutputMatchRule) Validate(ctx ValidationContext) error {
	inputPath := filepath.Join(ctx.FolderPath, "input")
	outputPath := filepath.Join(ctx.FolderPath, "output")

	inputEntries, err := os.ReadDir(inputPath)
	if err != nil {
		logger.Error("Failed to read input directory in the archive.", "error", err)
		return &ValidationError{
			RuleName: r.Name(),
			Message:  "failed to read input directory in the archive",
			Cause:    err,
			Context: map[string]interface{}{
				"input_path": inputPath,
			},
		}
	}

	outputEntries, err := os.ReadDir(outputPath)
	if err != nil {
		logger.Error("Failed to read output directory in the archive.", "error", err)
		return &ValidationError{
			RuleName: r.Name(),
			Message:  "failed to read output directory in the archive",
			Cause:    err,
			Context: map[string]interface{}{
				"output_path": outputPath,
			},
		}
	}

	if len(inputEntries) != len(outputEntries) || len(inputEntries) == 0 {
		context := map[string]interface{}{
			"input_count":  len(inputEntries),
			"output_count": len(outputEntries),
		}
		return &ValidationError{
			RuleName: r.Name(),
			Message:  "input/ and output/ directories should have the same non-zero number of test files",
			Cause:    nil,
			Context:  context,
		}
	}

	return nil
}
