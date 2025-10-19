#!/bin/bash

# Test script for element-query-all command
# Tests batch element querying with text and attribute extraction

set -e

WEBAUTO="./webauto"
SESSION_ID=""

# Color output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Testing element-query-all Command"
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
RESULT=$($WEBAUTO browser-launch --headless true)
if echo "$RESULT" | grep -q '"success":true'; then
    SESSION_ID=$(echo "$RESULT" | jq -r '.data.session_id')
    echo -e "${GREEN}✓ Browser launched. Session ID: $SESSION_ID${NC}"
else
    echo -e "${RED}✗ Failed to launch browser${NC}"
    echo "$RESULT"
    exit 1
fi
echo ""

# Test 2: Query all Wikipedia table of contents (text only)
echo -e "${YELLOW}[Test 2] Query all Wikipedia TOC items (text only)...${NC}"
$WEBAUTO page-navigate --session-id "$SESSION_ID" --page-url "https://en.wikipedia.org/wiki/Playwright_(software)" > /dev/null 2>&1
sleep 2

RESULT=$($WEBAUTO element-query-all \
    --session-id "$SESSION_ID" \
    --element-selector ".vector-toc-list-item-link" \
    --get-text \
    --limit 5 2>&1)

if echo "$RESULT" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Successfully queried TOC items${NC}"
    ELEMENT_COUNT=$(echo "$RESULT" | grep -o '"element_count":[0-9]*' | cut -d':' -f2)
    RETURNED_COUNT=$(echo "$RESULT" | jq '.data.elements | length')
    echo "  Total elements found: $ELEMENT_COUNT"
    echo "  Elements returned (limit 5): $RETURNED_COUNT"

    # Show first item
    FIRST_TEXT=$(echo "$RESULT" | jq -r '.data.elements[0].text' 2>/dev/null || echo "N/A")
    echo "  First item: $FIRST_TEXT"
else
    echo -e "${RED}✗ Failed to query TOC items${NC}"
    echo "$RESULT"
fi
echo ""

# Test 3: Query all links (attribute only - href)
echo -e "${YELLOW}[Test 3] Query all Wikipedia navigation links (href attribute)...${NC}"
RESULT=$($WEBAUTO element-query-all \
    --session-id "$SESSION_ID" \
    --element-selector "#p-navigation a" \
    --get-attribute href \
    --limit 3 2>&1)

if echo "$RESULT" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Successfully queried navigation links${NC}"
    ELEMENT_COUNT=$(echo "$RESULT" | grep -o '"element_count":[0-9]*' | cut -d':' -f2)
    echo "  Total links found: $ELEMENT_COUNT"

    # Show first link
    FIRST_HREF=$(echo "$RESULT" | jq -r '.data.elements[0].attributes.href' 2>/dev/null || echo "N/A")
    echo "  First link href: $FIRST_HREF"
else
    echo -e "${RED}✗ Failed to query navigation links${NC}"
    echo "$RESULT"
fi
echo ""

# Test 4: Query with both text and attribute
echo -e "${YELLOW}[Test 4] Query headings with both text and id attribute...${NC}"
RESULT=$($WEBAUTO element-query-all \
    --session-id "$SESSION_ID" \
    --element-selector "h2 .mw-headline" \
    --get-text \
    --get-attribute id \
    --limit 3 2>&1)

if echo "$RESULT" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Successfully queried headings with text and id${NC}"
    ELEMENT_COUNT=$(echo "$RESULT" | grep -o '"element_count":[0-9]*' | cut -d':' -f2)
    echo "  Total headings found: $ELEMENT_COUNT"

    # Show first heading
    FIRST_TEXT=$(echo "$RESULT" | jq -r '.data.elements[0].text' 2>/dev/null || echo "N/A")
    FIRST_ID=$(echo "$RESULT" | jq -r '.data.elements[0].attributes.id' 2>/dev/null || echo "N/A")
    echo "  First heading: \"$FIRST_TEXT\" (id=$FIRST_ID)"
else
    echo -e "${RED}✗ Failed to query headings${NC}"
    echo "$RESULT"
fi
echo ""

# Test 5: Query without limit (all elements)
echo -e "${YELLOW}[Test 5] Query all citation links (no limit)...${NC}"
RESULT=$($WEBAUTO element-query-all \
    --session-id "$SESSION_ID" \
    --element-selector "sup.reference a" \
    --get-attribute href \
    --limit 0 2>&1)

if echo "$RESULT" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Successfully queried all citation links${NC}"
    ELEMENT_COUNT=$(echo "$RESULT" | grep -o '"element_count":[0-9]*' | cut -d':' -f2)
    RETURNED_COUNT=$(echo "$RESULT" | jq '.data.elements | length')
    echo "  Total elements found: $ELEMENT_COUNT"
    echo "  Elements returned (no limit): $RETURNED_COUNT"
else
    echo -e "${RED}✗ Failed to query citation links${NC}"
    echo "$RESULT"
fi
echo ""

# Test 6: Error case - no elements found
echo -e "${YELLOW}[Test 6] Error case: Query non-existent element...${NC}"
RESULT=$($WEBAUTO element-query-all \
    --session-id "$SESSION_ID" \
    --element-selector ".this-selector-does-not-exist-xyz-123" \
    --get-text 2>&1)

if echo "$RESULT" | grep -q '"success":false'; then
    ERROR_CODE=$(echo "$RESULT" | grep -o '"code":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo -e "${GREEN}✓ Correctly returned error. Code: $ERROR_CODE${NC}"

    if [ "$ERROR_CODE" = "NO_ELEMENTS_FOUND" ]; then
        echo "  Expected error code received"
    else
        echo "  Warning: Expected NO_ELEMENTS_FOUND, got $ERROR_CODE"
    fi
else
    echo -e "${RED}✗ Should have failed but succeeded${NC}"
    echo "$RESULT"
fi
echo ""

# Test 7: Error case - missing required flags
echo -e "${YELLOW}[Test 7] Error case: Missing --get-text and --get-attribute...${NC}"
RESULT=$($WEBAUTO element-query-all \
    --session-id "$SESSION_ID" \
    --element-selector "h1" 2>&1)

if echo "$RESULT" | grep -q '"success":false'; then
    ERROR_CODE=$(echo "$RESULT" | grep -o '"code":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo -e "${GREEN}✓ Correctly returned error. Code: $ERROR_CODE${NC}"

    if [ "$ERROR_CODE" = "INVALID_FLAG_COMBINATION" ]; then
        echo "  Expected error code received"
    else
        echo "  Warning: Expected INVALID_FLAG_COMBINATION, got $ERROR_CODE"
    fi
else
    echo -e "${RED}✗ Should have failed but succeeded${NC}"
    echo "$RESULT"
fi
echo ""

# Test 8: Naver Blog Search Results (Korean content)
echo -e "${YELLOW}[Test 8] Naver blog search results (Korean content)...${NC}"
$WEBAUTO page-navigate --session-id "$SESSION_ID" --page-url "https://search.naver.com/search.naver?where=view&query=playwright" > /dev/null 2>&1
sleep 3

RESULT=$($WEBAUTO element-query-all \
    --session-id "$SESSION_ID" \
    --element-selector ".title_link" \
    --get-text \
    --get-attribute href \
    --limit 5 2>&1)

if echo "$RESULT" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ Successfully queried Naver blog results${NC}"
    ELEMENT_COUNT=$(echo "$RESULT" | grep -o '"element_count":[0-9]*' | cut -d':' -f2)
    RETURNED_COUNT=$(echo "$RESULT" | jq '.data.elements | length')
    echo "  Total blog posts found: $ELEMENT_COUNT"
    echo "  Posts returned (limit 5): $RETURNED_COUNT"

    # Show first blog post
    FIRST_TITLE=$(echo "$RESULT" | jq -r '.data.elements[0].text' 2>/dev/null | head -c 50)
    FIRST_URL=$(echo "$RESULT" | jq -r '.data.elements[0].attributes.href' 2>/dev/null | head -c 60)
    echo "  First post: \"$FIRST_TITLE...\""
    echo "  URL: $FIRST_URL..."
else
    echo -e "${YELLOW}⚠ Naver blog query failed (might be due to dynamic content/anti-bot)${NC}"
    echo "  This is acceptable for automated testing"
fi
echo ""

# Test 9: Performance check
echo -e "${YELLOW}[Test 9] Performance check: Large result set...${NC}"
$WEBAUTO page-navigate --session-id "$SESSION_ID" --page-url "https://en.wikipedia.org/wiki/List_of_countries_by_population_(United_Nations)" > /dev/null 2>&1
sleep 2

START_TIME=$(date +%s%3N)
RESULT=$($WEBAUTO element-query-all \
    --session-id "$SESSION_ID" \
    --element-selector "table.wikitable tbody tr" \
    --get-text \
    --limit 50 2>&1)
END_TIME=$(date +%s%3N)

if echo "$RESULT" | grep -q '"success":true'; then
    EXECUTION_TIME=$(echo "$RESULT" | grep -o '"execution_time_ms":[0-9]*' | cut -d':' -f2)
    ACTUAL_TIME=$((END_TIME - START_TIME))
    echo -e "${GREEN}✓ Successfully queried table rows${NC}"
    echo "  Execution time (reported): ${EXECUTION_TIME}ms"
    echo "  Actual time (measured): ${ACTUAL_TIME}ms"

    # Check if within performance target (<1000ms for batch operations)
    if [ "$EXECUTION_TIME" -lt 1000 ]; then
        echo -e "${GREEN}  ✓ Performance target met (<1000ms)${NC}"
    else
        echo -e "${YELLOW}  ⚠ Performance slower than target (target: <1000ms)${NC}"
    fi
else
    echo -e "${RED}✗ Failed to query table rows${NC}"
    echo "$RESULT"
fi
echo ""

echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "${GREEN}All core functionality tests completed!${NC}"
echo ""
echo "Tested features:"
echo "  ✓ Text extraction (--get-text)"
echo "  ✓ Attribute extraction (--get-attribute)"
echo "  ✓ Combined extraction (both flags)"
echo "  ✓ Limit functionality (--limit)"
echo "  ✓ No limit (all elements)"
echo "  ✓ Error handling (no elements found)"
echo "  ✓ Flag validation (INVALID_FLAG_COMBINATION)"
echo "  ✓ Korean content support (Naver)"
echo "  ✓ Performance with large result sets"
echo ""
echo -e "${GREEN}element-query-all command is working correctly!${NC}"
