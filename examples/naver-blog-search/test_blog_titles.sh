#!/bin/bash
# Test script for extracting Naver blog search results
# Demonstrates the use case from GitHub Issue #29

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== Naver Blog Search - Title Extraction Test ===${NC}\n"

# Build path to webauto binary
WEBAUTO="../../webauto"

if [ ! -f "$WEBAUTO" ]; then
    echo -e "${RED}Error: webauto binary not found${NC}"
    exit 1
fi

# Launch browser
echo -e "${YELLOW}1. Launching browser...${NC}"
launch_result=$($WEBAUTO browser-launch --headless true)
SESSION_ID=$(echo "$launch_result" | jq -r '.data.session_id')
echo -e "${GREEN}✓${NC} Session: $SESSION_ID\n"

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    $WEBAUTO browser-close --session-id "$SESSION_ID" >/dev/null 2>&1 || true
    echo -e "${GREEN}✓${NC} Done"
}
trap cleanup EXIT

# Navigate to Naver Blog search
SEARCH_QUERY="playwright"
SEARCH_URL="https://search.naver.com/search.naver?where=blog&query=${SEARCH_QUERY}"

echo -e "${YELLOW}2. Navigating to Naver Blog search...${NC}"
echo "   URL: $SEARCH_URL"

nav_result=$($WEBAUTO page-navigate \
    --page-url "$SEARCH_URL" \
    --session-id "$SESSION_ID")

if echo "$nav_result" | grep -q '"success": true'; then
    page_title=$(echo "$nav_result" | jq -r '.data.title')
    echo -e "${GREEN}✓${NC} Page loaded: $page_title\n"
else
    echo -e "${RED}✗${NC} Navigation failed"
    echo "$nav_result"
    exit 1
fi

# Wait for search results to load
echo -e "${YELLOW}3. Waiting for search results...${NC}"
sleep 2

# Try different selectors for blog titles (Naver structure may vary)
SELECTORS=(
    ".title_link"
    ".api_txt_lines"
    ".total_tit"
    "a.title_link"
    ".sh_blog_title"
)

echo -e "${YELLOW}4. Extracting blog titles...${NC}\n"

SUCCESS=false

for selector in "${SELECTORS[@]}"; do
    echo -e "${BLUE}Trying selector: $selector${NC}"

    result=$($WEBAUTO element-get-text \
        --session-id "$SESSION_ID" \
        --element-selector "$selector" 2>&1)

    if echo "$result" | grep -q '"success": true'; then
        text=$(echo "$result" | jq -r '.data.text')
        count=$(echo "$result" | jq -r '.data.element_count')

        echo -e "${GREEN}✓${NC} Found $count element(s)"

        # Check if we got multiple results (array)
        if echo "$text" | jq -e '. | type == "array"' >/dev/null 2>&1; then
            echo -e "${GREEN}✓${NC} Multiple blog titles found:"
            echo "$text" | jq -r '.[] | "  - " + .'
            SUCCESS=true
            break
        else
            echo -e "${YELLOW}!${NC} Single element (text: $text)"
            if [ "$count" -eq 1 ]; then
                echo "$text"
                SUCCESS=true
                break
            fi
        fi
    else
        echo -e "${RED}✗${NC} Selector failed"
    fi

    echo ""
done

# Take screenshot for debugging
echo -e "${YELLOW}5. Taking screenshot for reference...${NC}"
screenshot_path="naver_blog_search_result.png"

screenshot_result=$($WEBAUTO page-screenshot \
    --session-id "$SESSION_ID" \
    --image-path "$screenshot_path" \
    --full-page false)

if echo "$screenshot_result" | grep -q '"success": true'; then
    echo -e "${GREEN}✓${NC} Screenshot saved: $screenshot_path\n"
else
    echo -e "${YELLOW}!${NC} Screenshot failed (non-critical)\n"
fi

# Summary
echo -e "${BLUE}=== Test Summary ===${NC}"

if [ "$SUCCESS" = true ]; then
    echo -e "${GREEN}✓ Successfully extracted blog titles from Naver search${NC}"
    echo -e "${GREEN}✓ This demonstrates the use case from GitHub Issue #29${NC}"
    exit 0
else
    echo -e "${YELLOW}! Could not extract blog titles with tested selectors${NC}"
    echo -e "${YELLOW}! Naver's HTML structure may have changed${NC}"
    echo -e "${YELLOW}! Check the screenshot: $screenshot_path${NC}"
    exit 1
fi
