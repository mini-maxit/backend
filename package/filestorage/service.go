package filestorage

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mini-maxit/backend/package/utils"
	"github.com/mini-maxit/file-storage/pkg/filestorage"
	"go.uber.org/zap"
)

const descriptionFilename = "description.pdf"

type UploadedFile struct {
	Path       string `json:"path"`
	Filename   string `json:"filename"`
	Bucket     string `json:"bucket"`
	ServerType string `json:"serverType"`
}

type UploadedTaskFiles struct {
	DescriptionFile UploadedFile   `json:"descriptionFile"`
	InputFiles      []UploadedFile `json:"inputFiles"`
	OutputFiles     []UploadedFile `json:"outputFiles"`
}

type FileStorageService interface {
	// ValidateArchiveStructure checks if the archive structure is valid.
	// For more details, refer to this documentation:
	// https://github.com/mini-maxit/backend/wiki/Task-archive
	ValidateArchiveStructure(archivePath string) error
	// UploadTask uploads an archive content to the file storage keeping the original structure.
	UploadTask(taskID int64, archivePath string) (*UploadedTaskFiles, error)

	UploadSolutionFile(taskID, userID int64, newOrder int, filePath string) (*UploadedFile, error)

	//
	GetFileURL(path string) string

	GetTestResultStdoutPath(taskID, userID int64, submissionOrder, testCaseOrder int) *UploadedFile
	GetTestResultStderrPath(taskID, userID int64, submissionOrder, testCaseOrder int) *UploadedFile
	ServerType() string
}

type Decompressor interface {
	DecompressArchive(archivePath string, pattern string) (string, error)
}

type decompressor struct{}

func (d *decompressor) DecompressArchive(archivePath string, pattern string) (string, error) {
	folderPath, err := os.MkdirTemp(os.TempDir(), pattern)
	if err != nil {
		return "", &DecompressionError{
			ArchivePath: archivePath,
			Message:     "failed to create temporary directory",
			Cause:       err,
			Context: map[string]any{
				"pattern": pattern,
			},
		}
	}

	if strings.HasSuffix(archivePath, ".gz") {
		err := d.decompressGzip(archivePath, folderPath)
		if err != nil {
			return "", &DecompressionError{
				ArchivePath: archivePath,
				Message:     "failed to decompress gzip archive",
				Cause:       err,
				Context: map[string]any{
					"destination": folderPath,
				},
			}
		}
	} else if strings.HasSuffix(archivePath, ".zip") {
		err := d.decompressZip(archivePath, folderPath)
		if err != nil {
			return "", &DecompressionError{
				ArchivePath: archivePath,
				Message:     "failed to decompress zip archive",
				Cause:       err,
				Context: map[string]any{
					"destination": folderPath,
				},
			}
		}
	} else {
		return "", &DecompressionError{
			ArchivePath: archivePath,
			Message:     fmt.Sprintf("unsupported archive type: %s", archivePath),
			Cause:       nil,
			Context: map[string]any{
				"supported_types": []string{".zip", ".tar.gz"},
			},
		}
	}

	return folderPath, nil
}

// decompressGzip decompresses a Gzip archive from archivePath to a new directory in the newPath
func (d *decompressor) decompressGzip(archivePath string, newPath string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return &DecompressionError{
			ArchivePath: archivePath,
			Message:     "failed to open archive file",
			Cause:       err,
			Context: map[string]any{
				"destination": newPath,
			},
		}
	}
	defer closeIO(file)

	uncompressedStream, err := gzip.NewReader(file)
	if err != nil {
		return &DecompressionError{
			ArchivePath: archivePath,
			Message:     "failed to create gzip reader",
			Cause:       err,
			Context: map[string]any{
				"destination": newPath,
			},
		}
	}
	defer closeIO(uncompressedStream)

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return &DecompressionError{
				ArchivePath: archivePath,
				Message:     "failed to read tar entry",
				Cause:       err,
				Context: map[string]any{
					"destination": newPath,
				},
			}
		}

		switch header.Typeflag {
		case tar.TypeDir:
			dirPath := path.Join(newPath, header.Name)
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				return &DecompressionError{
					ArchivePath: archivePath,
					Message:     "failed to create directory",
					Cause:       err,
					Context: map[string]any{
						"directory_path": dirPath,
						"header_name":    header.Name,
					},
				}
			}

		case tar.TypeReg:
			filePath := path.Join(newPath, header.Name)
			if err := os.MkdirAll(path.Dir(filePath), 0755); err != nil {
				return &DecompressionError{
					ArchivePath: archivePath,
					Message:     "failed to create parent directory",
					Cause:       err,
					Context: map[string]any{
						"parent_directory": path.Dir(filePath),
						"file_path":        filePath,
					},
				}
			}

			outFile, err := os.Create(filePath)
			if err != nil {
				return &DecompressionError{
					ArchivePath: archivePath,
					Message:     "failed to create file",
					Cause:       err,
					Context: map[string]any{
						"file_path":   filePath,
						"header_name": header.Name,
					},
				}
			}
			defer closeIO(outFile)

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return &DecompressionError{
					ArchivePath: archivePath,
					Message:     "failed to write file content",
					Cause:       err,
					Context: map[string]any{
						"file_path": filePath,
					},
				}
			}

		default:
			return &DecompressionError{
				ArchivePath: archivePath,
				Message:     "unsupported file type in archive",
				Cause:       nil,
				Context: map[string]any{
					"file_type": header.Typeflag,
					"file_name": header.Name,
				},
			}
		}
	}
	return nil
}

// decompressZip decompresses a Zip archive from archivePath to a new directory in the newPath
func (d *decompressor) decompressZip(archivePath string, newPath string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return &DecompressionError{
			ArchivePath: archivePath,
			Message:     "failed to open zip archive",
			Cause:       err,
			Context: map[string]any{
				"destination": newPath,
			},
		}
	}
	defer closeIO(r)

	for _, f := range r.File {
		filePath := filepath.Join(newPath, f.Name)

		if f.FileInfo().IsDir() {
			err := os.MkdirAll(filePath, 0755)
			if err != nil {
				return &DecompressionError{
					ArchivePath: archivePath,
					Message:     "failed to create directory",
					Cause:       err,
					Context: map[string]any{
						"directory_path": filePath,
						"zip_entry":      f.Name,
					},
				}
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				return &DecompressionError{
					ArchivePath: archivePath,
					Message:     "failed to create parent directory",
					Cause:       err,
					Context: map[string]any{
						"parent_directory": filepath.Dir(filePath),
						"file_path":        filePath,
					},
				}
			}

			inFile, err := f.Open()
			if err != nil {
				return &DecompressionError{
					ArchivePath: archivePath,
					Message:     fmt.Sprintf("failed to open file in zip: %s", f.Name),
					Cause:       err,
					Context: map[string]any{
						"zip_entry": f.Name,
					},
				}
			}
			defer closeIO(inFile)

			outFile, err := os.Create(filePath)
			if err != nil {
				return &DecompressionError{
					ArchivePath: archivePath,
					Message:     "failed to create file",
					Cause:       err,
					Context: map[string]any{
						"file_path": filePath,
						"zip_entry": f.Name,
					},
				}
			}
			defer closeIO(outFile)

			if _, err := io.Copy(outFile, inFile); err != nil {
				return &DecompressionError{
					ArchivePath: archivePath,
					Message:     "failed to write file content",
					Cause:       err,
					Context: map[string]any{
						"file_path": filePath,
						"zip_entry": f.Name,
					},
				}
			}
		}
	}
	return nil
}

type fileStorageService struct {
	decompressor Decompressor
	validator    ArchiveValidator
	storage      filestorage.FileStorage
	bucketName   string
	logger       *zap.SugaredLogger
}

func NewFileStorageService(fileStorageURL string) (FileStorageService, error) {
	validator := NewArchiveValidator()

	// Configure validation rules
	validator.AddRule(&NonEmptyArchiveRule{})
	validator.AddRule(&RequiredEntriesRule{
		RequiredEntries: []string{"input", "output", descriptionFilename},
	})
	validator.AddRule(&InputOutputMatchRule{})
	validator.AddRule(&DirectoryFilesRule{
		Config: DirectoryConfig{
			Name:               "input",
			AcceptedExtensions: []string{".txt", ".in"},
			RequireSequential:  true,
		},
	})
	validator.AddRule(&DirectoryFilesRule{
		Config: DirectoryConfig{
			Name:               "output",
			AcceptedExtensions: []string{".txt", ".out"},
			RequireSequential:  true,
		},
	})

	config := filestorage.FileStorageConfig{
		URL:     fileStorageURL,
		Version: "v1",
	}
	storage, err := filestorage.NewFileStorage(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create file storage: %w", err)
	}
	return &fileStorageService{
		decompressor: &decompressor{},
		validator:    validator,
		storage:      storage,
		bucketName:   "maxit",
		logger:       utils.NewNamedLogger("file-storage"),
	}, nil
}

func (f *fileStorageService) ValidateArchiveStructure(archivePath string) error {
	if archivePath == "" {
		return errors.New("archive path cannot be empty")
	}

	folderPath, err := f.decompressor.DecompressArchive(archivePath, "validate-archive-structure")
	if err != nil {
		return err
	}

	defer func() {
		if err := removeDirectory(folderPath); err != nil {
			f.logger.Error("Failed to cleanup temporary directory", "path", folderPath, "error", err)
		}
	}()

	folderPath, err = f.normalizeFolderPath(folderPath)
	if err != nil {
		return fmt.Errorf("failed to normalize folder path: %w", err)
	}

	ctx := ValidationContext{
		ArchivePath: archivePath,
		FolderPath:  folderPath,
	}

	return f.validator.Validate(ctx)
}

func (f *fileStorageService) UploadTask(taskID int64, archivePath string) (*UploadedTaskFiles, error) {
	folderPath, err := f.decompressor.DecompressArchive(archivePath, "upload-task")
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := removeDirectory(folderPath); err != nil {
			f.logger.Error("Failed to cleanup temporary directory", "path", folderPath, "error", err)
		}
	}()

	// Handle single root directory case
	folderPath, err = f.normalizeFolderPath(folderPath)
	if err != nil {
		return nil, err
	}

	if err := f.ensureBucketExists(); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	taskBasePath := fmt.Sprintf("task/%d", taskID)

	// Upload description file
	descriptionFile, err := f.uploadDescriptionFile(folderPath, taskBasePath)
	if err != nil {
		return nil, err
	}

	// Upload input files
	inputFiles, err := f.uploadDirectoryFiles(folderPath, "input", taskBasePath)
	if err != nil {
		return nil, err
	}

	// Upload output files
	outputFiles, err := f.uploadDirectoryFiles(folderPath, "output", taskBasePath)
	if err != nil {
		return nil, err
	}

	return &UploadedTaskFiles{
		DescriptionFile: *descriptionFile,
		InputFiles:      inputFiles,
		OutputFiles:     outputFiles,
	}, nil
}

// normalizeFolderPath handles the case where the archive contains a single root directory
func (f *fileStorageService) normalizeFolderPath(folderPath string) (string, error) {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return "", fmt.Errorf("failed to read archive directory: %w", err)
	}

	if len(entries) == 1 && entries[0].IsDir() {
		return filepath.Join(folderPath, entries[0].Name()), nil
	}

	return folderPath, nil
}

// uploadDescriptionFile uploads the [descriptionFilename] file and returns the uploaded file info
func (f *fileStorageService) uploadDescriptionFile(folderPath, taskBasePath string) (*UploadedFile, error) {
	descriptionPath := filepath.Join(folderPath, descriptionFilename)

	file, err := os.Open(descriptionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open description file: %w", err)
	}
	defer closeIO(file)

	remotePath := filepath.Join(taskBasePath, descriptionFilename)
	if err := f.storage.UploadFile(f.bucketName, remotePath, file); err != nil {
		return nil, fmt.Errorf("failed to upload description file: %w", err)
	}

	return &UploadedFile{
		Path:       remotePath,
		Filename:   descriptionFilename,
		Bucket:     f.bucketName,
		ServerType: f.ServerType(),
	}, nil
}

// uploadDirectoryFiles uploads all files from a directory and returns the uploaded file info
func (f *fileStorageService) uploadDirectoryFiles(basePath, dirName, taskBasePath string) ([]UploadedFile, error) {
	dirPath := filepath.Join(basePath, dirName)

	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil, err // Return empty slice if directory doesn't exist
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s directory: %w", dirName, err)
	}

	var files []*os.File
	var uploadedFiles []UploadedFile
	remotePrefix := filepath.Join(taskBasePath, dirName)

	// Ensure cleanup of all opened files
	defer func() {
		for _, file := range files {
			closeIO(file)
		}
	}()

	// Open all files and prepare upload info
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open %s file %s: %w", dirName, entry.Name(), err)
		}

		files = append(files, file)
		uploadedFiles = append(uploadedFiles, UploadedFile{
			Path:       filepath.Join(remotePrefix, entry.Name()),
			Filename:   entry.Name(),
			Bucket:     f.bucketName,
			ServerType: f.ServerType(),
		})
	}

	// Upload all files
	if len(files) > 0 {
		if err := f.storage.UploadMultipleFiles(f.bucketName, remotePrefix, files); err != nil {
			return nil, fmt.Errorf("failed to upload %s files: %w", dirName, err)
		}
	}

	sort.Slice(uploadedFiles, func(i, j int) bool {
		return uploadedFiles[i].Filename < uploadedFiles[j].Filename
	})

	return uploadedFiles, nil
}

func (f *fileStorageService) ensureBucketExists() error {
	f.logger.Info("Ensuring bucket exists", "bucket_name", f.bucketName)
	err := f.storage.CreateBucket(f.bucketName)
	if err != nil {
		var apiErr *filestorage.ErrAPI
		if errors.As(err, &apiErr) {
			if apiErr.StatusCode != http.StatusConflict {
				return fmt.Errorf("failed to create bucket: %w", apiErr)
			}
		} else {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}
	return nil
}

func (f *fileStorageService) GetFileURL(path string) string {
	return f.storage.GetFileURL(f.bucketName, path)
}

func (f *fileStorageService) UploadSolutionFile(taskID, userID int64, order int, filePath string) (*UploadedFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer closeIO(file)
	fileExtension := filepath.Ext(filePath)

	remotePrefix := fmt.Sprintf("solution/%d/%d/%d", taskID, userID, order)
	remoteFilename := fmt.Sprintf("solution%s", fileExtension)
	uploadedFile := &UploadedFile{
		Path:     filepath.Join(remotePrefix, remoteFilename),
		Filename: remoteFilename,
	}

	if err := f.ensureBucketExists(); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	if err := f.storage.UploadFile(f.bucketName, uploadedFile.Path, file); err != nil {
		return nil, fmt.Errorf("failed to upload solution file: %w", err)
	}

	return uploadedFile, nil
}

func (f *fileStorageService) GetTestResultStdoutPath(taskID, userID int64, submissionOrder, testCaseOrder int) *UploadedFile {
	stdoutPath := fmt.Sprintf("solution/%d/%d/%d/stdout/%d.out", taskID, userID, submissionOrder, testCaseOrder)
	return &UploadedFile{
		Filename:   path.Base(stdoutPath),
		Path:       stdoutPath,
		Bucket:     f.bucketName,
		ServerType: f.ServerType(),
	}
}

func (f *fileStorageService) GetTestResultStderrPath(taskID, userID int64, submissionOrder, testCaseOrder int) *UploadedFile {
	stderrPath := fmt.Sprintf("solution/%d/%d/%d/stderr/%d.err", taskID, userID, submissionOrder, testCaseOrder)
	return &UploadedFile{
		Filename:   path.Base(stderrPath),
		Path:       stderrPath,
		Bucket:     f.bucketName,
		ServerType: f.ServerType(),
	}
}

func (f *fileStorageService) ServerType() string {
	return "filestorage"
}

func removeDirectory(path string) error {
	if path == "" {
		return errors.New("tried to remove an empty path")
	}
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return nil
}

// closeIO tries to close any io.Closer and logs an error if one occurs.
func closeIO(c io.Closer) {
	logger := utils.NewNamedLogger("closeIO")
	if err := c.Close(); err != nil {
		logger.Error("Error closing file.", "error", err)
	}
}
