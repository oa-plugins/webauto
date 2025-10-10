package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	templateName       = "plugin-name"
	templateModule     = "github.com/oa-plugins/plugin-name"
	templateAuthor     = "your-github-username"
	defaultModuleOrgPrefix = "github.com/oa-plugins/"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] PLUGIN_NAME\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Create a new OA plugin from template\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  PLUGIN_NAME    Name of the plugin (kebab-case, e.g., my-awesome-plugin)\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s my-plugin\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --module github.com/myorg/my-plugin --author myusername my-plugin\n", os.Args[0])
	}

	moduleFlag := flag.String("module", "", "Go module path (default: github.com/oa-plugins/PLUGIN_NAME)")
	authorFlag := flag.String("author", "", "Author GitHub username (default: from git config or prompt)")
	outputFlag := flag.String("output", "", "Output directory (default: PLUGIN_NAME)")
	flag.Parse()

	// Get plugin name from args
	var pluginName string
	if flag.NArg() > 0 {
		pluginName = flag.Arg(0)
	} else {
		// Interactive mode
		pluginName = promptInput("Plugin name (kebab-case)", "")
	}

	if pluginName == "" {
		fmt.Fprintln(os.Stderr, "Error: plugin name is required")
		flag.Usage()
		os.Exit(1)
	}

	// Validate plugin name
	if !isValidPluginName(pluginName) {
		fmt.Fprintf(os.Stderr, "Error: invalid plugin name '%s'\n", pluginName)
		fmt.Fprintln(os.Stderr, "Plugin name must:")
		fmt.Fprintln(os.Stderr, "  - Use kebab-case (lowercase with hyphens)")
		fmt.Fprintln(os.Stderr, "  - Start with a letter")
		fmt.Fprintln(os.Stderr, "  - Contain only letters, numbers, and hyphens")
		os.Exit(1)
	}

	// Get module path
	modulePath := *moduleFlag
	if modulePath == "" {
		defaultModule := defaultModuleOrgPrefix + pluginName
		modulePath = promptInput("Go module path", defaultModule)
	}

	// Get author
	author := *authorFlag
	if author == "" {
		gitAuthor := getGitUsername()
		author = promptInput("Author GitHub username", gitAuthor)
	}

	// Get output directory
	outputDir := *outputFlag
	if outputDir == "" {
		outputDir = pluginName
	}

	// Check if output directory exists
	if _, err := os.Stat(outputDir); err == nil {
		fmt.Fprintf(os.Stderr, "Error: directory '%s' already exists\n", outputDir)
		os.Exit(1)
	}

	// Find template directory
	templateDir, err := findTemplateDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Creating plugin '%s'...\n", pluginName)
	fmt.Printf("  Module: %s\n", modulePath)
	fmt.Printf("  Author: %s\n", author)
	fmt.Printf("  Output: %s\n", outputDir)
	fmt.Println()

	// Create plugin
	if err := createPlugin(templateDir, outputDir, pluginName, modulePath, author); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating plugin: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Plugin '%s' created successfully!\n\n", pluginName)
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", outputDir)
	fmt.Println("  go build -o", pluginName, "./cmd/"+pluginName)
	fmt.Println("  ./"+pluginName, "--help")
	fmt.Println()
	fmt.Println("To customize further:")
	fmt.Println("  - Edit plugin.yaml for metadata")
	fmt.Println("  - Implement commands in cmd/" + pluginName + "/commands_*.go")
	fmt.Println("  - Update README.md")
}

func findTemplateDir() (string, error) {
	// Try current directory first (when running from template repo)
	if isTemplateDir(".") {
		return ".", nil
	}

	// Try parent directories (when running go run github.com/...)
	// Go downloads modules to GOPATH/pkg/mod
	// We need to find where this code is running from
	executable, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(executable)
		for i := 0; i < 5; i++ {
			if isTemplateDir(dir) {
				return dir, nil
			}
			dir = filepath.Dir(dir)
		}
	}

	return "", fmt.Errorf("could not find template directory. Please run this from the plugin-template repository or use 'go run ./cmd/create'")
}

func isTemplateDir(dir string) bool {
	// Check if this looks like template directory
	markers := []string{
		"plugin.yaml",
		"cmd/plugin-name/main.go",
		"TEMPLATE.md",
	}
	for _, marker := range markers {
		if _, err := os.Stat(filepath.Join(dir, marker)); err != nil {
			return false
		}
	}
	return true
}

func createPlugin(templateDir, outputDir, pluginName, modulePath, author string) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Copy template files
	if err := copyDir(templateDir, outputDir, pluginName); err != nil {
		return fmt.Errorf("failed to copy template: %w", err)
	}

	// Rename directories
	oldCmdDir := filepath.Join(outputDir, "cmd", templateName)
	newCmdDir := filepath.Join(outputDir, "cmd", pluginName)
	if err := os.Rename(oldCmdDir, newCmdDir); err != nil {
		return fmt.Errorf("failed to rename cmd directory: %w", err)
	}

	// Replace content in files
	replacements := map[string]string{
		templateName:   pluginName,
		"Plugin Name":  toDisplayName(pluginName),
		templateModule: modulePath,
		templateAuthor: author,
	}

	filesToProcess := []string{
		"go.mod",
		"plugin.yaml",
		"README.md",
		".gitignore",
		".github/workflows/release.yml",
		filepath.Join("cmd", pluginName, "main.go"),
		filepath.Join("cmd", pluginName, "commands_windows.go"),
		filepath.Join("cmd", pluginName, "commands_darwin.go"),
		filepath.Join("cmd", pluginName, "commands_linux.go"),
	}

	for _, file := range filesToProcess {
		fullPath := filepath.Join(outputDir, file)
		if err := replaceInFile(fullPath, replacements); err != nil {
			return fmt.Errorf("failed to process %s: %w", file, err)
		}
	}

	// Remove template-specific files
	toRemove := []string{
		"TEMPLATE.md",
		"cmd/create",
	}
	for _, file := range toRemove {
		fullPath := filepath.Join(outputDir, file)
		os.RemoveAll(fullPath) // Ignore errors for files that might not exist
	}

	// Remove go.sum to avoid conflicts, will be regenerated by go mod tidy
	goSumPath := filepath.Join(outputDir, "go.sum")
	os.Remove(goSumPath)

	// Initialize go modules
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = outputDir
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Warning: go mod tidy failed: %v\n%s\n", err, output)
	}

	return nil
}

func copyDir(src, dst, pluginName string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// Skip certain files/directories
		if shouldSkip(entry.Name()) {
			continue
		}

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath, pluginName); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func shouldSkip(name string) bool {
	skipList := []string{
		".git",
		".DS_Store",
		"node_modules",
		"vendor",
	}
	for _, skip := range skipList {
		if name == skip {
			return true
		}
	}
	return false
}

func replaceInFile(filePath string, replacements map[string]string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	newContent := string(content)

	// Replace in specific order: longer strings first to avoid partial replacements
	orderedKeys := []string{
		templateModule,   // Must come before templateName
		"Plugin Name",    // Display name
		templateAuthor,   // Author
		templateName,     // Plugin name (last to avoid conflicts)
	}

	for _, key := range orderedKeys {
		if newValue, exists := replacements[key]; exists {
			newContent = strings.ReplaceAll(newContent, key, newValue)
		}
	}

	return os.WriteFile(filePath, []byte(newContent), 0644)
}

func isValidPluginName(name string) bool {
	// Must be kebab-case: lowercase letters, numbers, hyphens
	// Must start with letter
	matched, _ := regexp.MatchString(`^[a-z][a-z0-9-]*$`, name)
	return matched
}

func toDisplayName(pluginName string) string {
	// Convert kebab-case to Title Case
	// e.g., "my-awesome-plugin" -> "My Awesome Plugin"
	words := strings.Split(pluginName, "-")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

func getGitUsername() string {
	cmd := exec.Command("git", "config", "user.name")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func promptInput(prompt, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)

	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}
