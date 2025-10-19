#!/bin/bash
# Test script for element-get-attribute command across multiple websites
# This demonstrates attribute extraction from various real-world scenarios

set -e  # Exit on error

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== element-get-attribute Multi-Site Test Suite ===${NC}\n"

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
    local attribute="$4"
    local expected_pattern="$5"

    echo -e "${YELLOW}Test: ${test_name}${NC}"
    echo "  URL: $url"
    echo "  Selector: $selector"
    echo "  Attribute: $attribute"

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

    # Wait a moment for page to settle
    sleep 1

    # Get attribute
    result=$($WEBAUTO element-get-attribute --session-id "$SESSION_ID" --element-selector "$selector" --attribute-name "$attribute" 2>&1)

    if echo "$result" | grep -q '"success": true'; then
        attribute_value=$(echo "$result" | jq -r '.data.attribute_value' 2>/dev/null || echo "")
        element_count=$(echo "$result" | jq -r '.data.element_count' 2>/dev/null || echo "0")

        echo -e "  ${GREEN}✓${NC} Attribute extraction successful"
        echo "  Element count: $element_count"
        echo "  Attribute value: $attribute_value"

        # Validate expected pattern if provided
        if [ -n "$expected_pattern" ]; then
            if echo "$attribute_value" | grep -q "$expected_pattern"; then
                echo -e "  ${GREEN}✓${NC} Attribute value matches expected pattern"
                TESTS_PASSED=$((TESTS_PASSED + 1))
            else
                echo -e "  ${RED}✗${NC} Attribute value does not match expected pattern: $expected_pattern"
                TESTS_FAILED=$((TESTS_FAILED + 1))
                return 1
            fi
        else
            TESTS_PASSED=$((TESTS_PASSED + 1))
        fi
    else
        echo -e "  ${RED}✗${NC} Attribute extraction failed"
        echo "$result"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
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
# TEST 1: Simple link href extraction (example.com)
# =============================================================================
run_test \
    "Simple Link href - example.com" \
    "https://example.com" \
    "a" \
    "href" \
    "iana.org"

# =============================================================================
# TEST 2: Multiple paragraph class attributes (example.com)
# =============================================================================
run_test \
    "Paragraph class - example.com" \
    "https://example.com" \
    "p" \
    "class" \
    ""

# =============================================================================
# TEST 3: Wikipedia content link href
# =============================================================================
run_test \
    "Wikipedia Content Link" \
    "https://en.wikipedia.org/wiki/Web_scraping" \
    "a[href='/wiki/Data_scraping']" \
    "href" \
    "/wiki/Data_scraping"

# =============================================================================
# TEST 4: Wikipedia heading id attribute
# =============================================================================
run_test \
    "Wikipedia Heading id" \
    "https://en.wikipedia.org/wiki/Web_scraping" \
    "h1#firstHeading" \
    "id" \
    "firstHeading"

# =============================================================================
# TEST 5: Error case - non-existent selector
# =============================================================================
echo -e "${YELLOW}Test: Non-existent Selector (Error Case)${NC}"
echo "  URL: https://example.com"
echo "  Selector: .this-does-not-exist"
echo "  Attribute: href"

$WEBAUTO page-navigate --page-url "https://example.com" --session-id "$SESSION_ID" >/dev/null 2>&1
sleep 1

result=$($WEBAUTO element-get-attribute --session-id "$SESSION_ID" --element-selector ".this-does-not-exist" --attribute-name "href" 2>&1)

if echo "$result" | grep -q '"success": false'; then
    error_code=$(echo "$result" | jq -r '.error.code' 2>/dev/null || echo "")
    if [ "$error_code" = "ELEMENT_NOT_FOUND" ]; then
        echo -e "  ${GREEN}✓${NC} Correctly returned error for non-existent element"
        echo "  Error code: $error_code"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "  ${RED}✗${NC} Unexpected error code: $error_code"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
else
    echo -e "  ${RED}✗${NC} Should have failed for non-existent selector"
    echo "$result"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

echo ""

# =============================================================================
# TEST 6: Korean content test (Naver - link href)
# =============================================================================
run_test \
    "Naver Link href (Korean)" \
    "https://www.naver.com" \
    "a" \
    "href" \
    ""

# =============================================================================
# TEST 7: Null attribute test (element without specified attribute)
# =============================================================================
echo -e "${YELLOW}Test: Null Attribute (element without attribute)${NC}"
echo "  URL: https://example.com"
echo "  Selector: h1"
echo "  Attribute: href (doesn't exist on h1)"

$WEBAUTO page-navigate --page-url "https://example.com" --session-id "$SESSION_ID" >/dev/null 2>&1
sleep 1

result=$($WEBAUTO element-get-attribute --session-id "$SESSION_ID" --element-selector "h1" --attribute-name "href" 2>&1)

if echo "$result" | grep -q '"success": true'; then
    attribute_value=$(echo "$result" | jq -r '.data.attribute_value' 2>/dev/null || echo "")
    if [ "$attribute_value" = "null" ] || [ -z "$attribute_value" ]; then
        echo -e "  ${GREEN}✓${NC} Correctly returned null for non-existent attribute"
        echo "  Attribute value: $attribute_value"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "  ${RED}✗${NC} Unexpected attribute value: $attribute_value"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
else
    echo -e "  ${RED}✗${NC} Should have succeeded with null value"
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
