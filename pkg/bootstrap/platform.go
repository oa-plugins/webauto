package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	// Node.js version to download
	NodeVersion = "v22.11.0"

	// Base URL for Node.js downloads
	NodeBaseURL = "https://nodejs.org/dist"
)

// PlatformInfo holds platform-specific information
type PlatformInfo struct {
	OS           string // darwin, windows, linux
	Arch         string // arm64, x64
	DownloadURL  string
	ArchiveExt   string // .tar.gz, .zip
	NodeDirName  string // node-v22.11.0-darwin-arm64
	BinaryPath   string // bin/node or node.exe
}

// GetPlatformInfo returns platform information for current system
func GetPlatformInfo() (*PlatformInfo, error) {
	info := &PlatformInfo{
		OS: runtime.GOOS,
	}

	// Determine architecture
	switch runtime.GOARCH {
	case "amd64":
		info.Arch = "x64"
	case "arm64":
		info.Arch = "arm64"
	default:
		return nil, fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	// Platform-specific configuration
	switch info.OS {
	case "darwin":
		info.NodeDirName = fmt.Sprintf("node-%s-darwin-%s", NodeVersion, info.Arch)
		info.ArchiveExt = ".tar.gz"
		info.BinaryPath = "bin/node"
		info.DownloadURL = fmt.Sprintf("%s/%s/%s.tar.gz", NodeBaseURL, NodeVersion, info.NodeDirName)
	case "windows":
		info.NodeDirName = fmt.Sprintf("node-%s-win-%s", NodeVersion, info.Arch)
		info.ArchiveExt = ".zip"
		info.BinaryPath = "node.exe"
		info.DownloadURL = fmt.Sprintf("%s/%s/%s.zip", NodeBaseURL, NodeVersion, info.NodeDirName)
	case "linux":
		info.NodeDirName = fmt.Sprintf("node-%s-linux-%s", NodeVersion, info.Arch)
		info.ArchiveExt = ".tar.xz"
		info.BinaryPath = "bin/node"
		info.DownloadURL = fmt.Sprintf("%s/%s/%s.tar.xz", NodeBaseURL, NodeVersion, info.NodeDirName)
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", info.OS)
	}

	return info, nil
}

// GetCacheDir returns the cache directory path
func GetCacheDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "oa", "webauto")
	default:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".cache", "oa", "webauto")
	}
}

// GetRuntimeDir returns the runtime directory path (where Node.js is installed)
func GetRuntimeDir() string {
	return filepath.Join(GetCacheDir(), "runtime")
}

// GetNodeBinaryPath returns the full path to node executable
func GetNodeBinaryPath(info *PlatformInfo) string {
	return filepath.Join(GetRuntimeDir(), info.NodeDirName, info.BinaryPath)
}

// GetNodeModulesDir returns the node_modules directory path
func GetNodeModulesDir() string {
	return filepath.Join(GetCacheDir(), "node_modules")
}

// GetBrowsersDir returns the Playwright browsers directory path
func GetBrowsersDir() string {
	return filepath.Join(GetCacheDir(), "browsers")
}
