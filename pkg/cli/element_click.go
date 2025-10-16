package cli

import (
	"context"
	"time"

	"github.com/oa-plugins/webauto/pkg/playwright"
	"github.com/oa-plugins/webauto/pkg/response"
	"github.com/spf13/cobra"
)

var (
	elementSelector string
	clickTimeout    int
)

var elementClickCmd = &cobra.Command{
	Use:   "element-click",
	Short: "Click an element on the page",
	Long:  `Click an element identified by a CSS selector.`,
	Run:   runElementClick,
}

func init() {
	elementClickCmd.Flags().StringVar(&sessionID, "session-id", "", "Session ID (required)")
	elementClickCmd.Flags().StringVar(&elementSelector, "element-selector", "", "CSS selector for the element (required)")
	elementClickCmd.Flags().IntVar(&clickTimeout, "timeout", 30000, "Click timeout in milliseconds")

	elementClickCmd.MarkFlagRequired("session-id")
	elementClickCmd.MarkFlagRequired("element-selector")
}

func runElementClick(cmd *cobra.Command, args []string) {
	startTime := time.Now()
	ctx := context.Background()

	// Get global session manager (singleton pattern)
	sessionMgr := playwright.GetGlobalSessionManager()

	// Send click command to session
	clickCmd := map[string]interface{}{
		"command":  "click",
		"selector": elementSelector,
		"timeout":  clickTimeout,
	}

	result, err := sessionMgr.SendCommand(ctx, sessionID, clickCmd)
	if err != nil {
		resp := response.Error(
			response.ErrElementNotFound,
			"Failed to click element: "+err.Error(),
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
			"Click failed: "+result.Error,
			"Check if element is visible and clickable",
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
		"clicked":          true,
		"timeout_ms":       clickTimeout,
	}, startTime)
	resp.Print()
}
