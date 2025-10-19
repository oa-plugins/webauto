package cli

import (
	"context"
	"time"

	"github.com/oa-plugins/webauto/pkg/playwright"
	"github.com/oa-plugins/webauto/pkg/response"
	"github.com/spf13/cobra"
)

var (
	waitSelector   string
	waitCondition  string
	waitTimeoutMs  int
)

var elementWaitCmd = &cobra.Command{
	Use:   "element-wait",
	Short: "Wait for an element to meet a specific condition",
	Long:  `Wait for an element to become visible, hidden, attached to DOM, or detached from DOM. Eliminates unreliable sleep delays for AJAX/dynamic content.`,
	Run:   runElementWait,
}

func init() {
	elementWaitCmd.Flags().StringVar(&sessionID, "session-id", "", "Session ID (required)")
	elementWaitCmd.Flags().StringVar(&waitSelector, "element-selector", "", "CSS selector or XPath (required)")
	elementWaitCmd.Flags().StringVar(&waitCondition, "wait-for", "visible", "Wait condition: visible, hidden, attached, detached (default: visible)")
	elementWaitCmd.Flags().IntVar(&waitTimeoutMs, "timeout-ms", 30000, "Timeout in milliseconds")

	elementWaitCmd.MarkFlagRequired("session-id")
	elementWaitCmd.MarkFlagRequired("element-selector")
}

func runElementWait(cmd *cobra.Command, args []string) {
	startTime := time.Now()
	ctx := context.Background()

	// Validate wait condition
	validConditions := map[string]bool{
		"visible":  true,
		"hidden":   true,
		"attached": true,
		"detached": true,
	}

	if !validConditions[waitCondition] {
		resp := response.Error(
			"INVALID_WAIT_CONDITION",
			"Invalid wait condition: "+waitCondition,
			"Use one of: visible, hidden, attached, detached",
			map[string]interface{}{
				"session_id":       sessionID,
				"element_selector": waitSelector,
				"wait_condition":   waitCondition,
			},
			startTime,
		)
		resp.Print()
		return
	}

	// Get global session manager (singleton pattern)
	sessionMgr := playwright.GetGlobalSessionManager()

	// Send wait command to session
	waitCmd := map[string]interface{}{
		"command":       "wait",
		"selector":      waitSelector,
		"waitCondition": waitCondition,
		"timeout":       waitTimeoutMs,
	}

	result, err := sessionMgr.SendCommand(ctx, sessionID, waitCmd)
	if err != nil {
		resp := response.Error(
			response.ErrTimeoutExceeded,
			"Failed to wait for element: "+err.Error(),
			"Verify session ID, element selector, and timeout value",
			map[string]interface{}{
				"session_id":       sessionID,
				"element_selector": waitSelector,
				"wait_condition":   waitCondition,
				"timeout_ms":       waitTimeoutMs,
			},
			startTime,
		)
		resp.Print()
		return
	}

	if !result.Success {
		resp := response.Error(
			response.ErrTimeoutExceeded,
			"Wait failed: "+result.Error,
			"Element did not meet wait condition within timeout",
			map[string]interface{}{
				"session_id":       sessionID,
				"element_selector": waitSelector,
				"wait_condition":   waitCondition,
				"timeout_ms":       waitTimeoutMs,
			},
			startTime,
		)
		resp.Print()
		return
	}

	// Success response
	resp := response.Success(map[string]interface{}{
		"session_id":       sessionID,
		"element_selector": waitSelector,
		"wait_condition":   result.Data["wait_condition"],
		"waited_ms":        result.Data["waited_ms"],
		"element_found":    result.Data["element_found"],
	}, startTime)
	resp.Print()
}

// GetElementWaitCommand returns the element-wait command for registration
func GetElementWaitCommand() *cobra.Command {
	return elementWaitCmd
}
