# OA CLI 산술 연산 지원 제안서

**작성일:** 2025-10-23
**제안자:** webauto 플러그인 팀
**수신:** Codex (OA CLI 개발팀)
**우선순위:** 🔥 최우선 (Critical)

---

## 📋 Executive Summary

Issue #34로 `@while` 루프가 구현되었으나, **산술 연산 미지원**으로 인해 실용적 사용이 불가능합니다. 본 제안서는 **코드 간결성**을 최우선으로 고려한 해결 방안을 제시합니다.

**핵심 문제:**
```bash
@set COUNTER = ${COUNTER} + 1  # ❌ 에러: unexpected character '+'
```

**목표:**
```bash
@set COUNTER = ${COUNTER} + 1  # ✅ 정상 작동
```

---

## 🎯 설계 원칙: 코드 간결성 우선

### 설계 철학

사용자가 **가장 간결하게** 자동화 스크립트를 작성할 수 있어야 합니다.

**Bad (장황함):**
```bash
@set COUNTER = @lua(COUNTER + 1)           # @lua() 노이즈
@if @lua((COUNT > 0 and COUNT < 10) or FORCE)  # 반복되는 @lua()
```

**Good (간결함):**
```bash
@set COUNTER = ${COUNTER} + 1              # 직관적
@if (${COUNT} > 0 and ${COUNT} < 10) or ${FORCE}  # 깔끔함
```

### 간결성 비교 분석

#### 시나리오 1: 재시도 로직 (12줄 vs 11줄)

**하이브리드 방식 (Lua 표현식):**
```bash
@set RETRY = 0
@set MAX = 3
@set SUCCESS = false

@while @lua(RETRY < MAX and not SUCCESS)  # @lua() 노이즈
  @try
    oa plugin exec webauto element-click --session-id "${SID}" --element-selector "#btn"
    @set SUCCESS = true
  @catch
    @set RETRY = @lua(RETRY + 1)  # 또 @lua()
    @sleep 1000
  @endtry
@endwhile
```
**라인 수:** 12줄

**네이티브 지원 (제안 방식):**
```bash
@set RETRY = 0
@set SUCCESS = false

@while ${RETRY} < 3 and not ${SUCCESS}  # 깔끔함
  @try
    oa plugin exec webauto element-click --session-id "${SID}" --element-selector "#btn"
    @set SUCCESS = true
  @catch
    @set RETRY = ${RETRY} + 1  # 간결함
    @sleep 1000
  @endtry
@endwhile
```
**라인 수:** 11줄 ✅ **가장 간결**

#### 시나리오 2: OA 명령 호출 (1줄 vs 3-4줄)

**순수 Lua 방식:**
```lua
oa.plugin_exec("webauto", "browser-launch", {
    session_id = SID,
    headless = false
})
```
**라인 수:** 3-4줄

**네이티브 .oas (제안 방식):**
```bash
oa plugin exec webauto browser-launch --session-id "${SID}" --no-headless
```
**라인 수:** 1줄 ✅ **최고로 간결**

### 결론: 네이티브 표현식 지원이 최선

| 접근 방식 | 산술 연산 | OA 명령 | 학습 곡선 | 간결성 | 종합 |
|----------|----------|---------|----------|--------|------|
| 하이브리드 (@lua) | `@lua(X+1)` | 1줄 ✅ | 중간 | ⚠️ 노이즈 | △ |
| 순수 Lua 전환 | `x=x+1` | 3-4줄 | 높음 | ⚠️ 장황 | △ |
| **네이티브 지원** | `${X}+1` | 1줄 ✅ | 낮음 | ✅ 최고 | **⭐ 최선** |

---

## 🔧 기술 제안: expr-lang/expr

### 라이브러리 선정

**추천:** [`github.com/expr-lang/expr`](https://github.com/expr-lang/expr)

**선정 이유:**
- ✅ govaluate의 공식 후계자 (ARCHIVED.md에서 명시)
- ✅ 활발한 유지보수 (2024년 현재)
- ✅ Google, Uber, ByteDance 사용 (검증된 안정성)
- ✅ 메모리 안전, 부작용 없음, 항상 종료 보장
- ✅ 정적 타입 검증
- ✅ 6.1K+ stars, 활발한 커뮤니티

**대안 비교:**

| 라이브러리 | Stars | 상태 | 유지보수 | 평가 |
|-----------|-------|------|----------|------|
| govaluate | 4.3K | ❌ Archived | 중단됨 | 사용 불가 |
| **expr-lang/expr** | **6.1K** | ✅ Active | **활발** | **⭐ 추천** |
| casbin/govaluate | ~100 | ✅ Fork | 제한적 | 대안 |

### 설치

```bash
go get github.com/expr-lang/expr
```

### 기본 사용법

```go
package main

import (
    "fmt"
    "github.com/expr-lang/expr"
)

func main() {
    env := map[string]interface{}{
        "COUNTER": 5,
        "PRICE": 100,
        "QUANTITY": 3,
    }

    // 산술 연산
    program, _ := expr.Compile("COUNTER + 1", expr.Env(env))
    output, _ := expr.Run(program, env)
    fmt.Println(output)  // 6

    // 복잡한 표현식
    program, _ = expr.Compile("PRICE * QUANTITY + 10", expr.Env(env))
    output, _ = expr.Run(program, env)
    fmt.Println(output)  // 310

    // 조건식
    env["COUNT"] = 5
    env["FORCE"] = true
    program, _ = expr.Compile("(COUNT > 0 and COUNT < 10) or FORCE", expr.Env(env))
    output, _ = expr.Run(program, env)
    fmt.Println(output)  // true
}
```

---

## 💻 구현 계획

### Phase 1: 산술 연산 지원 (3-5일)

#### 1.1 표현식 평가기 추가

**파일:** `pkg/batch/variables.go`

```go
package batch

import (
    "fmt"
    "github.com/expr-lang/expr"
    "strings"
)

// evaluateExpression evaluates arithmetic and logical expressions
func evaluateExpression(expression string, vars map[string]string) (string, error) {
    // 변수 치환: ${VAR} -> VAR
    cleanExpr := substituteVariablesForExpr(expression, vars)

    // 산술/논리 연산자 감지
    if !containsOperators(cleanExpr) {
        // 단순 문자열 값
        return cleanExpr, nil
    }

    // expr로 평가
    env := make(map[string]interface{})
    for k, v := range vars {
        // 숫자 변환 시도
        if num, err := strconv.Atoi(v); err == nil {
            env[k] = num
        } else if b, err := strconv.ParseBool(v); err == nil {
            env[k] = b
        } else {
            env[k] = v
        }
    }

    program, err := expr.Compile(cleanExpr, expr.Env(env))
    if err != nil {
        return "", fmt.Errorf("expression compilation failed: %w", err)
    }

    output, err := expr.Run(program, env)
    if err != nil {
        return "", fmt.Errorf("expression evaluation failed: %w", err)
    }

    return fmt.Sprintf("%v", output), nil
}

func containsOperators(s string) bool {
    operators := []string{"+", "-", "*", "/", "%", "<", ">", "==", "!=", "and", "or", "not"}
    for _, op := range operators {
        if strings.Contains(s, " "+op+" ") || strings.Contains(s, op) {
            return true
        }
    }
    return false
}

func substituteVariablesForExpr(expr string, vars map[string]string) string {
    // ${VAR} -> VAR 변환
    result := expr
    for k := range vars {
        result = strings.ReplaceAll(result, "${"+k+"}", k)
    }
    return result
}
```

#### 1.2 @set 지시어 수정

**파일:** `pkg/batch/parser.go`

```go
func parseSetDirective(line string) (*SetDirective, error) {
    // "@set VAR = EXPR" 파싱
    parts := strings.SplitN(line, "=", 2)
    if len(parts) != 2 {
        return nil, fmt.Errorf("invalid @set syntax")
    }

    varName := strings.TrimSpace(strings.TrimPrefix(parts[0], "@set"))
    expr := strings.TrimSpace(parts[1])

    return &SetDirective{
        Variable: varName,
        Expression: expr,
    }, nil
}
```

#### 1.3 변수 할당 시 평가

**파일:** `pkg/batch/executor.go`

```go
func (e *Executor) executeSetDirective(directive *SetDirective) error {
    // 표현식 평가
    value, err := evaluateExpression(directive.Expression, e.variables)
    if err != nil {
        return fmt.Errorf("failed to evaluate expression '%s': %w",
            directive.Expression, err)
    }

    // 변수 저장
    e.variables[directive.Variable] = value
    return nil
}
```

### Phase 2: 조건식 개선 (1-2일)

**파일:** `pkg/batch/control_flow.go`

```go
func evaluateCondition(condition string, vars map[string]string) (bool, error) {
    // expr 사용하여 조건 평가
    result, err := evaluateExpression(condition, vars)
    if err != nil {
        return false, err
    }

    // boolean 변환
    return strconv.ParseBool(result)
}
```

### Phase 3: 에러 메시지 개선 (1일)

```go
func (e *Executor) executeSetDirective(directive *SetDirective) error {
    value, err := evaluateExpression(directive.Expression, e.variables)
    if err != nil {
        return &ScriptError{
            Code: "EXPRESSION_EVALUATION_ERROR",
            Message: fmt.Sprintf("Failed to evaluate expression: %s", directive.Expression),
            Line: directive.LineNumber,
            Context: directive.Expression,
            Suggestion: "Check variable names and operator syntax. Supported operators: +, -, *, /, %, <, >, ==, !=, and, or, not",
            Details: map[string]interface{}{
                "expression": directive.Expression,
                "error": err.Error(),
            },
        }
    }

    e.variables[directive.Variable] = value
    return nil
}
```

---

## 📊 지원 가능한 연산

### 산술 연산

```bash
@set RESULT = ${A} + ${B}       # 덧셈
@set RESULT = ${A} - ${B}       # 뺄셈
@set RESULT = ${A} * ${B}       # 곱셈
@set RESULT = ${A} / ${B}       # 나눗셈
@set RESULT = ${A} % ${B}       # 나머지
@set RESULT = (${A} + ${B}) * ${C}  # 괄호 그룹핑
```

### 비교 연산

```bash
@set RESULT = ${A} > ${B}       # 크다
@set RESULT = ${A} < ${B}       # 작다
@set RESULT = ${A} >= ${B}      # 크거나 같다
@set RESULT = ${A} <= ${B}      # 작거나 같다
@set RESULT = ${A} == ${B}      # 같다
@set RESULT = ${A} != ${B}      # 다르다
```

### 논리 연산

```bash
@set RESULT = ${A} and ${B}     # AND
@set RESULT = ${A} or ${B}      # OR
@set RESULT = not ${A}          # NOT
@set RESULT = (${A} and ${B}) or ${C}  # 복합 조건
```

### 조건식 (이미 구현?)

```bash
@if ${COUNT} > 0 and ${COUNT} < 10
@if (${A} == true and ${B} == false) or ${C}
@while ${RETRY} < ${MAX_RETRIES} and not ${SUCCESS}
```

---

## ✅ 예상 효과

### 1. 즉시 해결되는 문제

**Before (현재):**
```bash
@set COUNTER = 0
@while ${COUNTER} < 3
  @echo "Count: ${COUNTER}"
  @set COUNTER = 1  # ❌ 고정값만 가능, 무한 루프
@endwhile
```

**After (구현 후):**
```bash
@set COUNTER = 0
@while ${COUNTER} < 3
  @echo "Count: ${COUNTER}"
  @set COUNTER = ${COUNTER} + 1  # ✅ 정상 증가
@endwhile
```

### 2. 실행 가능해지는 스크립트

**Repository Examples:**
- ✅ `examples/while_batch_processing.oas` (재시도 로직)
- ✅ `examples/test_while_simple.oas` (카운터 테스트)
- ✅ `examples/test_while_condition.oas` (조건 평가)
- ✅ `examples/while_counter.oas` (카운터 반복)
- ✅ `examples/while_file_polling.oas` (파일 폴링)

**총 5개 스크립트가 즉시 작동 가능**

### 3. 사용 가능한 패턴

**재시도 로직:**
```bash
@set RETRY = 0
@while ${RETRY} < 3 and not ${SUCCESS}
  @try
    # 작업 시도
    @set SUCCESS = true
  @catch
    @set RETRY = ${RETRY} + 1  # ✅ 가능
    @sleep 1000
  @endtry
@endwhile
```

**페이지네이션:**
```bash
@set PAGE = 1
@set MAX_PAGES = 10
@while ${PAGE} <= ${MAX_PAGES}
  oa plugin exec webauto page-navigate --page-url "https://site.com?page=${PAGE}"
  @set PAGE = ${PAGE} + 1  # ✅ 가능
@endwhile
```

**누적 계산:**
```bash
@set TOTAL = 0
@foreach item in ${ITEMS}
  @set TOTAL = ${TOTAL} + ${item}  # ✅ 가능
@endforeach
@echo "Total: ${TOTAL}"
```

---

## 🎯 구현 우선순위

### 🔥 Critical (최우선): 산술 연산

**예상 기간:** 3-5일

**구현 항목:**
- `pkg/batch/variables.go`: `evaluateExpression()` 추가
- `pkg/batch/parser.go`: `parseSetDirective()` 수정
- `pkg/batch/executor.go`: `executeSetDirective()` 수정
- expr 라이브러리 통합
- 단위 테스트 작성

**검증 방법:**
```bash
# 간단한 테스트
@set X = 5
@set Y = ${X} + 1
@echo "${Y}"  # 6 출력

# 복잡한 테스트
@set RESULT = (${A} + ${B}) * ${C}
@if ${RESULT} > 100
  @echo "High value"
@endif
```

### ⚠️ Important (중요): 에러 메시지 개선

**예상 기간:** 1일

**구현 항목:**
- ScriptError에 표현식 평가 에러 추가
- 사용자 친화적 에러 메시지
- 해결 제안 포함

### ℹ️ Nice to Have (선택): 추가 기능

**예상 기간:** 1-2일 (선택 사항)

**구현 항목:**
- 문자열 함수 (len, substr, replace)
- 배열 함수 (length, append, contains)
- 수학 함수 (abs, min, max, round)

---

## 📋 테스트 계획

### 1. 단위 테스트

```go
// pkg/batch/variables_test.go
func TestEvaluateExpression(t *testing.T) {
    vars := map[string]string{
        "A": "5",
        "B": "3",
    }

    tests := []struct {
        expr     string
        expected string
    }{
        {"${A} + ${B}", "8"},
        {"${A} - ${B}", "2"},
        {"${A} * ${B}", "15"},
        {"${A} / ${B}", "1"},
        {"${A} % ${B}", "2"},
        {"(${A} + ${B}) * 2", "16"},
    }

    for _, tt := range tests {
        result, err := evaluateExpression(tt.expr, vars)
        assert.NoError(t, err)
        assert.Equal(t, tt.expected, result)
    }
}
```

### 2. 통합 테스트

```bash
# test_arithmetic.oas
@set COUNTER = 0
@set MAX = 3

@while ${COUNTER} < ${MAX}
  @echo "Iteration: ${COUNTER}"
  @set COUNTER = ${COUNTER} + 1
@endwhile

@echo "Final: ${COUNTER}"  # 3 출력 예상
```

### 3. 회귀 테스트

기존 5개 `examples/while_*` 스크립트 모두 실행하여 검증

---

## 📚 문서화 계획

### 업데이트 필요 문서

1. **OA CLI User Guide**
   - 산술 연산 섹션 추가
   - 지원 연산자 목록
   - 예제 코드

2. **CHANGELOG.md**
   ```markdown
   ## [1.1.0] - 2025-10-XX

   ### Added
   - Arithmetic operations in @set directive (+, -, *, /, %)
   - Complex expression evaluation using expr-lang/expr
   - Improved error messages for expression evaluation

   ### Fixed
   - @while loops now support counter increment patterns
   ```

3. **README.md**
   - 새로운 기능 하이라이트
   - 간단한 사용 예제

---

## 🔄 대안 검토 및 기각 이유

### 대안 1: @inc / @dec 지시어

```bash
@inc COUNTER        # COUNTER += 1
@dec COUNTER        # COUNTER -= 1
```

**기각 이유:**
- ⚠️ 증가/감소만 가능, 곱셈/나눗셈 불가
- ⚠️ 복잡한 계산 불가 (예: `TOTAL = PRICE * QUANTITY + TAX`)
- ⚠️ 새로운 문법 추가로 학습 곡선 증가
- ⚠️ 근본적 해결책 아님

### 대안 2: 하이브리드 (@lua 표현식)

```bash
@set COUNTER = @lua(COUNTER + 1)
@if @lua((COUNT > 0 and COUNT < 10) or FORCE)
```

**기각 이유:**
- ❌ `@lua()` 반복이 시각적 노이즈
- ❌ 코드 간결성 저하 (12줄 vs 11줄)
- ❌ Lua 의존성 추가
- ❌ 사용자가 Lua 문법 배워야 함

### 대안 3: 순수 Lua 전환

```lua
-- .oalua 스크립트
counter = 0
while counter < 10 do
    oa.plugin_exec("webauto", "browser-launch")
    counter = counter + 1
end
```

**기각 이유:**
- ❌ 기존 .oas 투자 손실 (파서, 예제, 문서)
- ❌ OA 명령 호출이 장황함 (1줄 → 3-4줄)
- ❌ 높은 학습 곡선 (새로운 언어)
- ❌ 10-15일 소요 (전면 재작성)
- ❌ 간결성 측면에서 열등

---

## 💰 비용-편익 분석

### 구현 비용

| 항목 | 예상 시간 | 복잡도 |
|------|----------|--------|
| 표현식 평가기 | 2일 | 중간 |
| @set 지시어 수정 | 1일 | 낮음 |
| 에러 처리 | 1일 | 낮음 |
| 테스트 작성 | 1일 | 낮음 |
| **총계** | **5일** | **중간** |

### 기대 효과

| 효과 | 평가 |
|------|------|
| 5개 예제 즉시 작동 | ✅ 높음 |
| 사용자 만족도 향상 | ✅ 높음 |
| 기능 완성도 | ✅ 높음 |
| 학습 곡선 | ✅ 낮음 유지 |
| 코드 간결성 | ✅ 최고 |

**ROI:** 5일 투자로 핵심 기능 완성 + 사용자 만족도 대폭 향상

---

## 🎯 액션 아이템

### Codex 팀 (OA CLI)

1. **의사 결정 (1일)**
   - 본 제안서 검토
   - expr-lang/expr 라이브러리 승인
   - 구현 일정 확정

2. **구현 (3-5일)**
   - expr 라이브러리 통합
   - 표현식 평가기 구현
   - @set 지시어 수정
   - 단위 테스트 작성

3. **검증 (1일)**
   - 통합 테스트 실행
   - 5개 while 예제 검증
   - 문서 업데이트

### webauto 팀

1. **테스트 지원**
   - 5개 while 예제로 회귀 테스트
   - 실제 사용 시나리오 검증
   - 피드백 제공

2. **문서 업데이트**
   - webauto 예제 문서 갱신
   - 새로운 패턴 예제 추가

---

## 📞 연락처

**제안 관련 문의:**
- GitHub Issue: pyhub-office-automation/oa#34
- 담당: webauto 플러그인 팀

**기술 지원:**
- expr-lang/expr 문서: https://expr-lang.org/
- GitHub: https://github.com/expr-lang/expr

---

## 부록 A: expr-lang/expr 기능

### 지원 연산자

**산술:**
- `+` (덧셈), `-` (뺄셈), `*` (곱셈), `/` (나눗셈), `%` (나머지)
- `**` (거듭제곱)

**비교:**
- `==`, `!=`, `<`, `>`, `<=`, `>=`

**논리:**
- `and`, `or`, `not`
- `in` (멤버십 테스트)

**문자열:**
- `+` (연결)
- `matches` (정규식)
- `contains`, `startsWith`, `endsWith`

**기타:**
- `? :` (삼항 연산자)
- `??` (null 병합)

### 내장 함수

**배열:**
- `len()`, `all()`, `none()`, `any()`, `one()`
- `filter()`, `map()`, `count()`

**수학:**
- `abs()`, `ceil()`, `floor()`, `round()`
- `max()`, `min()`

**문자열:**
- `lower()`, `upper()`, `trim()`

### 타입 안전성

```go
env := map[string]interface{}{
    "name": "John",
    "age": 30,
}

// ✅ 타입 체크
program, err := expr.Compile("name + age", expr.Env(env))
// err: invalid operation + (mismatched types string and int)
```

---

## 부록 B: 참고 자료

1. **govaluate ARCHIVED.md**
   - https://github.com/Knetic/govaluate/blob/master/ARCHIVED.md
   - expr-lang/expr를 공식 후계자로 언급

2. **expr-lang/expr 문서**
   - https://expr-lang.org/
   - https://github.com/expr-lang/expr

3. **사용 사례**
   - Google Cloud Platform
   - Uber Eats
   - GoDaddy Pro
   - ByteDance

4. **OA CLI Issue #34**
   - https://github.com/pyhub-office-automation/oa/issues/34
   - webauto 검증 결과: docs/OA_CLI_ISSUE_34_현황.md

---

**문서 버전:** 1.0
**최종 수정:** 2025-10-23
**상태:** 제안 대기중
