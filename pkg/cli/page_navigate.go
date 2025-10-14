package cli

import (
	"context"
	"time"

	"github.com/oa-plugins/webauto/pkg/config"
	"github.com/oa-plugins/webauto/pkg/playwright"
	"github.com/oa-plugins/webauto/pkg/response"
	"github.com/spf13/cobra"
)

var (
	pageURL     string
	waitUntil   string
	navTimeout  int
)

var pageNavigateCmd = &cobra.Command{
	Use:   "page-navigate",
	Short: "Navigate to a URL in an existing browser session",
	Long:  `Navigate to a URL and wait for the page to load.`,
	Run:   runPageNavigate,
}

func init() {
	pageNavigateCmd.Flags().StringVar(&sessionID, "session-id", "", "Session ID (required)")
	pageNavigateCmd.Flags().StringVar(&pageURL, "page-url", "", "URL to navigate to (required)")
	pageNavigateCmd.Flags().StringVar(&waitUntil, "wait-until", "load", "When to consider navigation successful (load|domcontentloaded|networkidle)")
	pageNavigateCmd.Flags().IntVar(&navTimeout, "timeout", 30000, "Navigation timeout in milliseconds")

	pageNavigateCmd.MarkFlagRequired("session-id")
	pageNavigateCmd.MarkFlagRequired("page-url")
}

func runPageNavigate(cmd *cobra.Command, args []string) {
	startTime := time.Now()
	ctx := context.Background()

	// Load configuration
	cfg := config.Load()

	// Initialize session manager
	sessionMgr := playwright.NewSessionManager(cfg)

	// Send navigate command to session
	navCmd := map[string]interface{}{
		"command":   "navigate",
		"url":       pageURL,
		"waitUntil": waitUntil,
		"timeout":   navTimeout,
	}

	result, err := sessionMgr.SendCommand(ctx, sessionID, navCmd)
	if err != nil {
		resp := response.Error(
			response.ErrPageNavigationFailed,
			"Failed to navigate: "+err.Error(),
			"Verify session ID and ensure URL is valid",
			map[string]interface{}{
				"session_id": sessionID,
				"page_url":   pageURL,
			},
			startTime,
		)
		resp.Print()
		return
	}

	if !result.Success {
		resp := response.Error(
			response.ErrPageNavigationFailed,
			"Navigation failed: "+result.Error,
			"Check URL and network connectivity",
			map[string]interface{}{
				"session_id": sessionID,
				"page_url":   pageURL,
			},
			startTime,
		)
		resp.Print()
		return
	}

	// Extract navigation results
	finalURL, _ := result.Data["url"].(string)
	pageTitle, _ := result.Data["title"].(string)

	// Success response
	resp := response.Success(map[string]interface{}{
		"session_id": sessionID,
		"url":        finalURL,
		"title":      pageTitle,
		"wait_until": waitUntil,
		"timeout_ms": navTimeout,
	}, startTime)
	resp.Print()
}
