package cli

import (
	"context"
	"time"

	"github.com/oa-plugins/webauto/pkg/playwright"
	"github.com/oa-plugins/webauto/pkg/response"
	"github.com/spf13/cobra"
)

var (
	evaluateScript  string
	evaluateTimeout int
)

var pageEvaluateCmd = &cobra.Command{
	Use:   "page-evaluate",
	Short: "Execute custom JavaScript in the page context",
	Long: `Execute custom JavaScript code in the browser page context and return the result.
The script runs in the page's JavaScript context and can access the DOM and page variables.
Only serializable values can be returned (primitives, objects, arrays). Functions and DOM nodes cannot be serialized.`,
	Run: runPageEvaluate,
}

func init() {
	pageEvaluateCmd.Flags().StringVar(&sessionID, "session-id", "", "Session ID (required)")
	pageEvaluateCmd.Flags().StringVar(&evaluateScript, "script", "", "JavaScript code to execute (required)")
	pageEvaluateCmd.Flags().IntVar(&evaluateTimeout, "timeout-ms", 30000, "Execution timeout in milliseconds")

	pageEvaluateCmd.MarkFlagRequired("session-id")
	pageEvaluateCmd.MarkFlagRequired("script")
}

func runPageEvaluate(cmd *cobra.Command, args []string) {
	startTime := time.Now()
	ctx := context.Background()

	// Get global session manager (singleton pattern)
	sessionMgr := playwright.GetGlobalSessionManager()

	// Send evaluate command to session
	evaluateCmd := map[string]interface{}{
		"command": "evaluate",
		"script":  evaluateScript,
		"timeout": evaluateTimeout,
	}

	result, err := sessionMgr.SendCommand(ctx, sessionID, evaluateCmd)
	if err != nil {
		resp := response.Error(
			response.ErrSessionNotFound,
			"Failed to execute script: "+err.Error(),
			"Verify session ID is valid and session is still active",
			map[string]interface{}{
				"session_id": sessionID,
			},
			startTime,
		)
		resp.Print()
		return
	}

	if !result.Success {
		resp := response.Error(
			response.ErrScriptExecutionFailed,
			"Script execution failed: "+result.Error,
			"Check JavaScript syntax and ensure script returns a serializable value",
			map[string]interface{}{
				"session_id": sessionID,
				"script":     evaluateScript,
			},
			startTime,
		)
		resp.Print()
		return
	}

	// Extract result and type from response
	scriptResult := result.Data["result"]
	resultType, _ := result.Data["result_type"].(string)

	// Success response
	resp := response.Success(map[string]interface{}{
		"session_id":  sessionID,
		"result":      scriptResult,
		"result_type": resultType,
	}, startTime)
	resp.Print()
}
