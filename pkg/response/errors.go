package response

// Common error codes
const (
	ErrNodeNotFound           = "NODE_NOT_FOUND"
	ErrPlaywrightNotInstalled = "PLAYWRIGHT_NOT_INSTALLED"
	ErrTimeoutExceeded        = "TIMEOUT_EXCEEDED"
	ErrSessionNotFound        = "SESSION_NOT_FOUND"
	ErrSessionLimitReached    = "SESSION_LIMIT_REACHED"
)

// Agent-related error codes
const (
	ErrPlannerFailed         = "PLANNER_FAILED"
	ErrGeneratorFailed       = "GENERATOR_FAILED"
	ErrHealerFailed          = "HEALER_FAILED"
	ErrScriptExecutionFailed = "SCRIPT_EXECUTION_FAILED"
)

// Browser-related error codes
const (
	ErrBrowserLaunchFailed   = "BROWSER_LAUNCH_FAILED"
	ErrBrowserConnectionLost = "BROWSER_CONNECTION_LOST"
	ErrPageLoadFailed        = "PAGE_LOAD_FAILED"
	ErrPageTimeout           = "PAGE_TIMEOUT"
)

// Element-related error codes
const (
	ErrElementNotFound      = "ELEMENT_NOT_FOUND"
	ErrElementNotVisible    = "ELEMENT_NOT_VISIBLE"
	ErrElementNotClickable  = "ELEMENT_NOT_CLICKABLE"
	ErrFormValidationFailed = "FORM_VALIDATION_FAILED"
)

// Anti-Bot related error codes
const (
	ErrCaptchaDetected       = "CAPTCHA_DETECTED"
	ErrBotDetectionTriggered = "BOT_DETECTION_TRIGGERED"
	ErrRateLimitExceeded     = "RATE_LIMIT_EXCEEDED"
	ErrAccessDenied          = "ACCESS_DENIED"
)
