# Performance Baseline - webauto

**Date**: 2025-10-16 09:38:52
**Platform**: Darwin arm64
**Go Version**: go version go1.25.0 darwin/arm64
**Node Version**: v24.3.0

## Benchmark Results

| Command | Iterations | Avg (ms) | Min (ms) | Max (ms) | Target (ms) | Status |
|---------|-----------|----------|----------|----------|-------------|--------|
| browser-launch | 5 | 🔄 Measuring: browser-launch --headless true
  ✅ [0;32mPASS[0m: avg=305ms (target=500ms, min=300ms, max=313ms)
305 | - | - | 500 | ❌ FAIL |
| browser-close | 1 | 🔄 Measuring: session-close --session-id ses_6f478f8a
  ✅ [0;32mPASS[0m: avg=11ms (target=100ms, min=11ms, max=11ms)
11 | - | - | 500 | ❌ FAIL |
| session-list | 10 | 🔄 Measuring: session-list 
  ✅ [0;32mPASS[0m: avg=10ms (target=100ms, min=10ms, max=12ms)
10 | - | - | 100 | ❌ FAIL |
| session-close | 1 | 🔄 Measuring: session-close --session-id ses_6f478f8a
  ✅ [0;32mPASS[0m: avg=11ms (target=100ms, min=11ms, max=11ms)
11 | - | - | 100 | ❌ FAIL |
| page-navigate | 5 | 🔄 Measuring: page-navigate --session-id ses_94d78b5e --page-url 'data:text/html,<html><body><h1>Test</h1></body></html>'
  ✅ [0;32mPASS[0m: avg=13ms (target=1000ms, min=13ms, max=15ms)
13 | - | - | 1000 | ❌ FAIL |
| element-click | 5 | 🔄 Measuring: element-click --session-id ses_94d78b5e --element-selector '#btn'
  ❌ [0;31mFAIL[0m: avg=30029ms (target=300ms, min=30013ms, max=30097ms)
30029 | - | - | 300 | ❌ FAIL |
| element-type | 5 | 🔄 Measuring: element-type --session-id ses_94d78b5e --element-selector '#input' --text-input 'test'
  ✅ [0;32mPASS[0m: avg=11ms (target=300ms, min=11ms, max=13ms)
11 | - | - | 300 | ❌ FAIL |
| page-screenshot | 5 | 🔄 Measuring: page-screenshot --session-id ses_94d78b5e --image-path /tmp/test_screenshot.png
  ✅ [0;32mPASS[0m: avg=41ms (target=1000ms, min=40ms, max=42ms)
41 | - | - | 1000 | ❌ FAIL |

## Raw Data

See `benchmark_results.csv` for detailed per-iteration results.

## Next Steps

1. Analyze bottlenecks using profiling:
   ```bash
   go test -cpuprofile=cpu.prof -bench=. tests/benchmarks/
   go tool pprof -http=:8080 cpu.prof
   ```

2. Review identified optimization opportunities in PERFORMANCE.md

3. Implement optimizations and re-run benchmarks
