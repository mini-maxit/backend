//nolint:testpackage // to access unexported identifiers for testing
package filestorage

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mini-maxit/backend/package/utils"
	"github.com/mini-maxit/file-storage/pkg/filestorage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecompressor_DecompressArchive(t *testing.T) {
	d := NewDecompressor()

	t.Run("Decompress zip archive successfully", func(t *testing.T) {
		zipPath := createTestZipArchive(t)

		folderPath, err := d.DecompressArchive(zipPath, "test-zip-*")
		require.NoError(t, err)
		defer os.RemoveAll(folderPath)

		assert.DirExists(t, folderPath)

		// Check that files were extracted
		files, err := os.ReadDir(folderPath)
		require.NoError(t, err)
		assert.NotEmpty(t, files)
	})

	t.Run("Decompress tar.gz archive successfully", func(t *testing.T) {
		targzPath := createTestTarGzArchive(t)

		folderPath, err := d.DecompressArchive(targzPath, "test-targz-*")
		require.NoError(t, err)
		defer os.RemoveAll(folderPath)

		assert.DirExists(t, folderPath)

		// Check that files were extracted
		files, err := os.ReadDir(folderPath)
		require.NoError(t, err)
		assert.NotEmpty(t, files)
	})

	t.Run("Fail on unsupported archive type", func(t *testing.T) {
		// Create a temporary file with unsupported extension
		tempFile, err := os.CreateTemp(t.TempDir(), "test-*.txt")
		require.NoError(t, err)
		tempFile.Close()

		_, err = d.DecompressArchive(tempFile.Name(), "test-*")
		require.Error(t, err)

		var decompressionErr *DecompressionError
		require.ErrorAs(t, err, &decompressionErr)
		assert.Contains(t, decompressionErr.Message, "unsupported archive type")
	})

	t.Run("Fail on nonexistent zip archive", func(t *testing.T) {
		_, err := d.DecompressArchive("/nonexistent/path.zip", "test-*")
		require.Error(t, err)

		var decompressionErr *DecompressionError
		require.ErrorAs(t, err, &decompressionErr)
	})

	t.Run("Fail on nonexistent tar.gz archive", func(t *testing.T) {
		_, err := d.DecompressArchive("/nonexistent/path.tar.gz", "test-*")
		require.Error(t, err)

		var decompressionErr *DecompressionError
		require.ErrorAs(t, err, &decompressionErr)
	})

	t.Run("Fail on corrupted zip archive", func(t *testing.T) {
		// Create a file with .zip extension but invalid content
		tempFile, err := os.CreateTemp(t.TempDir(), "corrupted-*.zip")
		require.NoError(t, err)
		_, err = tempFile.WriteString("not a valid zip file")
		require.NoError(t, err)
		tempFile.Close()

		_, err = d.DecompressArchive(tempFile.Name(), "test-*")
		require.Error(t, err)

		var decompressionErr *DecompressionError
		require.ErrorAs(t, err, &decompressionErr)
	})

	t.Run("Fail on corrupted tar.gz archive", func(t *testing.T) {
		// Create a file with .gz extension but invalid content
		tempFile, err := os.CreateTemp(t.TempDir(), "corrupted-*.gz")
		require.NoError(t, err)
		_, err = tempFile.WriteString("not a valid gzip file")
		require.NoError(t, err)
		tempFile.Close()

		_, err = d.DecompressArchive(tempFile.Name(), "test-*")
		require.Error(t, err)

		var decompressionErr *DecompressionError
		require.ErrorAs(t, err, &decompressionErr)
	})
}

func TestDecompressor_decompressZipWithDirs(t *testing.T) {
	d := NewDecompressor()

	t.Run("Decompress zip with directories", func(t *testing.T) {
		zipPath := createTestZipArchiveWithDirs(t)

		folderPath, err := d.DecompressArchive(zipPath, "test-zip-dirs-*")
		require.NoError(t, err)
		defer os.RemoveAll(folderPath)

		// Check directory structure
		assert.DirExists(t, filepath.Join(folderPath, "subdir"))
		assert.FileExists(t, filepath.Join(folderPath, "subdir", "file.txt"))
	})
}

func TestDecompressor_decompressTarGzWithDirs(t *testing.T) {
	d := NewDecompressor()

	t.Run("Decompress tar.gz with directories", func(t *testing.T) {
		targzPath := createTestTarGzArchiveWithDirs(t)

		folderPath, err := d.DecompressArchive(targzPath, "test-targz-dirs-*")
		require.NoError(t, err)
		defer os.RemoveAll(folderPath)

		// Check directory structure
		assert.DirExists(t, filepath.Join(folderPath, "subdir"))
		assert.FileExists(t, filepath.Join(folderPath, "subdir", "file.txt"))
	})
}

func TestRemoveDirectory(t *testing.T) {
	t.Run("Remove existing directory", func(t *testing.T) {
		tempDir := t.TempDir()
		testDir := filepath.Join(tempDir, "toremove")
		require.NoError(t, os.MkdirAll(testDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(testDir, "file.txt"), []byte("test"), 0644))

		err := RemoveDirectory(testDir)
		require.NoError(t, err)
		assert.NoDirExists(t, testDir)
	})

	t.Run("Remove nonexistent directory returns no error", func(t *testing.T) {
		err := RemoveDirectory("/nonexistent/path")
		require.NoError(t, err)
	})

	t.Run("Remove empty path returns error", func(t *testing.T) {
		err := RemoveDirectory("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty path")
	})
}

func TestFileStorageServiceGetTestResultPaths(t *testing.T) {
	// Create a testable service for testing the path generation methods
	service := NewTestableFileStorageService(nil, nil, "maxit")

	t.Run("GetTestResultStdoutPath returns correct path", func(t *testing.T) {
		result := service.GetTestResultStdoutPath(1, 2, 3, 4)

		assert.Equal(t, "solution/1/2/3/stdout/4.out", result.Path)
		assert.Equal(t, "4.out", result.Filename)
		assert.Equal(t, "maxit", result.Bucket)
		assert.Equal(t, "filestorage", result.ServerType)
	})

	t.Run("GetTestResultStderrPath returns correct path", func(t *testing.T) {
		result := service.GetTestResultStderrPath(1, 2, 3, 4)

		assert.Equal(t, "solution/1/2/3/stderr/4.err", result.Path)
		assert.Equal(t, "4.err", result.Filename)
		assert.Equal(t, "maxit", result.Bucket)
		assert.Equal(t, "filestorage", result.ServerType)
	})

	t.Run("GetTestResultDiffPath returns correct path", func(t *testing.T) {
		result := service.GetTestResultDiffPath(1, 2, 3, 4)

		assert.Equal(t, "solution/1/2/3/diff/4.diff", result.Path)
		assert.Equal(t, "4.diff", result.Filename)
		assert.Equal(t, "maxit", result.Bucket)
		assert.Equal(t, "filestorage", result.ServerType)
	})

	t.Run("ServerType returns filestorage", func(t *testing.T) {
		assert.Equal(t, "filestorage", service.ServerType())
	})
}

func TestFileStorageServiceNormalizeFolderPath(t *testing.T) {
	service := NewTestableFileStorageService(nil, nil, "maxit")

	t.Run("Single root directory", func(t *testing.T) {
		tempDir := t.TempDir()
		rootDir := filepath.Join(tempDir, "root")
		require.NoError(t, os.MkdirAll(rootDir, 0755))

		result, err := service.NormalizeFolderPath(tempDir)
		require.NoError(t, err)
		assert.Equal(t, rootDir, result)
	})

	t.Run("Multiple entries", func(t *testing.T) {
		tempDir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "dir1"), 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "dir2"), 0755))

		result, err := service.NormalizeFolderPath(tempDir)
		require.NoError(t, err)
		assert.Equal(t, tempDir, result)
	})

	t.Run("Single file entry", func(t *testing.T) {
		tempDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "file.txt"), []byte("test"), 0644))

		result, err := service.NormalizeFolderPath(tempDir)
		require.NoError(t, err)
		assert.Equal(t, tempDir, result)
	})

	t.Run("Empty directory", func(t *testing.T) {
		tempDir := t.TempDir()

		result, err := service.NormalizeFolderPath(tempDir)
		require.NoError(t, err)
		assert.Equal(t, tempDir, result)
	})

	t.Run("Nonexistent directory", func(t *testing.T) {
		_, err := service.NormalizeFolderPath("/nonexistent/path")
		require.Error(t, err)
	})
}

type capturingValidator struct {
	ArchiveValidator
	called bool
	gotCtx ValidationContext
}

// Implement ArchiveValidator
func (c *capturingValidator) Validate(ctx ValidationContext) error {
	c.called = true
	c.gotCtx = ctx
	return nil
}
func (c *capturingValidator) AddRule(rule ValidationRule) {}

type failingValidator struct{}

// Implement ArchiveValidator
func (f *failingValidator) Validate(ctx ValidationContext) error {
	return errors.New("validator failed")
}
func (f *failingValidator) AddRule(rule ValidationRule) {}
func TestFileStorageServiceValidateArchiveStructure(t *testing.T) {
	t.Run("Empty archive path returns error", func(t *testing.T) {
		service := NewTestableFileStorageService(nil, nil, "maxit")

		err := service.ValidateArchiveStructure("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
	})

	t.Run("Decompressor error is returned", func(t *testing.T) {
		service := NewTestableFileStorageService(nil, nil, "maxit")

		fd := &fakeDecompressor{folderPath: "", err: errors.New("decompress error")}
		service.SetDecompressor(fd)

		err := service.ValidateArchiveStructure("archive.zip")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "decompress")
	})

	t.Run("Validator error is propagated", func(t *testing.T) {
		service := NewTestableFileStorageService(nil, nil, "maxit")

		base := t.TempDir()
		fd := &fakeDecompressor{folderPath: base, err: nil}
		service.SetDecompressor(fd)

		service.SetValidator(&failingValidator{})

		err := service.ValidateArchiveStructure("ok.zip")
		require.Error(t, err)
	})

	t.Run("Decompressor error is returned", func(t *testing.T) {
		service := NewTestableFileStorageService(nil, nil, "maxit")
		// Use fake decompressor that returns an error
		fd := &fakeDecompressor{folderPath: "", err: errors.New("decompress error")}
		service.SetDecompressor(fd)

		err := service.ValidateArchiveStructure("archive.zip")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "decompress")
	})

	t.Run("Validator is invoked with proper context", func(t *testing.T) {
		service := NewTestableFileStorageService(nil, nil, "maxit")

		// Prepare a temp directory to act as decompressed content
		base := t.TempDir()
		// Create single root folder so normalizeFolderPath picks it
		root := filepath.Join(base, "root")
		require.NoError(t, os.MkdirAll(root, 0755))

		// Decompressor returns base directory
		fd := &fakeDecompressor{folderPath: base, err: nil}
		service.SetDecompressor(fd)

		// Create a real validator to ensure Validate is invoked without using mocks
		validator := &capturingValidator{}

		service.SetValidator(validator)

		expectedArchivePath := "test.zip"
		err := service.ValidateArchiveStructure(expectedArchivePath)
		require.NoError(t, err)
		assert.True(t, validator.called, "validator should be called")
		assert.Equal(t, expectedArchivePath, validator.gotCtx.ArchivePath)
	})

	t.Run("Validator error is propagated", func(t *testing.T) {
		service := NewTestableFileStorageService(nil, nil, "maxit")

		// Decompressor returns a valid folder
		base := t.TempDir()
		fd := &fakeDecompressor{folderPath: base, err: nil}
		service.SetDecompressor(fd)

		// Create a validator that will fail using existing rules by providing a folder with no required content
		validator := NewArchiveValidator()
		// Add a rule that requires the archive not to be empty
		validator.AddRule(&NonEmptyArchiveRule{})
		// Also add a rule that requires specific entries; with empty base, it'll fail
		validator.AddRule(&RequiredEntriesRule{
			RequiredEntries: []string{"description.pdf", "input/", "output/"},
		})

		service.SetValidator(validator)

		err := service.ValidateArchiveStructure("ok.zip")
		require.Error(t, err)
	})
}

// Helper functions to create test archives

func createTestZipArchive(t *testing.T) string {
	t.Helper()

	tempFile, err := os.CreateTemp(t.TempDir(), "test-*.zip")
	require.NoError(t, err)
	defer tempFile.Close()

	zipWriter := zip.NewWriter(tempFile)
	defer zipWriter.Close()

	// Add a file
	writer, err := zipWriter.Create("test.txt")
	require.NoError(t, err)
	_, err = writer.Write([]byte("test content"))
	require.NoError(t, err)

	return tempFile.Name()
}

type fakeStorage struct {
	filestorage.FileStorage
	createBucketErr    error
	uploadedFiles      map[string][]string // bucket -> paths
	lastUploadedPrefix string
}

func (f *fakeStorage) CreateBucket(bucket string) error {
	return f.createBucketErr
}

func (f *fakeStorage) UploadFile(bucket, path string, file *os.File) error {
	if f.uploadedFiles == nil {
		f.uploadedFiles = make(map[string][]string)
	}
	f.uploadedFiles[bucket] = append(f.uploadedFiles[bucket], path)
	return nil
}

func (f *fakeStorage) UploadMultipleFiles(bucket, prefix string, files []*os.File) error {
	f.lastUploadedPrefix = prefix
	if f.uploadedFiles == nil {
		f.uploadedFiles = make(map[string][]string)
	}
	for _, file := range files {
		if file == nil {
			continue
		}
		// Use the base filename as path suffix for testing
		name := filepath.Base(file.Name())
		f.uploadedFiles[bucket] = append(f.uploadedFiles[bucket], filepath.Join(prefix, name))
	}
	return nil
}

func (f *fakeStorage) GetFileURL(bucket, path string) string {
	return "http://fake/" + bucket + "/" + path
}

func TestEnsureBucketExists_SuccessAndError(t *testing.T) {
	svc := NewTestableFileStorageService(nil, nil, "maxit")
	// set a logger to avoid nil pointer deref in ensureBucketExists
	svc.logger = utils.NewNamedLogger("file-storage-test")

	t.Run("success when CreateBucket returns nil", func(t *testing.T) {
		fs := &fakeStorage{createBucketErr: nil}
		svc.storage = fs

		err := svc.ensureBucketExists()
		require.NoError(t, err)
	})

	t.Run("error when CreateBucket returns generic error", func(t *testing.T) {
		fs := &fakeStorage{createBucketErr: errors.New("create bucket failed")}
		svc.storage = fs

		err := svc.ensureBucketExists()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create bucket")
	})
}

func TestGetFileURL(t *testing.T) {
	svc := NewTestableFileStorageService(nil, nil, "maxit")
	fs := &fakeStorage{}
	svc.storage = fs

	url := svc.GetFileURL("task/1/description.pdf")
	assert.Equal(t, "http://fake/maxit/task/1/description.pdf", url)
}

func TestUploadDirectoryFiles_AllBranches(t *testing.T) {
	svc := NewTestableFileStorageService(nil, nil, "maxit")
	svc.logger = utils.NewNamedLogger("file-storage-test")
	fs := &fakeStorage{}
	svc.storage = fs

	t.Run("directory missing -> returns os.Stat error", func(t *testing.T) {
		base := t.TempDir()
		uploaded, err := svc.uploadDirectoryFiles(base, "missing", "task/1")
		require.Error(t, err)
		assert.Nil(t, uploaded)
	})

	t.Run("success -> uploads and sorts files", func(t *testing.T) {
		base := t.TempDir()
		dir := filepath.Join(base, "input")
		require.NoError(t, os.MkdirAll(dir, 0755))

		// create files in non-sorted order
		f2, _ := os.Create(filepath.Join(dir, "b.txt"))
		_ = f2.Close()
		f1, _ := os.Create(filepath.Join(dir, "a.txt"))
		_ = f1.Close()
		// create subdir which should be ignored
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "sub"), 0755))

		uploaded, err := svc.uploadDirectoryFiles(base, "input", "task/1")
		require.NoError(t, err)
		require.Len(t, uploaded, 2)
		assert.Equal(t, "a.txt", uploaded[0].Filename)
		assert.Equal(t, "b.txt", uploaded[1].Filename)
		assert.Equal(t, "task/1/input", fs.lastUploadedPrefix)
		assert.ElementsMatch(t, []string{
			"task/1/input/a.txt",
			"task/1/input/b.txt",
		}, fs.uploadedFiles["maxit"])
	})
}

func TestUploadDescriptionFile_ErrorAndSuccess(t *testing.T) {
	svc := NewTestableFileStorageService(nil, nil, "maxit")
	svc.logger = utils.NewNamedLogger("file-storage-test")
	fs := &fakeStorage{}
	svc.storage = fs

	t.Run("missing description.pdf -> open error", func(t *testing.T) {
		base := t.TempDir()
		uploaded, err := svc.uploadDescriptionFile(base, "task/99")
		require.Error(t, err)
		assert.Nil(t, uploaded)
		assert.Contains(t, err.Error(), "failed to open description file")
	})

	t.Run("success -> uploads description.pdf", func(t *testing.T) {
		base := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(base, "description.pdf"), []byte("pdf"), 0644))
		uploaded, err := svc.uploadDescriptionFile(base, "task/99")
		require.NoError(t, err)
		require.NotNil(t, uploaded)
		assert.Equal(t, "task/99/description.pdf", uploaded.Path)
		assert.Equal(t, "description.pdf", uploaded.Filename)
		assert.Equal(t, "maxit", uploaded.Bucket)
	})
}

func TestUploadSolutionFile_Branches(t *testing.T) {
	svc := NewTestableFileStorageService(nil, nil, "maxit")
	svc.logger = utils.NewNamedLogger("file-storage-test")

	t.Run("ensureBucketExists failure wraps", func(t *testing.T) {
		tmp, err := os.CreateTemp(t.TempDir(), "sol-*.cpp")
		require.NoError(t, err)
		_, _ = tmp.WriteString("// code")
		tmp.Close()

		fs := &fakeStorage{createBucketErr: errors.New("create error")}
		svc.storage = fs

		uploaded, err := svc.UploadSolutionFile(1, 2, 3, tmp.Name())
		require.Error(t, err)
		assert.Nil(t, uploaded)
		assert.Contains(t, err.Error(), "failed to ensure bucket exists")
	})

	t.Run("success flow", func(t *testing.T) {
		tmp, err := os.CreateTemp(t.TempDir(), "sol-*.cpp")
		require.NoError(t, err)
		_, _ = tmp.WriteString("// code")
		tmp.Close()

		fs := &fakeStorage{createBucketErr: nil}
		svc.storage = fs

		uploaded, err := svc.UploadSolutionFile(10, 20, 1, tmp.Name())
		require.NoError(t, err)
		require.NotNil(t, uploaded)

		assert.Equal(t, "solution/10/20/1/solution.cpp", uploaded.Path)
		assert.Equal(t, "solution.cpp", uploaded.Filename)
		assert.Equal(t, "maxit", uploaded.Bucket)
		assert.Equal(t, "filestorage", uploaded.ServerType)
	})
}

func TestUploadTask_EndToEnd_WithFakeDecompressor(t *testing.T) {
	svc := NewTestableFileStorageService(nil, nil, "maxit")
	svc.logger = utils.NewNamedLogger("file-storage-test")
	fs := &fakeStorage{}
	svc.storage = fs

	// Build a decompressed folder structure
	base := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(base, "description.pdf"), []byte("pdf"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(base, "input"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(base, "output"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(base, "input", "1.txt"), []byte("in1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(base, "output", "1.txt"), []byte("out1"), 0644))

	// Force decompressor to return our base dir
	svc.decompressor = &fakeDecompressor{folderPath: base, err: nil}

	files, err := svc.UploadTask(123, "/fake/archive.zip")
	require.NoError(t, err)
	require.NotNil(t, files)

	assert.Equal(t, "task/123/description.pdf", files.DescriptionFile.Path)
	require.Len(t, files.InputFiles, 1)
	require.Len(t, files.OutputFiles, 1)
	assert.Equal(t, "task/123/input/1.txt", files.InputFiles[0].Path)
	assert.Equal(t, "task/123/output/1.txt", files.OutputFiles[0].Path)
}

func createTestZipArchiveWithDirs(t *testing.T) string {
	t.Helper()

	tempFile, err := os.CreateTemp(t.TempDir(), "test-*.zip")
	require.NoError(t, err)
	defer tempFile.Close()

	zipWriter := zip.NewWriter(tempFile)
	defer zipWriter.Close()

	// Add a directory entry (directories end with /)
	_, err = zipWriter.Create("subdir/")
	require.NoError(t, err)

	// Add a file in the directory
	writer, err := zipWriter.Create("subdir/file.txt")
	require.NoError(t, err)
	_, err = writer.Write([]byte("test content"))
	require.NoError(t, err)

	return tempFile.Name()
}

func createTestTarGzArchive(t *testing.T) string {
	t.Helper()

	tempFile, err := os.CreateTemp(t.TempDir(), "test-*.tar.gz")
	require.NoError(t, err)
	defer tempFile.Close()

	gzWriter := gzip.NewWriter(tempFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add a file
	content := []byte("test content")
	header := &tar.Header{
		Name:     "test.txt",
		Mode:     0644,
		Size:     int64(len(content)),
		Typeflag: tar.TypeReg,
	}
	require.NoError(t, tarWriter.WriteHeader(header))
	_, err = tarWriter.Write(content)
	require.NoError(t, err)

	return tempFile.Name()
}

func TestDecompressor_decompressTarGz_UnsupportedType(t *testing.T) {
	d := NewDecompressor()

	// Create a tar.gz with a symlink entry to trigger unsupported type in decompressGzip
	tempFile, err := os.CreateTemp(t.TempDir(), "unsupported-*.tar.gz")
	require.NoError(t, err)
	defer tempFile.Close()

	gzWriter := gzip.NewWriter(tempFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add a symlink entry (unsupported by our decompressor)
	hdr := &tar.Header{
		Name:     "link",
		Linkname: "target",
		Mode:     0777,
		Typeflag: tar.TypeSymlink,
	}
	require.NoError(t, tarWriter.WriteHeader(hdr))

	// Attempt to decompress and expect unsupported file type error
	_, err = d.DecompressArchive(tempFile.Name(), "test-unsupported-*")
	require.Error(t, err)

	var decompressionErr *DecompressionError
	require.ErrorAs(t, err, &decompressionErr)
	assert.Contains(t, decompressionErr.Cause.Error(), "failed to read tar entry")
}

func createTestTarGzArchiveWithDirs(t *testing.T) string {
	t.Helper()

	tempFile, err := os.CreateTemp(t.TempDir(), "test-*.tar.gz")
	require.NoError(t, err)
	defer tempFile.Close()

	gzWriter := gzip.NewWriter(tempFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add directory
	dirHeader := &tar.Header{
		Name:     "subdir/",
		Mode:     0755,
		Typeflag: tar.TypeDir,
	}
	require.NoError(t, tarWriter.WriteHeader(dirHeader))

	// Add a file in the directory
	content := []byte("test content")
	fileHeader := &tar.Header{
		Name:     "subdir/file.txt",
		Mode:     0644,
		Size:     int64(len(content)),
		Typeflag: tar.TypeReg,
	}
	require.NoError(t, tarWriter.WriteHeader(fileHeader))
	_, err = tarWriter.Write(content)
	require.NoError(t, err)

	return tempFile.Name()
}

type fakeDecompressor struct {
	folderPath string
	err        error
}

func (d *fakeDecompressor) DecompressArchive(archivePath, pattern string) (string, error) {
	return d.folderPath, d.err
}
