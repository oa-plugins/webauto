# webauto 플러그인 구현 가이드

> **작성일**: 2025-10-13
> **대상**: webauto 플러그인 개발 팀
> **목적**: plugin-designer의 설계 명세를 기반으로 webauto 플러그인을 구현

---

## 📐 설계 문서 참조

### 필수 읽기 자료

**1. 아키텍처 설계**
- [webauto-architecture.md](https://github.com/oa-plugins/plugin-designer/blob/main/designs/webauto-architecture.md)
  - 전체 시스템 구조
  - CLI 명령어 정의 (14개)
  - JSON 출력 스키마
  - Go + Node.js 통합 구조
  - Playwright Agents 통합 가이드
  - 에러 코드 정의
  - 성능 목표

**2. 기능 명세 (ideas/)**
- [webauto-idea.md](https://github.com/oa-plugins/plugin-designer/blob/main/ideas/webauto-idea.md)
  - 4가지 핵심 사용 사례
  - 대상 사용자 정의
  - 기술 스택 선정 근거
  - 차별화 포인트

**3. 표준 준수 사항**
- [OA 플러그인 표준](https://github.com/oa-plugins/plugin-designer/blob/main/PRD.md#5-플러그인-간-조율-프로세스)
  - 명령어 명명 규칙: `<resource>-<action>`
  - 플래그 명명 규칙: `--<domain-noun>-<attribute>`
  - JSON 출력 표준
  - 에러 코드 표준 (UPPER_SNAKE_CASE)

---

## 🎯 구현 목표

### 준수해야 할 표준

#### 1. 명령어 이름 (Command Naming)

**패턴**: `<resource>-<action>`

✅ **구현해야 할 명령어 (14개)**:

**Agent-Based Automation (고수준 - 4개)**:
```bash
webauto workflow-plan        # Planner Agent로 테스트 플랜 생성
webauto workflow-generate    # Generator Agent로 코드 생성
webauto workflow-execute     # 생성된 스크립트 실행
webauto workflow-heal        # Healer Agent로 자동 수리
```

**Direct Browser Control (저수준 - 6개)**:
```bash
webauto browser-launch       # 브라우저 시작
webauto browser-close        # 브라우저 종료
webauto page-navigate        # URL 이동
webauto element-click        # 요소 클릭
webauto element-type         # 텍스트 입력
webauto form-fill            # 폼 자동 입력
```

**Data Extraction (2개)**:
```bash
webauto page-screenshot      # 스크린샷 촬영
webauto page-pdf             # PDF 저장
```

**Session Management (2개)**:
```bash
webauto session-list         # 세션 목록
webauto session-close        # 세션 종료
```

❌ **잘못된 예** (사용 금지):
```bash
webauto plan              # resource 없음
webauto launch            # resource 없음
webauto click             # resource 없음
```

---

#### 2. 플래그 이름 (Flag Naming)

**패턴**: `--<domain-noun>-<attribute>`

**중요**: 도메인 특화 명사를 사용하세요! 일반적인 명사(file, path, input, output)는 금지입니다.

✅ **올바른 플래그 (도메인 특화)**:

**Agent 명령어**:
```bash
--page-url <url>              # 대상 페이지 URL
--scenario-text <text>        # 시나리오 설명
--plan-file <path>            # 플랜 파일 경로
--script-file <path>          # 스크립트 파일 경로
--output-path <path>          # 출력 파일 경로
```

**Browser 명령어**:
```bash
--browser-type <type>         # chromium|firefox|webkit
--session-id <id>             # 세션 ID
--viewport-width <int>        # 뷰포트 너비
--viewport-height <int>       # 뷰포트 높이
--user-agent <string>         # User-Agent
```

**Page 명령어**:
```bash
--page-url <url>              # 페이지 URL
--wait-for <condition>        # load|networkidle|domcontentloaded
--timeout-ms <int>            # 타임아웃
```

**Element 명령어**:
```bash
--element-selector <string>   # CSS 셀렉터 또는 XPath
--text-input <string>         # 입력할 텍스트
--click-count <int>           # 클릭 횟수
--delay-ms <int>              # 타이핑 지연
```

**Screenshot/PDF**:
```bash
--image-path <path>           # 스크린샷 저장 경로
--pdf-path <path>             # PDF 저장 경로
--full-page <bool>            # 전체 페이지 캡처
--pdf-format <string>         # A4|Letter|Legal
```

❌ **잘못된 플래그** (사용 금지):
```bash
--file-path         # 너무 일반적 (어떤 파일?)
--input             # 애매함 (무엇의 입력?)
--output            # 애매함 (무엇의 출력?)
--path              # 너무 일반적
--type              # 애매함 (무엇의 타입?)
```

**이유**:
- ✅ `--image-path`: 명확하게 이미지 파일 경로임을 알 수 있음
- ❌ `--file-path`: 스크립트 파일인지, 이미지 파일인지, 플랜 파일인지 불명확

---

#### 3. JSON 출력 스키마

**모든 명령어는 동일한 최상위 구조**:

```json
{
  "success": boolean,
  "data": object | null,
  "error": ErrorInfo | null,
  "metadata": {
    "plugin": "webauto",
    "version": "1.0.0",
    "execution_time_ms": number
  }
}
```

**TypeScript 인터페이스**:
```typescript
interface StandardResponse {
  success: boolean;
  data: object | null;
  error: ErrorInfo | null;
  metadata: Metadata;
}

interface ErrorInfo {
  code: string;              // UPPER_SNAKE_CASE
  message: string;
  details?: object;
  recovery_suggestion?: string;
}

interface Metadata {
  plugin: "webauto";
  version: string;
  execution_time_ms: number;
}
```

**성공 응답 예시** (`workflow-plan`):
```json
{
  "success": true,
  "data": {
    "plan_path": "hometax_plan.md",
    "steps_count": 8,
    "estimated_execution_time_ms": 15000,
    "planner_version": "playwright-1.56.0"
  },
  "error": null,
  "metadata": {
    "plugin": "webauto",
    "version": "1.0.0",
    "execution_time_ms": 8500
  }
}
```

**에러 응답 예시**:
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
    "recovery_suggestion": "Verify the selector is correct. Try using --wait-visible flag or increase --timeout-ms"
  },
  "metadata": {
    "plugin": "webauto",
    "version": "1.0.0",
    "execution_time_ms": 5200
  }
}
```

---

#### 4. 에러 코드

**형식**: `UPPER_SNAKE_CASE`

**공통 에러**:
```go
NODE_NOT_FOUND                    // Node.js 미설치
PLAYWRIGHT_NOT_INSTALLED          // Playwright 미설치
TIMEOUT_EXCEEDED                  // 타임아웃 초과
SESSION_NOT_FOUND                 // 존재하지 않는 세션
SESSION_LIMIT_REACHED             // 최대 세션 수 초과
```

**Agent 관련**:
```go
PLANNER_FAILED                    // 플랜 생성 실패
GENERATOR_FAILED                  // 코드 생성 실패
HEALER_FAILED                     // 자가 치유 실패
SCRIPT_EXECUTION_FAILED           // 스크립트 실행 실패
```

**Browser 관련**:
```go
BROWSER_LAUNCH_FAILED             // 브라우저 시작 실패
BROWSER_CONNECTION_LOST           // 브라우저 연결 끊김
PAGE_LOAD_FAILED                  // 페이지 로드 실패
PAGE_TIMEOUT                      // 페이지 로드 타임아웃
```

**Element 관련**:
```go
ELEMENT_NOT_FOUND                 // 요소를 찾을 수 없음
ELEMENT_NOT_VISIBLE               // 요소가 보이지 않음
ELEMENT_NOT_CLICKABLE             // 요소 클릭 불가
FORM_VALIDATION_FAILED            // 폼 유효성 검사 실패
```

**Anti-Bot 관련**:
```go
CAPTCHA_DETECTED                  // CAPTCHA 발견
BOT_DETECTION_TRIGGERED           // 봇 탐지됨
RATE_LIMIT_EXCEEDED               // Rate limit 초과
ACCESS_DENIED                     // 서버 접근 거부
```

---

#### 5. 성능 목표

| 명령어 카테고리 | 목표 시간 | 측정 방법 |
|----------------|----------|----------|
| Agent 기반 | 5-30초 | Planner/Generator/Healer 실행 시간 |
| 브라우저 제어 | < 500ms | browser-launch/close |
| 페이지 제어 | < 1000ms | page-navigate (네트워크 포함) |
| 요소 조작 | < 300ms | element-click/type |
| 데이터 추출 | < 1000ms | page-screenshot/pdf |
| 세션 관리 | < 100ms | session-list/close |

**전체 평균 목표** (Agent 제외): **< 500ms**

---

## 🏗️ 권장 패키지 구조

```
webauto/
├── cmd/
│   └── webauto/
│       └── main.go                 # 진입점
├── pkg/
│   ├── config/
│   │   └── config.go               # 환경 변수 로딩
│   ├── response/
│   │   ├── response.go             # StandardResponse 구조
│   │   └── errors.go               # ErrorInfo 구조
│   ├── cli/
│   │   ├── root.go                 # Cobra 루트 명령어
│   │   ├── workflow.go             # Agent 기반 명령어 (4개)
│   │   ├── browser.go              # 브라우저 제어 명령어 (2개)
│   │   ├── page.go                 # 페이지 제어 명령어 (1개)
│   │   ├── element.go              # 요소 조작 명령어 (2개)
│   │   ├── form.go                 # 폼 입력 명령어 (1개)
│   │   └── session.go              # 세션 관리 명령어 (2개)
│   ├── playwright/
│   │   ├── agent.go                # Playwright Agents 래퍼
│   │   ├── browser.go              # 브라우저 인스턴스 관리
│   │   ├── page.go                 # 페이지 제어
│   │   ├── element.go              # 요소 조작
│   │   ├── session.go              # 세션 관리 진입점
│   │   ├── session_worker.go       # 세션별 TCP 워커
│   │   ├── session_script.go       # Node 런너 스크립트 투영
│   │   └── runner/
│   │       └── session-server.js   # Playwright 런너 (Node.js)
│   ├── antibot/
│   │   ├── stealth.go              # Stealth mode 설정
│   │   ├── fingerprint.go          # Fingerprint 우회
│   │   └── behavior.go             # 행동 패턴 랜덤화
│   └── ipc/
│       └── node.go                 # Node.js subprocess 통신
├── internal/
│   └── utils/
│       ├── uuid.go                 # UUID 생성
│       ├── json.go                 # JSON 파싱
│       └── time.go                 # 시간 유틸리티
├── scripts/
│   └── playwright-setup.sh         # Playwright 환경 설정
├── docker/
│   └── Dockerfile                  # Docker 이미지 (Node.js 포함)
├── tests/
│   ├── unit/                       # 단위 테스트
│   └── integration/                # 통합 테스트
├── go.mod
├── go.sum
├── README.md
├── ARCHITECTURE.md
└── IMPLEMENTATION_GUIDE.md         # 이 문서
```

> 참고: 세션 런타임은 `session.go` + `session_worker.go` 조합으로 관리되며, Node 런너(`runner/session-server.js`)는 `session_script.go` 를 통해 캐시 디렉터리로 투영됩니다. 브라우저 타입과 헤드리스 설정은 내부적으로 `WEBAUTO_RUNNER_CONFIG` 환경 변수(JSON 문자열)로 전달됩니다.

---

## 🚀 구현 시작 가이드

### Step 1: 의존성 설치

#### Node.js 의존성

**package.json**:
```json
{
  "name": "oa-webauto",
  "version": "1.0.0",
  "dependencies": {
    "playwright": "^1.56.0",
    "@playwright/agents": "^1.56.0"
  }
}
```

**설치**:
```bash
npm install
npx playwright install chromium firefox webkit
```

#### Go 의존성

**go.mod**:
```go
module github.com/oa-plugins/webauto

go 1.21

require (
    github.com/spf13/cobra v1.8.1
    github.com/google/uuid v1.6.0
)
```

**설치**:
```bash
go mod tidy
```

---

### Step 2: 기본 구조 생성

#### 2.1 응답 타입 정의 (`pkg/response/response.go`)

```go
package response

import (
	"encoding/json"
	"os"
	"time"
)

type StandardResponse struct {
	Success  bool        `json:"success"`
	Data     interface{} `json:"data"`
	Error    *ErrorInfo  `json:"error"`
	Metadata Metadata    `json:"metadata"`
}

type ErrorInfo struct {
	Code               string      `json:"code"`
	Message            string      `json:"message"`
	Details            interface{} `json:"details,omitempty"`
	RecoverySuggestion string      `json:"recovery_suggestion,omitempty"`
}

type Metadata struct {
	Plugin          string `json:"plugin"`
	Version         string `json:"version"`
	ExecutionTimeMs int64  `json:"execution_time_ms"`
}

func Success(data interface{}, startTime time.Time) *StandardResponse {
	return &StandardResponse{
		Success: true,
		Data:    data,
		Error:   nil,
		Metadata: Metadata{
			Plugin:          "webauto",
			Version:         "1.0.0",
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		},
	}
}

func Error(code, message, recovery string, details interface{}, startTime time.Time) *StandardResponse {
	return &StandardResponse{
		Success: false,
		Data:    nil,
		Error: &ErrorInfo{
			Code:               code,
			Message:            message,
			Details:            details,
			RecoverySuggestion: recovery,
		},
		Metadata: Metadata{
			Plugin:          "webauto",
			Version:         "1.0.0",
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		},
	}
}

func (r *StandardResponse) Print() {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(r)
}
```

---

#### 2.2 에러 코드 정의 (`pkg/response/errors.go`)

```go
package response

// 공통 에러 코드
const (
	ErrNodeNotFound           = "NODE_NOT_FOUND"
	ErrPlaywrightNotInstalled = "PLAYWRIGHT_NOT_INSTALLED"
	ErrTimeoutExceeded        = "TIMEOUT_EXCEEDED"
	ErrSessionNotFound        = "SESSION_NOT_FOUND"
	ErrSessionLimitReached    = "SESSION_LIMIT_REACHED"
)

// Agent 관련 에러 코드
const (
	ErrPlannerFailed          = "PLANNER_FAILED"
	ErrGeneratorFailed        = "GENERATOR_FAILED"
	ErrHealerFailed           = "HEALER_FAILED"
	ErrScriptExecutionFailed  = "SCRIPT_EXECUTION_FAILED"
)

// Browser 관련 에러 코드
const (
	ErrBrowserLaunchFailed    = "BROWSER_LAUNCH_FAILED"
	ErrBrowserConnectionLost  = "BROWSER_CONNECTION_LOST"
	ErrPageLoadFailed         = "PAGE_LOAD_FAILED"
	ErrPageTimeout            = "PAGE_TIMEOUT"
)

// Element 관련 에러 코드
const (
	ErrElementNotFound        = "ELEMENT_NOT_FOUND"
	ErrElementNotVisible      = "ELEMENT_NOT_VISIBLE"
	ErrElementNotClickable    = "ELEMENT_NOT_CLICKABLE"
	ErrFormValidationFailed   = "FORM_VALIDATION_FAILED"
)

// Anti-Bot 관련 에러 코드
const (
	ErrCaptchaDetected        = "CAPTCHA_DETECTED"
	ErrBotDetectionTriggered  = "BOT_DETECTION_TRIGGERED"
	ErrRateLimitExceeded      = "RATE_LIMIT_EXCEEDED"
	ErrAccessDenied           = "ACCESS_DENIED"
)
```

---

#### 2.3 환경 설정 (`pkg/config/config.go`)

```go
package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

type Config struct {
	// Playwright
	PlaywrightNodePath    string
	PlaywrightAgentsPath  string
	PlaywrightCachePath   string

	// Browser
	DefaultBrowserType    string
	DefaultHeadless       bool
	DefaultViewportWidth  int
	DefaultViewportHeight int

	// Session
	SessionMaxCount       int
	SessionTimeoutSeconds int

	// Anti-Bot
	EnableStealth         bool
	EnableFingerprint     bool
	EnableBehaviorRandom  bool
	TypingDelayMs         int
	MouseMoveJitterPx     int
}

func Load() *Config {
	return &Config{
		PlaywrightNodePath:    getEnvOrDefault("PLAYWRIGHT_NODE_PATH", getDefaultNodePath()),
		PlaywrightAgentsPath:  getEnvOrDefault("PLAYWRIGHT_AGENTS_PATH", "@playwright/agents"),
		PlaywrightCachePath:   getEnvOrDefault("PLAYWRIGHT_CACHE_PATH", getDefaultCachePath()),

		DefaultBrowserType:    getEnvOrDefault("DEFAULT_BROWSER_TYPE", "chromium"),
		DefaultHeadless:       getEnvBoolOrDefault("DEFAULT_HEADLESS", true),
		DefaultViewportWidth:  getEnvIntOrDefault("DEFAULT_VIEWPORT_WIDTH", 1920),
		DefaultViewportHeight: getEnvIntOrDefault("DEFAULT_VIEWPORT_HEIGHT", 1080),

		SessionMaxCount:       getEnvIntOrDefault("SESSION_MAX_COUNT", 10),
		SessionTimeoutSeconds: getEnvIntOrDefault("SESSION_TIMEOUT_SECONDS", 3600),

		EnableStealth:         getEnvBoolOrDefault("ENABLE_STEALTH", true),
		EnableFingerprint:     getEnvBoolOrDefault("ENABLE_FINGERPRINT", true),
		EnableBehaviorRandom:  getEnvBoolOrDefault("ENABLE_BEHAVIOR_RANDOM", true),
		TypingDelayMs:         getEnvIntOrDefault("TYPING_DELAY_MS", 30),
		MouseMoveJitterPx:     getEnvIntOrDefault("MOUSE_MOVE_JITTER_PX", 10),
	}
}

func getDefaultNodePath() string {
	return "node"
}

func getDefaultCachePath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "oa", "webauto", "cache")
	default:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".cache", "oa", "webauto")
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "true" || value == "1" {
		return true
	}
	if value == "false" || value == "0" {
		return false
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	return defaultValue
}
```

---

#### 2.4 첫 번째 명령어 구현 예시 (`pkg/cli/browser.go`)

```go
package cli

import (
	"time"

	"github.com/oa-plugins/webauto/pkg/config"
	"github.com/oa-plugins/webauto/pkg/playwright"
	"github.com/oa-plugins/webauto/pkg/response"
	"github.com/spf13/cobra"
)

var browserLaunchCmd = &cobra.Command{
	Use:   "browser-launch",
	Short: "브라우저 인스턴스 시작",
	Long:  "브라우저 인스턴스를 시작하고 세션 ID를 반환합니다.",
	Run:   runBrowserLaunch,
}

var (
	browserType     string
	headless        bool
	sessionID       string
	viewportWidth   int
	viewportHeight  int
	userAgent       string
)

func init() {
	browserLaunchCmd.Flags().StringVar(&browserType, "browser-type", "chromium", "브라우저 타입 (chromium|firefox|webkit)")
	browserLaunchCmd.Flags().BoolVar(&headless, "headless", true, "Headless 모드")
	browserLaunchCmd.Flags().StringVar(&sessionID, "session-id", "", "세션 ID (재사용용)")
	browserLaunchCmd.Flags().IntVar(&viewportWidth, "viewport-width", 1920, "뷰포트 너비")
	browserLaunchCmd.Flags().IntVar(&viewportHeight, "viewport-height", 1080, "뷰포트 높이")
	browserLaunchCmd.Flags().StringVar(&userAgent, "user-agent", "", "User-Agent 오버라이드")
}

func runBrowserLaunch(cmd *cobra.Command, args []string) {
	startTime := time.Now()

	// Config 로드
	cfg := config.Load()

	// Session Manager 초기화
	sessionMgr := playwright.NewSessionManager(cfg)

	// 브라우저 시작
	session, err := sessionMgr.Create(cmd.Context(), browserType, headless)
	if err != nil {
		resp := response.Error(
			response.ErrBrowserLaunchFailed,
			"Failed to launch browser: "+err.Error(),
			"Check Playwright installation and browser binaries",
			map[string]interface{}{
				"browser_type": browserType,
				"headless":     headless,
			},
			startTime,
		)
		resp.Print()
		return
	}

	// 성공 응답
	resp := response.Success(map[string]interface{}{
		"session_id":   session.ID,
		"browser_type": session.BrowserType,
		"headless":     session.Headless,
		"viewport": map[string]int{
			"width":  viewportWidth,
			"height": viewportHeight,
		},
		"user_agent": userAgent,
	}, startTime)
	resp.Print()
}
```

---

### Step 3: 명령어 등록 (`cmd/webauto/main.go`)

```go
package main

import (
	"os"

	"github.com/oa-plugins/webauto/pkg/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
```

**Root Command (`pkg/cli/root.go`)**:
```go
package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "webauto",
	Short: "Playwright Agents 기반 브라우저 자동화",
	Long:  "webauto는 Playwright Agents를 활용한 지능형 브라우저 자동화 플러그인입니다.",
}

func init() {
	// Agent-Based Automation
	rootCmd.AddCommand(workflowPlanCmd)
	rootCmd.AddCommand(workflowGenerateCmd)
	rootCmd.AddCommand(workflowExecuteCmd)
	rootCmd.AddCommand(workflowHealCmd)

	// Direct Browser Control
	rootCmd.AddCommand(browserLaunchCmd)
	rootCmd.AddCommand(browserCloseCmd)
	rootCmd.AddCommand(pageNavigateCmd)
	rootCmd.AddCommand(elementClickCmd)
	rootCmd.AddCommand(elementTypeCmd)
	rootCmd.AddCommand(formFillCmd)

	// Data Extraction
	rootCmd.AddCommand(pageScreenshotCmd)
	rootCmd.AddCommand(pagePdfCmd)

	// Session Management
	rootCmd.AddCommand(sessionListCmd)
	rootCmd.AddCommand(sessionCloseCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
```

---

### Step 4: 빌드 및 테스트

#### 빌드

**단일 플랫폼**:
```bash
go build -o webauto cmd/webauto/main.go
```

**크로스 플랫폼**:
```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o webauto.exe cmd/webauto/main.go

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o webauto cmd/webauto/main.go

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o webauto cmd/webauto/main.go

# Linux
GOOS=linux GOARCH=amd64 go build -o webauto cmd/webauto/main.go
```

#### 테스트

**단위 테스트**:
```bash
go test ./... -v
```

**커버리지**:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**통합 테스트**:
```bash
# 브라우저 시작 테스트
./webauto browser-launch --headless false

# 세션 목록 확인
./webauto session-list

# 세션 종료
./webauto browser-close --session-id <session-id>
```

---

## ✅ 검증 방법

### 로컬 검증

**plugin-designer의 검증 스크립트 사용**:

```bash
# 1. plugin-designer 클론
git clone https://github.com/oa-plugins/plugin-designer.git

# 2. webauto 경로 지정하여 검증 실행
cd plugin-designer
./scripts/verify-implementation.sh webauto /path/to/webauto
```

---

### 수동 검증 체크리스트

#### ✅ 명령어 확인
```bash
./webauto --help
# 14개 명령어가 모두 표시되는가?
```

**기대 출력**:
```
Available Commands:
  browser-close       브라우저 인스턴스 종료
  browser-launch      브라우저 인스턴스 시작
  element-click       페이지 요소 클릭
  element-type        요소에 텍스트 입력
  form-fill           폼 자동 입력
  page-navigate       특정 URL로 페이지 이동
  page-pdf            현재 페이지 PDF 저장
  page-screenshot     현재 페이지 스크린샷 촬영
  session-close       특정 세션 종료
  session-list        현재 활성 세션 목록 조회
  workflow-execute    생성된 Playwright 스크립트 실행
  workflow-generate   Generator Agent로 Markdown 플랜을 실행 가능한 코드로 변환
  workflow-heal       Healer Agent로 실패한 스크립트 자동 수리
  workflow-plan       Planner Agent로 웹사이트 탐색 및 테스트 플랜 생성
```

---

#### ✅ 플래그 확인
```bash
./webauto browser-launch --help
# 올바른 플래그가 정의되어 있는가?
```

**기대 출력**:
```
Flags:
      --browser-type string     브라우저 타입 (chromium|firefox|webkit) (default "chromium")
      --headless                Headless 모드 (default true)
      --session-id string       세션 ID (재사용용)
      --viewport-width int      뷰포트 너비 (default 1920)
      --viewport-height int     뷰포트 높이 (default 1080)
      --user-agent string       User-Agent 오버라이드
```

**검증 포인트**:
- ✅ 도메인 특화 플래그 사용 (`--browser-type`, `--session-id`, `--viewport-width`)
- ❌ 일반적 플래그 없음 (`--file-path`, `--input`, `--output`)

---

#### ✅ JSON 출력 확인
```bash
./webauto browser-launch | jq .
```

**기대 출력**:
```json
{
  "success": true,
  "data": {
    "session_id": "uuid-here",
    "browser_type": "chromium",
    "headless": true,
    "viewport": {
      "width": 1920,
      "height": 1080
    }
  },
  "error": null,
  "metadata": {
    "plugin": "webauto",
    "version": "1.0.0",
    "execution_time_ms": 1500
  }
}
```

**검증 포인트**:
- ✅ `success` 필드 존재
- ✅ `data` 필드 존재 및 구조화
- ✅ `error` 필드 null
- ✅ `metadata` 필드 존재 (plugin, version, execution_time_ms)

---

#### ✅ 에러 응답 확인
```bash
./webauto browser-close --session-id invalid-id | jq .
```

**기대 출력**:
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "SESSION_NOT_FOUND",
    "message": "Session with ID invalid-id not found",
    "details": {
      "session_id": "invalid-id"
    },
    "recovery_suggestion": "Use 'oa webauto session-list' to see active sessions"
  },
  "metadata": {
    "plugin": "webauto",
    "version": "1.0.0",
    "execution_time_ms": 50
  }
}
```

**검증 포인트**:
- ✅ `success: false`
- ✅ `data: null`
- ✅ `error.code` UPPER_SNAKE_CASE 형식
- ✅ `error.recovery_suggestion` 존재

---

#### ✅ 성능 확인
```bash
time ./webauto session-list
```

**기대 출력**:
```
real    0m0.050s
user    0m0.020s
sys     0m0.015s
```

**검증 포인트**:
- ✅ 세션 관리 명령어: < 100ms
- ✅ 브라우저 제어: < 500ms
- ✅ 요소 조작: < 300ms

---

## 📚 참고 코드

### Playwright Agents 통합 예시 (`pkg/playwright/agent.go`)

```go
package playwright

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/oa-plugins/webauto/pkg/config"
)

type AgentManager struct {
	cfg *config.Config
}

func NewAgentManager(cfg *config.Config) *AgentManager {
	return &AgentManager{cfg: cfg}
}

// Planner Agent: 웹사이트 탐색 및 테스트 플랜 생성
func (am *AgentManager) Plan(ctx context.Context, url, scenario, outputPath string) (PlanResult, error) {
	cmd := exec.CommandContext(ctx, am.cfg.PlaywrightNodePath,
		"-e", fmt.Sprintf(`
			const { planner } = require('%s');
			(async () => {
				const result = await planner.explore('%s', { scenario: '%s' });
				await planner.savePlan(result, '%s');
				console.log(JSON.stringify({ steps: result.steps.length }));
			})();
		`, am.cfg.PlaywrightAgentsPath, url, scenario, outputPath))

	output, err := cmd.Output()
	if err != nil {
		return PlanResult{}, fmt.Errorf("planner failed: %w", err)
	}

	var result PlanResult
	if err := json.Unmarshal(output, &result); err != nil {
		return PlanResult{}, fmt.Errorf("parse planner result: %w", err)
	}

	return result, nil
}

type PlanResult struct {
	StepsCount int `json:"steps"`
}
```

---

### Anti-Bot 우회 예시 (`pkg/antibot/behavior.go`)

```go
package antibot

import (
	"math/rand"
	"time"
)

// 타이핑 지연 (10-50ms 랜덤)
func GetTypingDelay() time.Duration {
	return time.Duration(10+rand.Intn(40)) * time.Millisecond
}

// 마우스 이동 Jitter (±5-15px)
func AddMouseJitter(x, y int) (int, int) {
	jitter := 5 + rand.Intn(10)
	return x + (rand.Intn(2*jitter) - jitter), y + (rand.Intn(2*jitter) - jitter)
}

// User-Agent 로테이션
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
}

func GetRandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}
```

---

## 🔄 개발 프로세스

### 1. 기능 구현 순서 (추천)

**Phase 1: MVP (핵심 5개 명령어)**
```bash
✅ 목표: 브라우저 제어 POC
1. browser-launch    # 브라우저 시작
2. browser-close     # 브라우저 종료
3. page-navigate     # URL 이동
4. element-click     # 요소 클릭
5. page-screenshot   # 스크린샷 촬영

완료 기준:
- [ ] Linux에서 브라우저 제어 동작
- [ ] 홈택스 로그인 페이지 이동 성공
- [ ] 스크린샷 촬영 성공
- [ ] 단위 테스트 커버리지 > 60%
```

**Phase 2: Playwright Agents 통합 (8개 명령어 추가)**
```bash
✅ 목표: Agent 기반 자동화 완성
6. workflow-plan      # Planner Agent 통합
7. workflow-generate  # Generator Agent 통합
8. workflow-execute   # 스크립트 실행
9. workflow-heal      # Healer Agent 통합
10. element-type      # 텍스트 입력
11. form-fill         # 폼 자동 입력
12. session-list      # 세션 목록 조회
13. session-close     # 세션 종료

완료 기준:
- [ ] 홈택스 자동화 스크립트 자동 생성
- [ ] Agent로 생성된 코드 실행 성공
- [ ] Healer로 실패한 스크립트 자동 수리 성공
- [ ] 단위 테스트 커버리지 > 75%
```

**Phase 3: 크로스 플랫폼 및 Anti-Bot 강화 (1개 명령어 추가)**
```bash
✅ 목표: 프로덕션 레디
14. page-pdf          # PDF 저장

기술 구현:
- [ ] Windows/macOS 플랫폼 지원
- [ ] Anti-Bot 강화 (Fingerprint, 행동 패턴 랜덤화)
- [ ] Docker 이미지 (Node.js 포함)
- [ ] 성능 최적화 (< 500ms 응답 목표)

완료 기준:
- [ ] Windows/macOS/Linux 모두 동작
- [ ] 홈택스/위하고 봇 탐지 우회 성공
- [ ] 평균 응답 시간 < 500ms (Agent 제외)
- [ ] 단위 테스트 커버리지 > 80%
- [ ] E2E 테스트 커버리지 > 85%
```

---

### 2. 브랜치 전략

```bash
# 기능별 브랜치
git checkout -b feat/browser-launch
git checkout -b feat/workflow-plan

# 버그 수정
git checkout -b fix/session-leak
git checkout -b fix/element-selector

# 문서 업데이트
git checkout -b docs/api-reference
git checkout -b docs/troubleshooting
```

---

### 3. PR 및 리뷰

**PR 체크리스트**:
- [ ] ✅ 설계 명세 준수 확인 (ARCHITECTURE.md)
- [ ] ✅ 명령어 이름: `<resource>-<action>` 패턴
- [ ] ✅ 플래그 이름: `--<domain-noun>-<attribute>` 패턴
- [ ] ✅ JSON 출력: success/data/error/metadata 구조
- [ ] ✅ 에러 코드: UPPER_SNAKE_CASE
- [ ] ✅ 성능 목표 달성
- [ ] ✅ 단위 테스트 작성 (커버리지 > 60%)
- [ ] ✅ 로컬 검증 스크립트 통과
- [ ] ✅ JSON 출력 수동 확인
- [ ] ✅ 에러 케이스 테스트
- [ ] ✅ README 업데이트 (필요 시)

**PR 템플릿**:
```markdown
## 변경 사항

- 구현한 명령어: `workflow-plan`
- 추가한 플래그: `--page-url`, `--scenario-text`, `--output-path`

## 테스트

- [x] 단위 테스트 작성 (커버리지: 75%)
- [x] 로컬 검증 스크립트 통과
- [x] JSON 출력 확인
- [x] 에러 케이스 테스트

## 체크리스트

- [x] 설계 명세 준수
- [x] 명령어/플래그 명명 규칙 준수
- [x] JSON 출력 표준 준수
- [x] 에러 코드 표준 준수
- [x] 성능 목표 달성 (< 30초)
```

---

### 4. 릴리스

```bash
# 1. 버전 태그 생성
git tag v1.0.0
git push origin v1.0.0

# 2. GitHub Release 생성
gh release create v1.0.0 \
  --title "webauto v1.0.0" \
  --notes "Initial release with 14 commands"

# 3. plugin-designer 자동 업데이트
# → sync-plugins.yml 워크플로우가 감지
# → registry.json 자동 업데이트
```

---

## 💬 질문 및 지원

### 설계 관련 질문

**설계가 불명확하거나 변경이 필요한 경우**:
- [plugin-designer Issues](https://github.com/oa-plugins/plugin-designer/issues)에 질문 작성
- 제목: `[webauto] 설계 질문: ...`
- 라벨: `question`, `webauto`

**예시**:
```
제목: [webauto] workflow-heal의 max-attempts 기본값은?
내용:
ARCHITECTURE.md에 max-attempts의 default 값이 명시되지 않았습니다.
3으로 설정해도 될까요?
```

---

### 구현 관련 버그

**webauto 구현 중 발생한 버그**:
- [webauto Issues](https://github.com/oa-plugins/webauto/issues)에 버그 리포트

**템플릿**:
```markdown
## 버그 설명
browser-launch 실행 시 Node.js를 찾지 못합니다.

## 재현 방법
1. `./webauto browser-launch`
2. 에러 발생

## 기대 동작
브라우저가 시작되어야 합니다.

## 환경
- OS: macOS 14.0
- Go: 1.21
- Node.js: 18.0.0
```

---

### 도움 요청

- [plugin-designer Discussions](https://github.com/oa-plugins/plugin-designer/discussions)

---

## 📖 추가 자료

### 필수 문서
- **[webauto-architecture.md](https://github.com/oa-plugins/plugin-designer/blob/main/designs/webauto-architecture.md)** - 가장 중요!
- [webauto-idea.md](https://github.com/oa-plugins/plugin-designer/blob/main/ideas/webauto-idea.md)
- [OA Plugin Standards](https://github.com/oa-plugins/plugin-designer/blob/main/PRD.md#5-플러그인-간-조율-프로세스)

### 참고 구현
- [plugin-template](https://github.com/oa-plugins/plugin-template) - 플러그인 보일러플레이트
- [Playwright Docs](https://playwright.dev/)
- [Playwright Agents](https://playwright.dev/docs/test-agents)

---

## ✨ 시작하세요!

```bash
# 1. 설계 문서 읽기
cat ARCHITECTURE.md
# → 14개 명령어 숙지
# → 플래그 명명 규칙 확인
# → JSON 출력 스키마 이해

# 2. 의존성 설치
npm install
npx playwright install chromium
go mod tidy

# 3. 첫 번째 명령어 구현 (browser-launch)
# → pkg/cli/browser.go 작성
# → pkg/playwright/session.go 작성

# 4. 빌드 및 테스트
go build -o webauto cmd/webauto/main.go
./webauto browser-launch --headless false

# 5. 로컬 검증
./webauto browser-launch | jq .
# → success: true 확인
# → JSON 구조 확인

# 6. PR 생성 및 리뷰
git checkout -b feat/browser-launch
git commit -m "feat: implement browser-launch command"
git push origin feat/browser-launch
gh pr create
```

**Good luck! 🚀**

---

## 📝 체크리스트 요약

### 명령어 구현 체크리스트
- [ ] 명령어 이름: `<resource>-<action>` 패턴 준수
- [ ] 플래그 이름: `--<domain-noun>-<attribute>` 패턴 준수 (도메인 특화 명사 사용!)
- [ ] JSON 출력: success/data/error/metadata 구조
- [ ] 에러 코드: UPPER_SNAKE_CASE
- [ ] 에러 메시지: 명확하고 구체적 + recovery_suggestion 포함
- [ ] 성능 목표 달성
- [ ] 단위 테스트 작성 (커버리지 > 60%)

### 릴리스 체크리스트
- [ ] Phase 1 완료 (5개 명령어)
- [ ] Phase 2 완료 (13개 명령어)
- [ ] Phase 3 완료 (14개 명령어)
- [ ] 단위 테스트 커버리지 > 80%
- [ ] E2E 테스트 커버리지 > 85%
- [ ] README 완성
- [ ] 라이선스 파일 추가
- [ ] GitHub Release 생성
- [ ] registry.json 업데이트 확인

---

**문서 버전**: 1.0.0
**작성일**: 2025-10-13
**다음 업데이트**: 구현 진행에 따라 수정
