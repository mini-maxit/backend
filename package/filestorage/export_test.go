package filestorage

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
