package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/oa-plugins/webauto/pkg/playwright"
	"github.com/oa-plugins/webauto/pkg/response"
	"github.com/spf13/cobra"
)

var (
	queryAllSelector   string
	queryAllGetText    bool
	queryAllAttribute  string
	queryAllLimit      int
	queryAllTimeout    int
)

var elementQueryAllCmd = &cobra.Command{
	Use:   "element-query-all",
	Short: "Query multiple elements and extract data in batch",
	Long: `Extract text or attributes from multiple elements matching a selector.
Efficient for lists, tables, search results, and bulk data collection.

Examples:
  # Get text from all blog titles (limit 10)
  webauto element-query-all --session-id ses_abc --element-selector ".blog-title" --get-text --limit 10

  # Get href attributes from all links
  webauto element-query-all --session-id ses_abc --element-selector "a.link" --get-attribute href

  # Get both text and href from search results
  webauto element-query-all --session-id ses_abc --element-selector ".result-item" --get-text --get-attribute href --limit 5`,
	Run: runElementQueryAll,
}

func init() {
	elementQueryAllCmd.Flags().StringVar(&sessionID, "session-id", "", "Session ID (required)")
	elementQueryAllCmd.Flags().StringVar(&queryAllSelector, "element-selector", "", "CSS selector or XPath (required)")
	elementQueryAllCmd.Flags().BoolVar(&queryAllGetText, "get-text", false, "Extract text content from each element")
	elementQueryAllCmd.Flags().StringVar(&queryAllAttribute, "get-attribute", "", "Attribute name to extract (href, src, class, etc.)")
	elementQueryAllCmd.Flags().IntVar(&queryAllLimit, "limit", 0, "Maximum number of elements to process (0 = all elements)")
	elementQueryAllCmd.Flags().IntVar(&queryAllTimeout, "timeout-ms", 30000, "Timeout in milliseconds")

	elementQueryAllCmd.MarkFlagRequired("session-id")
	elementQueryAllCmd.MarkFlagRequired("element-selector")
}

func runElementQueryAll(cmd *cobra.Command, args []string) {
	startTime := time.Now()
	ctx := context.Background()

	// Validate: at least one extraction flag must be set
	if !queryAllGetText && queryAllAttribute == "" {
		resp := response.Error(
			"INVALID_FLAG_COMBINATION",
			"At least one of --get-text or --get-attribute must be specified",
			"Specify --get-text, --get-attribute <name>, or both",
			map[string]interface{}{
				"session_id":       sessionID,
				"element_selector": queryAllSelector,
				"get_text":         queryAllGetText,
				"get_attribute":    queryAllAttribute,
			},
			startTime,
		)
		resp.Print()
		return
	}

	// Get global session manager (singleton pattern)
	sessionMgr := playwright.GetGlobalSessionManager()

	// Send query-all command to session
	queryAllCmd := map[string]interface{}{
		"command":  "query-all",
		"selector": queryAllSelector,
		"getText":  queryAllGetText,
		"limit":    queryAllLimit,
		"timeout":  queryAllTimeout,
	}

	// Add attribute name if specified
	if queryAllAttribute != "" {
		queryAllCmd["attributeName"] = queryAllAttribute
	}

	result, err := sessionMgr.SendCommand(ctx, sessionID, queryAllCmd)
	if err != nil {
		resp := response.Error(
			response.ErrElementNotFound,
			"Failed to query elements: "+err.Error(),
			"Verify session ID and element selector",
			map[string]interface{}{
				"session_id":       sessionID,
				"element_selector": queryAllSelector,
				"get_text":         queryAllGetText,
				"get_attribute":    queryAllAttribute,
				"limit":            queryAllLimit,
			},
			startTime,
		)
		resp.Print()
		return
	}

	if !result.Success {
		// Check if it's a "no elements found" error
		errorCode := response.ErrElementNotFound
		if result.Error == "No elements found: "+queryAllSelector {
			errorCode = "NO_ELEMENTS_FOUND"
		}

		resp := response.Error(
			errorCode,
			"Query all failed: "+result.Error,
			"Check if elements exist and are accessible",
			map[string]interface{}{
				"session_id":       sessionID,
				"element_selector": queryAllSelector,
				"get_text":         queryAllGetText,
				"get_attribute":    queryAllAttribute,
				"limit":            queryAllLimit,
			},
			startTime,
		)
		resp.Print()
		return
	}

	// Build success response
	data := map[string]interface{}{
		"session_id":       sessionID,
		"element_selector": queryAllSelector,
		"element_count":    result.Data["element_count"],
		"elements":         result.Data["elements"],
	}

	// Add limit info if it was applied
	if queryAllLimit > 0 {
		data["limit"] = result.Data["limit"]

		// Add helpful message if limit was reached
		totalCount := int(result.Data["element_count"].(float64))
		if totalCount > queryAllLimit {
			data["note"] = fmt.Sprintf("Returned %d of %d total elements (limited by --limit flag)", queryAllLimit, totalCount)
		}
	}

	resp := response.Success(data, startTime)
	resp.Print()
}

// GetElementQueryAllCommand returns the element-query-all command for registration
func GetElementQueryAllCommand() *cobra.Command {
	return elementQueryAllCmd
}
