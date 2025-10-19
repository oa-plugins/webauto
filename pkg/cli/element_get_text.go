package cli

import (
	"context"
	"time"

	"github.com/oa-plugins/webauto/pkg/playwright"
	"github.com/oa-plugins/webauto/pkg/response"
	"github.com/spf13/cobra"
)

var (
	getTextSelector string
	getTextTimeout  int
)

var elementGetTextCmd = &cobra.Command{
	Use:   "element-get-text",
	Short: "Get text content from an element",
	Long:  `Extract text content from an element identified by a CSS selector or XPath.`,
	Run:   runElementGetText,
}

func init() {
	elementGetTextCmd.Flags().StringVar(&sessionID, "session-id", "", "Session ID (required)")
	elementGetTextCmd.Flags().StringVar(&getTextSelector, "element-selector", "", "CSS selector or XPath (required)")
	elementGetTextCmd.Flags().IntVar(&getTextTimeout, "timeout-ms", 30000, "Timeout in milliseconds")

	elementGetTextCmd.MarkFlagRequired("session-id")
	elementGetTextCmd.MarkFlagRequired("element-selector")
}

func runElementGetText(cmd *cobra.Command, args []string) {
	startTime := time.Now()
	ctx := context.Background()

	// Get global session manager (singleton pattern)
	sessionMgr := playwright.GetGlobalSessionManager()

	// Send get-text command to session
	getTextCmd := map[string]interface{}{
		"command":  "get-text",
		"selector": getTextSelector,
		"timeout":  getTextTimeout,
	}

	result, err := sessionMgr.SendCommand(ctx, sessionID, getTextCmd)
	if err != nil {
		resp := response.Error(
			response.ErrElementNotFound,
			"Failed to get text: "+err.Error(),
			"Verify session ID and element selector",
			map[string]interface{}{
				"session_id":       sessionID,
				"element_selector": getTextSelector,
			},
			startTime,
		)
		resp.Print()
		return
	}

	if !result.Success {
		resp := response.Error(
			response.ErrElementNotFound,
			"Get text failed: "+result.Error,
			"Check if element exists and is accessible",
			map[string]interface{}{
				"session_id":       sessionID,
				"element_selector": getTextSelector,
			},
			startTime,
		)
		resp.Print()
		return
	}

	// Success response
	resp := response.Success(map[string]interface{}{
		"session_id":       sessionID,
		"element_selector": getTextSelector,
		"text":             result.Data["text"],
		"element_count":    result.Data["element_count"],
	}, startTime)
	resp.Print()
}

// GetElementGetTextCommand returns the element-get-text command for registration
func GetElementGetTextCommand() *cobra.Command {
	return elementGetTextCmd
}
