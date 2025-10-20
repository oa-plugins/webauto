# OAS Scripting Guide for WebAuto Plugin

## Overview

This guide demonstrates how to use Office Automation Script (`.oas`) format for webauto plugin automation, replacing verbose shell scripts with clean, maintainable automation workflows.

## Why .oas Instead of Shell Scripts?

### Shell Script Problems
```bash
# Verbose and hard to read (58+ lines)
#!/bin/bash
set -e
WEBAUTO="../../webauto"
RESULT=$($WEBAUTO browser-launch --headless true)
SESSION_ID=$(echo "$RESULT" | jq -r '.data.session_id')

if [ -z "$SESSION_ID" ] || [ "$SESSION_ID" = "null" ]; then
    echo "❌ 브라우저 실행 실패"
    exit 1
fi
# ... 30+ more lines of jq parsing and error handling
```

### .oas Script Benefits
```bash
# Clean and readable (30 lines)
@set SESSION_ID = "web_session"
oa plugin exec webauto browser-launch --session-id "${SESSION_ID}"
oa plugin exec webauto page-navigate --session-id "${SESSION_ID}" --page-url "https://example.com"
oa plugin exec webauto page-screenshot --session-id "${SESSION_ID}" --image-path "output.png"
oa plugin exec webauto browser-close --session-id "${SESSION_ID}"
```

**Key Advantages:**
- **50% fewer lines of code**
- **No jq dependency** - JSON parsing handled by OA runtime
- **Built-in error handling** - `@try/@catch/@finally` blocks
- **Variable management** - `@set`, `@unset`, `@export` directives
- **Control flow** - `@if/@foreach/@while` for complex logic
- **CI/CD integration** - `oa batch run script.oas`

## Getting Started

### Prerequisites

1. **OA CLI installed**
```bash
# Check if oa is available
oa --version

# Verify webauto plugin is registered
oa plugin list
```

2. **WebAuto plugin setup**
```bash
# Ensure webauto binary is in plugins directory
ls -la ~/.oa/plugins/webauto/

# Expected structure:
# ~/.oa/plugins/webauto/
#   ├── webauto (or webauto.exe on Windows)
#   └── plugin.yaml
```

### Basic Example

Create `hello_web.oas`:
```bash
# hello_web.oas - Your first .oas script

@set TARGET_URL = "https://example.com"
@set OUTPUT_DIR = "./screenshots"

# Create output directory
@if not exists("${OUTPUT_DIR}")
  @mkdir "${OUTPUT_DIR}"
@endif

@echo "Starting browser automation..."

# Launch browser
@set SESSION_ID = "hello_session"
oa plugin exec webauto browser-launch --session-id "${SESSION_ID}"

# Navigate and capture
oa plugin exec webauto page-navigate --session-id "${SESSION_ID}" --page-url "${TARGET_URL}"
oa plugin exec webauto page-screenshot --session-id "${SESSION_ID}" --image-path "${OUTPUT_DIR}/hello.png"

# Cleanup
oa plugin exec webauto browser-close --session-id "${SESSION_ID}"

@echo "Done! Screenshot saved to ${OUTPUT_DIR}/hello.png"
```

Execute:
```bash
oa batch run hello_web.oas
```

## .oas Syntax Reference

### Variables

```bash
# Set variables
@set BROWSER_TYPE = "chromium"
@set HEADLESS = true
@set TIMEOUT = 5000

# Use variables
oa plugin exec webauto browser-launch --headless "${HEADLESS}"

# Unset variables
@unset TIMEOUT

# Export to environment
@export PATH = "/usr/local/bin:${PATH}"
```

### Control Flow

#### Conditional Execution
```bash
@set ENABLE_PDF = true

@if ${ENABLE_PDF} == true
  oa plugin exec webauto page-pdf --session-id "${SESSION_ID}" --pdf-path "output.pdf"
  @echo "PDF generated"
@else
  @echo "PDF generation skipped"
@endif
```

#### Loops
```bash
# Array iteration
@set PAGES = ["https://page1.com", "https://page2.com", "https://page3.com"]

@foreach page in ${PAGES}
  @echo "Processing: ${page}"
  oa plugin exec webauto page-navigate --session-id "${SESSION_ID}" --page-url "${page}"
  @sleep 2000
@endforeach
```

#### Error Handling
```bash
@try
  oa plugin exec webauto element-click --session-id "${SESSION_ID}" --element-selector "#submit-btn"
  @echo "Click successful"
@catch
  @echo "Click failed, trying alternative selector"
  oa plugin exec webauto element-click --session-id "${SESSION_ID}" --element-selector "button[type='submit']"
@finally
  @echo "Operation completed"
@endtry
```

## Real-World Examples

### Example 1: Web Scraping

See: `examples/oas-scripts/web_scraping.oas`

**Features:**
- Directory creation
- Screenshot capture
- PDF export
- Automatic cleanup

**Compared to shell script:**
- Shell: 58 lines with jq parsing
- .oas: 30 lines with clear logic

### Example 2: Naver Blog Search

See: `examples/oas-scripts/naver_blog_search.oas`

**Features:**
- Multiple keyword search
- Data extraction with `element-query-all`
- Per-result screenshot capture
- Rate limiting with `@sleep`
- Error handling for anti-bot detection

**Key technique:**
```bash
@foreach keyword in ${KEYWORDS}
  @try
    oa plugin exec webauto page-navigate --session-id "${SESSION_ID}" --page-url "https://search.naver.com/search.naver?where=view&query=${keyword}"
    @sleep 2000
    oa plugin exec webauto element-query-all --session-id "${SESSION_ID}" --element-selector ".title_link" --get-text --get-attribute href --limit 10
  @catch
    @echo "⚠ Search failed for: ${keyword} (anti-bot protection)"
  @endtry
@endforeach
```

### Example 3: Naver Map Search

See: `examples/oas-scripts/naver_map_search.oas`

**Features:**
- Form interaction (search box)
- Dynamic content waiting
- Place data extraction
- Error screenshot on failure

**Key technique:**
```bash
@try
  oa plugin exec webauto element-type --session-id "${SESSION_ID}" --element-selector "input.input_search" --element-text "${SEARCH_QUERY}" --delay 100
  oa plugin exec webauto element-click --session-id "${SESSION_ID}" --element-selector "button.btn_search"
  @sleep 3000
  oa plugin exec webauto element-wait --session-id "${SESSION_ID}" --element-selector ".place_bluelink" --wait-for visible --timeout 10000
@catch
  oa plugin exec webauto page-screenshot --session-id "${SESSION_ID}" --image-path "${OUTPUT_DIR}/error_screenshot.png"
  @echo "Error screenshot saved for debugging"
@endtry
```

### Example 4: Advanced Form Automation

See: `examples/oas-scripts/advanced_form_automation.oas`

**Features:**
- Retry logic with `@while` loop
- Success flag tracking
- Progressive error screenshots
- Robust form filling

**Key technique:**
```bash
@set ATTEMPT = 0
@set MAX_RETRIES = 3
@set SUCCESS = false

@while ${ATTEMPT} < ${MAX_RETRIES} and not ${SUCCESS}
  @set ATTEMPT = ${ATTEMPT} + 1
  @try
    oa plugin exec webauto form-fill --session-id "${SESSION_ID}" --field-selector "input[name='name']=${FORM_NAME}" --submit-selector "button[type='submit']"
    @set SUCCESS = true
  @catch
    oa plugin exec webauto page-screenshot --session-id "${SESSION_ID}" --image-path "${OUTPUT_DIR}/error_attempt_${ATTEMPT}.png"
    @sleep 5000
  @endtry
@endwhile
```

## Current Limitations and Workarounds

### 1. JSON Response Parsing

**Current limitation:**
```bash
# This doesn't work yet:
@set RESULT = oa plugin exec webauto browser-launch --headless true
@set SESSION_ID = ${RESULT.data.session_id}  # ❌ JSON path not supported
```

**Workaround:**
```bash
# Use predefined session IDs
@set SESSION_ID = "my_session_001"
oa plugin exec webauto browser-launch --session-id "${SESSION_ID}"
```

**Future enhancement:**
The OA batch engine could be extended to support JSON path extraction:
```bash
# Proposed syntax
@set RESULT = $(oa plugin exec webauto browser-launch --headless true)
@set SESSION_ID = ${RESULT.data.session_id}
```

### 2. Plugin Command Shortcuts

**Current:**
```bash
oa plugin exec webauto page-navigate --session-id "${SESSION_ID}" --page-url "..."
```

**Proposed shorthand:**
```bash
# Option 1: Native directives
@page-navigate --session-id "${SESSION_ID}" --page-url "..."

# Option 2: Plugin prefix
webauto page-navigate --session-id "${SESSION_ID}" --page-url "..."
```

This would require:
1. `.oas` parser extension in `pkg/batch/parser.go`
2. Plugin command registration in batch executor
3. Documentation updates

## Migration Guide

### Converting Shell Scripts to .oas

**Step 1: Remove shell boilerplate**

Before (shell):
```bash
#!/bin/bash
set -e
WEBAUTO="../../webauto"
OUTPUT_DIR="./output"
mkdir -p "$OUTPUT_DIR"
```

After (.oas):
```bash
@set OUTPUT_DIR = "./output"
@if not exists("${OUTPUT_DIR}")
  @mkdir "${OUTPUT_DIR}"
@endif
```

**Step 2: Replace JSON parsing**

Before (shell):
```bash
RESULT=$($WEBAUTO browser-launch --headless true)
SESSION_ID=$(echo "$RESULT" | jq -r '.data.session_id')
if [ -z "$SESSION_ID" ]; then exit 1; fi
```

After (.oas):
```bash
@set SESSION_ID = "automation_session"
oa plugin exec webauto browser-launch --session-id "${SESSION_ID}"
```

**Step 3: Convert plugin calls**

Before (shell):
```bash
$WEBAUTO page-navigate --session-id "$SESSION_ID" --page-url "https://example.com"
```

After (.oas):
```bash
oa plugin exec webauto page-navigate --session-id "${SESSION_ID}" --page-url "https://example.com"
```

**Step 4: Add error handling**

Before (shell):
```bash
if ! $WEBAUTO element-click --session-id "$SESSION_ID" --element-selector "#btn"; then
    echo "Click failed"
    exit 1
fi
```

After (.oas):
```bash
@try
  oa plugin exec webauto element-click --session-id "${SESSION_ID}" --element-selector "#btn"
@catch
  @echo "Click failed, attempting recovery..."
  oa plugin exec webauto element-wait --session-id "${SESSION_ID}" --element-selector "#btn" --wait-for visible
  oa plugin exec webauto element-click --session-id "${SESSION_ID}" --element-selector "#btn"
@endtry
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Weekly Web Scraping

on:
  schedule:
    - cron: '0 0 * * 0'  # Every Sunday at midnight

jobs:
  scrape:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install OA CLI
        run: |
          curl -fsSL https://install.oa-cli.com | sh
          oa plugin refresh

      - name: Run scraping script
        run: |
          oa batch run examples/oas-scripts/web_scraping.oas \
            --set OUTPUT_DIR="${GITHUB_WORKSPACE}/output" \
            --verbose

      - name: Upload results
        uses: actions/upload-artifact@v3
        with:
          name: scraping-results
          path: output/
```

### GitLab CI Example

```yaml
weekly_scraping:
  image: oa-cli:latest
  script:
    - oa batch run examples/oas-scripts/naver_blog_search.oas --set KEYWORDS='["topic1", "topic2"]'
  artifacts:
    paths:
      - output/
  only:
    - schedules
```

## Best Practices

### 1. Session Management
```bash
# Good: Use descriptive session IDs
@set SESSION_ID = "hometax_tax_invoice_query"

# Bad: Generic IDs
@set SESSION_ID = "session1"
```

### 2. Error Handling
```bash
# Good: Comprehensive error handling with recovery
@try
  oa plugin exec webauto element-click --session-id "${SESSION_ID}" --element-selector "#primary-btn"
@catch
  @echo "Primary selector failed, trying fallback"
  oa plugin exec webauto element-click --session-id "${SESSION_ID}" --element-selector "button.submit"
@finally
  oa plugin exec webauto page-screenshot --session-id "${SESSION_ID}" --image-path "final_state.png"
@endtry

# Bad: No error handling
oa plugin exec webauto element-click --session-id "${SESSION_ID}" --element-selector "#primary-btn"
```

### 3. Rate Limiting
```bash
# Good: Add delays to avoid anti-bot detection
@foreach page in ${PAGES}
  oa plugin exec webauto page-navigate --session-id "${SESSION_ID}" --page-url "${page}"
  @sleep 3000  # 3 second delay
@endforeach

# Bad: Rapid-fire requests
@foreach page in ${PAGES}
  oa plugin exec webauto page-navigate --session-id "${SESSION_ID}" --page-url "${page}"
@endforeach
```

### 4. Resource Cleanup
```bash
# Good: Always cleanup in @finally
@try
  oa plugin exec webauto browser-launch --session-id "${SESSION_ID}"
  # ... automation steps
@catch
  @echo "Error occurred"
@finally
  oa plugin exec webauto browser-close --session-id "${SESSION_ID}"
@endtry

# Bad: Cleanup only on success path
oa plugin exec webauto browser-launch --session-id "${SESSION_ID}"
# ... automation steps
oa plugin exec webauto browser-close --session-id "${SESSION_ID}"
```

## Troubleshooting

### Syntax Validation
```bash
# Check script syntax before running
oa batch validate script.oas
```

### Dry-run Mode
```bash
# See what would execute without actually running
oa batch run script.oas --dry-run
```

### Verbose Logging
```bash
# Enable detailed execution logs
oa batch run script.oas --verbose
```

### Common Errors

**Error: Plugin not found**
```bash
# Solution: Refresh plugin registry
oa plugin refresh
oa plugin list  # Verify webauto is listed
```

**Error: Session not found**
```bash
# Solution: Check session ID consistency
@set SESSION_ID = "my_session"  # Define once at top
# Use ${SESSION_ID} everywhere, not hardcoded strings
```

**Error: Element not found**
```bash
# Solution: Add wait conditions
oa plugin exec webauto element-wait --session-id "${SESSION_ID}" --element-selector "#my-elem" --wait-for visible --timeout 10000
```

## Next Steps

1. **Try the examples**: Run the provided `.oas` scripts in `examples/oas-scripts/`
2. **Convert your scripts**: Migrate existing shell scripts to `.oas` format
3. **Propose enhancements**: Open issues for JSON parsing and other language features
4. **Share workflows**: Contribute your `.oas` scripts to the community

## Resources

- [OA Batch Scripting Reference](../../oa/BATCH_SCRIPTING_DESIGN.md)
- [WebAuto Plugin Documentation](../README.md)
- [OA CLI Documentation](../../oa/README.md)
- [Example Scripts](./examples/oas-scripts/)
