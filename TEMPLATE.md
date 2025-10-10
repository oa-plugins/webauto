# Template Customization Guide

> **Step-by-step guide to customize this template for your plugin**

---

## 📋 Checklist

Use this checklist to ensure you've customized everything:

- [ ] **Step 1**: Rename plugin
- [ ] **Step 2**: Update module path
- [ ] **Step 3**: Customize plugin.yaml
- [ ] **Step 4**: Implement commands
- [ ] **Step 5**: Update README
- [ ] **Step 6**: Test build
- [ ] **Step 7**: Create first release

---

## Step 1: Rename Plugin

### Files to Rename

```bash
# Rename directory
mv cmd/plugin-name cmd/your-plugin-name

# Update in .github/workflows/release.yml
# Find and replace: plugin-name → your-plugin-name

# Update in .gitignore
# Add your binary names
```

### Search and Replace

Replace `plugin-name` with your actual plugin name in:
- `go.mod`
- `plugin.yaml`
- `.github/workflows/release.yml`
- `.gitignore`
- `README.md`

---

## Step 2: Update Module Path

**go.mod**:
```go
module github.com/oa-plugins/your-plugin-name

go 1.21

require github.com/spf13/cobra v1.8.1
```

**cmd/your-plugin-name/main.go**:
```go
const (
    pluginName    = "your-plugin-name"
    pluginVersion = "0.1.0"
)

var (
    rootCmd = &cobra.Command{
        Use:     pluginName,
        Short:   "Your plugin short description",
        Long:    `Your plugin long description`,
        Version: pluginVersion,
    }
)
```

---

## Step 3: Customize plugin.yaml

Update all TODO fields:

```yaml
name: your-plugin-name
version: 0.1.0
display_name: Your Plugin Name
description: Detailed description of what your plugin does
short_description: Brief one-liner

author: your-github-username

# Update repository URLs
homepage: https://github.com/oa-plugins/your-plugin-name
repository: https://github.com/oa-plugins/your-plugin-name
documentation: https://github.com/oa-plugins/your-plugin-name#readme
issues: https://github.com/oa-plugins/your-plugin-name/issues

# Update platform support based on what you'll implement
platform_support:
  windows:
    supported: true  # Change to false if not implementing Windows support
  darwin:
    supported: true  # Change to false if not implementing macOS support
  linux:
    supported: true  # Change to false if not implementing Linux support

# Update commands list
commands:
  - name: your-command
    description: What this command does
    platforms:
      - windows
      - darwin
      - linux

# Update tags
tags:
  - your
  - relevant
  - tags
```

---

## Step 4: Implement Commands

### Windows-Specific Commands

**cmd/your-plugin-name/commands_windows.go**:
```go
//go:build windows

package main

import (
    "github.com/spf13/cobra"
)

func registerPlatformCommands(rootCmd *cobra.Command) {
    // Register your Windows-specific commands
    rootCmd.AddCommand(yourWindowsCmd)
}

var yourWindowsCmd = &cobra.Command{
    Use:   "your-command",
    Short: "Description",
    Run: func(cmd *cobra.Command, args []string) {
        // TODO: Implement Windows-specific logic
        // Example: Windows API calls, COM automation, etc.
    },
}
```

### macOS-Specific Commands

**cmd/your-plugin-name/commands_darwin.go**:
```go
//go:build darwin

package main

import (
    "github.com/spf13/cobra"
)

func registerPlatformCommands(rootCmd *cobra.Command) {
    // Register your macOS-specific commands
    rootCmd.AddCommand(yourMacCmd)
}

var yourMacCmd = &cobra.Command{
    Use:   "your-command",
    Short: "Description",
    Run: func(cmd *cobra.Command, args []string) {
        // TODO: Implement macOS-specific logic
        // Example: AppleScript, osascript, etc.
    },
}
```

### Linux-Specific Commands

**cmd/your-plugin-name/commands_linux.go**:
```go
//go:build linux

package main

import (
    "github.com/spf13/cobra"
)

func registerPlatformCommands(rootCmd *cobra.Command) {
    // Register your Linux-specific commands
    rootCmd.AddCommand(yourLinuxCmd)
}

var yourLinuxCmd = &cobra.Command{
    Use:   "your-command",
    Short: "Description",
    Run: func(cmd *cobra.Command, args []string) {
        // TODO: Implement Linux-specific logic
        // Example: DBus, CLI tools, etc.
    },
}
```

### Adding Platform-Specific Dependencies

**go.mod**:
```go
module github.com/oa-plugins/your-plugin-name

go 1.21

require (
    github.com/spf13/cobra v1.8.1
    // Add Windows-specific deps here (won't affect other platforms)
    github.com/lxn/win v0.0.0-20210218163916-a377121e959e
)
```

---

## Step 5: Update README

Replace the template README.md with your plugin's documentation:

```markdown
# Your Plugin Name

> Brief description

## Features

- Feature 1
- Feature 2

## Installation

\`\`\`bash
oa plugin install your-plugin-name
\`\`\`

## Usage

\`\`\`bash
your-plugin-name command --options
\`\`\`

## Platform-Specific Commands

### Windows
- `command1` - Description

### macOS
- `command2` - Description

### Linux
- `command3` - Description
```

---

## Step 6: Test Build

### Build for All Platforms

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o your-plugin.exe ./cmd/your-plugin-name

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o your-plugin-darwin-amd64 ./cmd/your-plugin-name

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o your-plugin-darwin-arm64 ./cmd/your-plugin-name

# Linux
GOOS=linux GOARCH=amd64 go build -o your-plugin-linux-amd64 ./cmd/your-plugin-name
```

### Test on Current Platform

```bash
# Build
go build -o your-plugin ./cmd/your-plugin-name

# Test
./your-plugin --help
./your-plugin --version
./your-plugin your-command
```

---

## Step 7: Create First Release

### Push to GitHub

```bash
git init
git add .
git commit -m "feat: initial commit"
git remote add origin https://github.com/oa-plugins/your-plugin-name.git
git branch -M main
git push -u origin main
```

### Create Release Tag

```bash
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

### GitHub Actions will:
1. Build binaries for all platforms
2. Create archives (zip for Windows, tar.gz for Unix)
3. Calculate SHA256 checksums
4. Create GitHub Release with all artifacts

---

## 🎯 Next Steps

After your first release:

1. **Get SHA256 checksums** from GitHub Release (`SHA256SUMS.txt`)

2. **Create registry manifest** at `oa-plugins/registry`

3. **Submit PR** to registry

4. **Users can install** via `oa plugin install your-plugin-name`

---

## 💡 Tips

### Platform-Specific Implementation Tips

**Windows**:
- Use Windows API via `github.com/lxn/win`
- COM automation for Office apps
- PowerShell execution for complex tasks

**macOS**:
- Use `osascript` for AppleScript
- Objective-C bridges for native APIs
- Shell out to macOS CLI tools

**Linux**:
- DBus for desktop integration
- CLI tools (xdotool, wmctrl, etc.)
- Shell scripting

### Command Organization

```
your-plugin/
├── cmd/your-plugin/
│   ├── main.go                 # Common entry point
│   ├── commands_windows.go     # Windows commands
│   ├── commands_darwin.go      # macOS commands
│   ├── commands_linux.go       # Linux commands
│   └── common.go               # Shared utilities
├── internal/
│   ├── windows/                # Windows-specific packages
│   ├── darwin/                 # macOS-specific packages
│   └── linux/                  # Linux-specific packages
```

### Testing Strategy

1. **Unit tests**: Test common logic
2. **Platform-specific tests**: Use build tags
3. **Integration tests**: Test on actual platforms (use GitHub Actions matrix)

---

## 📚 Additional Resources

- [Go Build Tags](https://golang.org/pkg/go/build/#hdr-Build_Constraints)
- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [Cross-Compilation](https://golang.org/doc/install/source#environment)

---

**Questions?** Open an issue at https://github.com/oa-plugins/plugin-template/issues
