#!/bin/bash
# Manual test script for element-get-text command
# Simple step-by-step demonstration

WEBAUTO="../../webauto"

echo "=== Manual Test: element-get-text ==="
echo ""

# Step 1: Launch browser
echo "Step 1: Launching browser (headless=false for visibility)..."
$WEBAUTO browser-launch --headless false | jq .

echo ""
read -p "Copy the session_id from above and press Enter to continue..."
read -p "Enter session ID: " SESSION_ID

echo ""
echo "Step 2: Navigating to example.com..."
$WEBAUTO page-navigate \
    --page-url "https://example.com" \
    --session-id "$SESSION_ID" | jq .

echo ""
echo "Step 3: Getting text from <h1> element..."
$WEBAUTO element-get-text \
    --session-id "$SESSION_ID" \
    --element-selector "h1" | jq .

echo ""
echo "Step 4: Getting text from all <p> elements..."
$WEBAUTO element-get-text \
    --session-id "$SESSION_ID" \
    --element-selector "p" | jq .

echo ""
echo "Step 5: Testing with non-existent selector (should fail)..."
$WEBAUTO element-get-text \
    --session-id "$SESSION_ID" \
    --element-selector ".does-not-exist" | jq .

echo ""
echo "Step 6: Taking screenshot..."
$WEBAUTO page-screenshot \
    --session-id "$SESSION_ID" \
    --image-path "example_com.png" \
    --full-page true | jq .

echo ""
read -p "Press Enter to close browser and exit..."

echo "Step 7: Closing browser..."
$WEBAUTO browser-close --session-id "$SESSION_ID" | jq .

echo ""
echo "✓ Test complete!"
