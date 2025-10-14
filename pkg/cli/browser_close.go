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
	sessionID string
)

var browserCloseCmd = &cobra.Command{
	Use:   "browser-close",
	Short: "Close a browser session",
	Long:  `Close a browser session and release all associated resources.`,
	Run:   runBrowserClose,
}

func init() {
	browserCloseCmd.Flags().StringVar(&sessionID, "session-id", "", "Session ID to close (required)")
	browserCloseCmd.MarkFlagRequired("session-id")
}

func runBrowserClose(cmd *cobra.Command, args []string) {
	startTime := time.Now()
	ctx := context.Background()

	// Load configuration
	cfg := config.Load()

	// Initialize session manager
	sessionMgr := playwright.NewSessionManager(cfg)

	// Validate session ID format
	if sessionID == "" {
		resp := response.Error(
			response.ErrSessionNotFound,
			"Session ID is required",
			"Provide --session-id flag with a valid session ID",
			map[string]interface{}{
				"session_id": sessionID,
			},
			startTime,
		)
		resp.Print()
		return
	}

	// Close browser session
	err := sessionMgr.Close(sessionID)
	if err != nil {
		resp := response.Error(
			response.ErrSessionNotFound,
			"Failed to close browser session: "+err.Error(),
			"Verify session ID with session-list command",
			map[string]interface{}{
				"session_id": sessionID,
			},
			startTime,
		)
		resp.Print()
		return
	}

	// Success response
	resp := response.Success(map[string]interface{}{
		"session_id": sessionID,
		"closed_at":  time.Now().Format(time.RFC3339),
	}, startTime)
	resp.Print()

	_ = ctx // Avoid unused variable warning
}
