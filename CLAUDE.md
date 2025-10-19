# CLAUDE.md

This guide tells Claude Code (claude.ai/code) how to work inside the **webauto** repository.

## Project Overview

- **Purpose**: browser automation helper for the OA CLI focused on Korean tax/accounting portals (Hometax, Wehago).
- **Stack**: Go CLI (Cobra) that shells out to a Node.js Playwright runner via JSON IPC.
- **Current surfaced commands**: 14 lifecycle, navigation, element, extraction, and session utilities (no workflow/agent commands are compiled yet).
- **Planned extensions**: Agent-based workflows and richer anti-bot tooling remain design targets in `ARCHITECTURE.md`, but are not merged as of this snapshot.

## Single Source Of Truth

- 모든 버그·기능 논의, 결정 사항은 **반드시 GitHub Issues**에 기록합니다.
- 이 문서는 절차/규칙만 요약합니다. 이슈 히스토리나 중복 설명을 추가하지 말고 관련 이슈 번호를 참조 링크로 남기세요.
- 새 작업을 시작할 때는 열려 있는 Issue, 혹은 연결된 PR의 TODO를 우선 확인하세요.

## Critical Standards

### Command Naming
- Pattern: `<resource>-<action>`
- ✅ `browser-launch`, `page-navigate`, `element-get-text`
- ❌ `launch`, `navigate`, `get` (resource missing)

### Flag Naming
- Pattern: `--<domain-noun>-<attribute>`
- ✅ `--page-url`, `--session-id`, `--element-selector`
- ❌ `--input`, `--path`, `--value` (too generic)

### JSON Output Schema
All commands **must** return:
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
Error codes stay in `UPPER_SNAKE_CASE` (`ELEMENT_NOT_FOUND`, `SESSION_NOT_FOUND`, `INVALID_WAIT_CONDITION`, ...).

## Build & Test Commands

### Setup
```bash
npm install
npx playwright install chromium firefox webkit
go mod tidy
```

### Build
```bash
go build -o webauto cmd/webauto/main.go
GOOS=windows GOARCH=amd64 go build -o webauto.exe cmd/webauto/main.go
GOOS=darwin GOARCH=arm64 go build -o webauto cmd/webauto/main.go
GOOS=linux GOARCH=amd64 go build -o webauto cmd/webauto/main.go
```

### Test
```bash
go test ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Quick integration smoke (JSON contract)
./webauto browser-launch | jq .
./webauto session-list
```

## Architecture

### Package Layout
```
webauto/
├── cmd/webauto/          # Cobra entry + platform wiring
├── internal/utils/       # UUID, JSON, time helpers
├── pkg/
│   ├── cli/              # CLI command implementations (13 commands)
│   ├── config/           # Environment-driven defaults and toggles
│   ├── ipc/              # TCP/IPC helpers between Go and Node
│   ├── playwright/       # Session manager + Playwright bridge
│   └── response/         # StandardResponse builder utilities
└── webauto/              # Node runner code (generated/maintained separately)
```

### Go ↔ Node.js Integration
1. Go spawns `node` with `exec.CommandContext`.
2. Commands stream JSON over TCP sockets managed by `pkg/ipc`.
3. Responses resolve to `response.StandardResponse` and propagate back to the CLI.

Example (`pkg/playwright/session.go`) snippet:
```go
cmd := exec.CommandContext(ctx, nodePath,
	"-e", fmt.Sprintf(`(async () => {
		const { %s } = require('playwright');
		/* ... */
	})();`, browserType),
)
output, err := cmd.Output()
```
Keep scripts compatible with plain `playwright`. Add new npm deps only when `package.json` declares them.

## Command Surface (14)

**Browser Lifecycle**
- `browser-launch`: boot a browser session, returns `session_id`
- `browser-close`: close session and clean up cache

**Navigation & Form Actions**
- `page-navigate`: load URL with optional wait config
- `page-evaluate`: execute custom JavaScript in page context (Issue #34)
- `element-click`: click selector
- `element-type`: type text with optional delay
- `element-wait`: wait for visibility/hidden/attached/detached
- `element-query-all`: query multiple elements with optional data extraction
- `form-fill`: map of field selectors to values and optional submit

**Element Inspection**
- `element-get-text`: fetch inner text
- `element-get-attribute`: fetch attribute value

**Data Extraction**
- `page-screenshot`: save PNG to disk
- `page-get-html`: get HTML source from page or element
- `page-pdf`: save PDF (Chromium only)

**Session Management**
- `session-list`: enumerate open sessions
- `session-close`: close by ID

When adding commands, register them in `pkg/cli/root.go` and expose shared helpers via `pkg/playwright`.

## Implementation Guidelines

- Respect defaults from `pkg/config/config.go`. Anti-bot toggles (`ENABLE_STEALTH`, etc.) exist as configuration but still require handler code in the Playwright bridge—add behavior there when implementing enhancements.
- Favor streaming JSON over stdout/stdin or TCP; avoid writing temporary files unless unavoidable.
- Observe performance envelopes:
  - `browser-launch` / `browser-close`: <500 ms nominal
  - `page-navigate`: <1000 ms (network bound)
  - Element interactions: <300 ms
  - Screenshot/PDF: <1000 ms

## Error Handling

Common failure modes and recoveries:
- `NODE_NOT_FOUND`: ensure Node.js is on `PATH`
- `PLAYWRIGHT_NOT_INSTALLED`: `npm install && npx playwright install chromium`
- `SESSION_NOT_FOUND`: user passed stale `session_id`; suggest `session-list`
- `ELEMENT_NOT_FOUND`: bad selector; advise wait condition or different locator

Standard error response template:
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
    "recovery_suggestion": "Verify selector or adjust wait condition"
  },
  "metadata": {
    "plugin": "webauto",
    "version": "1.0.0",
    "execution_time_ms": 5200
  }
}
```

## Usage Walkthrough

```bash
# 1. Launch browser (Chromium headless by default)
oa webauto browser-launch --headless false
# => {"success":true,"data":{"session_id":"ses_abcd1234",...}}

# 2. Navigate to target page
oa webauto page-navigate \
  --session-id ses_abcd1234 \
  --page-url "https://hometax.go.kr"

# 3. Wait and interact with elements
oa webauto element-wait \
  --session-id ses_abcd1234 \
  --element-selector "#loginBtn" \
  --wait-for visible

oa webauto element-click \
  --session-id ses_abcd1234 \
  --element-selector "#loginBtn"

# 4. Capture evidence
oa webauto page-screenshot \
  --session-id ses_abcd1234 \
  --image-path evidence.png

# 5. Close session
oa webauto browser-close --session-id ses_abcd1234
```

## Platform Support

- Windows 10/11 (Chromium, Firefox)
- macOS 11+ Intel/Apple Silicon (Chromium, Firefox, WebKit)
- Ubuntu 20.04+ (Chromium, Firefox; WebKit limited)

Platform notes:
- Windows default node path: `C:\Program Files\nodejs\node.exe`
- macOS: `/usr/local/bin/node` or asdf shim
- Linux: ensure `libnss3`, `libatk1.0-0`, `libatk-bridge2.0-0`

## Roadmap Alignment

- Follow staged goals in `ARCHITECTURE.md` for agent workflows and advanced anti-bot behavior.
- Gate new features behind Issues/PRs; update this document only when functionality lands in main.

## Key References

- [ARCHITECTURE.md](ARCHITECTURE.md)
- [docs/implementation-guide.md](docs/implementation-guide.md)
- [PRODUCTION_READINESS_CHECKLIST.md](PRODUCTION_READINESS_CHECKLIST.md)
- Original idea: [webauto-idea.md](https://github.com/oa-plugins/plugin-designer/blob/main/ideas/webauto-idea.md)
- OA standards PRD: <https://github.com/oa-plugins/plugin-designer/blob/main/PRD.md>
- Playwright docs: <https://playwright.dev/>

## Legal Notice

**Personal use only**: Automate your own accounts/data.

**Prohibited**:
- Unauthorized access to third-party accounts
- Violation of site terms of service
- Commercial scraping/resale
- Excessive automated traffic

Compliance responsibility lies with the end user.
