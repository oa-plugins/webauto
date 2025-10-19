#!/bin/bash
# Test script for element-wait command across multiple scenarios
# This demonstrates element waiting for AJAX/dynamic content scenarios

set -e  # Exit on error

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== element-wait Multi-Scenario Test Suite ===${NC}\n"

# Build path to webauto binary
WEBAUTO="../../webauto"

if [ ! -f "$WEBAUTO" ]; then
    echo -e "${RED}Error: webauto binary not found at $WEBAUTO${NC}"
    echo "Please run: go build -o webauto cmd/webauto/main.go"
    exit 1
fi

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Function to run a test
run_test() {
    local test_name="$1"
    local url="$2"
    local selector="$3"
    local wait_condition="$4"
    local timeout="$5"
    local expect_success="$6"

    echo -e "${YELLOW}Test: ${test_name}${NC}"
    echo "  URL: $url"
    echo "  Selector: $selector"
    echo "  Wait Condition: $wait_condition"
    echo "  Timeout: ${timeout}ms"

    # Navigate to URL
    result=$($WEBAUTO page-navigate --page-url "$url" --session-id "$SESSION_ID" 2>&1)

    if echo "$result" | grep -q '"success": true'; then
        echo -e "  ${GREEN}✓${NC} Navigation successful"
    else
        echo -e "  ${RED}✗${NC} Navigation failed"
        echo "$result"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi

    # Wait for element
    result=$($WEBAUTO element-wait \
        --session-id "$SESSION_ID" \
        --element-selector "$selector" \
        --wait-for "$wait_condition" \
        --timeout-ms "$timeout" 2>&1)

    if [ "$expect_success" = "true" ]; then
        if echo "$result" | grep -q '"success": true'; then
            waited_ms=$(echo "$result" | jq -r '.data.waited_ms' 2>/dev/null || echo "0")
            element_found=$(echo "$result" | jq -r '.data.element_found' 2>/dev/null || echo "false")

            echo -e "  ${GREEN}✓${NC} Wait successful"
            echo "  Waited: ${waited_ms}ms"
            echo "  Element found: $element_found"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            echo -e "  ${RED}✗${NC} Wait failed unexpectedly"
            echo "$result"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            return 1
        fi
    else
        if echo "$result" | grep -q '"success": false'; then
            error_code=$(echo "$result" | jq -r '.error.code' 2>/dev/null || echo "")
            echo -e "  ${GREEN}✓${NC} Correctly failed with timeout"
            echo "  Error code: $error_code"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            echo -e "  ${RED}✗${NC} Should have failed but succeeded"
            echo "$result"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            return 1
        fi
    fi

    echo ""
}

# Launch browser session
echo -e "${YELLOW}Launching browser...${NC}"
launch_result=$($WEBAUTO browser-launch --headless true 2>&1)

if echo "$launch_result" | grep -q '"success": true'; then
    SESSION_ID=$(echo "$launch_result" | jq -r '.data.session_id')
    echo -e "${GREEN}✓${NC} Browser launched (Session: $SESSION_ID)\n"
else
    echo -e "${RED}✗ Failed to launch browser${NC}"
    echo "$launch_result"
    exit 1
fi

# Trap to ensure browser cleanup on script exit
cleanup() {
    echo -e "\n${YELLOW}Cleaning up browser session...${NC}"
    $WEBAUTO browser-close --session-id "$SESSION_ID" >/dev/null 2>&1 || true
    echo -e "${GREEN}✓${NC} Cleanup complete"
}
trap cleanup EXIT

# =============================================================================
# TEST 1: Visible - Wait for element that is immediately visible
# =============================================================================
run_test \
    "Visible - Immediate (Example.com link)" \
    "https://example.com" \
    "a" \
    "visible" \
    5000 \
    "true"

# =============================================================================
# TEST 2: Attached - Wait for element attached to DOM
# =============================================================================
run_test \
    "Attached - DOM element (Example.com h1)" \
    "https://example.com" \
    "h1" \
    "attached" \
    5000 \
    "true"

# =============================================================================
# TEST 3: Visible - Fast response for already visible element
# =============================================================================
run_test \
    "Visible - Fast response (Example.com div)" \
    "https://example.com" \
    "div" \
    "visible" \
    5000 \
    "true"

# =============================================================================
# TEST 4: Timeout - Element never appears (should fail)
# =============================================================================
run_test \
    "Timeout - Non-existent element (error case)" \
    "https://example.com" \
    ".this-element-does-not-exist" \
    "visible" \
    2000 \
    "false"

# =============================================================================
# TEST 5: Korean Content - Naver homepage search box
# =============================================================================
run_test \
    "Visible - Korean content (Naver search)" \
    "https://www.naver.com" \
    "#query" \
    "visible" \
    5000 \
    "true"

# =============================================================================
# TEST 6: Attached - Wikipedia article content
# =============================================================================
run_test \
    "Attached - Wikipedia content div" \
    "https://en.wikipedia.org/wiki/Web_scraping" \
    "#mw-content-text" \
    "attached" \
    5000 \
    "true"

# =============================================================================
# TEST 7: Hidden condition - Wait for visible element to timeout when checking hidden
# =============================================================================
echo -e "${YELLOW}Test: Hidden - Visible element timeout (error case)${NC}"
echo "  URL: https://example.com"
echo "  Selector: h1 (visible element, should timeout for hidden)"
echo "  Wait Condition: hidden"
echo "  Timeout: 2000ms"

$WEBAUTO page-navigate --page-url "https://example.com" --session-id "$SESSION_ID" >/dev/null 2>&1

result=$($WEBAUTO element-wait \
    --session-id "$SESSION_ID" \
    --element-selector "h1" \
    --wait-for "hidden" \
    --timeout-ms 2000 2>&1)

if echo "$result" | grep -q '"success": false'; then
    error_code=$(echo "$result" | jq -r '.error.code' 2>/dev/null || echo "")
    echo -e "  ${GREEN}✓${NC} Correctly timed out waiting for visible element to hide"
    echo "  Error code: $error_code"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}✗${NC} Should have timed out"
    echo "$result"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

echo ""

# =============================================================================
# Print Summary
# =============================================================================
echo -e "${YELLOW}=== Test Summary ===${NC}"
echo -e "Total tests: $((TESTS_PASSED + TESTS_FAILED))"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}✗ Some tests failed${NC}"
    exit 1
fi
