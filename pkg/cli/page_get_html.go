package cli

import (
	"context"
	"os"
	"time"

	"github.com/oa-plugins/webauto/pkg/playwright"
	"github.com/oa-plugins/webauto/pkg/response"
	"github.com/spf13/cobra"
)

var (
	getHtmlSelector   string
	getHtmlOutputPath string
	getHtmlTimeout    int
)

var pageGetHtmlCmd = &cobra.Command{
	Use:   "page-get-html",
	Short: "Get HTML source from page or element",
	Long:  `Extract HTML source from the entire page or a specific element identified by a CSS selector or XPath.`,
	Run:   runPageGetHtml,
}

func init() {
	pageGetHtmlCmd.Flags().StringVar(&sessionID, "session-id", "", "Session ID (required)")
	pageGetHtmlCmd.Flags().StringVar(&getHtmlSelector, "element-selector", "", "CSS selector or XPath (optional, omit for full page)")
	pageGetHtmlCmd.Flags().StringVar(&getHtmlOutputPath, "output-path", "", "Output file path (optional, omit to return HTML in JSON)")
	pageGetHtmlCmd.Flags().IntVar(&getHtmlTimeout, "timeout-ms", 30000, "Timeout in milliseconds")

	pageGetHtmlCmd.MarkFlagRequired("session-id")
}

func runPageGetHtml(cmd *cobra.Command, args []string) {
	startTime := time.Now()
	ctx := context.Background()

	// Get global session manager (singleton pattern)
	sessionMgr := playwright.GetGlobalSessionManager()

	// Send get-html command to session
	getHtmlCmd := map[string]interface{}{
		"command": "get-html",
		"timeout": getHtmlTimeout,
	}

	// Add selector only if provided
	if getHtmlSelector != "" {
		getHtmlCmd["selector"] = getHtmlSelector
	}

	result, err := sessionMgr.SendCommand(ctx, sessionID, getHtmlCmd)
	if err != nil {
		resp := response.Error(
			response.ErrPageLoadFailed,
			"Failed to get HTML: "+err.Error(),
			"Verify session ID and element selector",
			map[string]interface{}{
				"session_id":       sessionID,
				"element_selector": getHtmlSelector,
			},
			startTime,
		)
		resp.Print()
		return
	}

	if !result.Success {
		resp := response.Error(
			response.ErrElementNotFound,
			"Get HTML failed: "+result.Error,
			"Check if element exists and is accessible",
			map[string]interface{}{
				"session_id":       sessionID,
				"element_selector": getHtmlSelector,
			},
			startTime,
		)
		resp.Print()
		return
	}

	// Extract HTML from response
	html, ok := result.Data["html"].(string)
	if !ok {
		resp := response.Error(
			response.ErrPageLoadFailed,
			"Failed to get HTML data from response",
			"Internal error",
			map[string]interface{}{
				"session_id": sessionID,
			},
			startTime,
		)
		resp.Print()
		return
	}

	htmlLength := len(html)

	// Prepare response data
	responseData := map[string]interface{}{
		"session_id":  sessionID,
		"html_length": htmlLength,
	}

	if getHtmlSelector != "" {
		responseData["selector"] = getHtmlSelector
	}

	// Write to file if output path is provided
	if getHtmlOutputPath != "" {
		if err := os.WriteFile(getHtmlOutputPath, []byte(html), 0644); err != nil {
			resp := response.Error(
				response.ErrPageLoadFailed,
				"Failed to write HTML file: "+err.Error(),
				"Check file path and permissions",
				map[string]interface{}{
					"session_id":  sessionID,
					"output_path": getHtmlOutputPath,
				},
				startTime,
			)
			resp.Print()
			return
		}
		responseData["output_path"] = getHtmlOutputPath
	} else {
		// Return HTML in response if no output path
		responseData["html"] = html
	}

	// Success response
	resp := response.Success(responseData, startTime)
	resp.Print()
}

// GetPageGetHtmlCommand returns the page-get-html command for registration
func GetPageGetHtmlCommand() *cobra.Command {
	return pageGetHtmlCmd
}
