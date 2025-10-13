# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**webauto** is a Playwright Agents-based intelligent browser automation plugin for the OA (Office Automation) CLI system. It targets Korean tax/accounting services (Hometax, Wehago) with sophisticated UI automation and anti-bot capabilities.

**Architecture**: Go CLI wrapper + Node.js Playwright Agents integration via subprocess IPC
**Target**: 14 commands across 4 categories (Agent-based, Browser Control, Data Extraction, Session Management)

## Critical Standards

### Command Naming Convention
**Pattern**: `<resource>-<action>`

Examples:
- ✅ `browser-launch`, `page-navigate`, `workflow-plan`
- ❌ `launch`, `navigate`, `plan` (missing resource)

### Flag Naming Convention
**Pattern**: `--<domain-noun>-<attribute>`

**IMPORTANT**: Use domain-specific nouns, NOT generic terms.

Examples:
- ✅ `--page-url`, `--script-file`, `--browser-type`, `--element-selector`, `--image-path`
- ❌ `--file-path`, `--input`, `--output`, `--path` (too generic)

Rationale: `--image-path` is explicit (screenshot file), while `--file-path` is ambiguous (script? plan? image?).

### JSON Output Schema
**All commands MUST return**:
```json
{
  "success": boolean,
  "data": object | null,
  "error": {
    "code": "UPPER_SNAKE_CASE",
    "message": "string",
    "details": {},
    "recovery_suggestion": "string"
  } | null,
  "metadata": {
    "plugin": "webauto",
    "version": "1.0.0",
    "execution_time_ms": number
  }
}
```

Error codes must be `UPPER_SNAKE_CASE` (e.g., `ELEMENT_NOT_FOUND`, `SESSION_NOT_FOUND`).

## Build & Test Commands

### Setup
```bash
# Install Node.js dependencies (Playwright Agents)
npm install
npx playwright install chromium firefox webkit

# Install Go dependencies
go mod tidy
```

### Build
```bash
# Single platform
go build -o webauto cmd/webauto/main.go

# Cross-platform builds
GOOS=windows GOARCH=amd64 go build -o webauto.exe cmd/webauto/main.go
GOOS=darwin GOARCH=amd64 go build -o webauto cmd/webauto/main.go
GOOS=darwin GOARCH=arm64 go build -o webauto cmd/webauto/main.go
GOOS=linux GOARCH=amd64 go build -o webauto cmd/webauto/main.go
```

### Test
```bash
# Unit tests
go test ./...

# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Integration test (verify JSON output)
./webauto browser-launch | jq .
./webauto session-list
```

## Architecture

### Package Structure
```
webauto/
├── cmd/webauto/            # CLI entry point
│   ├── main.go            # Root command, platform detection
│   ├── commands_darwin.go  # macOS-specific commands
│   ├── commands_linux.go   # Linux-specific commands
│   └── commands_windows.go # Windows-specific commands
├── pkg/
│   ├── config/            # Environment variable loading
│   ├── response/          # StandardResponse types, error codes
│   ├── cli/              # Cobra command definitions (14 commands)
│   ├── playwright/       # Playwright Agents wrappers
│   ├── antibot/          # Stealth mode, fingerprint, behavior randomization
│   └── ipc/              # Node.js subprocess communication
└── internal/utils/       # UUID, JSON, time utilities
```

### Command Categories (14 total)

**Agent-Based Automation (4 commands)**:
- `workflow-plan`: Planner Agent generates test plans from scenarios
- `workflow-generate`: Generator Agent converts plans to Playwright code
- `workflow-execute`: Execute generated scripts
- `workflow-heal`: Healer Agent auto-repairs failed scripts

**Browser Control (6 commands)**:
- `browser-launch`, `browser-close`: Browser lifecycle
- `page-navigate`: URL navigation
- `element-click`, `element-type`: DOM interaction
- `form-fill`: Multi-field form automation

**Data Extraction (2 commands)**:
- `page-screenshot`: Screenshot capture
- `page-pdf`: PDF export

**Session Management (2 commands)**:
- `session-list`: List active browser sessions
- `session-close`: Close specific session

### Go ↔ Node.js Integration

**Communication Pattern**:
1. Go spawns Node.js subprocess with `exec.CommandContext`
2. Pass JSON via stdin, receive JSON from stdout
3. Timeouts: Planner 60s, Generator 30s, Healer 90s
4. Session persistence: In-memory map with UUID keys

**Example** (Planner Agent):
```go
cmd := exec.CommandContext(ctx, nodePath,
    "-e", `const { planner } = require('@playwright/agents');
           (async () => {
             const result = await planner.explore('${url}', { scenario: '${scenario}' });
             console.log(JSON.stringify(result));
           })();`)
output, err := cmd.Output()
```

### Anti-Bot Strategy

**Implemented Techniques**:
1. **Playwright Stealth Mode**: Auto-hide WebDriver flags
2. **Fingerprint Rotation**: User-Agent randomization from pool
3. **Behavior Randomization**: 10-50ms typing delays, ±5-15px mouse jitter
4. **Rate Limiting**: 500ms minimum interval between requests

**Configuration** (via environment variables):
- `ENABLE_STEALTH=true`
- `ENABLE_FINGERPRINT=true`
- `TYPING_DELAY_MS=30`
- `MOUSE_MOVE_JITTER_PX=10`

## Implementation Guidelines

### Performance Targets
| Category | Target | Measurement |
|----------|--------|-------------|
| Agent-based | 5-30s | Planner/Generator/Healer execution |
| Browser control | <500ms | browser-launch/close |
| Page control | <1000ms | page-navigate (includes network) |
| Element ops | <300ms | element-click/type |
| Data extraction | <1000ms | screenshot/PDF |
| Session mgmt | <100ms | session-list/close |

**Average (excluding Agents)**: <500ms

### Error Handling

**Common Errors**:
- `NODE_NOT_FOUND`: Node.js not in PATH → Install from nodejs.org
- `PLAYWRIGHT_NOT_INSTALLED`: Missing Playwright → `npm install playwright @playwright/agents`
- `SESSION_NOT_FOUND`: Invalid session ID → Use `session-list` to verify
- `ELEMENT_NOT_FOUND`: Selector invalid → Add `--wait-visible` or increase timeout

**Error Response Example**:
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "ELEMENT_NOT_FOUND",
    "message": "Element with selector '#submit-btn' not found",
    "details": {
      "selector": "#submit-btn",
      "page_url": "https://hometax.go.kr",
      "timeout_ms": 5000
    },
    "recovery_suggestion": "Verify selector or use --wait-visible flag"
  },
  "metadata": {
    "plugin": "webauto",
    "version": "1.0.0",
    "execution_time_ms": 5200
  }
}
```

## Workflow Examples

### Agent-Based Automation (Recommended)
```bash
# 1. Generate plan from scenario
oa webauto workflow-plan \
  --page-url "https://hometax.go.kr" \
  --scenario-text "로그인 → 세금계산서 조회 → CSV 다운로드" \
  --output-path hometax_plan.md

# 2. Generate executable code
oa webauto workflow-generate \
  --plan-file hometax_plan.md \
  --output-path hometax_automation.ts

# 3. Execute automation
oa webauto workflow-execute \
  --script-file hometax_automation.ts \
  --headless false

# 4. Auto-heal failures
oa webauto workflow-heal \
  --script-file hometax_automation.ts \
  --max-attempts 5
```

### Direct Browser Control
```bash
# 1. Launch browser
oa webauto browser-launch --headless false
# → {"success": true, "data": {"session_id": "ses_abc123", ...}}

# 2. Navigate
oa webauto page-navigate \
  --page-url "https://hometax.go.kr" \
  --session-id ses_abc123

# 3. Fill form
oa webauto form-fill \
  --form-data '{"username":"user1","password":"pass123"}' \
  --session-id ses_abc123 \
  --submit true

# 4. Screenshot
oa webauto page-screenshot \
  --image-path result.png \
  --session-id ses_abc123

# 5. Close
oa webauto browser-close --session-id ses_abc123
```

## Platform Support

- ✅ **Windows** 10/11: Chromium, Firefox (WebKit limited)
- ✅ **macOS** 11+ (Intel/Apple Silicon): All browsers supported
- ✅ **Linux** Ubuntu 20.04+: Chromium, Firefox (WebKit limited)

**Platform-specific considerations**:
- Windows: Node.js path typically `C:\Program Files\nodejs\node.exe`
- macOS: Node.js path typically `/usr/local/bin/node`
- Linux: Requires `libnss3`, `libatk1.0-0` dependencies for Playwright

## Development Phases

**Phase 1 (MVP)**: 5 core commands (browser-launch, browser-close, page-navigate, element-click, page-screenshot)
- Target: Basic browser control POC, Hometax login page navigation

**Phase 2 (Agents)**: 8 additional commands (all workflow-* + element-type, form-fill, session-*)
- Target: Full Playwright Agents integration, auto-generated scripts

**Phase 3 (Production)**: Remaining command (page-pdf) + cross-platform hardening
- Target: Windows/macOS support, anti-bot enhancement, <500ms avg performance

## Key References

- **Primary Design Doc**: [ARCHITECTURE.md](ARCHITECTURE.md) - Complete system design, all 14 commands, JSON schemas, error codes
- **Implementation Guide**: [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md) - Step-by-step coding instructions
- **Original Idea**: [webauto-idea.md](https://github.com/oa-plugins/plugin-designer/blob/main/ideas/webauto-idea.md)
- **OA Standards**: [plugin-designer PRD](https://github.com/oa-plugins/plugin-designer/blob/main/PRD.md)
- **Playwright Docs**: https://playwright.dev/
- **Playwright Agents**: https://playwright.dev/docs/test-agents

## Legal Notice

**Personal Use Only**: This plugin is for automating YOUR OWN tax/accounting data.

**Prohibited**:
- ❌ Unauthorized access to others' accounts
- ❌ Terms of service violations
- ❌ Commercial scraping
- ❌ Excessive requests (rate limit abuse)

Users are solely responsible for legal compliance.
