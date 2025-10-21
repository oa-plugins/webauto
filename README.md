# OA webauto Plugin

Playwright Agents를 활용한 지능형 브라우저 자동화 플러그인으로, 한국 세무/회계 서비스(홈택스, 위하고 등)의 복잡한 UI 자동화를 지원합니다.

## 🎯 핵심 기능

### Agent-Based Automation (고수준 자동화)
- **workflow-plan**: Planner Agent로 웹사이트 탐색 및 테스트 플랜 생성
- **workflow-generate**: Generator Agent로 플랜을 Playwright 코드로 변환
- **workflow-execute**: 생성된 자동화 스크립트 실행
- **workflow-heal**: Healer Agent로 실패한 스크립트 자동 수리

### Direct Browser Control (저수준 제어)
- **browser-launch**: 브라우저 시작 및 세션 생성
- **browser-close**: 브라우저 종료
- **page-navigate**: URL 이동
- **element-click**: 요소 클릭
- **element-type**: 텍스트 입력
- **form-fill**: 폼 자동 입력

### Data Extraction
- **page-screenshot**: 스크린샷 촬영
- **page-pdf**: PDF 저장

### Session Management
- **session-list**: 활성 세션 목록
- **session-close**: 세션 종료

**총 14개 명령어**

## 🚀 빠른 시작

### 사전 요구사항

1. **Go 1.22+**
2. **Node.js 18+** (Playwright 실행용)
3. **OA CLI** (플러그인 호스트)

### 설치

```bash
# 1. Playwright 및 브라우저 설치
npm install playwright @playwright/agents
npx playwright install chromium firefox webkit

# 2. webauto 플러그인 빌드
go build -o webauto cmd/webauto/main.go

# 3. OA CLI에 등록
oa plugin install ./webauto
```

## 📖 사용 예시

### Agent 기반 자동화 (권장)

```bash
# 1. 시나리오로부터 플랜 생성
oa webauto workflow-plan \
  --page-url "https://hometax.go.kr" \
  --scenario-text "로그인 → 세금계산서 조회 → CSV 다운로드" \
  --output-path hometax_plan.md

# 2. 플랜을 실행 가능한 코드로 변환
oa webauto workflow-generate \
  --plan-file hometax_plan.md \
  --output-path hometax_automation.ts

# 3. 자동화 실행
oa webauto workflow-execute \
  --script-file hometax_automation.ts \
  --headless false

# 4. 실패 시 자동 수리
oa webauto workflow-heal \
  --script-file hometax_automation.ts \
  --max-attempts 5
```

### Direct Control (수동 제어)

```bash
# 1. 브라우저 시작 (자동 세션 ID)
oa plugin exec webauto browser-launch --no-headless
# 출력: {"success":true,"data":{"session_id":"ses_abc123",...}}

# 또는 커스텀 세션 ID 지정
oa plugin exec webauto browser-launch \
  --session-id "my_session" \
  --no-headless
# 출력: {"success":true,"data":{"session_id":"my_session",...}}

# 2. 페이지 이동
oa plugin exec webauto page-navigate \
  --session-id "my_session" \
  --page-url "https://hometax.go.kr"

# 3. 폼 입력
oa plugin exec webauto form-fill \
  --session-id "my_session" \
  --form-data '{"username":"user1","password":"pass123"}' \
  --submit true

# 4. 스크린샷 촬영
oa plugin exec webauto page-screenshot \
  --session-id "my_session" \
  --image-path hometax_result.png

# 5. 브라우저 종료
oa plugin exec webauto browser-close --session-id "my_session"
```

## 🛡️ Anti-Bot 전략

webauto는 다음 기술로 봇 탐지를 우회합니다:

1. **Playwright Stealth Mode**: WebDriver 플래그 자동 숨김
2. **Fingerprint 우회**: User-Agent 로테이션
3. **행동 패턴 랜덤화**: 타이핑 지연, 마우스 이동 Jitter
4. **Rate Limiting**: 요청 간격 제어

### 환경 변수 설정

```bash
export ENABLE_STEALTH=true
export ENABLE_FINGERPRINT=true
export ENABLE_BEHAVIOR_RANDOM=true
export TYPING_DELAY_MS=30
export MOUSE_MOVE_JITTER_PX=10
```

## 🌍 플랫폼 지원

- ✅ **Windows** 10/11 (amd64)
- ✅ **macOS** 11+ (Intel/Apple Silicon)
- ✅ **Linux** Ubuntu 20.04+ (amd64, arm64)

**상세 설치 가이드**: [Platform Guide](docs/platform-guide.md)

## 📊 성능 목표

| 명령어 카테고리 | 목표 시간 |
|----------------|----------|
| Agent 기반 | 5-30초 |
| 브라우저 제어 | < 500ms |
| 페이지 제어 | < 1000ms |
| 요소 조작 | < 300ms |
| 데이터 추출 | < 1000ms |
| 세션 관리 | < 100ms |

## 🧪 테스트

```bash
# 단위 테스트
go test ./...

# 커버리지
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📝 OAS Scripting (.oas 스크립트 지원)

webauto 플러그인은 **Office Automation Script (.oas)** 포맷을 지원하여 Shell 스크립트보다 **45-69% 적은 코드**로 자동화를 구현할 수 있습니다.

### .oas vs Shell Script 비교

| 특징 | Shell Script | .oas Script | 개선도 |
|------|-------------|-------------|--------|
| 코드 라인 수 | 58-259줄 | 30-80줄 | **45-69% 감소** |
| 외부 의존성 | bash, jq, grep | oa CLI만 | **1개만 필요** |
| JSON 파싱 | jq 수동 파싱 | 내장 지원 | **자동화** |
| 에러 처리 | 수동 체크 | @try/@catch | **안전성 향상** |
| 가독성 | 중간 | 높음 | **유지보수 쉬움** |

### 빠른 예시

**Shell Script (58줄):**
```bash
#!/bin/bash
set -e
WEBAUTO="../../webauto"
RESULT=$($WEBAUTO browser-launch --headless true)
SESSION_ID=$(echo "$RESULT" | jq -r '.data.session_id')
if [ -z "$SESSION_ID" ]; then exit 1; fi
# ... 50+ more lines
```

**.oas Script (30줄):**
```bash
# web_scraping.oas
@set SESSION_ID = "web_session"

oa plugin exec webauto browser-launch \
  --session-id "${SESSION_ID}" \
  --no-headless

oa plugin exec webauto page-navigate \
  --session-id "${SESSION_ID}" \
  --page-url "https://example.com"

oa plugin exec webauto page-screenshot \
  --session-id "${SESSION_ID}" \
  --image-path "output.png"

oa plugin exec webauto browser-close --session-id "${SESSION_ID}"
```

### 실행 방법

```bash
# .oas 스크립트 실행
oa batch run examples/basic/manual_test.oas

# Dry-run (실행하지 않고 확인만)
oa batch run examples/basic/manual_test.oas --dry-run

# Verbose 모드 (상세 로그)
oa batch run examples/basic/manual_test.oas --verbose

# 변수 오버라이드
oa batch run examples/advanced/retry.oas --set TARGET_URL="https://example.com"
```

### 제공 예제 (18개)

**기본 (examples/basic/):**
- **manual_test.oas**: 브라우저 수동 테스트 헬퍼
- **screenshot.oas**: 스크린샷 캡처
- **element_interaction.oas**: 요소 클릭 및 입력
- **form_fill.oas**: 폼 자동 입력
- **session_list.oas**: 세션 목록 조회

**고급 (examples/advanced/):**
- **retry.oas**: 재시도 로직 구현
- **wait_conditions.oas**: 다양한 대기 조건
- **multi_element.oas**: 여러 요소 조작
- **conditional.oas**: 조건부 실행
- **batch_operations.oas**: 일괄 작업

**실사용 (examples/real-world/):**
- **naver_blog_search.oas**: 네이버 블로그 검색
- **naver_map_search.oas**: 네이버 지도 검색
- **hometax_screenshot.oas**: 홈택스 스크린샷

**상세 예제는 [examples/README.md](examples/README.md) 참조**

### 상세 문서

- **[OAS Scripting Guide](docs/oas-scripting-guide.md)**: 전체 .oas 문법 및 고급 예제
- **[Migration Guide](docs/oas-migration-guide.md)**: Shell Script → .oas 변환 가이드

## 📚 기타 문서

- [플랫폼별 설치 가이드](docs/platform-guide.md)
- [아키텍처 설계](ARCHITECTURE.md)
- [구현 가이드](docs/implementation-guide.md)
- [성능 가이드](docs/performance-guide.md)
- [아이디어 제안서](https://github.com/oa-plugins/plugin-designer/blob/main/ideas/webauto-idea.md)
- [API 문서](https://github.com/oa-plugins/plugin-designer/blob/main/designs/webauto-architecture.md)

## ⚖️ 법적 고지

**개인 정보 자동화 전용**: 이 플러그인은 본인의 세금/회계 정보를 자동화하기 위한 목적으로만 사용하세요.

**금지 사항**:
- ❌ 타인의 계정 무단 접근
- ❌ 서비스 약관 위반
- ❌ 상업적 스크래핑
- ❌ 과도한 요청 (Rate Limit 초과)

**책임**: 사용자는 이 플러그인 사용으로 인한 법적 책임을 스스로 부담합니다.

## 🤝 기여

Pull Request 환영합니다! 기여 전 [CONTRIBUTING.md](CONTRIBUTING.md)를 확인하세요.

## 📄 라이선스

MIT License

## 🔗 관련 링크

- [OA CLI](https://github.com/oa-plugins/oa)
- [Plugin Designer](https://github.com/oa-plugins/plugin-designer)
- [Playwright Docs](https://playwright.dev/)
- [Playwright Agents](https://playwright.dev/docs/test-agents)

---

**버전**: 1.0.0
**작성**: 2025-10-13
**문의**: [GitHub Issues](https://github.com/oa-plugins/webauto/issues)
