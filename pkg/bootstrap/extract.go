package bootstrap

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractTarGz extracts a .tar.gz archive to destDir
func ExtractTarGz(archivePath, destDir string) error {
	fmt.Printf("   ▸ Extracting Node.js runtime...\n")

	// Open archive file
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Create tar reader
	tr := tar.NewReader(gzr)

	// Extract all files
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Construct target path
		target := filepath.Join(destDir, header.Name)

		// Prevent path traversal attacks
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", header.Name)
		}

		// Handle different file types
		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", target, err)
			}

		case tar.TypeReg:
			// Create regular file
			if err := extractFile(tr, target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to extract file %s: %w", target, err)
			}

		case tar.TypeSymlink:
			// Create symbolic link
			if err := os.Symlink(header.Linkname, target); err != nil {
				// Ignore symlink errors on Windows
				if !os.IsExist(err) {
					fmt.Printf("   ⚠ Warning: failed to create symlink %s: %v\n", target, err)
				}
			}

		default:
			// Skip other types (devices, fifos, etc.)
			fmt.Printf("   ⚠ Warning: skipping unsupported file type: %c in %s\n", header.Typeflag, header.Name)
		}
	}

	fmt.Printf("     ✓ Complete\n")
	return nil
}

// extractFile extracts a single file from tar reader
func extractFile(tr *tar.Reader, target string, mode os.FileMode) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Create file
	f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	// Copy content
	if _, err := io.Copy(f, tr); err != nil {
		return fmt.Errorf("failed to copy content: %w", err)
	}

	return nil
}

// ExtractZip extracts a .zip archive to destDir (for Windows)
func ExtractZip(archivePath, destDir string) error {
	fmt.Printf("   ▸ Extracting Node.js runtime...\n")

	// Open zip file
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	// Extract all files
	for _, f := range r.File {
		// Construct target path
		target := filepath.Join(destDir, f.Name)

		// Prevent path traversal attacks
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			// Create directory
			if err := os.MkdirAll(target, f.Mode()); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", target, err)
			}
		} else {
			// Extract file
			if err := extractZipFile(f, target); err != nil {
				return fmt.Errorf("failed to extract file %s: %w", target, err)
			}
		}
	}

	fmt.Printf("     ✓ Complete\n")
	return nil
}

// extractZipFile extracts a single file from zip
func extractZipFile(f *zip.File, target string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Open source file
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("failed to open zip entry: %w", err)
	}
	defer rc.Close()

	// Create target file
	outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	// Copy content
	if _, err := io.Copy(outFile, rc); err != nil {
		return fmt.Errorf("failed to copy content: %w", err)
	}

	return nil
}

// CleanupArchive removes the downloaded archive file
func CleanupArchive(archivePath string) error {
	return os.Remove(archivePath)
}
