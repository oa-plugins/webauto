# webauto 자동화 예제 모음

webauto 플러그인을 사용한 실제 웹 자동화 스크립트 예제 모음입니다.

## 📁 디렉토리 구조

```
examples/
├── README.md                              # 이 문서
├── IMPROVEMENTS.md                        # webauto 개선사항 목록
├── basic/                                 # 기본 예제
│   ├── web_scraping.sh                   # 웹 스크래핑 기본
│   ├── multi_site_crawler.sh             # 멀티 사이트 크롤러
│   └── sites_config.json                 # 멀티 사이트 설정 예시
├── naver-blog-search/                    # 네이버 블로그 검색
│   ├── search_blogs.sh                   # 블로그 검색 자동화
│   └── screenshots/                      # 스크린샷 저장 폴더
├── naver-map-search/                     # 네이버 지도 검색
│   ├── search_places.sh                  # 플레이스 검색 자동화
│   └── screenshots/                      # 스크린샷 저장 폴더
├── hometax/                              # 홈택스 자동화
│   └── tax_invoice_query.sh              # 세금계산서 조회
└── wehago/                               # 위하고 자동화
    └── accounting_data_export.sh         # 회계 데이터 조회
```

---

## 🚀 Quick Start

### 사전 준비

1. **webauto 빌드**
   ```bash
   cd ../..  # webauto 루트 디렉토리로 이동
   go build -o webauto cmd/webauto/main.go
   ```

2. **Node.js 및 Playwright 설치**
   ```bash
   npm install
   npx playwright install chromium
   ```

3. **예제 스크립트 실행 권한 부여**
   ```bash
   chmod +x examples/**/*.sh
   ```

---

## 📖 예제 목록

### 1. 기본 웹 스크래핑 (`basic/web_scraping.sh`)

**목적**: 웹사이트에 접속하여 스크린샷과 PDF를 저장합니다.

**사용법**:
```bash
cd examples/basic
./web_scraping.sh
```

**수행 작업**:
- ✅ 브라우저 실행 (headless 모드)
- ✅ example.com 접속
- ✅ 페이지 스크린샷 저장
- ✅ 페이지 PDF 저장
- ✅ 브라우저 종료

**출력 파일**:
- `output/example_screenshot.png`
- `output/example_page.pdf`

---

### 2. 멀티 사이트 크롤러 (`basic/multi_site_crawler.sh`)

**목적**: JSON 설정 파일에 정의된 여러 웹사이트를 순회하며 데이터를 수집합니다.

**사용법**:
```bash
cd examples/basic
./multi_site_crawler.sh sites_config.json
```

**수행 작업**:
- ✅ 설정 파일에서 사이트 목록 읽기
- ✅ 각 사이트 순회
- ✅ 스크린샷 및 PDF 저장
- ✅ 크롤링 리포트 생성

**설정 파일 형식** (`sites_config.json`):
```json
{
  "sites": [
    {
      "name": "example",
      "url": "https://example.com",
      "wait_seconds": 2
    }
  ]
}
```

**출력 파일**:
- `output/batch_YYYYMMDD_HHMMSS/`
  - `{site_name}_screenshot.png`
  - `{site_name}_page.pdf`
  - `crawl_report.json`

---

### 3. 네이버 모바일 블로그 검색 (`naver-blog-search/search_blogs.sh`)

**목적**: 네이버 모바일에서 블로그를 검색하고 결과 화면을 캡처합니다.

**사용법**:
```bash
cd examples/naver-blog-search
./search_blogs.sh "검색어"
```

**수행 작업**:
- ✅ 브라우저 실행 (모바일 User-Agent)
- ✅ 네이버 모바일 접속
- ✅ 검색어 입력
- ✅ 검색 실행
- ✅ 블로그 탭 이동
- ✅ 스크린샷 캡처 (PNG, PDF)
- ❌ **블로그 제목/URL 데이터 수집** (현재 불가능)

**출력 파일**:
- `screenshots/01_main_page.png` - 네이버 메인 페이지
- `screenshots/02_blog_results.png` - 블로그 검색 결과 (전체 페이지)
- `screenshots/02_blog_results.pdf` - 검색 결과 PDF

**제한사항**:
- 현재 webauto는 텍스트 추출 기능(`element-get-text`)이 없어 **데이터 수집 불가**
- 스크린샷과 PDF로만 결과 저장 가능
- 개선사항은 `IMPROVEMENTS.md` 참고

---

### 4. 네이버 지도 플레이스 검색 (`naver-map-search/search_places.sh`)

**목적**: 네이버 지도에서 플레이스를 검색하고 결과 화면을 캡처합니다.

**사용법**:
```bash
cd examples/naver-map-search
./search_places.sh "강남역 맛집"
```

**수행 작업**:
- ✅ 브라우저 실행
- ✅ 네이버 지도 접속
- ✅ 검색어 입력
- ✅ 검색 실행
- ✅ 스크린샷 캡처 (PNG, PDF)
- ❌ **플레이스 정보 데이터 수집** (현재 불가능)

**출력 파일**:
- `screenshots/01_map_main.png` - 네이버 지도 메인
- `screenshots/02_place_results.png` - 플레이스 검색 결과 (전체 페이지)
- `screenshots/02_place_results.pdf` - 검색 결과 PDF

**제한사항**:
- 현재 webauto는 텍스트/속성 추출 기능이 없어 **데이터 수집 불가**
- 필요한 명령어: `element-get-text`, `element-get-attribute`, `element-query-all`
- 개선사항은 `IMPROVEMENTS.md` 참고

---

### 5. 홈택스 세금계산서 조회 (`hometax/tax_invoice_query.sh`)

**목적**: 홈택스에서 세금계산서를 조회하고 결과를 저장합니다.

**사용법**:
```bash
cd examples/hometax
./tax_invoice_query.sh "123-45-67890" "20250101" "20250131"
```

**인자**:
1. `사업자등록번호`: 조회할 사업자등록번호 (예: "123-45-67890")
2. `시작일자`: 조회 시작 날짜 (형식: YYYYMMDD)
3. `종료일자`: 조회 종료 날짜 (형식: YYYYMMDD)

**수행 작업**:
1. ✅ 브라우저 실행 (headless=false, 로그인 필요)
2. ✅ 홈택스 접속
3. ⚠️ **수동 로그인 필요** (공동인증서 또는 간편인증)
4. ✅ 세금계산서 조회 메뉴 이동
5. ✅ 조회 조건 입력 (사업자번호, 기간)
6. ✅ 조회 실행
7. ✅ 결과 스크린샷 및 PDF 저장
8. ✅ 엑셀 다운로드 시도

**출력 파일**:
- `output/hometax_result_YYYYMMDD_HHMMSS.png`
- `output/hometax_result_YYYYMMDD_HHMMSS.pdf`
- 다운로드 폴더에 엑셀 파일 (브라우저 기본 경로)

**주의사항**:
- 로그인은 수동으로 진행해야 합니다
- 실제 셀렉터는 홈택스 페이지 구조에 따라 조정이 필요할 수 있습니다
- 보안을 위해 브라우저를 열어둡니다 (수동 종료 또는 세션 ID로 종료)

---

### 4. 위하고 회계 데이터 조회 (`wehago/accounting_data_export.sh`)

**목적**: 위하고에서 회계 장부 데이터를 조회하고 결과를 저장합니다.

**사용법**:
```bash
cd examples/wehago
./accounting_data_export.sh "COMP001" "2025-01-01" "2025-01-31"
```

**인자**:
1. `회사코드`: 조회할 회사 코드 (예: "COMP001")
2. `시작일자`: 조회 시작 날짜 (형식: YYYY-MM-DD)
3. `종료일자`: 조회 종료 날짜 (형식: YYYY-MM-DD)

**수행 작업**:
1. ✅ 브라우저 실행 (headless=false, 로그인 필요)
2. ✅ 위하고 접속
3. ⚠️ **수동 로그인 필요** (ID/PW 또는 간편 로그인)
4. ✅ 회계 메뉴 이동
5. ✅ 회사 선택
6. ✅ 조회 기간 설정
7. ✅ 조회 실행
8. ✅ 결과 스크린샷 및 PDF 저장
9. ✅ 엑셀 다운로드 시도

**출력 파일**:
- `output/wehago_ledger_YYYYMMDD_HHMMSS.png`
- `output/wehago_ledger_YYYYMMDD_HHMMSS.pdf`
- 다운로드 폴더에 엑셀 파일 (브라우저 기본 경로)

**주의사항**:
- 로그인은 수동으로 진행해야 합니다
- 실제 셀렉터는 위하고 페이지 구조에 따라 조정이 필요할 수 있습니다
- 회사 코드는 실제 위하고 계정의 회사 코드를 사용해야 합니다

---

## 🔧 고급 활용

### 스크립트 커스터마이징

각 예제 스크립트를 자신의 요구사항에 맞게 수정할 수 있습니다:

**1. 셀렉터 변경**
```bash
# 원본
--element-selector "#searchBtn"

# 수정 (실제 페이지의 셀렉터로 변경)
--element-selector ".search-button"
```

**2. 대기 시간 조정**
```bash
sleep 3  # 페이지 로딩 대기 시간 조정
```

**3. 출력 경로 변경**
```bash
OUTPUT_DIR="./my_output"
```

### 에러 처리 추가

```bash
# 명령어 실패 시 계속 진행
$WEBAUTO element-click \
    --session-id "$SESSION_ID" \
    --element-selector "#btn" \
    2>/dev/null || echo "⚠️  버튼 클릭 실패 - 계속 진행"

# 명령어 실패 시 스크립트 종료
$WEBAUTO element-click \
    --session-id "$SESSION_ID" \
    --element-selector "#btn" || {
        echo "❌ 버튼 클릭 실패 - 종료"
        $WEBAUTO browser-close --session-id "$SESSION_ID"
        exit 1
    }
```

---

## 🎯 실전 활용 시나리오

### 1. 일일 세금계산서 자동 조회

**cron 작업으로 등록**:
```bash
# 매일 오전 9시에 전날 세금계산서 조회
0 9 * * * cd /path/to/examples/hometax && ./tax_invoice_query.sh "123-45-67890" "$(date -d yesterday +\%Y\%m\%d)" "$(date -d yesterday +\%Y\%m\%d)"
```

### 2. 월말 회계 데이터 자동 수집

**월말 자동 실행**:
```bash
# 매월 말일 오후 6시에 당월 회계 데이터 조회
0 18 28-31 * * cd /path/to/examples/wehago && [ "$(date -d tomorrow +\%d)" -eq "01" ] && ./accounting_data_export.sh "COMP001" "$(date +\%Y-\%m-01)" "$(date +\%Y-\%m-\%d)"
```

### 3. 멀티 사이트 모니터링

**주기적 웹사이트 모니터링**:
```bash
# 1시간마다 주요 사이트 상태 체크
0 * * * * cd /path/to/examples/basic && ./multi_site_crawler.sh sites_config.json
```

---

## ⚠️ 주의사항

### 보안

1. **인증 정보 관리**
   - 스크립트에 패스워드를 하드코딩하지 마세요
   - 환경 변수 또는 별도 설정 파일 사용 권장
   ```bash
   # 환경 변수 사용 예시
   export HOMETAX_USER_ID="your_id"
   export HOMETAX_PASSWORD="your_password"
   ```

2. **세션 관리**
   - 작업 완료 후 반드시 브라우저 세션을 종료하세요
   - 민감한 정보가 캐시에 남지 않도록 주의하세요

### 법적 준수

1. **서비스 약관 확인**
   - 자동화를 허용하는지 각 서비스의 이용약관을 확인하세요
   - robots.txt 파일을 존중하세요

2. **개인 정보 보호**
   - 본인 또는 권한이 있는 계정만 사용하세요
   - 수집한 데이터의 보관 및 관리에 주의하세요

3. **서버 부하**
   - 과도한 요청으로 서버에 부하를 주지 마세요
   - 적절한 대기 시간을 설정하세요

### 안티봇 대응

1. **User-Agent 설정**
   ```bash
   # webauto는 기본적으로 User-Agent를 설정합니다
   # 필요시 환경 변수로 커스터마이징 가능
   export ENABLE_STEALTH=true
   export ENABLE_FINGERPRINT=true
   ```

2. **요청 간격 조정**
   ```bash
   # 각 요청 사이에 충분한 대기 시간 추가
   sleep $((RANDOM % 5 + 2))  # 2-7초 랜덤 대기
   ```

---

## 🐛 트러블슈팅

### 문제: "세션을 찾을 수 없습니다"

**원인**: 브라우저 세션이 만료되었거나 종료됨

**해결**:
```bash
# 세션 목록 확인
../../webauto session-list

# 새 브라우저 실행
../../webauto browser-launch --headless false
```

### 문제: "요소를 찾을 수 없습니다"

**원인**: 페이지 구조 변경 또는 로딩 시간 부족

**해결**:
1. 대기 시간 증가
2. 셀렉터 확인 및 업데이트
3. 브라우저 개발자 도구로 실제 셀렉터 확인

### 문제: "Node.js를 찾을 수 없습니다"

**원인**: Node.js가 설치되지 않았거나 PATH에 없음

**해결**:
```bash
# Node.js 설치 확인
node --version

# PATH 설정 확인
which node
```

---

## 📚 추가 리소스

- [webauto 아키텍처 문서](../ARCHITECTURE.md)
- [webauto 성능 최적화 문서](../PERFORMANCE.md)
- [Playwright 공식 문서](https://playwright.dev/)

---

## 🤝 기여

예제 스크립트 개선이나 새로운 예제 추가는 언제나 환영합니다!

1. Fork the repository
2. Create your feature branch
3. Add your example script
4. Update this README
5. Submit a pull request

---

## 📝 라이선스

이 예제들은 개인 사용 목적으로 제공됩니다. 상업적 사용 시 각 서비스의 이용약관을 확인하세요.

**면책 조항**: 이 스크립트들은 예제 목적으로만 제공됩니다. 사용자는 각 서비스의 이용약관 준수 및 법적 책임을 스스로 부담합니다.
