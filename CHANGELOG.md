# Changelog

All notable changes to the webauto plugin will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `--session-id` flag for `browser-launch` command to support custom session IDs (Issue #36)
  - Enables predictable session management in .oas scripts
  - Auto-generates session ID if not provided (backward compatible)
  - Validates against duplicate session IDs
- `--no-headless` flag for `browser-launch` command to support visible browser mode (Issue #36)
  - Complements existing `--headless` flag
  - Follows Cobra boolean flag best practices
  - Defaults to headless mode if not specified

### Changed
- Migrated all 18 shell script examples to .oas format (Issue #36)
  - 43% code reduction (2,156 → 1,235 lines)
  - Improved maintainability and readability
  - Better error handling with @try/@catch blocks
  - Consistent variable management with @set
- Updated `SessionManager.Create()` signature to accept optional `customSessionID` parameter
  - Maintains backward compatibility with auto-generated IDs
  - Supports both use cases: automated workflows and manual testing

### Dependencies
- Requires OA CLI >= 1.0.0 (commit 7538fc9) for:
  - Boolean flag support via `oa plugin exec` (Issue #31)
  - Backslash line continuation in .oas scripts (Issue #32)
  - Enhanced batch command execution

### Fixed
- Session ID management now supports both auto-generated and custom IDs
- Browser visibility control now properly supported via `--no-headless` flag
- Multi-line commands in .oas scripts now work correctly with backslash continuation

## [1.0.0] - 2025-10-21

### Added
- Initial release with 14 core commands:
  - Browser lifecycle: `browser-launch`, `browser-close`
  - Navigation: `page-navigate`, `page-evaluate`
  - Element interaction: `element-click`, `element-type`, `element-wait`, `element-query-all`
  - Form handling: `form-fill`
  - Element inspection: `element-get-text`, `element-get-attribute`
  - Data extraction: `page-screenshot`, `page-get-html`, `page-pdf`
  - Session management: `session-list`, `session-close`
- Cross-platform support: Windows, macOS (Intel/Apple Silicon), Linux
- Multi-browser support: Chromium, Firefox, WebKit
- Automatic Node.js and Playwright bootstrapping
- Persistent session management with file-based storage
- JSON-based IPC between Go CLI and Node.js Playwright runner
- Comprehensive documentation and examples

[Unreleased]: https://github.com/oa-plugins/webauto/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/oa-plugins/webauto/releases/tag/v1.0.0
