package cli

import (
	"context"
	"time"

	"github.com/oa-plugins/webauto/pkg/playwright"
	"github.com/oa-plugins/webauto/pkg/response"
	"github.com/spf13/cobra"
)

var (
	browserType     string
	headless        bool
	viewportWidth   int
	viewportHeight  int
	userAgent       string
)

var browserLaunchCmd = &cobra.Command{
	Use:   "browser-launch",
	Short: "Launch a browser instance",
	Long:  `Launch a browser instance and return a session ID for subsequent commands.`,
	Run:   runBrowserLaunch,
}

func init() {
	browserLaunchCmd.Flags().StringVar(&browserType, "browser-type", "chromium", "Browser type (chromium|firefox|webkit)")
	browserLaunchCmd.Flags().BoolVar(&headless, "headless", true, "Headless mode")
	browserLaunchCmd.Flags().IntVar(&viewportWidth, "viewport-width", 1920, "Viewport width")
	browserLaunchCmd.Flags().IntVar(&viewportHeight, "viewport-height", 1080, "Viewport height")
	browserLaunchCmd.Flags().StringVar(&userAgent, "user-agent", "", "User-Agent override")
}

func runBrowserLaunch(cmd *cobra.Command, args []string) {
	startTime := time.Now()
	ctx := context.Background()

	// Get global session manager (singleton pattern)
	sessionMgr := playwright.GetGlobalSessionManager()

	// Create browser session
	session, err := sessionMgr.Create(ctx, browserType, headless)
	if err != nil {
		resp := response.Error(
			response.ErrBrowserLaunchFailed,
			"Failed to launch browser: "+err.Error(),
			"Check Playwright installation and browser binaries",
			map[string]interface{}{
				"browser_type": browserType,
				"headless":     headless,
			},
			startTime,
		)
		resp.Print()
		return
	}

	// Success response
	resp := response.Success(map[string]interface{}{
		"session_id":   session.ID,
		"browser_type": session.BrowserType,
		"headless":     session.Headless,
		"viewport": map[string]int{
			"width":  viewportWidth,
			"height": viewportHeight,
		},
		"user_agent":  userAgent,
		"created_at":  session.CreatedAt.Format(time.RFC3339),
	}, startTime)
	resp.Print()
}
