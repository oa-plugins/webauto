# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **GitHub Template Repository** for creating Go-based plugins for [pyhub-office-automation](https://github.com/pyhub-kr/pyhub-office-automation). Despite the "pyhub" naming, plugins are written in **Go** and compiled to platform-specific binaries.

The template implements **platform-specific command filtering** using Go build tags, allowing different commands to be exposed on Windows, macOS, and Linux.

## Build Commands

### Local Development

```bash
# Build for current platform
go build -o plugin-name ./cmd/plugin-name

# Test the binary
./plugin-name --help
./plugin-name --version
```

### Cross-Platform Builds

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o plugin-name.exe ./cmd/plugin-name

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o plugin-name-darwin-amd64 ./cmd/plugin-name

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o plugin-name-darwin-arm64 ./cmd/plugin-name

# Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o plugin-name-linux-amd64 ./cmd/plugin-name

# Linux (arm64)
GOOS=linux GOARCH=arm64 go build -o plugin-name-linux-arm64 ./cmd/plugin-name
```

### Release Process

```bash
# Create and push a tag to trigger automated release build
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions will automatically build binaries for all platforms, create archives, calculate SHA256 checksums, and publish a GitHub Release.

## Architecture

### Platform-Specific Command System

The core architectural pattern is **build tag-based command registration**:

1. **main.go** - Common entry point that calls `registerPlatformCommands(rootCmd)`
2. **commands_windows.go** (`//go:build windows`) - Implements `registerPlatformCommands()` for Windows
3. **commands_darwin.go** (`//go:build darwin`) - Implements `registerPlatformCommands()` for macOS
4. **commands_linux.go** (`//go:build linux`) - Implements `registerPlatformCommands()` for Linux

Each platform file provides its own implementation of `registerPlatformCommands()`. The Go compiler automatically includes only the file matching the target platform's build tags.

### Key Design Principles

**Build Tags Control Compilation**: Only one `commands_*.go` file is compiled per platform. This means:
- Windows builds only include `commands_windows.go`
- macOS builds only include `commands_darwin.go`
- Linux builds only include `commands_linux.go`

**Single Function Signature**: All platform files must implement:
```go
func registerPlatformCommands(rootCmd *cobra.Command)
```

**Command Registration Pattern**: Each platform registers its commands inside `registerPlatformCommands()`:
```go
func registerPlatformCommands(rootCmd *cobra.Command) {
    rootCmd.AddCommand(exampleWindowsCmd)
    rootCmd.AddCommand(anotherWindowsCmd)
}
```

### Cobra CLI Framework

This template uses [spf13/cobra](https://github.com/spf13/cobra) for CLI structure:
- `rootCmd` is defined in `main.go` with version info and global flags
- Platform-specific commands are added via `rootCmd.AddCommand()` in each `commands_*.go` file
- Commands use standard Cobra structure with `Use`, `Short`, `Long`, and `Run` fields

## Customization Workflow

When creating a new plugin from this template:

1. **Rename**: Replace all instances of `plugin-name` with actual plugin name
   - Directory: `cmd/plugin-name` → `cmd/your-plugin-name`
   - Files: `go.mod`, `plugin.yaml`, `.github/workflows/release.yml`, `.gitignore`
   - Code: Update `pluginName` constant in `main.go`

2. **Update Metadata**: Customize `plugin.yaml` with:
   - Plugin name, version, description
   - Author information
   - Repository URLs
   - Platform support flags
   - Command list
   - Tags for discoverability

3. **Implement Commands**: Edit `commands_*.go` files to add actual platform-specific logic:
   - Windows: Windows API calls, COM automation, PowerShell
   - macOS: AppleScript, osascript, native macOS APIs
   - Linux: DBus, xdotool, CLI tools

4. **Update Module Path**: Change `go.mod` module from `github.com/oa-plugins/plugin-name` to your repository path

5. **Test Builds**: Verify cross-compilation works for all target platforms before tagging a release

## Plugin Distribution

After releasing, plugins are submitted to the [oa-plugins/registry](https://github.com/oa-plugins/registry):

1. GitHub Actions builds and publishes release artifacts
2. Extract SHA256 checksums from `SHA256SUMS.txt` in the release
3. Create a manifest in the registry repository with download URIs and checksums
4. Users install via: `oa plugin install your-plugin-name`

## Important Files

- **plugin.yaml**: Plugin metadata consumed by the plugin manager
- **.github/workflows/release.yml**: Automated multi-platform build and release pipeline
- **TEMPLATE.md**: Step-by-step customization guide for template users
- **go.mod**: Go module definition (update module path when forking)

## Common Pitfalls

- **Forgetting Build Tags**: All platform-specific files MUST have `//go:build <platform>` at the top
- **Inconsistent Function Signatures**: All `commands_*.go` files must implement the same `registerPlatformCommands()` signature
- **Module Path**: Update `go.mod` module path to match your repository, not the template's path
- **Binary Names**: Update binary names in GitHub Actions workflow to match your plugin name
