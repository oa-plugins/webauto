package cli

import (
	"context"
	"time"

	"github.com/oa-plugins/webauto/pkg/playwright"
	"github.com/oa-plugins/webauto/pkg/response"
	"github.com/spf13/cobra"
)

var (
	elementText    string
	typeTimeout    int
)

var elementTypeCmd = &cobra.Command{
	Use:   "element-type",
	Short: "Type text into an element on the page",
	Long:  `Type text into an input field or textarea identified by a CSS selector.`,
	Run:   runElementType,
}

func init() {
	elementTypeCmd.Flags().StringVar(&sessionID, "session-id", "", "Session ID (required)")
	elementTypeCmd.Flags().StringVar(&elementSelector, "element-selector", "", "CSS selector for the element (required)")
	elementTypeCmd.Flags().StringVar(&elementText, "element-text", "", "Text to type (required)")
	elementTypeCmd.Flags().IntVar(&typeTimeout, "timeout", 30000, "Type timeout in milliseconds")

	elementTypeCmd.MarkFlagRequired("session-id")
	elementTypeCmd.MarkFlagRequired("element-selector")
	elementTypeCmd.MarkFlagRequired("element-text")
}

func runElementType(cmd *cobra.Command, args []string) {
	startTime := time.Now()
	ctx := context.Background()


	// Get global session manager (singleton pattern)
	sessionMgr := playwright.GetGlobalSessionManager()

	// Send type command to session
	typeCmd := map[string]interface{}{
		"command":  "type",
		"selector": elementSelector,
		"text":     elementText,
		"timeout":  typeTimeout,
	}

	result, err := sessionMgr.SendCommand(ctx, sessionID, typeCmd)
	if err != nil {
		resp := response.Error(
			response.ErrElementNotFound,
			"Failed to type into element: "+err.Error(),
			"Verify session ID and element selector",
			map[string]interface{}{
				"session_id":       sessionID,
				"element_selector": elementSelector,
			},
			startTime,
		)
		resp.Print()
		return
	}

	if !result.Success {
		resp := response.Error(
			response.ErrElementNotClickable,
			"Type failed: "+result.Error,
			"Check if element is visible and editable",
			map[string]interface{}{
				"session_id":       sessionID,
				"element_selector": elementSelector,
			},
			startTime,
		)
		resp.Print()
		return
	}

	// Success response
	resp := response.Success(map[string]interface{}{
		"session_id":       sessionID,
		"element_selector": elementSelector,
		"element_text":     elementText,
		"typed":            true,
		"timeout_ms":       typeTimeout,
	}, startTime)
	resp.Print()
}
