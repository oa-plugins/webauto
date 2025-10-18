#!/bin/bash

# ============================================
# 네이버 지도 플레이스 검색 자동화
# ============================================
#
# 목적: webauto 명령어만 사용하여 네이버 지도에서 플레이스 검색 수행
#
# 제한사항:
# - 현재 webauto는 텍스트 추출 기능이 없어 데이터 수집 불가
# - 스크린샷과 PDF 저장만 가능
# - 개선사항은 ../../IMPROVEMENTS.md 참고
#
# ============================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WEBAUTO="$SCRIPT_DIR/../../webauto"
QUERY="${1:-강남역 맛집}"
SCREENSHOT_DIR="$SCRIPT_DIR/screenshots"

# 색상 출력
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "============================================"
echo "  네이버 지도 플레이스 검색 자동화"
echo "============================================"
echo ""
echo "검색어: $QUERY"
echo ""

# 1. 브라우저 시작
echo -e "${YELLOW}[1/6]${NC} 브라우저 시작 중..."
SESSION_ID=$($WEBAUTO browser-launch \
  --headless false \
  2>&1 | jq -r '.data.session_id')

if [ -z "$SESSION_ID" ] || [ "$SESSION_ID" == "null" ]; then
  echo -e "${RED}✗ 브라우저 시작 실패${NC}"
  exit 1
fi

echo -e "${GREEN}✓ 세션 ID: $SESSION_ID${NC}"
echo ""

# 2. 네이버 지도 접속
echo -e "${YELLOW}[2/6]${NC} 네이버 지도 접속 중..."
$WEBAUTO page-navigate \
  --session-id "$SESSION_ID" \
  --page-url "https://map.naver.com/v5/" \
  > /dev/null 2>&1

echo -e "${GREEN}✓ 페이지 로드 완료${NC}"
sleep 3

# 3. 초기 화면 캡처
echo -e "${YELLOW}[3/6]${NC} 초기 화면 캡처 중..."
$WEBAUTO page-screenshot \
  --session-id "$SESSION_ID" \
  --image-path "$SCREENSHOT_DIR/01_map_main.png" \
  > /dev/null 2>&1

echo -e "${GREEN}✓ 스크린샷 저장: 01_map_main.png${NC}"
echo ""

# 4. 검색어 입력
echo -e "${YELLOW}[4/6]${NC} 검색어 입력 중: \"$QUERY\"..."

# 네이버 지도 검색창 셀렉터 (2025년 기준, 변경될 수 있음)
SEARCH_INPUT_RESULT=$($WEBAUTO element-type \
  --session-id "$SESSION_ID" \
  --element-selector "input.input_search" \
  --text-input "$QUERY" \
  2>&1)

if echo "$SEARCH_INPUT_RESULT" | jq -e '.success' > /dev/null 2>&1; then
  echo -e "${GREEN}✓ 검색어 입력 완료${NC}"
else
  echo -e "${RED}✗ 검색어 입력 실패${NC}"
  echo "셀렉터가 변경되었을 수 있습니다. 네이버 지도 페이지 구조를 확인하세요."
  $WEBAUTO session-close --session-id "$SESSION_ID" > /dev/null 2>&1
  exit 1
fi

sleep 1

# 5. 검색 버튼 클릭
echo -e "${YELLOW}[5/6]${NC} 검색 실행 중..."

SEARCH_BTN_RESULT=$($WEBAUTO element-click \
  --session-id "$SESSION_ID" \
  --element-selector "button.btn_search" \
  2>&1)

if echo "$SEARCH_BTN_RESULT" | jq -e '.success' > /dev/null 2>&1; then
  echo -e "${GREEN}✓ 검색 실행 완료${NC}"
else
  echo -e "${RED}✗ 검색 버튼 클릭 실패${NC}"
  $WEBAUTO session-close --session-id "$SESSION_ID" > /dev/null 2>&1
  exit 1
fi

sleep 4
echo ""

# 6. 플레이스 검색 결과 캡처
echo -e "${YELLOW}[6/6]${NC} 검색 결과 캡처 중..."

# 스크린샷 (전체 페이지)
SCREENSHOT_RESULT=$($WEBAUTO page-screenshot \
  --session-id "$SESSION_ID" \
  --image-path "$SCREENSHOT_DIR/02_place_results.png" \
  --full-page \
  2>&1)

if echo "$SCREENSHOT_RESULT" | jq -e '.success' > /dev/null 2>&1; then
  WIDTH=$(echo "$SCREENSHOT_RESULT" | jq -r '.data.image_width')
  HEIGHT=$(echo "$SCREENSHOT_RESULT" | jq -r '.data.image_height')
  SIZE=$(echo "$SCREENSHOT_RESULT" | jq -r '.data.file_size')
  echo -e "${GREEN}✓ 스크린샷 저장: 02_place_results.png${NC}"
  echo "  해상도: ${WIDTH}x${HEIGHT} px"
  echo "  파일 크기: $SIZE bytes"
else
  echo -e "${RED}✗ 스크린샷 저장 실패${NC}"
fi

# PDF 저장
PDF_RESULT=$($WEBAUTO page-pdf \
  --session-id "$SESSION_ID" \
  --pdf-path "$SCREENSHOT_DIR/02_place_results.pdf" \
  2>&1)

if echo "$PDF_RESULT" | jq -e '.success' > /dev/null 2>&1; then
  echo -e "${GREEN}✓ PDF 저장: 02_place_results.pdf${NC}"
else
  echo -e "${RED}✗ PDF 저장 실패${NC}"
fi

echo ""

# 7. 세션 종료
echo "브라우저 세션 종료 중..."
$WEBAUTO session-close --session-id "$SESSION_ID" > /dev/null 2>&1
echo -e "${GREEN}✓ 세션 종료 완료${NC}"
echo ""

# 결과 요약
echo "============================================"
echo "  검색 완료"
echo "============================================"
echo ""
echo -e "${GREEN}✅ 성공한 작업:${NC}"
echo "  - 네이버 지도 접속"
echo "  - 검색어 입력"
echo "  - 검색 실행"
echo "  - 스크린샷 캡처 (PNG, PDF)"
echo ""
echo -e "${RED}❌ 실패한 작업:${NC}"
echo "  - 플레이스 이름/주소/평점 데이터 수집"
echo "  - 필요한 명령: element-get-text, element-get-attribute"
echo "  - 필요한 명령: element-query-all (여러 플레이스 일괄 조회)"
echo "  - 개선사항: ../../IMPROVEMENTS.md 참고"
echo ""
echo -e "${YELLOW}📁 결과 파일:${NC}"
echo "  - $SCREENSHOT_DIR/01_map_main.png"
echo "  - $SCREENSHOT_DIR/02_place_results.png"
echo "  - $SCREENSHOT_DIR/02_place_results.pdf"
echo ""
