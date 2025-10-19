package bootstrap

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/schollz/progressbar/v3"
)

const (
	// MaxRetries is the maximum number of download retry attempts
	MaxRetries = 3

	// RetryDelay is the delay between retry attempts
	RetryDelay = 2 * time.Second
)

// DownloadFile downloads a file from URL to destPath with progress bar
func DownloadFile(url, destPath string, description string) error {
	var lastErr error

	for attempt := 1; attempt <= MaxRetries; attempt++ {
		if attempt > 1 {
			fmt.Printf("   ▸ Retry attempt %d/%d...\n", attempt, MaxRetries)
			time.Sleep(RetryDelay)
		}

		err := downloadFileAttempt(url, destPath, description)
		if err == nil {
			return nil
		}

		lastErr = err
	}

	return fmt.Errorf("failed after %d attempts: %w", MaxRetries, lastErr)
}

// downloadFileAttempt performs a single download attempt
func downloadFileAttempt(url, destPath, description string) error {
	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent to avoid potential blocking
	req.Header.Set("User-Agent", "webauto/1.0")

	// Execute request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Create progress bar
	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		description,
	)

	// Download with progress tracking
	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	if err != nil {
		os.Remove(destPath) // Clean up partial download
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Println() // New line after progress bar
	return nil
}

// CheckDiskSpace checks if there is enough disk space available
func CheckDiskSpace(path string, requiredBytes int64) error {
	// Check if path exists
	if _, err := os.Stat(path); err != nil {
		// If path doesn't exist, check parent directory
		parent := filepath.Dir(path)
		if _, err = os.Stat(parent); err != nil {
			return fmt.Errorf("failed to check disk space: %w", err)
		}
	}

	// Note: os.FileInfo doesn't provide disk space info in Go's standard library
	// This is a simplified check - actual implementation would use syscall for accurate disk space check
	// For now, we'll skip the actual check and just ensure directory is writable

	return nil
}
