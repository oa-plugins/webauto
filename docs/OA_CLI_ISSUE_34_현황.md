# OA CLI Issue #34 구현 현황 및 제한사항

**검증일:** 2025-10-21
**OA CLI 커밋:** 7c7ab37 (Implement batch parser enhancements #34)
**검증자:** webauto 플러그인 팀

---

## 요약

codex가 Issue #34를 통해 `@while` 루프, 다중 라인 배열, 복합 조건식 지원을 구현했습니다. 기본 기능은 작동하지만, **산술 연산 미지원**으로 인해 실용적인 `@while` 루프 사용에 제약이 있습니다.

**핵심 이슈:**
- ✅ `@while` 루프 구문 인식 및 실행
- ❌ **산술 연산 미지원** (`+`, `-`, `*`, `/`)
- ⚠️ 이로 인해 카운터 증가 로직 구현 불가

---

## 구현 완료 항목

### 1. `@while` 루프 지원 ✅

**상태:** 구문 인식 및 기본 실행 작동

**구현 내용:**
- `@while <condition>` ~ `@endwhile` 파싱
- 조건 평가 및 반복 실행
- 무한 루프 방지 (10,000회 제한)
- `__WHILE_ITERATION__` 내장 변수
- `@break`, `@continue` 지시어

**테스트 코드:**
```bash
@set COUNTER = 0

@while ${COUNTER} < 3
  @echo "Count: ${COUNTER}"
  @set COUNTER = 1  # 고정값만 가능, 증가 불가
@endwhile
```

**실행 결과:**
```
Count: 0
Count: 1
Count: 1
Count: 1
... (10,000번 반복)
Lines 4-7: @while loop exceeded max iterations (10000) [RUNTIME_ERROR_INFINITE_LOOP]
```

**문제:**
- ✅ `@while` 인식됨
- ✅ 조건 평가 작동
- ✅ 루프 실행됨
- ❌ **카운터 증가 불가** (산술 연산 미지원)

### 2. 개선된 에러 메시지 ✅

**상태:** 구조화된 에러 메시지 작동

**구현 내용:**
- `ScriptError` 구조체
- 라인 범위 표시
- 문맥 정보 (주변 코드)
- 해결 제안 (Suggestion)
- 에러 코드 분류 (`RUNTIME_ERROR_*`, `SYNTAX_ERROR_*`)

**예시:**
```
Lines 4-7: @while loop exceeded max iterations (10000) [RUNTIME_ERROR_INFINITE_LOOP]
  Command: @while ${COUNTER} < 3
  Context:
       4 | @while ${COUNTER} < 3
       5 |   @echo "Count: ${COUNTER}"
       6 |   @set COUNTER = 1
       7 | @endwhile
  Suggestion: Ensure the loop updates its exit condition or adjust OA_MAX_WHILE_ITERATIONS.
```

**평가:** ✅ 매우 유용한 에러 메시지

### 3. `@break` / `@continue` 지원 ✅

**상태:** 구현됨 (Issue #34 구현 노트에 명시)

**기능:**
- `@break`: 루프 즉시 종료
- `@continue`: 현재 반복 건너뛰고 다음 반복

**미테스트:** 아직 동작 확인 안 함

---

## 발견된 제한사항

### 🚨 제한사항 1: 산술 연산 미지원 (Critical)

**상태:** ❌ 미구현

**문제:**
`@set` 지시어에서 산술 연산을 지원하지 않아 **카운터 증가/감소가 불가능**합니다.

**재현 코드:**
```bash
@set COUNTER = 0
@set COUNTER = ${COUNTER} + 1  # ❌ 파싱 에러
```

**에러 메시지:**
```
Lines 34-90: unexpected character '+' [SYNTAX_ERROR_CONDITION]
  Command: @while ${ATTEMPT} < ${MAX_RETRIES} and not ${SUCCESS}
  Context:
      34 | @while ${ATTEMPT} < ${MAX_RETRIES} and not ${SUCCESS}
      35 |   @set ATTEMPT = ${ATTEMPT} + 1
      36 |   @echo ""
  Suggestion: Verify the @while condition syntax and ensure parentheses are balanced.
```

**영향:**
- ❌ `@while` 루프에서 카운터 증가 불가
- ❌ 재시도 로직 구현 불가
- ❌ 페이지네이션 구현 불가
- ❌ 모든 `examples/while_*` 스크립트 실행 불가

**회피 방법 (현재 불가능):**
```bash
# 시도 1: 산술 연산
@set COUNTER = ${COUNTER} + 1  # ❌ 에러

# 시도 2: 표현식
@set COUNTER = $((${COUNTER} + 1))  # ❌ 에러

# 시도 3: 고정값만 가능
@set COUNTER = 1  # ✅ 가능하지만 증가 불가
```

**필요한 구현:**

**옵션 1: 기본 산술 연산 지원 (권장)**
```bash
@set COUNTER = ${COUNTER} + 1
@set COUNTER = ${COUNTER} - 1
@set TOTAL = ${COUNT} * ${PRICE}
@set AVERAGE = ${SUM} / ${COUNT}
@set REMAINDER = ${NUM} % 10
```

**옵션 2: `@expr` 지시어 추가**
```bash
@expr COUNTER = ${COUNTER} + 1
@expr TOTAL = (${PRICE} * ${QUANTITY}) + ${TAX}
```

**옵션 3: `@inc` / `@dec` 지시어 추가 (최소 구현)**
```bash
@inc COUNTER        # COUNTER += 1
@dec COUNTER        # COUNTER -= 1
@inc COUNTER 5      # COUNTER += 5
@dec COUNTER 2      # COUNTER -= 2
```

**우선순위:** 🔥 **최우선** (없으면 `@while` 루프가 실용적으로 사용 불가)

---

### ⚠️ 제한사항 2: 다중 라인 배열 (미검증)

**상태:** ⚠️ 구현되었다고 하지만 미검증

**Issue #34 Scope 명시:**
> Add multi-line array support during parsing (`pkg/batch/parser.go`, `pkg/batch/variables.go`).

**테스트 필요:**
```bash
@set URLS = [
  "https://example.com",
  "https://github.com",
  "https://playwright.dev"
]

@foreach url in ${URLS}
  @echo "URL: ${url}"
@endforeach
```

**예상 결과:** ✅ 정상 작동 (구현되었다고 명시됨)

**검증 필요:** 실제 테스트 필요

---

### ⚠️ 제한사항 3: 복합 조건식 (미검증)

**상태:** ⚠️ 구현되었다고 하지만 미검증

**Issue #34 Scope 명시:**
> Harden condition evaluation so chained `and`/`or`/`not` expressions and parentheses behave consistently.

**구현 노트:**
> Boolean expressions run through the new precedence-aware parser so nested parentheses and mixed operators behave predictably.

**테스트 필요:**
```bash
# 괄호 그룹핑
@if (${COUNT} > 0 and ${COUNT} < 10) or ${FORCE_MODE} == true
  @echo "Valid"
@endif

# 연산자 우선순위
@if ${A} == true and ${B} == true or ${C} == false
  @echo "Complex condition"
@endif

# 중첩 NOT
@if not (${ERROR} == true or ${WARNING} == true)
  @echo "All good"
@endif
```

**예상 결과:** ✅ 정상 작동 (우선순위 인식 파서 구현됨)

**검증 필요:** 실제 테스트 필요

---

### ⚠️ 제한사항 4: 문자열 연산 (미확인)

**상태:** ❓ 지원 여부 불명

**필요한 기능:**
```bash
# 문자열 연결
@set FULL_NAME = "${FIRST_NAME} ${LAST_NAME}"  # ✅ 변수 치환으로 가능?

# 문자열 길이
@set LENGTH = len("${TEXT}")  # ❓

# 부분 문자열
@set SUBSTR = substr("${TEXT}", 0, 5)  # ❓

# 문자열 치환
@set REPLACED = replace("${TEXT}", "old", "new")  # ❓
```

**우선순위:** 낮음 (기본 자동화에는 불필요)

---

### ⚠️ 제한사항 5: 배열 조작 (미확인)

**상태:** ❓ 지원 여부 불명

**필요한 기능:**
```bash
@set ITEMS = ["a", "b", "c"]

# 배열 길이
@set COUNT = ${ITEMS.length}  # ❓ 지원 여부 불명

# 배열 추가
@append ITEMS "d"  # ❓

# 배열 인덱스 접근
@set FIRST = ${ITEMS[0]}  # ❓
```

**우선순위:** 중간 (편의 기능)

---

## 실제 사용 시나리오별 평가

### 시나리오 1: 재시도 로직

**요구사항:**
```bash
@set RETRY_COUNT = 0
@set MAX_RETRIES = 3
@set SUCCESS = false

@while ${RETRY_COUNT} < ${MAX_RETRIES} and not ${SUCCESS}
  @try
    oa plugin exec webauto element-click --session-id "${SID}" --element-selector "#submit"
    @set SUCCESS = true
  @catch
    @set RETRY_COUNT = ${RETRY_COUNT} + 1  # ❌ 산술 연산 필요
    @sleep 1000
  @endtry
@endwhile
```

**현재 상태:** ❌ **불가능** (산술 연산 미지원)

**회피 방법:** 없음

---

### 시나리오 2: 파일 폴링 (고정 횟수)

**요구사항:**
```bash
@set MAX_ATTEMPTS = 10
@set ATTEMPT = 0
@set FILE_FOUND = false

@while ${ATTEMPT} < ${MAX_ATTEMPTS} and not ${FILE_FOUND}
  @if exists("${TARGET_FILE}")
    @set FILE_FOUND = true
  @endif
  @set ATTEMPT = ${ATTEMPT} + 1  # ❌ 산술 연산 필요
  @sleep 1000
@endwhile
```

**현재 상태:** ❌ **불가능** (산술 연산 미지원)

**회피 방법:**
```bash
# 방법 1: @foreach 사용 (산술 연산 불필요)
@set FILE_FOUND = false

@foreach attempt in [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
  @if not ${FILE_FOUND}
    @if exists("${TARGET_FILE}")
      @set FILE_FOUND = true
    @else
      @sleep 1000
    @endif
  @endif
@endforeach
```

**평가:** ⚠️ 회피 가능하지만 불편함

---

### 시나리오 3: 조건 기반 루프 (산술 불필요)

**요구사항:**
```bash
@set HAS_NEXT_PAGE = true

@while ${HAS_NEXT_PAGE}
  @echo "Processing page..."

  # 페이지 처리

  @try
    oa plugin exec webauto element-click --session-id "${SID}" --element-selector ".next-page"
  @catch
    @set HAS_NEXT_PAGE = false
  @endtry
@endwhile
```

**현재 상태:** ✅ **가능** (산술 연산 불필요)

**평가:** 이런 패턴은 작동함

---

### 시나리오 4: 카운터 기반 반복

**요구사항:**
```bash
@set COUNTER = 0

@while ${COUNTER} < 10
  @echo "Iteration ${COUNTER}"
  @set COUNTER = ${COUNTER} + 1  # ❌ 산술 연산 필요
@endwhile
```

**현재 상태:** ❌ **불가능** (산술 연산 미지원)

**회피 방법:**
```bash
# @foreach 사용
@foreach counter in [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
  @echo "Iteration ${counter}"
@endforeach
```

**평가:** ⚠️ 회피 가능 (하지만 `@foreach`가 더 적합)

---

## 권장 사항

### 1. 산술 연산 지원 추가 (최우선) 🔥

**현재 상황:**
- `@while` 루프는 구현되었지만 실용적으로 사용 불가
- 5개 `examples/while_*` 스크립트 모두 실행 불가

**제안:**

**Phase 1: 최소 구현 (1-2일)**
```bash
@inc COUNTER        # COUNTER += 1
@dec COUNTER        # COUNTER -= 1
@inc COUNTER 5      # COUNTER += 5
@dec COUNTER 2      # COUNTER -= 2
```

**장점:**
- 빠른 구현 가능
- 가장 흔한 사용 사례(카운터 증가) 해결
- 기존 파서 큰 수정 불필요

**Phase 2: 산술 표현식 (3-5일)**
```bash
@set COUNTER = ${COUNTER} + 1
@set TOTAL = ${PRICE} * ${QUANTITY}
@set AVERAGE = ${SUM} / ${COUNT}
```

**장점:**
- 자연스러운 문법
- 다른 스크립트 언어와 일관성
- 복잡한 계산 가능

**구현 위치:**
- `pkg/batch/parser.go`: `parseSetDirective()` 수정
- `pkg/batch/variables.go`: 산술 표현식 평가기 추가

**파싱 로직:**
```go
func parseSetDirective(line string) (*SetDirective, error) {
    // "@set VAR = EXPR" 파싱
    parts := strings.SplitN(line, "=", 2)
    varName := strings.TrimSpace(strings.TrimPrefix(parts[0], "@set"))
    expr := strings.TrimSpace(parts[1])

    // 산술 연산자 감지
    if containsArithmeticOperator(expr) {
        value, err := evaluateArithmeticExpression(expr, variableScope)
        if err != nil {
            return nil, err
        }
        return &SetDirective{Variable: varName, Value: value}, nil
    }

    // 기존 로직 (문자열 할당)
    return &SetDirective{Variable: varName, Value: expr}, nil
}

func evaluateArithmeticExpression(expr string, vars map[string]string) (string, error) {
    // ${VAR} 치환
    expr = substituteVariables(expr, vars)

    // 산술 연산 평가
    // 예: "5 + 3" -> "8"
    result, err := evaluateExpression(expr)
    return result, err
}
```

**우선순위:** 🔥 **최우선**

### 2. 다중 라인 배열 검증 (1일)

**필요 작업:**
- 실제 테스트 스크립트 작성
- 동작 확인
- 엣지 케이스 검증 (빈 배열, 주석, 중첩 등)

**우선순위:** 중간

### 3. 복합 조건식 검증 (1일)

**필요 작업:**
- 괄호 그룹핑 테스트
- 연산자 우선순위 테스트
- 중첩 조건 테스트

**우선순위:** 중간

### 4. 문서 업데이트

**필요 작업:**
- `SHELL_USER_GUIDE.md` 업데이트
- `@while` 사용 예제 추가
- 제한사항 명시
- 회피 방법 문서화

**우선순위:** 낮음 (기능 완성 후)

---

## 결론

### 구현 상태 요약

| 기능 | 상태 | 실용성 | 비고 |
|------|------|--------|------|
| `@while` 구문 | ✅ 완료 | ❌ 낮음 | 산술 연산 없어 사용 제한적 |
| 에러 메시지 | ✅ 완료 | ✅ 높음 | 매우 유용함 |
| `@break`/`@continue` | ✅ 완료 | ⚠️ 미검증 | 테스트 필요 |
| 다중 라인 배열 | ✅ 완료? | ⚠️ 미검증 | 테스트 필요 |
| 복합 조건식 | ✅ 완료? | ⚠️ 미검증 | 테스트 필요 |
| **산술 연산** | ❌ **미구현** | 🔥 **필수** | **최우선 추가 필요** |

### 긴급 조치 필요

**1순위: 산술 연산 지원 추가**
- 현재 `@while` 루프가 실용적으로 사용 불가
- 최소 구현: `@inc`, `@dec` 지시어 (1-2일)
- 이상적 구현: `@set VAR = EXPR` 산술 표현식 (3-5일)

**2순위: 기능 검증**
- 다중 라인 배열 테스트
- 복합 조건식 테스트
- `@break`/`@continue` 테스트

**3순위: 예제 수정**
- `examples/while_*` 스크립트 수정
  - 산술 연산 구현 전: `@foreach` 회피 방법 사용
  - 산술 연산 구현 후: 원래 의도대로 수정
- webauto 확장 QA 세트 재테스트

---

## 제안: Codex에게 전달할 내용

### 📋 간략 요약

Issue #34 구현 감사드립니다. `@while` 루프 기본 구조는 잘 작동하지만, **산술 연산 미지원**으로 인해 실용적 사용이 불가능합니다.

**핵심 문제:**
```bash
@set COUNTER = ${COUNTER} + 1  # ❌ 에러: unexpected character '+'
```

**제안 방향:**
- ✅ **코드 간결성 우선** 원칙
- ✅ **네이티브 표현식 지원** (가장 간결함)
- ✅ **검증된 라이브러리 사용** (expr-lang/expr)

### 📄 상세 제안서

**문서:** `docs/OA_CLI_산술연산_제안서.md`

**주요 내용:**
1. **설계 원칙:** 코드 간결성 최우선 (간결성 비교 분석 포함)
2. **기술 제안:** expr-lang/expr 라이브러리 (govaluate 공식 후계자)
3. **구현 계획:** 3-5일 예상 (상세 코드 예제 포함)
4. **예상 효과:** 5개 while 예제 즉시 작동
5. **대안 검토:** @inc/@dec, 하이브리드, 순수 Lua (모두 기각, 이유 포함)

### 🎯 권장 구현

**네이티브 표현식 지원 (3-5일):**
```bash
@set COUNTER = ${COUNTER} + 1
@set TOTAL = ${PRICE} * ${QUANTITY} + ${TAX}
@if (${COUNT} > 0 and ${COUNT} < 10) or ${FORCE}
@while ${RETRY} < 3 and not ${SUCCESS}
```

**기술 스택:**
- `github.com/expr-lang/expr`: Google, Uber, ByteDance 사용 (6.1K stars)
- 메모리 안전, 부작용 없음, 항상 종료 보장
- 정적 타입 검증, 사용자 친화적 에러 메시지

### 간결성 비교 (핵심 근거)

| 접근 방식 | 재시도 로직 | OA 명령 | 학습 곡선 | 간결성 |
|----------|------------|---------|----------|--------|
| 하이브리드 (@lua) | 12줄 | 1줄 | 중간 | ⚠️ |
| 순수 Lua | 14줄 | 3-4줄 | 높음 | ⚠️ |
| **네이티브** | **11줄** ✅ | **1줄** ✅ | **낮음** ✅ | **최고** ✅ |

### 추가 검증 필요

1. 다중 라인 배열 실제 테스트
2. 복합 조건식 실제 테스트 (구현되었다고 명시되어 있으나 미검증)
3. `@break`/`@continue` 동작 확인

### 📞 다음 단계

1. **검토:** `docs/OA_CLI_산술연산_제안서.md` 문서 확인
2. **의사결정:** 구현 방향 및 일정 확정
3. **구현:** expr-lang/expr 통합 및 테스트
4. **검증:** webauto 팀이 5개 예제로 회귀 테스트 지원

---

**문서 버전:** 1.0
**마지막 업데이트:** 2025-10-21
**검증 환경:** macOS, OA CLI commit 7c7ab37
