package cli

import (
	"context"
	"encoding/base64"
	"os"
	"time"

	"github.com/oa-plugins/webauto/pkg/config"
	"github.com/oa-plugins/webauto/pkg/playwright"
	"github.com/oa-plugins/webauto/pkg/response"
	"github.com/spf13/cobra"
)

var (
	imagePath        string
	screenshotType   string
	fullPage         bool
	screenshotTimeout int
)

var pageScreenshotCmd = &cobra.Command{
	Use:   "page-screenshot",
	Short: "Take a screenshot of the current page",
	Long:  `Capture the current page as an image file (PNG or JPEG).`,
	Run:   runPageScreenshot,
}

func init() {
	pageScreenshotCmd.Flags().StringVar(&sessionID, "session-id", "", "Session ID (required)")
	pageScreenshotCmd.Flags().StringVar(&imagePath, "image-path", "screenshot.png", "Output image file path")
	pageScreenshotCmd.Flags().StringVar(&screenshotType, "type", "png", "Screenshot type (png|jpeg)")
	pageScreenshotCmd.Flags().BoolVar(&fullPage, "full-page", false, "Capture full scrollable page")
	pageScreenshotCmd.Flags().IntVar(&screenshotTimeout, "timeout", 30000, "Screenshot timeout in milliseconds")

	pageScreenshotCmd.MarkFlagRequired("session-id")
}

func runPageScreenshot(cmd *cobra.Command, args []string) {
	startTime := time.Now()
	ctx := context.Background()

	// Load configuration
	cfg := config.Load()

	// Initialize session manager
	sessionMgr := playwright.NewSessionManager(cfg)

	// Send screenshot command to session
	screenshotCmd := map[string]interface{}{
		"command":  "screenshot",
		"type":     screenshotType,
		"fullPage": fullPage,
		"timeout":  screenshotTimeout,
	}

	result, err := sessionMgr.SendCommand(ctx, sessionID, screenshotCmd)
	if err != nil {
		resp := response.Error(
			response.ErrPageLoadFailed,
			"Failed to take screenshot: "+err.Error(),
			"Verify session ID and page is loaded",
			map[string]interface{}{
				"session_id": sessionID,
				"image_path": imagePath,
			},
			startTime,
		)
		resp.Print()
		return
	}

	if !result.Success {
		resp := response.Error(
			response.ErrPageLoadFailed,
			"Screenshot failed: "+result.Error,
			"Check if page is ready",
			map[string]interface{}{
				"session_id": sessionID,
				"image_path": imagePath,
			},
			startTime,
		)
		resp.Print()
		return
	}

	// Decode base64 screenshot
	screenshotBase64, ok := result.Data["screenshot"].(string)
	if !ok {
		resp := response.Error(
			response.ErrPageLoadFailed,
			"Failed to get screenshot data from response",
			"Internal error",
			map[string]interface{}{
				"session_id": sessionID,
			},
			startTime,
		)
		resp.Print()
		return
	}

	screenshotBytes, err := base64.StdEncoding.DecodeString(screenshotBase64)
	if err != nil {
		resp := response.Error(
			response.ErrPageLoadFailed,
			"Failed to decode screenshot: "+err.Error(),
			"Internal error",
			map[string]interface{}{
				"session_id": sessionID,
			},
			startTime,
		)
		resp.Print()
		return
	}

	// Write screenshot to file
	if err := os.WriteFile(imagePath, screenshotBytes, 0644); err != nil {
		resp := response.Error(
			response.ErrPageLoadFailed,
			"Failed to write screenshot file: "+err.Error(),
			"Check file path and permissions",
			map[string]interface{}{
				"session_id": sessionID,
				"image_path": imagePath,
			},
			startTime,
		)
		resp.Print()
		return
	}

	// Get file info
	fileInfo, _ := os.Stat(imagePath)
	fileSize := int64(0)
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	}

	// Success response
	resp := response.Success(map[string]interface{}{
		"session_id": sessionID,
		"image_path": imagePath,
		"type":       screenshotType,
		"full_page":  fullPage,
		"file_size":  fileSize,
	}, startTime)
	resp.Print()
}
