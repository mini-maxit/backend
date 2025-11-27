package filestorage_test

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/mini-maxit/backend/package/filestorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecompressor_DecompressArchive(t *testing.T) {
	d := filestorage.NewDecompressor()

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

		var decompressionErr *filestorage.DecompressionError
		require.ErrorAs(t, err, &decompressionErr)
		assert.Contains(t, decompressionErr.Message, "unsupported archive type")
	})

	t.Run("Fail on nonexistent zip archive", func(t *testing.T) {
		_, err := d.DecompressArchive("/nonexistent/path.zip", "test-*")
		require.Error(t, err)

		var decompressionErr *filestorage.DecompressionError
		require.ErrorAs(t, err, &decompressionErr)
	})

	t.Run("Fail on nonexistent tar.gz archive", func(t *testing.T) {
		_, err := d.DecompressArchive("/nonexistent/path.tar.gz", "test-*")
		require.Error(t, err)

		var decompressionErr *filestorage.DecompressionError
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

		var decompressionErr *filestorage.DecompressionError
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

		var decompressionErr *filestorage.DecompressionError
		require.ErrorAs(t, err, &decompressionErr)
	})
}

func TestDecompressor_decompressZipWithDirs(t *testing.T) {
	d := filestorage.NewDecompressor()

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
	d := filestorage.NewDecompressor()

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

		err := filestorage.RemoveDirectory(testDir)
		require.NoError(t, err)
		assert.NoDirExists(t, testDir)
	})

	t.Run("Remove nonexistent directory returns no error", func(t *testing.T) {
		err := filestorage.RemoveDirectory("/nonexistent/path")
		require.NoError(t, err)
	})

	t.Run("Remove empty path returns error", func(t *testing.T) {
		err := filestorage.RemoveDirectory("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty path")
	})
}

func TestFileStorageServiceGetTestResultPaths(t *testing.T) {
	// Create a testable service for testing the path generation methods
	service := filestorage.NewTestableFileStorageService(nil, nil, "maxit")

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
	service := filestorage.NewTestableFileStorageService(nil, nil, "maxit")

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

func TestFileStorageServiceValidateArchiveStructure(t *testing.T) {
	t.Run("Empty archive path returns error", func(t *testing.T) {
		service := filestorage.NewTestableFileStorageService(nil, nil, "maxit")

		err := service.ValidateArchiveStructure("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
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
