package bootstrap

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// InstallPlaywright installs Playwright via npm
func InstallPlaywright(nodeExePath string) error {
	cacheDir := GetCacheDir()
	nodeModulesDir := GetNodeModulesDir()
	browsersDir := GetBrowsersDir()

	fmt.Printf("   ▸ Installing Playwright library...\n")

	// Ensure directories exist
	if err := os.MkdirAll(nodeModulesDir, 0755); err != nil {
		return fmt.Errorf("failed to create node_modules directory: %w", err)
	}

	// Get npm path (should be in same directory as node)
	npmPath := filepath.Join(filepath.Dir(nodeExePath), "npm")
	if _, err := os.Stat(npmPath); os.IsNotExist(err) {
		// On Windows, npm might be npm.cmd
		npmPath = filepath.Join(filepath.Dir(nodeExePath), "npm.cmd")
		if _, err := os.Stat(npmPath); os.IsNotExist(err) {
			return fmt.Errorf("npm not found in Node.js installation")
		}
	}

	// Run npm install playwright
	cmd := exec.Command(npmPath, "install", "playwright", "@playwright/test")
	cmd.Dir = cacheDir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PLAYWRIGHT_BROWSERS_PATH=%s", browsersDir),
	)

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("npm install failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("     ✓ Playwright installed\n")
	return nil
}

// InstallPlaywrightBrowsers downloads and installs Playwright browsers
func InstallPlaywrightBrowsers(nodeExePath string) error {
	browsersDir := GetBrowsersDir()
	nodeModulesDir := GetNodeModulesDir()

	fmt.Printf("   ▸ Installing Playwright browsers...\n")

	// Ensure browsers directory exists
	if err := os.MkdirAll(browsersDir, 0755); err != nil {
		return fmt.Errorf("failed to create browsers directory: %w", err)
	}

	// Get npx path
	npxPath := filepath.Join(filepath.Dir(nodeExePath), "npx")
	if _, err := os.Stat(npxPath); os.IsNotExist(err) {
		// On Windows, npx might be npx.cmd
		npxPath = filepath.Join(filepath.Dir(nodeExePath), "npx.cmd")
		if _, err := os.Stat(npxPath); os.IsNotExist(err) {
			return fmt.Errorf("npx not found in Node.js installation")
		}
	}

	// Run npx playwright install chromium
	cmd := exec.Command(npxPath, "playwright", "install", "chromium", "--with-deps")
	cmd.Dir = filepath.Dir(nodeModulesDir)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PLAYWRIGHT_BROWSERS_PATH=%s", browsersDir),
	)

	// Show output in real-time for long-running browser download
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("playwright install failed: %w", err)
	}

	fmt.Printf("     ✓ Chromium browser installed\n")
	return nil
}

// VerifyNodeInstallation checks if Node.js is properly installed
func VerifyNodeInstallation(nodePath string) error {
	// Check if node executable exists
	if _, err := os.Stat(nodePath); os.IsNotExist(err) {
		return fmt.Errorf("node executable not found at %s", nodePath)
	}

	// Run node --version to verify it works
	cmd := exec.Command(nodePath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to verify node installation: %w", err)
	}

	// Silent verification for subsequent runs, verbose for first install
	version := string(output)
	if len(version) > 0 && version[len(version)-1] == '\n' {
		version = version[:len(version)-1]
	}

	fmt.Printf("     ✓ Node.js %s verified\n", version)
	return nil
}
