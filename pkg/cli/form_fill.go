package cli

import (
	"context"
	"encoding/json"
	"time"

	"github.com/oa-plugins/webauto/pkg/playwright"
	"github.com/oa-plugins/webauto/pkg/response"
	"github.com/spf13/cobra"
)

var (
	formData       string
	submitForm     bool
	submitSelector string
	formTimeout    int
)

var formFillCmd = &cobra.Command{
	Use:   "form-fill",
	Short: "Fill multiple form fields at once",
	Long:  `Fill multiple form fields with values provided in JSON format. Optionally submit the form after filling.`,
	Run:   runFormFill,
}

func init() {
	formFillCmd.Flags().StringVar(&sessionID, "session-id", "", "Session ID (required)")
	formFillCmd.Flags().StringVar(&formData, "form-data", "", "JSON object with selector:value pairs (required)")
	formFillCmd.Flags().BoolVar(&submitForm, "submit", false, "Submit the form after filling")
	formFillCmd.Flags().StringVar(&submitSelector, "submit-selector", "", "CSS selector for submit button (required if --submit is true)")
	formFillCmd.Flags().IntVar(&formTimeout, "timeout", 30000, "Timeout for each field in milliseconds")

	formFillCmd.MarkFlagRequired("session-id")
	formFillCmd.MarkFlagRequired("form-data")
}

func runFormFill(cmd *cobra.Command, args []string) {
	startTime := time.Now()
	ctx := context.Background()


	// Get global session manager (singleton pattern)
	sessionMgr := playwright.GetGlobalSessionManager()

	// Parse form data JSON
	var fields map[string]string
	if err := json.Unmarshal([]byte(formData), &fields); err != nil {
		resp := response.Error(
			response.ErrPageLoadFailed,
			"Failed to parse form-data: "+err.Error(),
			"Provide valid JSON object with selector:value pairs",
			map[string]interface{}{
				"session_id": sessionID,
				"form_data":  formData,
			},
			startTime,
		)
		resp.Print()
		return
	}

	if len(fields) == 0 {
		resp := response.Error(
			response.ErrPageLoadFailed,
			"Form data is empty",
			"Provide at least one field in form-data",
			map[string]interface{}{
				"session_id": sessionID,
			},
			startTime,
		)
		resp.Print()
		return
	}

	// Fill each field
	filledFields := make([]map[string]interface{}, 0, len(fields))
	for selector, value := range fields {
		typeCmd := map[string]interface{}{
			"command":  "type",
			"selector": selector,
			"text":     value,
			"timeout":  formTimeout,
		}

		result, err := sessionMgr.SendCommand(ctx, sessionID, typeCmd)
		if err != nil {
			resp := response.Error(
				response.ErrElementNotFound,
				"Failed to fill field: "+err.Error(),
				"Verify selector and session ID",
				map[string]interface{}{
					"session_id": sessionID,
					"selector":   selector,
					"value":      value,
				},
				startTime,
			)
			resp.Print()
			return
		}

		if !result.Success {
			resp := response.Error(
				response.ErrElementNotClickable,
				"Field fill failed: "+result.Error,
				"Check if element is visible and editable",
				map[string]interface{}{
					"session_id": sessionID,
					"selector":   selector,
				},
				startTime,
			)
			resp.Print()
			return
		}

		filledFields = append(filledFields, map[string]interface{}{
			"selector": selector,
			"value":    value,
			"filled":   true,
		})
	}

	// Submit form if requested
	var submitted bool
	if submitForm {
		if submitSelector == "" {
			resp := response.Error(
				response.ErrPageLoadFailed,
				"Submit selector is required when --submit is true",
				"Provide --submit-selector flag with a valid CSS selector",
				map[string]interface{}{
					"session_id": sessionID,
					"submit":     submitForm,
				},
				startTime,
			)
			resp.Print()
			return
		}

		clickCmd := map[string]interface{}{
			"command":  "click",
			"selector": submitSelector,
			"timeout":  formTimeout,
		}

		result, err := sessionMgr.SendCommand(ctx, sessionID, clickCmd)
		if err != nil {
			resp := response.Error(
				response.ErrElementNotFound,
				"Failed to click submit button: "+err.Error(),
				"Verify submit-selector",
				map[string]interface{}{
					"session_id":      sessionID,
					"submit_selector": submitSelector,
				},
				startTime,
			)
			resp.Print()
			return
		}

		if !result.Success {
			resp := response.Error(
				response.ErrElementNotClickable,
				"Submit button click failed: "+result.Error,
				"Check if submit button is visible and clickable",
				map[string]interface{}{
					"session_id":      sessionID,
					"submit_selector": submitSelector,
				},
				startTime,
			)
			resp.Print()
			return
		}

		submitted = true
	}

	// Success response
	resp := response.Success(map[string]interface{}{
		"session_id":    sessionID,
		"fields_filled": len(filledFields),
		"fields":        filledFields,
		"submitted":     submitted,
		"timeout_ms":    formTimeout,
	}, startTime)
	resp.Print()
}
