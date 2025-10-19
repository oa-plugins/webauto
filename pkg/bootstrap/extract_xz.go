package bootstrap

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExtractTarXz extracts a .tar.xz archive to destDir (for Linux)
// Uses system 'xz' command for decompression
func ExtractTarXz(archivePath, destDir string) error {
	fmt.Printf("   ▸ Extracting Node.js runtime...\n")

	// Check if xz is available
	if _, err := exec.LookPath("xz"); err != nil {
		return fmt.Errorf("xz command not found (required for .tar.xz extraction): %w\nPlease install: apt-get install xz-utils (Ubuntu/Debian) or yum install xz (RHEL/CentOS)", err)
	}

	// Decompress .xz to .tar using xz command
	xzCmd := exec.Command("xz", "-d", "-c", archivePath)
	xzOutput, err := xzCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create xz stdout pipe: %w", err)
	}

	if err := xzCmd.Start(); err != nil {
		return fmt.Errorf("failed to start xz decompression: %w", err)
	}

	// Create tar reader from xz output
	tr := tar.NewReader(xzOutput)

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
				if !os.IsExist(err) {
					fmt.Printf("   ⚠ Warning: failed to create symlink %s: %v\n", target, err)
				}
			}

		default:
			// Skip other types
			fmt.Printf("   ⚠ Warning: skipping unsupported file type: %c in %s\n", header.Typeflag, header.Name)
		}
	}

	// Wait for xz command to complete
	if err := xzCmd.Wait(); err != nil {
		return fmt.Errorf("xz decompression failed: %w", err)
	}

	fmt.Printf("     ✓ Complete\n")
	return nil
}
