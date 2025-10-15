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
	pdfPath         string
	pdfFormat       string
	landscape       bool
	printBackground bool
	pdfTimeout      int
)

var pagePdfCmd = &cobra.Command{
	Use:   "page-pdf",
	Short: "Save the current page as PDF",
	Long:  `Export the current page to a PDF file with configurable format and options.`,
	Run:   runPagePdf,
}

func init() {
	pagePdfCmd.Flags().StringVar(&sessionID, "session-id", "", "Session ID (required)")
	pagePdfCmd.Flags().StringVar(&pdfPath, "pdf-path", "page.pdf", "Output PDF file path")
	pagePdfCmd.Flags().StringVar(&pdfFormat, "pdf-format", "A4", "PDF page format (A4|Letter|Legal)")
	pagePdfCmd.Flags().BoolVar(&landscape, "landscape", false, "Landscape orientation")
	pagePdfCmd.Flags().BoolVar(&printBackground, "print-background", true, "Print background graphics")
	pagePdfCmd.Flags().IntVar(&pdfTimeout, "timeout", 30000, "PDF generation timeout in milliseconds")

	pagePdfCmd.MarkFlagRequired("session-id")
}

func runPagePdf(cmd *cobra.Command, args []string) {
	startTime := time.Now()
	ctx := context.Background()

	// Load configuration
	cfg := config.Load()

	// Initialize session manager
	sessionMgr := playwright.NewSessionManager(cfg)

	// Send PDF command to session
	pdfCmd := map[string]interface{}{
		"command":         "pdf",
		"format":          pdfFormat,
		"landscape":       landscape,
		"printBackground": printBackground,
		"timeout":         pdfTimeout,
	}

	result, err := sessionMgr.SendCommand(ctx, sessionID, pdfCmd)
	if err != nil {
		resp := response.Error(
			response.ErrPageLoadFailed,
			"Failed to generate PDF: "+err.Error(),
			"Verify session ID and page is loaded",
			map[string]interface{}{
				"session_id": sessionID,
				"pdf_path":   pdfPath,
			},
			startTime,
		)
		resp.Print()
		return
	}

	if !result.Success {
		resp := response.Error(
			response.ErrPageLoadFailed,
			"PDF generation failed: "+result.Error,
			"Check if page is ready",
			map[string]interface{}{
				"session_id": sessionID,
				"pdf_path":   pdfPath,
			},
			startTime,
		)
		resp.Print()
		return
	}

	// Decode base64 PDF
	pdfBase64, ok := result.Data["pdf"].(string)
	if !ok {
		resp := response.Error(
			response.ErrPageLoadFailed,
			"Failed to get PDF data from response",
			"Internal error",
			map[string]interface{}{
				"session_id": sessionID,
			},
			startTime,
		)
		resp.Print()
		return
	}

	pdfBytes, err := base64.StdEncoding.DecodeString(pdfBase64)
	if err != nil {
		resp := response.Error(
			response.ErrPageLoadFailed,
			"Failed to decode PDF: "+err.Error(),
			"Internal error",
			map[string]interface{}{
				"session_id": sessionID,
			},
			startTime,
		)
		resp.Print()
		return
	}

	// Write PDF to file
	if err := os.WriteFile(pdfPath, pdfBytes, 0644); err != nil {
		resp := response.Error(
			response.ErrPageLoadFailed,
			"Failed to write PDF file: "+err.Error(),
			"Check file path and permissions",
			map[string]interface{}{
				"session_id": sessionID,
				"pdf_path":   pdfPath,
			},
			startTime,
		)
		resp.Print()
		return
	}

	// Get file info
	fileInfo, _ := os.Stat(pdfPath)
	fileSize := int64(0)
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	}

	// Success response
	resp := response.Success(map[string]interface{}{
		"session_id":       sessionID,
		"pdf_path":         pdfPath,
		"pdf_format":       pdfFormat,
		"landscape":        landscape,
		"print_background": printBackground,
		"file_size":        fileSize,
	}, startTime)
	resp.Print()
}
