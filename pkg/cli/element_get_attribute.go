package cli

import (
	"context"
	"time"

	"github.com/oa-plugins/webauto/pkg/playwright"
	"github.com/oa-plugins/webauto/pkg/response"
	"github.com/spf13/cobra"
)

var (
	getAttributeSelector string
	getAttributeName     string
	getAttributeTimeout  int
)

var elementGetAttributeCmd = &cobra.Command{
	Use:   "element-get-attribute",
	Short: "Get attribute value from an element",
	Long:  `Extract attribute value (href, src, class, id, data-*, aria-label, etc.) from an element identified by a CSS selector or XPath.`,
	Run:   runElementGetAttribute,
}

func init() {
	elementGetAttributeCmd.Flags().StringVar(&sessionID, "session-id", "", "Session ID (required)")
	elementGetAttributeCmd.Flags().StringVar(&getAttributeSelector, "element-selector", "", "CSS selector or XPath (required)")
	elementGetAttributeCmd.Flags().StringVar(&getAttributeName, "attribute-name", "", "Attribute name to extract (required)")
	elementGetAttributeCmd.Flags().IntVar(&getAttributeTimeout, "timeout-ms", 30000, "Timeout in milliseconds")

	elementGetAttributeCmd.MarkFlagRequired("session-id")
	elementGetAttributeCmd.MarkFlagRequired("element-selector")
	elementGetAttributeCmd.MarkFlagRequired("attribute-name")
}

func runElementGetAttribute(cmd *cobra.Command, args []string) {
	startTime := time.Now()
	ctx := context.Background()

	// Get global session manager (singleton pattern)
	sessionMgr := playwright.GetGlobalSessionManager()

	// Send get-attribute command to session
	getAttributeCmd := map[string]interface{}{
		"command":       "get-attribute",
		"selector":      getAttributeSelector,
		"attributeName": getAttributeName,
		"timeout":       getAttributeTimeout,
	}

	result, err := sessionMgr.SendCommand(ctx, sessionID, getAttributeCmd)
	if err != nil {
		resp := response.Error(
			response.ErrElementNotFound,
			"Failed to get attribute: "+err.Error(),
			"Verify session ID, element selector, and attribute name",
			map[string]interface{}{
				"session_id":       sessionID,
				"element_selector": getAttributeSelector,
				"attribute_name":   getAttributeName,
			},
			startTime,
		)
		resp.Print()
		return
	}

	if !result.Success {
		resp := response.Error(
			response.ErrElementNotFound,
			"Get attribute failed: "+result.Error,
			"Check if element exists and has the specified attribute",
			map[string]interface{}{
				"session_id":       sessionID,
				"element_selector": getAttributeSelector,
				"attribute_name":   getAttributeName,
			},
			startTime,
		)
		resp.Print()
		return
	}

	// Success response
	resp := response.Success(map[string]interface{}{
		"session_id":        sessionID,
		"element_selector":  getAttributeSelector,
		"attribute_name":    getAttributeName,
		"attribute_value":   result.Data["attribute_value"],
		"element_count":     result.Data["element_count"],
	}, startTime)
	resp.Print()
}

// GetElementGetAttributeCommand returns the element-get-attribute command for registration
func GetElementGetAttributeCommand() *cobra.Command {
	return elementGetAttributeCmd
}
