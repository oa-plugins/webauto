# OA CLI Batch Parser 개선 제안서

**작성일:** 2025-10-21
**작성자:** webauto 플러그인 팀
**대상:** OA CLI 개발팀
**목적:** .oas 스크립트 파서 기능 개선 요청

---

## 요약

webauto 플러그인의 .oas 예제 스크립트 검증 결과, **저장소 내 11개 예제 중 5개(45.5%)가 `@while` 미지원으로 실패**했으며, **사내 확장 QA 세트 7개가 다중 라인 배열, 복합 조건식, anti-bot 이슈로 실패**했습니다.

**현재 상황:**
- ✅ 저장소 내 예제: 6/11 통과 (54.5%)
- ❌ `@while` 미지원: 5개 실패 (45.5%)
- ❌ 확장 QA 세트: 7개 중 대부분 실패

**개선 시 기대효과:**
- `@while` 구현: 5개 예제 즉시 통과 → 11/11 (100%)
- 다중 라인 배열 지원: 확장 세트 2개 추가 통과
- 조건식 개선: 확장 세트 1개 추가 통과

---

## 제안 1: `@while` 루프 구현 (최우선)

### 현재 상황 ❌

**저장소 내 5개 예제 모두 실패:**
1. `examples/while_batch_processing.oas`
2. `examples/test_while_simple.oas`
3. `examples/test_while_condition.oas`
4. `examples/while_counter.oas`
5. `examples/while_file_polling.oas`

**확장 QA 세트 1개 실패:**
- `oas-scripts/advanced_form_automation.oas`

**문법:**
```bash
@set ATTEMPT = 0
@set MAX_RETRIES = 3
@set SUCCESS = false

@while ${ATTEMPT} < ${MAX_RETRIES} and not ${SUCCESS}
  @set ATTEMPT = ${ATTEMPT} + 1
  @echo "Attempt ${ATTEMPT} of ${MAX_RETRIES}..."

  @try
    oa plugin exec webauto form-fill --session-id "${SESSION_ID}" --form-data "${FORM_DATA}"
    @set SUCCESS = true
  @catch
    @echo "Attempt ${ATTEMPT} failed"
    @sleep 2000
  @endtry
@endwhile
```

**현재 동작:**
```
Warning: Unknown directive (skipped): @while ${ATTEMPT} < ${MAX_RETRIES} and not ${SUCCESS}
Warning: Unknown directive (skipped): @endwhile
```

**결과:** 루프 블록 전체가 무시됨, 인덱스 증가 로직 미실행

### 예상 동작 ✅

조건이 참인 동안 반복 실행되어야 합니다.

**동작 방식:**
1. 조건식 평가 (`${ATTEMPT} < ${MAX_RETRIES} and not ${SUCCESS}`)
2. 참이면 블록 내부 실행
3. `@endwhile` 도달 시 조건 재평가
4. 거짓이면 루프 종료

### 주요 사용 시나리오

#### 시나리오 1: 재시도 로직 (가장 흔함)
```bash
@set RETRY_COUNT = 0
@set MAX_RETRIES = 5
@set SUCCESS = false

@while ${RETRY_COUNT} < ${MAX_RETRIES} and not ${SUCCESS}
  @try
    oa plugin exec webauto element-click --session-id "${SID}" --element-selector "#submit"
    @set SUCCESS = true
  @catch
    @set RETRY_COUNT = ${RETRY_COUNT} + 1
    @sleep 1000
  @endtry
@endwhile
```

#### 시나리오 2: 파일 폴링/대기
```bash
@set FILE_FOUND = false
@set WAIT_TIME = 0
@set MAX_WAIT = 30000  # 30초

@while not ${FILE_FOUND} and ${WAIT_TIME} < ${MAX_WAIT}
  @try
    @if exists("${TARGET_FILE}")
      @set FILE_FOUND = true
    @endif
  @catch
    @sleep 1000
    @set WAIT_TIME = ${WAIT_TIME} + 1000
  @endtry
@endwhile
```

#### 시나리오 3: 페이지네이션
```bash
@set HAS_NEXT_PAGE = true
@set PAGE_NUMBER = 1

@while ${HAS_NEXT_PAGE}
  @echo "Processing page ${PAGE_NUMBER}..."

  # 페이지 데이터 추출
  oa plugin exec webauto element-query-all \
    --session-id "${SID}" \
    --element-selector ".item"

  # 다음 페이지 이동
  @try
    oa plugin exec webauto element-click \
      --session-id "${SID}" \
      --element-selector ".next-page"
    @set PAGE_NUMBER = ${PAGE_NUMBER} + 1
  @catch
    @set HAS_NEXT_PAGE = false
  @endtry
@endwhile
```

#### 시나리오 4: 카운터 기반 반복
```bash
@set COUNTER = 0

@while ${COUNTER} < 10
  @echo "Iteration ${COUNTER}"
  @set COUNTER = ${COUNTER} + 1
@endwhile
```

### 현재 회피 방법의 한계

**방법 1: `@foreach` + 고정 범위**
```bash
@foreach i in [1, 2, 3]
  @if not ${SUCCESS}
    @set ATTEMPT = ${ATTEMPT} + 1
    @echo "Attempt ${ATTEMPT} of 3..."

    @try
      oa plugin exec webauto form-fill --session-id "${SESSION_ID}" --form-data "${FORM_DATA}"
      @set SUCCESS = true
    @catch
      @echo "Attempt ${ATTEMPT} failed"
      @sleep 2000
    @endtry
  @endif
@endforeach
```

**문제점:**
- ❌ 최대 반복 횟수를 하드코딩해야 함
- ❌ 조기 종료 시에도 모든 반복 실행 (성능 낭비)
- ❌ 동적 반복 횟수 불가능
- ❌ 가독성 저하 (의도가 명확하지 않음)

**방법 2: 재귀 스크립트 호출**
```bash
# retry_logic.oas
@if ${ATTEMPT} < ${MAX_RETRIES} and not ${SUCCESS}
  @set ATTEMPT = ${ATTEMPT} + 1
  # ... retry logic ...
  oa batch run retry_logic.oas  # 재귀 호출
@endif
```

**문제점:**
- ❌ 파일 분할 필요 (단일 스크립트로 불가능)
- ❌ 변수 전달 복잡 (환경변수 또는 파일 공유)
- ❌ 스택 오버플로 위험
- ❌ 디버깅 매우 어려움

### 기술적 구현 제안

**파일:** `pkg/batch/parser.go`, `pkg/batch/executor.go`

**1. 파서 단계 (parser.go):**
```go
type WhileStatement struct {
    Condition string
    Body      []Statement
    LineStart int
    LineEnd   int
}

func (p *Parser) parseWhileBlock(line string) (*WhileStatement, error) {
    // Extract condition from "@while <condition>"
    condition := strings.TrimPrefix(line, "@while")
    condition = strings.TrimSpace(condition)

    // Collect body until @endwhile
    body := []Statement{}
    depth := 1  // Track nested @while blocks

    for p.scanner.Scan() {
        line := p.scanner.Text()
        trimmed := strings.TrimSpace(line)

        if strings.HasPrefix(trimmed, "@while") {
            depth++
        } else if strings.HasPrefix(trimmed, "@endwhile") {
            depth--
            if depth == 0 {
                break
            }
        }

        stmt, err := p.parseStatement(line)
        if err != nil {
            return nil, err
        }
        body = append(body, stmt)
    }

    return &WhileStatement{
        Condition: condition,
        Body:      body,
    }, nil
}
```

**2. 실행 단계 (executor.go):**
```go
func (e *Executor) executeWhileLoop(whileStmt *WhileStatement) error {
    maxIterations := 10000 // Safety limit to prevent infinite loops
    iteration := 0

    for {
        // Check iteration limit
        if iteration >= maxIterations {
            return fmt.Errorf("@while loop exceeded max iterations (%d). Possible infinite loop?", maxIterations)
        }

        // Evaluate condition
        conditionResult, err := e.evaluateCondition(whileStmt.Condition)
        if err != nil {
            return fmt.Errorf("failed to evaluate @while condition: %w", err)
        }

        if !conditionResult {
            break // Exit loop
        }

        // Execute loop body
        for _, stmt := range whileStmt.Body {
            err := e.executeStatement(stmt)
            if err != nil {
                return err
            }
        }

        iteration++

        // Optional: Set built-in variable for debugging
        e.setVariable("__WHILE_ITERATION__", iteration)
    }

    return nil
}
```

**3. 조건 평가 (기존 `@if` 로직 재사용):**
```go
func (e *Executor) evaluateCondition(condition string) (bool, error) {
    // Substitute variables
    condition = e.substituteVariables(condition)

    // Parse and evaluate
    // (기존 @if 조건 평가 로직 재사용)
    return e.evaluateBooleanExpression(condition)
}
```

### 무한 루프 방지 전략

**1. 최대 반복 횟수 제한:**
```go
maxIterations := 10000  // 기본값
if configMaxIter := os.Getenv("OA_MAX_WHILE_ITERATIONS"); configMaxIter != "" {
    maxIterations = parseInt(configMaxIter)
}
```

**2. 타임아웃 설정:**
```go
startTime := time.Now()
maxDuration := 5 * time.Minute  // 5분 제한

for {
    if time.Since(startTime) > maxDuration {
        return fmt.Errorf("@while loop timeout after %v", maxDuration)
    }
    // ...
}
```

**3. 명시적 break 지시어 (선택 사항):**
```bash
@while true
  @if ${CONDITION_MET}
    @break
  @endif
@endwhile
```

### 영향 받는 예제 및 기대효과

**즉시 통과 예정:**
1. ✅ `examples/while_batch_processing.oas`
2. ✅ `examples/test_while_simple.oas`
3. ✅ `examples/test_while_condition.oas`
4. ✅ `examples/while_counter.oas`
5. ✅ `examples/while_file_polling.oas`
6. ✅ `oas-scripts/advanced_form_automation.oas` (확장 QA)

**성공률 변화:**
- 현재: 6/11 (54.5%)
- `@while` 구현 후: 11/11 (100%) ⬆️ +45.5%

### 우선순위

**최우선** ⚡⚡⚡

**이유:**
- 저장소 내 5개 예제가 모두 이 기능 대기 중
- 재시도 로직은 실전 자동화의 핵심 패턴
- 현재 회피 방법이 매우 불편하고 제한적
- 다른 스크립트 언어(Bash, Python, PowerShell)의 기본 기능
- 사용자 요구가 가장 많은 기능

---

## 제안 2: 다중 라인 배열 정의 지원

### 현재 상황 ❌

**확장 QA 세트 2개 실패:**
- `advanced/data_extraction_pipeline.oas`
- `advanced/parallel_session_management.oas`

**문법:**
```bash
@set TARGET_URLS = [
  "https://en.wikipedia.org/wiki/Web_scraping",
  "https://en.wikipedia.org/wiki/Browser_automation",
  "https://en.wikipedia.org/wiki/Playwright_(software)"
]
```

**에러:**
```
Line 5: https://en.wikipedia.org/wiki/Web_scraping,
Error: command not found: https://en.wikipedia.org/wiki/Web_scraping,
```

**원인:** 파서가 각 라인을 별도 명령으로 해석

### 예상 동작 ✅

배열 정의가 여러 라인에 걸쳐 있어도 하나의 변수 할당으로 인식되어야 합니다.

**동작 방식:**
1. `[` 토큰 발견 시 배열 파싱 모드 진입
2. `]` 토큰까지 모든 라인을 하나의 값으로 수집
3. 공백 정규화 및 변수 할당 완료

### 현재 회피 방법

**단일 라인으로 강제 변환:**
```bash
@set TARGET_URLS = ["https://en.wikipedia.org/wiki/Web_scraping", "https://en.wikipedia.org/wiki/Browser_automation", "https://en.wikipedia.org/wiki/Playwright_(software)"]
```

**문제점:**
- ❌ 가독성 저하 (120+ 문자 한 줄)
- ❌ 유지보수 어려움 (항목 추가/제거 시 실수 유발)
- ❌ 배열 요소 파악 어려움
- ❌ Linter/Formatter 사용 불가
- ❌ Git diff 가독성 악화

**예시 (실제 사례):**
```bash
# 나쁜 예 (현재 강제되는 방식)
@set SITES = ["https://www.naver.com", "https://www.daum.net", "https://www.google.com", "https://www.github.com", "https://stackoverflow.com", "https://www.reddit.com"]

# 좋은 예 (구현 필요)
@set SITES = [
  "https://www.naver.com",
  "https://www.daum.net",
  "https://www.google.com",
  "https://www.github.com",
  "https://stackoverflow.com",
  "https://www.reddit.com"
]
```

### 기술적 구현 제안

**파일:** `pkg/batch/parser.go`

```go
func (p *Parser) parseSetDirective(line string) (*SetDirective, error) {
    // Extract variable name and value
    parts := strings.SplitN(line, "=", 2)
    if len(parts) != 2 {
        return nil, fmt.Errorf("invalid @set syntax")
    }

    varName := strings.TrimSpace(strings.TrimPrefix(parts[0], "@set"))
    value := strings.TrimSpace(parts[1])

    // Check for multi-line array
    if strings.HasPrefix(value, "[") && !strings.HasSuffix(value, "]") {
        return p.parseMultiLineArray(varName, value)
    }

    // Single-line processing
    return &SetDirective{
        Variable: varName,
        Value:    value,
    }, nil
}

func (p *Parser) parseMultiLineArray(varName, firstLine string) (*SetDirective, error) {
    arrayLines := []string{firstLine}

    // Collect lines until closing bracket
    for p.scanner.Scan() {
        line := strings.TrimSpace(p.scanner.Text())

        // Skip empty lines and comments
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }

        arrayLines = append(arrayLines, line)

        if strings.Contains(line, "]") {
            break
        }

        // Safety: prevent infinite loop
        if len(arrayLines) > 1000 {
            return nil, fmt.Errorf("array definition too long (>1000 lines)")
        }
    }

    // Join and normalize
    fullArray := strings.Join(arrayLines, " ")
    fullArray = strings.ReplaceAll(fullArray, "\n", " ")
    fullArray = normalizeWhitespace(fullArray)

    return &SetDirective{
        Variable: varName,
        Value:    fullArray,
    }, nil
}

func normalizeWhitespace(s string) string {
    // Replace multiple spaces with single space
    re := regexp.MustCompile(`\s+`)
    return re.ReplaceAllString(s, " ")
}
```

### 영향 받는 예제

**확장 QA 세트:**
1. ✅ `advanced/data_extraction_pipeline.oas` (Line 4-8)
2. ✅ `advanced/parallel_session_management.oas` (Line 5-9)

### 우선순위

**높음** ⚡⚡

**이유:**
- 가독성 크게 개선
- 유지보수 편의성 향상
- Git diff 품질 개선
- 다중 URL 처리는 자동화의 핵심 시나리오

---

## 제안 3: 복합 조건식 평가 개선

### 현재 상황 ⚠️

**확장 QA 세트 1개 실패:**
- `advanced/conditional_workflows.oas`

**일부 동작하지만 불안정:**
```bash
# 단일 조건 - ✅ 동작
@if ${ENABLE_PDF} == true
  @echo "PDF enabled"
@endif

# AND 조건 - ✅ 일부 동작 (검증됨: batch_control_flow_example.oas:81)
@if ${COUNT} > 0 and ${ENABLE_MODE} == true
  @echo "Both conditions met"
@endif

# OR 조건 - ⚠️ 일부 실패
@if ${ERROR_COUNT} > 0 or ${WARNING_COUNT} > 5
  @echo "Issues detected"
@endif

# NOT 조건 - ✅ 동작
@if not ${SUCCESS}
  @echo "Operation failed"
@endif

# 복합 조건 - ❌ 대부분 실패
@if (${COUNT} > 0 and ${COUNT} < 10) or ${FORCE_MODE} == true
  @echo "Valid range or forced"
@endif
```

**문제:**
- 조건식 파싱이 불완전
- 연산자 우선순위 처리 미흡
- 괄호 그룹핑 미지원
- 에러 메시지 불명확

### 예상 동작 ✅

표준 프로그래밍 언어 수준의 조건식 평가가 되어야 합니다.

**지원되어야 할 연산자:**
- 비교: `==`, `!=`, `>`, `<`, `>=`, `<=`
- 논리: `and`, `or`, `not`
- 그룹: `(`, `)`

**연산자 우선순위:**
1. `( )` - 그룹핑
2. `not` - 논리 부정
3. `>`, `<`, `>=`, `<=`, `==`, `!=` - 비교
4. `and` - 논리 AND
5. `or` - 논리 OR

### 기술적 구현 제안

**옵션 1: 표현식 파서 라이브러리 사용 (권장)**
```go
import "github.com/Knetic/govaluate"

func evaluateCondition(expr string, variables map[string]interface{}) (bool, error) {
    expression, err := govaluate.NewEvaluableExpression(expr)
    if err != nil {
        return false, fmt.Errorf("invalid condition expression: %w", err)
    }

    result, err := expression.Evaluate(variables)
    if err != nil {
        return false, fmt.Errorf("condition evaluation failed: %w", err)
    }

    boolResult, ok := result.(bool)
    if !ok {
        return false, fmt.Errorf("condition did not evaluate to boolean")
    }

    return boolResult, nil
}
```

**옵션 2: 재귀 하강 파서 직접 구현**
```go
type ConditionParser struct {
    tokens []Token
    pos    int
}

func (p *ConditionParser) parseExpression() (bool, error) {
    return p.parseOr()
}

func (p *ConditionParser) parseOr() (bool, error) {
    left, err := p.parseAnd()
    if err != nil {
        return false, err
    }

    for p.match("or") {
        right, err := p.parseAnd()
        if err != nil {
            return false, err
        }
        left = left || right
    }

    return left, nil
}

func (p *ConditionParser) parseAnd() (bool, error) {
    left, err := p.parseNot()
    if err != nil {
        return false, err
    }

    for p.match("and") {
        right, err := p.parseNot()
        if err != nil {
            return false, err
        }
        left = left && right
    }

    return left, nil
}

// ... parseNot, parseComparison, parsePrimary, etc.
```

### 우선순위

**중간** ⚡

**이유:**
- 기본 조건식은 대부분 동작
- 회피 방법 존재 (중첩 `@if` 사용)
- 영향받는 사례가 상대적으로 적음
- 제안 1, 2보다 우선순위 낮음

---

## 제안 4: 개선된 에러 메시지

### 현재 상황 ❌

**에러 메시지 예시:**
```
Line 5: https://en.wikipedia.org/wiki/Web_scraping,
Error: command not found: https://en.wikipedia.org/wiki/Web_scraping,
```

또는:
```
Warning: Unknown directive (skipped): @while ${ATTEMPT} < ${MAX_RETRIES}
Warning: Unknown directive (skipped): @endwhile
```

**문제점:**
- ❌ 실제 원인을 알 수 없음
- ❌ 해결 방법 제시 없음
- ❌ 초보자가 이해하기 어려움
- ❌ Warning으로만 표시되어 실패 원인 파악 어려움

### 개선 방향 ✅

**더 나은 에러 메시지:**
```
Line 4-8: Syntax Error in @set directive [SYNTAX_ERROR_001]
Error: Multi-line array definition is not currently supported

Current code:
  4 | @set TARGET_URLS = [
  5 |   "https://en.wikipedia.org/wiki/Web_scraping",
  6 |   "https://en.wikipedia.org/wiki/Browser_automation"
  7 | ]

Suggestion: Use single-line array definition
  @set TARGET_URLS = ["https://...", "https://...", "https://..."]

Documentation: https://docs.oa.com/batch-scripting#arrays
Roadmap: Multi-line arrays planned for v1.1.0
```

또는:
```
Line 15-30: Unknown directive @while [FEATURE_NOT_IMPLEMENTED]
Error: @while loops are not yet implemented in OA CLI

Current code:
  15 | @while ${ATTEMPT} < ${MAX_RETRIES} and not ${SUCCESS}
  16 |   @set ATTEMPT = ${ATTEMPT} + 1
  ...
  30 | @endwhile

Workaround: Use @foreach with fixed range
  @foreach i in [1, 2, 3]
    @if not ${SUCCESS}
      # your retry logic
    @endif
  @endforeach

Documentation: https://docs.oa.com/batch-scripting#loops
Issue Tracker: https://github.com/oa/oa/issues/32
```

### 제안 내용

**1. 구조화된 에러 타입:**
```go
type ScriptError struct {
    Code        string   // "SYNTAX_ERROR_001", "FEATURE_NOT_IMPLEMENTED", etc.
    Line        int
    LineEnd     int
    Message     string
    Context     []string // Surrounding code lines
    Suggestion  string   // Fix suggestion
    DocLink     string   // Documentation URL
    IssueLink   string   // GitHub issue tracking this limitation
}
```

**2. 문맥 정보 제공:**
- 에러가 발생한 라인 범위
- 주변 코드 5-10줄 표시
- 라인 번호 포함

**3. 원인 분석 및 해결 방법:**
- 명확한 에러 타입 분류
- 사용자 의도 추론
- 대안 문법 예시
- 관련 문서 링크

**4. 카테고리별 에러 코드:**
```
SYNTAX_ERROR_001: Multi-line array definition
SYNTAX_ERROR_002: Unterminated string
SYNTAX_ERROR_003: Invalid variable name

FEATURE_NOT_IMPLEMENTED_001: @while loop
FEATURE_NOT_IMPLEMENTED_002: @break directive
FEATURE_NOT_IMPLEMENTED_003: Complex regex in conditions

RUNTIME_ERROR_001: Variable undefined
RUNTIME_ERROR_002: Type mismatch
RUNTIME_ERROR_003: Division by zero

VALIDATION_WARNING_001: Variable defined but not used
VALIDATION_WARNING_002: Long sleep duration
```

### 구현 예시

```go
func (e *ScriptError) Error() string {
    var buf strings.Builder

    // Header
    fmt.Fprintf(&buf, "Line %d", e.Line)
    if e.LineEnd > e.Line {
        fmt.Fprintf(&buf, "-%d", e.LineEnd)
    }
    fmt.Fprintf(&buf, ": %s [%s]\n", e.Message, e.Code)

    // Context
    if len(e.Context) > 0 {
        buf.WriteString("\nCurrent code:\n")
        for i, line := range e.Context {
            lineNum := e.Line + i
            fmt.Fprintf(&buf, "  %3d | %s\n", lineNum, line)
        }
    }

    // Suggestion
    if e.Suggestion != "" {
        fmt.Fprintf(&buf, "\nSuggestion:\n  %s\n", e.Suggestion)
    }

    // Documentation
    if e.DocLink != "" {
        fmt.Fprintf(&buf, "\nDocumentation: %s\n", e.DocLink)
    }

    // Issue tracking
    if e.IssueLink != "" {
        fmt.Fprintf(&buf, "Issue Tracker: %s\n", e.IssueLink)
    }

    return buf.String()
}
```

### 우선순위

**높음** ⚡⚡

**이유:**
- 모든 사용자에게 영향
- 학습 곡선 크게 완화
- 디버깅 시간 단축
- 구현 난이도 낮음

---

## 구현 우선순위 요약

| 제안 | 중요도 | 난이도 | 영향도 | 우선순위 | 예상 공수 |
|------|--------|--------|--------|----------|-----------|
| 1. `@while` 루프 | 최우선 ⚡⚡⚡ | 높음 | 5개 예제 | **1순위** | 3-5일 |
| 2. 다중 라인 배열 | 높음 ⚡⚡ | 중간 | 2개 예제 | **2순위** | 1-2일 |
| 3. 에러 메시지 개선 | 높음 ⚡⚡ | 낮음 | 모든 사용자 | **2순위** | 2-3일 |
| 4. 복합 조건식 | 중간 ⚡ | 높음 | 1개 예제 | 3순위 | 5-7일 |

---

## 검증 데이터

### 저장소 내 예제 (11개)

**✅ 통과 (6개):**
1. `examples/batch_example.oas`
2. `examples/batch_variables_example.oas`
3. `examples/batch_plugin_test.oas`
4. `examples/batch_control_flow_example.oas`
5. `examples/batch_error_handling_example.oas`
6. `examples/batch_advanced_test.oas`

**❌ 실패 (5개) - 모두 `@while` 미지원:**
1. `examples/while_batch_processing.oas`
2. `examples/test_while_simple.oas`
3. `examples/test_while_condition.oas`
4. `examples/while_counter.oas`
5. `examples/while_file_polling.oas`

### 확장 QA 세트 (7개)

**❌ 실패 (7개):**
1. `advanced/data_extraction_pipeline.oas` - 다중 라인 배열
2. `advanced/parallel_session_management.oas` - 다중 라인 배열
3. `advanced/conditional_workflows.oas` - 복합 조건식
4. `oas-scripts/advanced_form_automation.oas` - `@while` 미지원
5. `basic/test_element_operations.oas` - 외부 사이트 셀렉터 변경
6. `oas-scripts/naver_blog_search.oas` - anti-bot 감지
7. `oas-scripts/naver_map_search.oas` - 동적 콘텐츠 타임아웃

**상세 결과:** `docs/OAS_VERIFICATION_RESULTS.md` 참조

---

## 기대 효과

### Phase 1: `@while` 구현 (1순위)

**즉시 효과:**
- 저장소 예제: 6/11 (54.5%) → 11/11 (100%) ⬆️ +45.5%
- 확장 QA: +1개 통과

**사용자 경험:**
- ✅ 재시도 로직 간편 구현
- ✅ 파일 폴링/대기 패턴 지원
- ✅ 페이지네이션 자동화 가능
- ✅ 동적 반복 제어 가능

### Phase 2: 다중 라인 배열 + 에러 메시지 (2순위)

**즉시 효과:**
- 확장 QA: +2개 통과 (배열 관련)
- 모든 사용자 디버깅 시간 단축

**사용자 경험:**
- ✅ 가독성 높은 배열 정의
- ✅ 유지보수 편의성 향상
- ✅ 명확한 에러 메시지로 학습 곡선 완화

### Phase 3: 복합 조건식 (3순위)

**즉시 효과:**
- 확장 QA: +1개 통과

**사용자 경험:**
- ✅ 복잡한 비즈니스 로직 구현 가능
- ✅ 전문가 수준 자동화 지원

### 장기 효과

**생태계 성장:**
- 더 많은 플러그인 개발자 유입
- 복잡한 자동화 시나리오 구현 가능
- 커뮤니티 예제 품질 향상
- 타 도구에서 마이그레이션 용이

**유지보수성:**
- 가독성 높은 스크립트
- 재사용 가능한 패턴
- 팀 협업 용이

---

## 타 언어/도구 비교

### Bash
```bash
# 다중 라인 배열 - ✅ 지원
URLS=(
  "url1"
  "url2"
)

# while 루프 - ✅ 지원
while [ $count -lt 5 ] && [ "$success" != "true" ]; do
  echo "Attempt $count"
  count=$((count + 1))
done
```

### Python
```python
# 다중 라인 배열 - ✅ 지원
urls = [
    "url1",
    "url2",
]

# while 루프 - ✅ 지원
while count < 5 and not success:
    print(f"Attempt {count}")
    count += 1
```

### PowerShell
```powershell
# 다중 라인 배열 - ✅ 지원
$urls = @(
  "url1",
  "url2"
)

# while 루프 - ✅ 지원
while ($count -lt 5 -and -not $success) {
  Write-Host "Attempt $count"
  $count++
}
```

**기대 수준:**
- .oas 파일도 현대적인 스크립트 언어 수준의 문법 지원 필요
- 사용자가 타 도구에서 .oas로 마이그레이션 시 학습 곡선 완화
- `@while`과 다중 라인 배열은 모든 주요 스크립트 언어의 기본 기능

---

## 제안 일정

### Phase 1: Critical Features (1-2주)
**목표:** 저장소 예제 100% 통과

1. ✅ **제안 1: `@while` 루프 구현** (3-5일)
   - 파서 수정 (1일)
   - 실행기 구현 (2일)
   - 테스트 및 디버깅 (1-2일)

2. ✅ **제안 4: 에러 메시지 개선** (2-3일)
   - ScriptError 구조 설계 (1일)
   - 에러 코드 체계 정의 (0.5일)
   - 구현 및 기존 에러 마이그레이션 (1-1.5일)

**예상 효과:**
- 저장소 예제: 54.5% → 100% ✅
- 사용자 만족도 크게 향상

### Phase 2: Quality Improvements (2-3주)
**목표:** 확장 QA 세트 개선

3. ✅ **제안 2: 다중 라인 배열 지원** (1-2일)
   - 파서 수정 (0.5일)
   - 공백 정규화 (0.5일)
   - 테스트 (1일)

**예상 효과:**
- 확장 QA: +2개 통과
- 가독성 크게 개선

### Phase 3: Advanced Features (4-6주)
**목표:** 전문가 수준 자동화 지원

4. ✅ **제안 3: 복합 조건식 개선** (5-7일)
   - 표현식 파서 선정/통합 (2-3일)
   - 조건 평가기 개선 (2일)
   - 테스트 (1-2일)

**예상 효과:**
- 확장 QA: +1개 통과
- 복잡한 비즈니스 로직 구현 가능

---

## 협업 요청

### webauto 팀 지원 가능 사항

1. **테스트 케이스 제공:**
   - `@while` 루프 다양한 시나리오
   - 다중 라인 배열 edge cases
   - 복합 조건식 예제

2. **구현 리뷰:**
   - PR 코드 리뷰 참여
   - 통합 테스트 수행
   - 문서 작성 지원

3. **피드백 수집:**
   - 사용자 의견 전달
   - 버그 리포트
   - 개선 제안

### 논의 필요 사항

**제안 1 (`@while` 루프):**
- ✅ 구현 방향 동의 여부
- ❓ 무한 루프 방지: 기본 최대 반복 횟수 10,000회 적절한가?
- ❓ `@break`, `@continue` 지시어 필요 여부
- ❓ 내장 변수 `__WHILE_ITERATION__` 필요 여부

**제안 2 (다중 라인 배열):**
- ✅ 구현 방향 동의 여부
- ❓ 최대 배열 라인 수 제한 (현재 1,000줄)
- ❓ 주석 처리 방식 (배열 내부 주석 허용 여부)

**제안 4 (에러 메시지):**
- ✅ 에러 코드 체계 동의 여부
- ❓ 문서 링크 URL 패턴
- ❓ GitHub Issue 자동 링크 여부
- ❓ 다국어 지원 계획 (한글/영문)

---

## 연락처

**담당:** webauto 플러그인 팀
**문의:** GitHub Issues
**문서 위치:** `/Users/allieus/Apps/pyhub-office-automation/plugins/webauto/docs/`

**관련 문서:**
- `docs/OAS_VERIFICATION_RESULTS.md` - 상세 검증 결과
- `docs/IMPLEMENTATION_SUMMARY_2025-10-21.md` - 최근 구현 내역

---

## 부록: 구현 예시 코드

### A. `@while` 루프 파싱 구현

```go
// pkg/batch/parser.go

type WhileStatement struct {
    Condition string
    Body      []Statement
    LineStart int
    LineEnd   int
}

func (p *Parser) parseWhileBlock(line string) (*WhileStatement, error) {
    lineStart := p.currentLine

    // Extract condition from "@while <condition>"
    condition := strings.TrimPrefix(line, "@while")
    condition = strings.TrimSpace(condition)

    if condition == "" {
        return nil, fmt.Errorf("@while requires a condition")
    }

    // Collect body until @endwhile
    body := []Statement{}
    depth := 1  // Track nested @while blocks

    for p.scanner.Scan() {
        p.currentLine++
        line := p.scanner.Text()
        trimmed := strings.TrimSpace(line)

        // Handle nested @while
        if strings.HasPrefix(trimmed, "@while") {
            depth++
        } else if strings.HasPrefix(trimmed, "@endwhile") {
            depth--
            if depth == 0 {
                break
            }
        }

        // Parse statement
        stmt, err := p.parseStatement(line)
        if err != nil {
            return nil, err
        }
        body = append(body, stmt)
    }

    if depth != 0 {
        return nil, fmt.Errorf("@while block not closed (missing @endwhile)")
    }

    return &WhileStatement{
        Condition: condition,
        Body:      body,
        LineStart: lineStart,
        LineEnd:   p.currentLine,
    }, nil
}
```

### B. `@while` 루프 실행 구현

```go
// pkg/batch/executor.go

func (e *Executor) executeWhileLoop(whileStmt *WhileStatement) error {
    maxIterations := e.getMaxIterations() // Default: 10000
    iteration := 0
    startTime := time.Now()

    for {
        // Check iteration limit
        if iteration >= maxIterations {
            return &ScriptError{
                Code:    "RUNTIME_ERROR_INFINITE_LOOP",
                Line:    whileStmt.LineStart,
                LineEnd: whileStmt.LineEnd,
                Message: fmt.Sprintf("@while loop exceeded max iterations (%d)", maxIterations),
                Suggestion: "Check your loop condition or increase OA_MAX_WHILE_ITERATIONS",
            }
        }

        // Check timeout (5 minutes default)
        if time.Since(startTime) > 5*time.Minute {
            return &ScriptError{
                Code:    "RUNTIME_ERROR_TIMEOUT",
                Line:    whileStmt.LineStart,
                Message: "@while loop timeout after 5 minutes",
            }
        }

        // Evaluate condition
        conditionResult, err := e.evaluateCondition(whileStmt.Condition)
        if err != nil {
            return fmt.Errorf("failed to evaluate @while condition: %w", err)
        }

        if !conditionResult {
            break // Exit loop
        }

        // Execute loop body
        for _, stmt := range whileStmt.Body {
            err := e.executeStatement(stmt)
            if err != nil {
                return err
            }
        }

        iteration++

        // Set built-in variable for debugging
        e.setVariable("__WHILE_ITERATION__", strconv.Itoa(iteration))
    }

    return nil
}

func (e *Executor) getMaxIterations() int {
    if max := os.Getenv("OA_MAX_WHILE_ITERATIONS"); max != "" {
        if val, err := strconv.Atoi(max); err == nil && val > 0 {
            return val
        }
    }
    return 10000 // Default
}
```

### C. 다중 라인 배열 파싱 구현

```go
// pkg/batch/parser.go

func (p *Parser) parseMultiLineArray(varName, firstLine string) (*SetDirective, error) {
    arrayLines := []string{firstLine}
    lineStart := p.currentLine

    // Collect lines until closing bracket
    for p.scanner.Scan() {
        p.currentLine++
        line := strings.TrimSpace(p.scanner.Text())

        // Skip empty lines and comments
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }

        arrayLines = append(arrayLines, line)

        if strings.Contains(line, "]") {
            break
        }

        // Safety: prevent infinite loop
        if len(arrayLines) > 1000 {
            return nil, &ScriptError{
                Code:    "SYNTAX_ERROR_ARRAY_TOO_LONG",
                Line:    lineStart,
                Message: "Array definition too long (>1000 lines)",
                Suggestion: "Consider splitting into multiple arrays or using external file",
            }
        }
    }

    // Join and normalize whitespace
    fullArray := strings.Join(arrayLines, " ")
    fullArray = normalizeWhitespace(fullArray)

    return &SetDirective{
        Variable: varName,
        Value:    fullArray,
    }, nil
}

func normalizeWhitespace(s string) string {
    // Replace multiple spaces with single space
    re := regexp.MustCompile(`\s+`)
    normalized := re.ReplaceAllString(s, " ")
    return strings.TrimSpace(normalized)
}
```

---

**이 제안서를 OA CLI 팀에게 전달해주시면:**
- 저장소 예제 성공률: 54.5% → 100% (Phase 1 완료 시)
- 전체 예제 성공률: 33% → 78% (모든 Phase 완료 시)
- 사용자 경험 크게 개선
- 실전 자동화 시나리오 구현 가능

**우선순위:** `@while` 루프 구현이 가장 시급합니다. (5개 예제 대기 중)
