# webauto 개선사항

> **작성일**: 2025-10-18
> **작성 이유**: 네이버 검색 자동화 예제 작성 중 발견된 필수 기능 부족

---

## 📋 개요

현재 webauto는 브라우저 제어와 스크린샷/PDF 저장은 가능하지만, **데이터 추출 기능**이 없어 실제 자동화 활용에 한계가 있습니다.

**예제 작업 중 발견된 주요 문제**:
- 네이버 블로그 검색 결과의 제목/URL 추출 불가
- 네이버 지도 플레이스 정보 (상호명, 평점, 주소) 추출 불가
- 여러 검색 결과를 반복 처리할 방법 없음

---

## 🚨 우선순위 1: 필수 데이터 추출 명령어

### 1. `element-get-text` - 요소 텍스트 추출

**현재 상황**:
- `element-click`, `element-type`은 있지만 텍스트 **읽기** 불가
- 스크린샷으로 시각적 확인만 가능

**필요한 기능**:
```bash
webauto element-get-text \
  --session-id ses_abc123 \
  --element-selector ".blog-title"
```

**기대 출력**:
```json
{
  "success": true,
  "data": {
    "session_id": "ses_abc123",
    "element_selector": ".blog-title",
    "text": "Playwright로 웹 자동화하기",
    "element_count": 1
  },
  "error": null,
  "metadata": {
    "plugin": "webauto",
    "version": "1.0.0",
    "execution_time_ms": 15
  }
}
```

**사용 사례**:
- 블로그 제목 수집
- 플레이스 상호명 추출
- 검색 결과 개수 확인
- 에러 메시지 읽기

**Playwright 구현 참고**:
```javascript
const text = await page.locator('.blog-title').textContent();
```

---

### 2. `element-get-attribute` - 요소 속성 추출

**현재 상황**:
- 링크의 `href`, 이미지의 `src` 등 속성값 추출 불가

**필요한 기능**:
```bash
webauto element-get-attribute \
  --session-id ses_abc123 \
  --element-selector "a.blog-link" \
  --attribute-name "href"
```

**기대 출력**:
```json
{
  "success": true,
  "data": {
    "session_id": "ses_abc123",
    "element_selector": "a.blog-link",
    "attribute_name": "href",
    "attribute_value": "https://blog.naver.com/example",
    "element_count": 1
  },
  "error": null,
  "metadata": {
    "plugin": "webauto",
    "version": "1.0.0",
    "execution_time_ms": 12
  }
}
```

**사용 사례**:
- 블로그 URL 수집
- 이미지 src 추출
- data-* 속성 읽기
- aria-label 접근성 정보 추출

**Playwright 구현 참고**:
```javascript
const href = await page.locator('a.blog-link').getAttribute('href');
```

---

## 🔄 우선순위 2: 다중 요소 처리

### 3. `element-query-all` - 여러 요소 일괄 조회

**현재 상황**:
- 현재는 단일 요소만 처리 가능
- 검색 결과 10개를 순회하려면 10번 호출 필요 (비효율적)

**필요한 기능**:
```bash
webauto element-query-all \
  --session-id ses_abc123 \
  --element-selector ".blog-item" \
  --get-text \
  --get-attribute href \
  --limit 10
```

**기대 출력**:
```json
{
  "success": true,
  "data": {
    "session_id": "ses_abc123",
    "element_selector": ".blog-item",
    "element_count": 10,
    "elements": [
      {
        "index": 0,
        "text": "Playwright 자동화 가이드",
        "attributes": {
          "href": "https://blog.naver.com/example1"
        }
      },
      {
        "index": 1,
        "text": "웹 테스팅 자동화",
        "attributes": {
          "href": "https://blog.naver.com/example2"
        }
      }
    ]
  },
  "error": null,
  "metadata": {
    "plugin": "webauto",
    "version": "1.0.0",
    "execution_time_ms": 45
  }
}
```

**사용 사례**:
- 검색 결과 목록 수집
- 플레이스 목록 크롤링
- 테이블 데이터 추출
- 메뉴 항목 수집

**Playwright 구현 참고**:
```javascript
const elements = await page.locator('.blog-item').all();
const results = await Promise.all(
  elements.map(async (el, idx) => ({
    index: idx,
    text: await el.textContent(),
    href: await el.getAttribute('href')
  }))
);
```

---

## ⏱️ 우선순위 3: 동적 페이지 대응

### 4. `element-wait` - 요소 대기

**현재 상황**:
- 스크립트에서 `sleep 2` 같은 고정 대기 사용
- 페이지 로딩 속도에 따라 실패 가능

**필요한 기능**:
```bash
webauto element-wait \
  --session-id ses_abc123 \
  --element-selector ".search-results" \
  --wait-for visible \
  --timeout 5000
```

**기대 출력**:
```json
{
  "success": true,
  "data": {
    "session_id": "ses_abc123",
    "element_selector": ".search-results",
    "wait_condition": "visible",
    "waited_ms": 1234,
    "element_found": true
  },
  "error": null,
  "metadata": {
    "plugin": "webauto",
    "version": "1.0.0",
    "execution_time_ms": 1250
  }
}
```

**대기 조건**:
- `visible` - 요소가 화면에 보일 때까지
- `hidden` - 요소가 사라질 때까지
- `attached` - 요소가 DOM에 추가될 때까지
- `detached` - 요소가 DOM에서 제거될 때까지

**사용 사례**:
- AJAX 로딩 후 결과 대기
- 모달 팝업 표시 대기
- 로딩 스피너 사라질 때까지 대기

**Playwright 구현 참고**:
```javascript
await page.locator('.search-results').waitFor({ state: 'visible', timeout: 5000 });
```

---

## 📊 우선순위 4: 페이지 정보 조회

### 5. `page-get-html` - HTML 소스 가져오기

**현재 상황**:
- 페이지 소스 추출 불가
- DOM 구조 분석 불가

**필요한 기능**:
```bash
webauto page-get-html \
  --session-id ses_abc123 \
  --selector "#content" \
  --output-path page_source.html
```

**기대 출력**:
```json
{
  "success": true,
  "data": {
    "session_id": "ses_abc123",
    "html_length": 15234,
    "output_path": "page_source.html"
  },
  "error": null,
  "metadata": {
    "plugin": "webauto",
    "version": "1.0.0",
    "execution_time_ms": 8
  }
}
```

**사용 사례**:
- 페이지 구조 분석
- 오프라인 아카이빙
- DOM 디버깅

---

### 6. `page-evaluate` - JavaScript 실행 및 결과 반환

**현재 상황**:
- 사용자 정의 JavaScript 실행 불가
- 복잡한 데이터 추출 불가

**필요한 기능**:
```bash
webauto page-evaluate \
  --session-id ses_abc123 \
  --script "document.querySelectorAll('.blog-item').length"
```

**기대 출력**:
```json
{
  "success": true,
  "data": {
    "session_id": "ses_abc123",
    "result": 10,
    "result_type": "number"
  },
  "error": null,
  "metadata": {
    "plugin": "webauto",
    "version": "1.0.0",
    "execution_time_ms": 5
  }
}
```

**사용 사례**:
- 복잡한 데이터 구조 추출
- 페이지 내 변수 읽기
- DOM 조작 및 검증

---

## 🎯 구현 우선순위 요약

| 우선순위 | 명령어 | 중요도 | 난이도 | 비고 |
|---------|--------|--------|--------|------|
| **1** | `element-get-text` | 🔴 필수 | 하 | 가장 기본적인 데이터 추출 |
| **1** | `element-get-attribute` | 🔴 필수 | 하 | URL, 속성 추출 필수 |
| **2** | `element-query-all` | 🟡 중요 | 중 | 효율적 데이터 수집 |
| **3** | `element-wait` | 🟡 중요 | 중 | 안정적 자동화 |
| **4** | `page-get-html` | 🟢 선택 | 하 | 디버깅/분석용 |
| **4** | `page-evaluate` | 🟢 선택 | 상 | 고급 사용자용 |

---

## 💡 구현 제안

### Phase 1: 기본 데이터 추출 (1-2주)
- `element-get-text` 구현
- `element-get-attribute` 구현
- 예제 스크립트 업데이트 (실제 데이터 수집 가능)

### Phase 2: 효율성 개선 (2-3주)
- `element-query-all` 구현
- `element-wait` 구현
- 성능 최적화

### Phase 3: 고급 기능 (선택적)
- `page-get-html` 구현
- `page-evaluate` 구현
- 에러 처리 개선

---

## 📖 관련 예제

현재 제한된 기능으로 작성된 예제:
- `examples/naver-blog-search/` - 블로그 검색 (스크린샷만 가능)
- `examples/naver-map-search/` - 지도 검색 (스크린샷만 가능)

개선 후 가능한 작업:
- 블로그 제목/URL 10개 JSON 파일로 저장
- 플레이스 정보 (이름, 평점, 주소) 구조화된 데이터로 추출
- 검색 결과를 데이터베이스에 저장하는 워크플로우

---

## 🔗 참고 자료

- [Playwright Locator API](https://playwright.dev/docs/api/class-locator)
- [Playwright 요소 조회](https://playwright.dev/docs/locators)
- [Playwright 대기 전략](https://playwright.dev/docs/actionability)

---

**작성자**: Claude Code
**검토 필요**: webauto 개발팀
**업데이트**: 구현 완료 시 이 문서 업데이트 필요
