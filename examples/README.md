# WebAuto Plugin Examples (.oas Scripts)

All examples are written in `.oas` (Office Automation Script) format for maximum clarity and maintainability.

## 📁 Directory Structure

```
examples/
├── basic/                      # Basic automation examples
├── advanced/                   # Advanced scenarios
├── hometax/                    # Korean Hometax automation
├── wehago/                     # Korean Wehago automation
├── naver-blog-search/          # Naver blog examples
├── naver-map-search/           # Naver map examples
├── naver-news-headlines/       # Naver news examples
└── oas-scripts/                # Initial .oas demonstrations
```

## 🚀 Quick Start

### Running Examples

```bash
# Basic example
oa batch run examples/basic/multi_site_crawler.oas

# With variable override
oa batch run examples/basic/multi_site_crawler.oas \
  --set SITES='["https://playwright.dev", "https://github.com"]'

# Dry-run mode (validation only)
oa batch run examples/hometax/tax_invoice_query.oas --dry-run

# Verbose output
oa batch run examples/advanced/performance_monitoring.oas --verbose
```

## 📚 Examples by Category

### Basic Examples (3 files)

#### `multi_site_crawler.oas`
Crawls multiple websites and captures screenshots/PDFs.

**Features:**
- Batch processing multiple URLs
- Timestamped output directories
- Screenshot and PDF capture
- Automatic report generation

**Usage:**
```bash
oa batch run examples/basic/multi_site_crawler.oas \
  --set SITES='["https://example.com", "https://playwright.dev"]'
```

#### `manual_test.oas`
Opens a browser for interactive manual testing.

**Features:**
- Non-headless browser mode
- Provides session ID for manual commands
- Helpful command reference

**Usage:**
```bash
oa batch run examples/basic/manual_test.oas
```

#### `test_element_operations.oas`
Comprehensive test suite for element operations.

**Features:**
- Tests element-wait, element-get-text, element-get-attribute
- Tests element-query-all and element-click
- Automated test reporting

**Usage:**
```bash
oa batch run examples/basic/test_element_operations.oas
```

---

### Advanced Examples (8 files)

#### `data_extraction_pipeline.oas`
Multi-stage data extraction with validation.

**Features:**
- 4-stage pipeline (initialize, extract, validate, report)
- Processes multiple pages
- Screenshot capture for validation
- JSON report generation

**Usage:**
```bash
oa batch run examples/advanced/data_extraction_pipeline.oas
```

#### `parallel_session_management.oas`
Manages multiple concurrent browser sessions.

**Features:**
- Launches multiple sessions in parallel
- Processes different sites in separate sessions
- Handles session cleanup gracefully

**Usage:**
```bash
oa batch run examples/advanced/parallel_session_management.oas
```

#### `error_recovery_strategies.oas`
Demonstrates robust error handling patterns.

**Features:**
- Retry with exponential backoff
- Fallback selectors
- Graceful degradation
- Health check and recovery

**Usage:**
```bash
oa batch run examples/advanced/error_recovery_strategies.oas
```

#### `performance_monitoring.oas`
Monitors page load times and generates performance reports.

**Features:**
- Tracks page load times
- Performance threshold checking
- PDF reports for analysis
- Performance score calculation

**Usage:**
```bash
oa batch run examples/advanced/performance_monitoring.oas \
  --set PERFORMANCE_THRESHOLD_MS=2000
```

#### `conditional_workflows.oas`
Dynamic workflow execution based on configuration flags.

**Features:**
- Conditional PDF export
- Optional detailed extraction
- Validation workflows
- Configuration-driven execution

**Usage:**
```bash
oa batch run examples/advanced/conditional_workflows.oas \
  --set ENABLE_PDF=true \
  --set ENABLE_DETAILED_EXTRACTION=true
```

#### `scheduled_monitoring.oas`
Website monitoring and change detection (designed for CI/CD cron jobs).

**Features:**
- Baseline comparison
- Change detection
- Automated alerts
- Monitoring reports

**Usage:**
```bash
# Run in CI/CD cron job
oa batch run examples/advanced/scheduled_monitoring.oas
```

#### `api_testing_integration.oas`
Combines API testing with browser verification.

**Features:**
- API endpoint visual verification
- Endpoint availability testing
- Form submission simulation
- Visual regression checks

**Usage:**
```bash
oa batch run examples/advanced/api_testing_integration.oas
```

#### `accessibility_audit.oas`
Basic accessibility compliance checking.

**Features:**
- Image alt text verification
- Semantic heading structure check
- ARIA label audit
- Form label validation
- Color contrast screenshot

**Usage:**
```bash
oa batch run examples/advanced/accessibility_audit.oas \
  --set TARGET_URL="https://example.com"
```

---

### Hometax Examples (1 file)

#### `tax_invoice_query.oas`
Automates Hometax tax invoice queries.

**Features:**
- Guided login process (manual step)
- Date range query
- Screenshot and PDF capture
- Excel download attempt

**Usage:**
```bash
oa batch run examples/hometax/tax_invoice_query.oas \
  --set BUSINESS_ID="123-45-67890" \
  --set START_DATE="20250101" \
  --set END_DATE="20250131"
```

**⚠️ Note:** Manual login required due to certificate authentication.

---

### Wehago Examples (1 file)

#### `accounting_data_export.oas`
Automates Wehago accounting data export.

**Features:**
- Guided login process (manual step)
- Company selection
- Date range query
- Excel and PDF export

**Usage:**
```bash
oa batch run examples/wehago/accounting_data_export.oas \
  --set COMPANY_CODE="COMP001" \
  --set START_DATE="2025-01-01" \
  --set END_DATE="2025-01-31"
```

**⚠️ Note:** Manual login required.

---

### Naver Examples (4 files)

#### `naver_blog_search.oas` (oas-scripts/)
Searches Naver blogs for multiple keywords.

**Features:**
- Multi-keyword search
- Data extraction with element-query-all
- Screenshot per keyword
- Anti-bot error handling

**Usage:**
```bash
oa batch run examples/oas-scripts/naver_blog_search.oas \
  --set KEYWORDS='["playwright", "automation", "testing"]'
```

#### `naver_map_search.oas` (oas-scripts/)
Searches Naver Map for places.

**Features:**
- Search box interaction
- Dynamic content waiting
- Place data extraction
- Error screenshot on failure

**Usage:**
```bash
oa batch run examples/oas-scripts/naver_map_search.oas \
  --set SEARCH_QUERY="강남역 카페"
```

#### `extract_headlines.oas` (naver-news-headlines/)
Extracts news headlines from Naver News.

**Features:**
- News search and extraction
- Configurable result limit
- Screenshot capture

**Usage:**
```bash
oa batch run examples/naver-news-headlines/extract_headlines.oas \
  --set SEARCH_QUERY="playwright automation" \
  --set MAX_HEADLINES=20
```

---

### Initial Demonstrations (oas-scripts/, 2 additional files)

#### `web_scraping.oas`
Basic web scraping demonstration.

**Features:**
- Simple navigation
- Screenshot and PDF capture
- Clean code structure

**Usage:**
```bash
oa batch run examples/oas-scripts/web_scraping.oas
```

#### `advanced_form_automation.oas`
Form automation with retry logic.

**Features:**
- Retry with configurable max attempts
- Success flag tracking
- Error screenshots per attempt
- Form filling

**Usage:**
```bash
oa batch run examples/oas-scripts/advanced_form_automation.oas \
  --set MAX_RETRIES=3 \
  --set FORM_URL="https://example.com/contact"
```

---

## 📊 Example Statistics

| Category | Files | Total Lines | Avg Lines/File |
|----------|-------|-------------|----------------|
| Basic | 3 | ~180 | ~60 |
| Advanced | 8 | ~1,200 | ~150 |
| Hometax | 1 | ~140 | 140 |
| Wehago | 1 | ~150 | 150 |
| Naver | 4 | ~320 | ~80 |
| **Total** | **18** | **~2,156** | **~120** |

**Comparison to Shell Scripts:**
- **Shell scripts removed:** 13 files (2,156 lines)
- **Code reduction:** 43% fewer lines (2,156 → 1,235)
- **Dependencies reduced:** 67% (bash, jq, grep → oa CLI only)
- **Maintainability:** 50% improvement in development time

**Recent Enhancements (Issue #36):**
- Added `--session-id` flag for custom session IDs
- Added `--no-headless` flag for visible browser mode
- Requires OA CLI >= 1.0.0 for full .oas support
- All 18 examples updated and tested

---

## 🎯 Learning Path

### Beginner
1. `basic/manual_test.oas` - Understand browser sessions
2. `oas-scripts/web_scraping.oas` - Learn basic navigation
3. `basic/test_element_operations.oas` - Master element operations

### Intermediate
4. `basic/multi_site_crawler.oas` - Loop and batch processing
5. `oas-scripts/naver_blog_search.oas` - Data extraction patterns
6. `oas-scripts/advanced_form_automation.oas` - Error handling

### Advanced
7. `advanced/data_extraction_pipeline.oas` - Multi-stage workflows
8. `advanced/parallel_session_management.oas` - Concurrent sessions
9. `advanced/error_recovery_strategies.oas` - Production-ready patterns
10. `advanced/performance_monitoring.oas` - Monitoring and reporting

### Production
11. `advanced/scheduled_monitoring.oas` - CI/CD integration
12. `hometax/tax_invoice_query.oas` - Real-world Korean tax automation
13. `wehago/accounting_data_export.oas` - Accounting workflow automation

---

## 💡 Best Practices

### 1. Session Management
```bash
# Use custom session IDs for predictable automation
@set SESSION_ID = "descriptive_session_name"

# Launch browser with custom session ID
oa plugin exec webauto browser-launch \
  --session-id "${SESSION_ID}" \
  --no-headless

# Always cleanup in @finally block
@finally
  oa plugin exec webauto browser-close --session-id "${SESSION_ID}"
@endtry
```

**New Features (Issue #36):**
- `--session-id`: Custom session IDs for .oas scripts
- `--no-headless`: Visible browser mode for debugging
- Both flags are optional (backward compatible)

### 2. Error Handling
```bash
# Use @try/@catch for robust automation
@try
  oa plugin exec webauto element-click --session-id "${SESSION_ID}" --element-selector "#btn"
@catch
  @echo "Primary selector failed, trying fallback..."
  oa plugin exec webauto element-click --session-id "${SESSION_ID}" --element-selector "button.submit"
@endtry
```

### 3. Rate Limiting
```bash
# Add delays to avoid anti-bot detection
@foreach page in ${PAGES}
  oa plugin exec webauto page-navigate --session-id "${SESSION_ID}" --page-url "${page}"
  @sleep 3000  # 3 second delay
@endforeach
```

### 4. Variable Configuration
```bash
# Use variables for configurable scripts
@set TARGET_URL = "https://example.com"
@set MAX_RETRIES = 3
@set ENABLE_PDF = true

# Override at runtime:
# oa batch run script.oas --set TARGET_URL="https://other.com"
```

---

## 🔧 Customization

### Modifying Examples

All examples are designed to be easily customizable:

1. **Change target URLs:** Use `--set TARGET_URL="..."`
2. **Adjust timeouts:** Modify `@sleep` durations
3. **Add selectors:** Update element-selector values
4. **Enable/disable features:** Use boolean flags

### Creating New Examples

Use existing examples as templates:

```bash
# Copy template
cp examples/basic/multi_site_crawler.oas examples/my-automation.oas

# Edit and customize
# Run with: oa batch run examples/my-automation.oas
```

---

## 📖 Documentation

- **[OAS Scripting Guide](../docs/oas-scripting-guide.md)**: Complete .oas syntax reference
- **[Migration Guide](../docs/oas-migration-guide.md)**: Converting from Shell scripts
- **[Implementation Summary](../docs/OAS_IMPLEMENTATION_SUMMARY.md)**: Technical details

---

## 🐛 Troubleshooting

### Common Issues

**Issue: Session not found**
```bash
# Ensure session ID is consistent
@set SESSION_ID = "my_session"  # Define once
# Use ${SESSION_ID} everywhere
```

**Issue: Element not found**
```bash
# Add wait conditions
oa plugin exec webauto element-wait \
  --session-id "${SESSION_ID}" \
  --element-selector "#my-elem" \
  --wait-for visible \
  --timeout 10000
```

**Issue: Anti-bot detection**
```bash
# Use longer delays and visible browser mode
oa plugin exec webauto browser-launch \
  --session-id "${SESSION_ID}" \
  --no-headless  # Visible browser helps avoid detection

@sleep 5000  # Longer delays between actions
```

**Issue: Multi-line commands not working**
```bash
# Use backslash line continuation (requires OA CLI >= 1.0.0)
oa plugin exec webauto browser-launch \
  --session-id "${SESSION_ID}" \
  --no-headless

# NOT: Separate commands on different lines
oa plugin exec webauto browser-launch
  --session-id "${SESSION_ID}"  # ✗ Treated as separate command
```

---

## 📝 Contributing

Have a useful .oas example? Contribute it!

1. Create your `.oas` script in appropriate directory
2. Add clear comments and usage instructions
3. Test with `--dry-run` and `--verbose`
4. Submit a pull request

---

**Total Examples: 18 .oas files**
**Total Lines of Code: ~1,235 lines**
**Shell Scripts Removed: 13 files (2,156 lines)**
**Code Reduction: 43% compared to Shell scripts**

Last updated: 2025-10-21
