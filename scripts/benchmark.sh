#!/bin/bash

# benchmark.sh - Performance benchmark script for webauto commands
# Measures actual command execution times against targets

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
WEBAUTO_BIN="$PROJECT_ROOT/webauto"
RESULTS_FILE="$PROJECT_ROOT/benchmark_results.csv"
BASELINE_FILE="$PROJECT_ROOT/PERFORMANCE_BASELINE.md"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Performance targets (in milliseconds)
# Function to get target for a command
get_target() {
    case "$1" in
        "browser-launch") echo 500 ;;
        "browser-close") echo 500 ;;
        "page-navigate") echo 1000 ;;
        "element-click") echo 300 ;;
        "element-type") echo 300 ;;
        "page-screenshot") echo 1000 ;;
        "session-list") echo 100 ;;
        "session-close") echo 100 ;;
        *) echo 0 ;;
    esac
}

# Check if webauto binary exists
if [ ! -f "$WEBAUTO_BIN" ]; then
    echo "❌ webauto binary not found at $WEBAUTO_BIN"
    echo "Please build it first: go build -o webauto cmd/webauto/main.go"
    exit 1
fi

# Initialize results CSV
echo "command,iteration,duration_ms,target_ms,status" > "$RESULTS_FILE"

# Helper: Measure command execution time
measure_command() {
    local cmd=$1
    local args=$2
    local iterations=${3:-5}

    local total_time=0
    local min_time=999999
    local max_time=0

    echo "🔄 Measuring: $cmd $args"

    for i in $(seq 1 $iterations); do
        local start=$(perl -MTime::HiRes -e 'printf("%.0f\n",Time::HiRes::time()*1000)')
        $WEBAUTO_BIN $cmd $args > /dev/null 2>&1
        local end=$(perl -MTime::HiRes -e 'printf("%.0f\n",Time::HiRes::time()*1000)')
        local duration=$((end - start))

        total_time=$((total_time + duration))

        if [ $duration -lt $min_time ]; then
            min_time=$duration
        fi

        if [ $duration -gt $max_time ]; then
            max_time=$duration
        fi

        # Write to CSV
        local target=$(get_target "$cmd")
        local status="PASS"
        if [ $duration -gt $target ]; then
            status="FAIL"
        fi

        echo "$cmd,$i,$duration,$target,$status" >> "$RESULTS_FILE"
    done

    local avg_time=$((total_time / iterations))
    local target=$(get_target "$cmd")

    # Check if within target
    if [ $avg_time -le $target ]; then
        echo -e "  ✅ ${GREEN}PASS${NC}: avg=${avg_time}ms (target=${target}ms, min=${min_time}ms, max=${max_time}ms)"
    else
        echo -e "  ❌ ${RED}FAIL${NC}: avg=${avg_time}ms (target=${target}ms, min=${min_time}ms, max=${max_time}ms)"
    fi

    # Return average time for summary
    echo $avg_time
}

# Helper: Clean up sessions
cleanup_sessions() {
    echo "🧹 Cleaning up sessions..."
    $WEBAUTO_BIN session-list 2>/dev/null | jq -r '.data.sessions[]?.session_id' 2>/dev/null | while read session_id; do
        if [ ! -z "$session_id" ]; then
            $WEBAUTO_BIN session-close --session-id "$session_id" > /dev/null 2>&1 || true
        fi
    done
}

# Main benchmark suite
echo "================================================"
echo "  webauto Performance Benchmark Suite"
echo "================================================"
echo ""

# Clean up before starting
cleanup_sessions

# Category 1: Browser Control (Target: < 500ms)
echo "📊 Category 1: Browser Control (Target: < 500ms)"
echo "------------------------------------------------"

# browser-launch
SESSION_ID=$($WEBAUTO_BIN browser-launch --headless true 2>/dev/null | jq -r '.data.session_id')
if [ -z "$SESSION_ID" ]; then
    echo "❌ Failed to launch browser for testing"
    exit 1
fi

# browser-close (using the session we just created)
avg_close=$(measure_command "browser-close" "--session-id $SESSION_ID" 1)

# browser-launch (5 iterations)
avg_launch=$(measure_command "browser-launch" "--headless true" 5)
echo ""

# Clean up sessions created during browser-launch benchmarks
cleanup_sessions

# Category 2: Session Management (Target: < 100ms)
echo "📊 Category 2: Session Management (Target: < 100ms)"
echo "------------------------------------------------"

# Create 3 sessions for testing
SESSION_IDS=()
for i in {1..3}; do
    sid=$($WEBAUTO_BIN browser-launch --headless true 2>/dev/null | jq -r '.data.session_id')
    SESSION_IDS+=("$sid")
done

avg_list=$(measure_command "session-list" "" 10)
avg_close=$(measure_command "session-close" "--session-id ${SESSION_IDS[0]}" 1)
echo ""

# Clean up remaining sessions
cleanup_sessions

# Category 3: Page Control (Target: < 1000ms)
echo "📊 Category 3: Page Control (Target: < 1000ms)"
echo "------------------------------------------------"

# Create a session for page operations
SESSION_ID=$($WEBAUTO_BIN browser-launch --headless true 2>/dev/null | jq -r '.data.session_id')

avg_nav=$(measure_command "page-navigate" "--session-id $SESSION_ID --page-url 'data:text/html,<html><body><h1>Test</h1></body></html>'" 5)
echo ""

# Category 4: Element Operations (Target: < 300ms)
echo "📊 Category 4: Element Operations (Target: < 300ms)"
echo "------------------------------------------------"

# Benchmark element-click (with page navigation before each iteration)
echo "🔄 Measuring: element-click --session-id $SESSION_ID --element-selector '#btn'"
total_time=0
min_time=999999
max_time=0
iterations=5

for i in $(seq 1 $iterations); do
    # Navigate to test page before each click to ensure clean state
    $WEBAUTO_BIN page-navigate --session-id $SESSION_ID --page-url 'data:text/html,<html><body><button id="btn" onclick="console.log(\"clicked\"); return false;">Click</button></body></html>' > /dev/null 2>&1

    start=$(perl -MTime::HiRes -e 'printf("%.0f\n",Time::HiRes::time()*1000)')
    $WEBAUTO_BIN element-click --session-id $SESSION_ID --element-selector '#btn' > /dev/null 2>&1
    end=$(perl -MTime::HiRes -e 'printf("%.0f\n",Time::HiRes::time()*1000)')
    duration=$((end - start))

    total_time=$((total_time + duration))

    if [ $duration -lt $min_time ]; then
        min_time=$duration
    fi

    if [ $duration -gt $max_time ]; then
        max_time=$duration
    fi

    # Write to CSV
    cmd_target=$(get_target "element-click")
    status="PASS"
    if [ $duration -gt $cmd_target ]; then
        status="FAIL"
    fi

    echo "element-click,$i,$duration,$cmd_target,$status" >> "$RESULTS_FILE"
done

avg_click=$((total_time / iterations))
target=$(get_target "element-click")

if [ $avg_click -le $target ]; then
    echo -e "  ✅ ${GREEN}PASS${NC}: avg=${avg_click}ms (target=${target}ms, min=${min_time}ms, max=${max_time}ms)"
else
    echo -e "  ❌ ${RED}FAIL${NC}: avg=${avg_click}ms (target=${target}ms, min=${min_time}ms, max=${max_time}ms)"
fi

# Navigate to test page for element-type (clean input field)
$WEBAUTO_BIN page-navigate --session-id $SESSION_ID --page-url 'data:text/html,<html><body><input id="input" /></body></html>' > /dev/null 2>&1

avg_type=$(measure_command "element-type" "--session-id $SESSION_ID --element-selector '#input' --text-input 'test'" 5)
echo ""

# Category 5: Data Extraction (Target: < 1000ms)
echo "📊 Category 5: Data Extraction (Target: < 1000ms)"
echo "------------------------------------------------"

avg_screenshot=$(measure_command "page-screenshot" "--session-id $SESSION_ID --image-path /tmp/test_screenshot.png" 5)
echo ""

# Clean up
cleanup_sessions
rm -f /tmp/test_screenshot.png

# Summary
echo "================================================"
echo "  Summary"
echo "================================================"
echo ""
echo "Results saved to: $RESULTS_FILE"
echo ""
echo "View detailed results:"
echo "  cat $RESULTS_FILE | column -t -s,"
echo ""
echo "Run analysis:"
echo "  go run scripts/performance_report.go"
echo ""

# Create baseline markdown if it doesn't exist
if [ ! -f "$BASELINE_FILE" ]; then
    echo "📝 Creating baseline performance report..."
    cat > "$BASELINE_FILE" <<EOF
# Performance Baseline - webauto

**Date**: $(date +"%Y-%m-%d %H:%M:%S")
**Platform**: $(uname -s) $(uname -m)
**Go Version**: $(go version)
**Node Version**: $(node --version)

## Benchmark Results

| Command | Iterations | Avg (ms) | Min (ms) | Max (ms) | Target (ms) | Status |
|---------|-----------|----------|----------|----------|-------------|--------|
| browser-launch | 5 | $avg_launch | - | - | 500 | $([ $avg_launch -le 500 ] && echo "✅ PASS" || echo "❌ FAIL") |
| browser-close | 1 | $avg_close | - | - | 500 | $([ $avg_close -le 500 ] && echo "✅ PASS" || echo "❌ FAIL") |
| session-list | 10 | $avg_list | - | - | 100 | $([ $avg_list -le 100 ] && echo "✅ PASS" || echo "❌ FAIL") |
| session-close | 1 | $avg_close | - | - | 100 | $([ $avg_close -le 100 ] && echo "✅ PASS" || echo "❌ FAIL") |
| page-navigate | 5 | $avg_nav | - | - | 1000 | $([ $avg_nav -le 1000 ] && echo "✅ PASS" || echo "❌ FAIL") |
| element-click | 5 | $avg_click | - | - | 300 | $([ $avg_click -le 300 ] && echo "✅ PASS" || echo "❌ FAIL") |
| element-type | 5 | $avg_type | - | - | 300 | $([ $avg_type -le 300 ] && echo "✅ PASS" || echo "❌ FAIL") |
| page-screenshot | 5 | $avg_screenshot | - | - | 1000 | $([ $avg_screenshot -le 1000 ] && echo "✅ PASS" || echo "❌ FAIL") |

## Raw Data

See \`benchmark_results.csv\` for detailed per-iteration results.

## Next Steps

1. Analyze bottlenecks using profiling:
   \`\`\`bash
   go test -cpuprofile=cpu.prof -bench=. tests/benchmarks/
   go tool pprof -http=:8080 cpu.prof
   \`\`\`

2. Review identified optimization opportunities in PERFORMANCE.md

3. Implement optimizations and re-run benchmarks
EOF

    echo "✅ Baseline report created: $BASELINE_FILE"
fi

echo "Done! 🎉"
