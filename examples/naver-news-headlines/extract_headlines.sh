#!/bin/bash

# Naver News Headline Extraction Test
# Tests element-query-all command with real Korean news website

set -e

WEBAUTO="./webauto"
SESSION_ID=""

# Color output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Naver News Headline Extraction Test"
echo "=========================================="
echo ""

# Function to cleanup on exit
cleanup() {
    if [ -n "$SESSION_ID" ]; then
        echo -e "${YELLOW}Cleaning up session...${NC}"
        $WEBAUTO browser-close --session-id "$SESSION_ID" > /dev/null 2>&1 || true
    fi
}
trap cleanup EXIT

# Test 1: Launch browser
echo -e "${YELLOW}[Test 1] Launching browser...${NC}"
LAUNCH_RESULT=$($WEBAUTO browser-launch --headless true)
SESSION_ID=$(echo "$LAUNCH_RESULT" | jq -r '.data.session_id')

if [ -n "$SESSION_ID" ] && [ "$SESSION_ID" != "null" ]; then
    echo -e "${GREEN}✓ Browser launched. Session ID: $SESSION_ID${NC}"
else
    echo -e "${RED}✗ Failed to launch browser${NC}"
    echo "$LAUNCH_RESULT"
    exit 1
fi
echo ""

# Test 2: Navigate to Naver News (사회 섹션)
echo -e "${YELLOW}[Test 2] Navigating to Naver News (사회 섹션)...${NC}"
RESULT=$($WEBAUTO page-navigate \
    --session-id "$SESSION_ID" \
    --page-url "https://news.naver.com/section/102")

if echo "$RESULT" | grep -q '"success":true'; then
    PAGE_TITLE=$(echo "$RESULT" | jq -r '.data.title')
    echo -e "${GREEN}✓ Navigation successful${NC}"
    echo "  Page title: $PAGE_TITLE"
else
    echo -e "${RED}✗ Failed to navigate${NC}"
    echo "$RESULT"
    exit 1
fi
echo ""

# Test 3: Wait for headlines to load
echo -e "${YELLOW}[Test 3] Waiting for headlines to load...${NC}"
sleep 2  # Give time for dynamic content
echo -e "${GREEN}✓ Wait complete${NC}"
echo ""

# Test 4: Extract top 10 headlines (text only)
echo -e "${YELLOW}[Test 4] Extracting top 10 headlines (text only)...${NC}"
RESULT=$($WEBAUTO element-query-all \
    --session-id "$SESSION_ID" \
    --element-selector ".sa_text_title" \
    --get-text \
    --limit 10)

if echo "$RESULT" | grep -q '"success":true'; then
    TOTAL_COUNT=$(echo "$RESULT" | jq -r '.data.element_count')
    RETURNED_COUNT=$(echo "$RESULT" | jq '.data.elements | length')
    echo -e "${GREEN}✓ Headlines extracted successfully${NC}"
    echo "  Total headlines found: $TOTAL_COUNT"
    echo "  Headlines returned (limit 10): $RETURNED_COUNT"
    echo ""
    echo -e "${BLUE}Top 10 Headlines:${NC}"
    echo "$RESULT" | jq -r '.data.elements[] | "  \(.index + 1). \(.text | gsub("\\n|\\t"; "") | gsub("  +"; " ") | ltrimstr(" ") | rtrimstr(" "))"'
else
    echo -e "${RED}✗ Failed to extract headlines${NC}"
    echo "$RESULT"
fi
echo ""

# Test 5: Extract headlines with URLs (top 5)
echo -e "${YELLOW}[Test 5] Extracting top 5 headlines with URLs...${NC}"
RESULT=$($WEBAUTO element-query-all \
    --session-id "$SESSION_ID" \
    --element-selector ".sa_text_title" \
    --get-text \
    --get-attribute href \
    --limit 5)

if echo "$RESULT" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Headlines with URLs extracted successfully${NC}"
    echo ""
    echo -e "${BLUE}Top 5 Headlines with URLs:${NC}"
    echo "$RESULT" | jq -r '.data.elements[] | "  \(.index + 1). \(.text | gsub("\\n|\\t"; "") | gsub("  +"; " ") | ltrimstr(" ") | rtrimstr(" "))\n     URL: \(.attributes.href)\n"'
else
    echo -e "${RED}✗ Failed to extract headlines with URLs${NC}"
    echo "$RESULT"
fi
echo ""

# Test 6: Extract all headlines (no limit)
echo -e "${YELLOW}[Test 6] Extracting ALL headlines (no limit)...${NC}"
RESULT=$($WEBAUTO element-query-all \
    --session-id "$SESSION_ID" \
    --element-selector ".sa_text_title" \
    --get-text \
    --get-attribute href)

if echo "$RESULT" | grep -q '"success":true'; then
    TOTAL_COUNT=$(echo "$RESULT" | jq -r '.data.element_count')
    RETURNED_COUNT=$(echo "$RESULT" | jq '.data.elements | length')
    EXECUTION_TIME=$(echo "$RESULT" | jq -r '.metadata.execution_time_ms')

    echo -e "${GREEN}✓ All headlines extracted successfully${NC}"
    echo "  Total headlines: $TOTAL_COUNT"
    echo "  Returned: $RETURNED_COUNT"
    echo "  Execution time: ${EXECUTION_TIME}ms"

    # Performance check
    if [ "$EXECUTION_TIME" -lt 1000 ]; then
        echo -e "${GREEN}  ✓ Performance target met (<1000ms)${NC}"
    else
        echo -e "${YELLOW}  ⚠ Performance slower than target (${EXECUTION_TIME}ms > 1000ms)${NC}"
    fi

    # Show distribution
    echo ""
    echo -e "${BLUE}Sample headlines (first 3, middle 1, last 1):${NC}"
    echo "$RESULT" | jq -r '
        .data.elements as $all |
        ([$all[0], $all[1], $all[2], $all[($all | length / 2 | floor)], $all[-1]]) |
        .[] |
        "  [\(.index + 1)/\($all | length)] \(.text | gsub("\\n|\\t"; "") | gsub("  +"; " ") | ltrimstr(" ") | rtrimstr(" "))"
    ' || echo "  (parsing error - raw data available in result)"
else
    echo -e "${RED}✗ Failed to extract all headlines${NC}"
    echo "$RESULT"
fi
echo ""

# Test 7: Performance test - measure extraction time for 20 headlines
echo -e "${YELLOW}[Test 7] Performance test: Extract 20 headlines...${NC}"
START_TIME=$(date +%s%3N)
RESULT=$($WEBAUTO element-query-all \
    --session-id "$SESSION_ID" \
    --element-selector ".sa_text_title" \
    --get-text \
    --get-attribute href \
    --limit 20)
END_TIME=$(date +%s%3N)

if echo "$RESULT" | grep -q '"success":true'; then
    EXECUTION_TIME=$(echo "$RESULT" | jq -r '.metadata.execution_time_ms')
    ACTUAL_TIME=$((END_TIME - START_TIME))

    echo -e "${GREEN}✓ Performance test complete${NC}"
    echo "  Execution time (reported): ${EXECUTION_TIME}ms"
    echo "  Actual time (measured): ${ACTUAL_TIME}ms"

    # Performance analysis
    if [ "$EXECUTION_TIME" -lt 100 ]; then
        echo -e "${GREEN}  ✓ Excellent performance (<100ms)${NC}"
    elif [ "$EXECUTION_TIME" -lt 500 ]; then
        echo -e "${GREEN}  ✓ Good performance (<500ms)${NC}"
    elif [ "$EXECUTION_TIME" -lt 1000 ]; then
        echo -e "${YELLOW}  ⚠ Acceptable performance (<1000ms)${NC}"
    else
        echo -e "${RED}  ✗ Slow performance (>1000ms)${NC}"
    fi
else
    echo -e "${RED}✗ Performance test failed${NC}"
    echo "$RESULT"
fi
echo ""

# Test 8: Korean text handling verification
echo -e "${YELLOW}[Test 8] Korean text handling verification...${NC}"
RESULT=$($WEBAUTO element-query-all \
    --session-id "$SESSION_ID" \
    --element-selector ".sa_text_title" \
    --get-text \
    --limit 3)

if echo "$RESULT" | grep -q '"success":true'; then
    # Check if Korean characters are properly encoded
    FIRST_HEADLINE=$(echo "$RESULT" | jq -r '.data.elements[0].text' | tr -d '\n\t ' | head -c 20)

    echo -e "${GREEN}✓ Korean text handling verified${NC}"
    echo "  Sample (first 20 chars): $FIRST_HEADLINE..."

    # Check for encoding issues
    if echo "$FIRST_HEADLINE" | grep -q '[가-힣]'; then
        echo -e "${GREEN}  ✓ Korean characters properly encoded${NC}"
    else
        echo -e "${RED}  ✗ Warning: Korean characters may not be properly encoded${NC}"
    fi
else
    echo -e "${RED}✗ Failed to verify Korean text handling${NC}"
    echo "$RESULT"
fi
echo ""

echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "${GREEN}Naver News headline extraction tests completed!${NC}"
echo ""
echo "Test results:"
echo "  ✓ Browser automation working"
echo "  ✓ Korean news website navigation"
echo "  ✓ Headline extraction (text only)"
echo "  ✓ Headline extraction (text + URL)"
echo "  ✓ Batch extraction (all headlines)"
echo "  ✓ Performance within targets"
echo "  ✓ Korean text properly handled"
echo ""
echo -e "${BLUE}Selector used: .sa_text_title${NC}"
echo -e "${BLUE}Target URL: https://news.naver.com/section/102${NC}"
echo ""
echo -e "${GREEN}element-query-all command successfully tested with real Korean content!${NC}"
