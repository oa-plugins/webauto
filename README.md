# OA Plugin Template

> **GitHub Template Repository for creating [pyhub-office-automation](https://github.com/pyhub-kr/pyhub-office-automation) plugins**

이 템플릿을 사용하여 multi-platform을 지원하는 OA 플러그인을 빠르게 시작할 수 있습니다.

---

## 🚀 Quick Start

### 1. Use This Template

Click the **"Use this template"** button at the top of this repository to create your own plugin repository.

### 2. Customize Your Plugin

See [TEMPLATE.md](./TEMPLATE.md) for detailed customization instructions.

**Quick checklist**:
- [ ] Rename `plugin-name` to your actual plugin name
- [ ] Update `go.mod` module path
- [ ] Customize `plugin.yaml`
- [ ] Implement platform-specific commands in `commands_*.go`
- [ ] Update README with your plugin documentation
- [ ] Test build for all platforms

### 3. Release

```bash
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions will automatically build binaries for all platforms and create a release.

---

## 🏗️ Architecture

### Multi-Platform Support

This template implements **platform-specific command filtering**:

```
Your Plugin
├── Windows → Shows Windows-specific commands
├── macOS   → Shows macOS-specific commands
└── Linux   → Shows Linux-specific commands
```

**How it works**:
- `main.go` - Common entry point
- `commands_windows.go` - Windows-only commands (`//go:build windows`)
- `commands_darwin.go` - macOS-only commands (`//go:build darwin`)
- `commands_linux.go` - Linux-only commands (`//go:build linux`)

### Example

```bash
# On Windows
plugin-name --help
  windows-example    # ✅ Visible

# On macOS
plugin-name --help
  macos-example      # ✅ Visible

# On Linux
plugin-name --help
  linux-example      # ✅ Visible
```

---

## 📦 Platform Support

This template supports building for:
- ✅ Windows (amd64)
- ✅ macOS (amd64, arm64)
- ✅ Linux (amd64, arm64)

Different platforms can expose different commands. Commands are automatically filtered based on build tags.

---

## 🛠️ Development

### Prerequisites

- Go 1.21+
- Git

### Build Locally

```bash
# Build for current platform
go build -o plugin-name ./cmd/plugin-name

# Build for specific platform
GOOS=windows GOARCH=amd64 go build -o plugin-name.exe ./cmd/plugin-name
GOOS=darwin GOARCH=arm64 go build -o plugin-name ./cmd/plugin-name
GOOS=linux GOARCH=amd64 go build -o plugin-name ./cmd/plugin-name
```

### Test

```bash
# Run your plugin
./plugin-name --help
./plugin-name --version

# Test platform-specific commands
./plugin-name windows-example  # Only on Windows
./plugin-name macos-example    # Only on macOS
./plugin-name linux-example    # Only on Linux
```

---

## 📝 Submitting to Registry

After releasing your plugin:

1. **Create manifest** in [oa-plugins/registry](https://github.com/oa-plugins/registry)

   ```yaml
   name: your-plugin
   version: 1.0.0
   platforms:
     windows-amd64:
       uri: https://github.com/oa-plugins/your-plugin/releases/download/v1.0.0/your-plugin-windows-amd64.zip
       sha256: "..."
       bin: your-plugin.exe
     darwin-amd64:
       uri: https://github.com/oa-plugins/your-plugin/releases/download/v1.0.0/your-plugin-darwin-amd64.tar.gz
       sha256: "..."
       bin: your-plugin
     # ... other platforms
   ```

2. **Submit PR** to registry repository

3. **Get SHA256 checksums** from your GitHub Release (see `SHA256SUMS.txt`)

---

## 📚 Resources

- [Plugin Development Guide](https://github.com/oa-plugins/registry/blob/main/docs/plugin-development-guide.md)
- [Multi-Platform Support](https://github.com/oa-plugins/registry/blob/main/docs/multi-platform-support.md)
- [Registry](https://github.com/oa-plugins/registry)
- [Example Plugin: kakaotalk-core](https://github.com/oa-plugins/kakaotalk-core)

---

## 🤝 Contributing

Found an issue with this template? Please [open an issue](https://github.com/oa-plugins/plugin-template/issues).

---

## 📝 License

MIT License - see [LICENSE](LICENSE) for details.

---

**© 2024 pyhub-office-automation** | [GitHub](https://github.com/oa-plugins)
