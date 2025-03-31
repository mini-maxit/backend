package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// DecompressArchive decompresses archive (either .zip or .tar.gzip) to the given newPath.
func DecompressArchive(archive *os.File, newPath string) error {
	_, err := archive.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to seek archive: %w", err)
	}
	name := archive.Name()
	switch {
	case strings.HasSuffix(name, ".gz"):
		return DecompressGzip(archive, newPath)
	case strings.HasSuffix(name, ".zip"):
		return DecompressZip(archive, newPath)
	}

	return fmt.Errorf("unsupported archive type: %s", name)
}

const maxCopyBytes = 10 << 20

// DecompressGzip decompresses a Gzip archive from passed as [file] to a new directory in the newPath.
func DecompressGzip(file *os.File, newPath string) error {
	uncompressedStream, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer uncompressedStream.Close()

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// TODO: fix path traversal vulnerability
			dirPath := path.Join(newPath, header.Name)
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				return err
			}

		case tar.TypeReg:
			// TODO: fix path traversal vulnerability
			filePath := path.Join(newPath, header.Name)
			if err := os.MkdirAll(path.Dir(filePath), 0755); err != nil {
				return err
			}

			outFile, err := os.Create(filePath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			if _, err := io.CopyN(outFile, tarReader, maxCopyBytes); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unsupported file type %s", string(header.Typeflag))
		}
	}
	return nil
}

// DecompressZip decompresses a zip archive from [file] to a new directory in the newPath.
func DecompressZip(file *os.File, newPath string) error {
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stat: %w", err)
	}
	r, err := zip.NewReader(file, stat.Size())
	if err != nil {
		return err
	}

	for _, f := range r.File {
		// TODO: fix path traversal vulnerability
		filePath := filepath.Join(newPath, f.Name)

		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}

		if !f.FileInfo().IsDir() {
			inFile, err := f.Open()
			if err != nil {
				return err
			}
			defer inFile.Close()

			outFile, err := os.Create(filePath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.CopyN(outFile, inFile, maxCopyBytes)
			if err != nil && errors.Is(err, io.EOF) {
				return err
			}
		}
	}
	return nil
}
