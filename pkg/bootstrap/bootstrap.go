package bootstrap

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// EnsureRuntime checks and installs Node.js runtime if needed
// Returns the path to node executable
func EnsureRuntime() (string, error) {
	// Get platform information
	platform, err := GetPlatformInfo()
	if err != nil {
		return "", fmt.Errorf("failed to detect platform: %w", err)
	}

	// Get expected node binary path
	nodePath := GetNodeBinaryPath(platform)

	// Check if already installed
	if _, err := os.Stat(nodePath); err == nil {
		// Node.js already installed, verify it works silently
		cmd := exec.Command(nodePath, "--version")
		if err := cmd.Run(); err == nil {
			// Verification passed, return immediately
			return nodePath, nil
		}
		// If verification fails, continue with reinstallation
		fmt.Printf("⚠ Existing Node.js installation appears corrupted, reinstalling...\n")
	}

	// Not installed, perform first-time setup
	fmt.Printf("\n📦 Setting up webauto runtime (one-time setup)...\n")
	fmt.Printf("   ▸ Detecting platform: %s %s\n", platform.OS, platform.Arch)

	// Create runtime directory
	runtimeDir := GetRuntimeDir()
	if err := os.MkdirAll(runtimeDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create runtime directory: %w", err)
	}

	// Download Node.js
	archivePath := filepath.Join(runtimeDir, platform.NodeDirName+platform.ArchiveExt)
	fmt.Printf("   ▸ Downloading Node.js %s (~30MB)...\n", NodeVersion)

	if err := DownloadFile(platform.DownloadURL, archivePath, "     "); err != nil {
		return "", fmt.Errorf("failed to download Node.js: %w\n\nSuggestions:\n   1. Check your internet connection\n   2. Retry: oa webauto browser-launch\n   3. Manual setup: export PLAYWRIGHT_NODE_PATH=/path/to/node", err)
	}

	// Extract archive
	if err := extractArchive(archivePath, runtimeDir, platform.ArchiveExt); err != nil {
		return "", fmt.Errorf("failed to extract Node.js: %w", err)
	}

	// Clean up archive
	if err := CleanupArchive(archivePath); err != nil {
		fmt.Printf("   ⚠ Warning: failed to cleanup archive: %v\n", err)
	}

	// Verify Node.js installation
	if err := VerifyNodeInstallation(nodePath); err != nil {
		return "", fmt.Errorf("Node.js installation verification failed: %w", err)
	}

	// Install Playwright
	if err := InstallPlaywright(nodePath); err != nil {
		return "", fmt.Errorf("failed to install Playwright: %w", err)
	}

	// Install Playwright browsers
	if err := InstallPlaywrightBrowsers(nodePath); err != nil {
		return "", fmt.Errorf("failed to install Playwright browsers: %w", err)
	}

	fmt.Printf("✓ Setup complete!\n\n")
	return nodePath, nil
}

// extractArchive extracts archive based on file extension
func extractArchive(archivePath, destDir, ext string) error {
	switch ext {
	case ".tar.gz":
		return ExtractTarGz(archivePath, destDir)
	case ".zip":
		return ExtractZip(archivePath, destDir)
	case ".tar.xz":
		return fmt.Errorf("tar.xz extraction not yet implemented (Linux support coming soon)")
	default:
		return fmt.Errorf("unsupported archive format: %s", ext)
	}
}

// IsRuntimeInstalled checks if Node.js runtime is already installed
func IsRuntimeInstalled() bool {
	platform, err := GetPlatformInfo()
	if err != nil {
		return false
	}

	nodePath := GetNodeBinaryPath(platform)
	_, err = os.Stat(nodePath)
	return err == nil
}

// GetInstalledNodePath returns the path to installed Node.js or empty string
func GetInstalledNodePath() string {
	platform, err := GetPlatformInfo()
	if err != nil {
		return ""
	}

	nodePath := GetNodeBinaryPath(platform)
	if _, err := os.Stat(nodePath); err == nil {
		return nodePath
	}

	return ""
}
