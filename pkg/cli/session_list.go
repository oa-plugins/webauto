package cli

import (
	"time"

	"github.com/oa-plugins/webauto/pkg/playwright"
	"github.com/oa-plugins/webauto/pkg/response"
	"github.com/spf13/cobra"
)

var sessionListCmd = &cobra.Command{
	Use:   "session-list",
	Short: "List all browser sessions",
	Long:  `Display all active and persisted browser sessions from memory and file system.`,
	Run:   runSessionList,
}

func init() {
	// No flags needed for session-list
}

func runSessionList(cmd *cobra.Command, args []string) {
	startTime := time.Now()


	// Get global session manager (singleton pattern)
	sessionMgr := playwright.GetGlobalSessionManager()

	// Get all sessions (memory + file system)
	sessions := sessionMgr.ListAll()

	// Build session list for response
	sessionList := make([]map[string]interface{}, 0, len(sessions))
	for _, session := range sessions {
		sessionList = append(sessionList, map[string]interface{}{
			"session_id":   session.ID,
			"browser_type": session.BrowserType,
			"headless":     session.Headless,
			"pid":          session.PID,
			"port":         session.Port,
			"created_at":   session.CreatedAt.Format(time.RFC3339),
			"last_used_at": session.LastUsedAt.Format(time.RFC3339),
		})
	}

	// Success response
	resp := response.Success(map[string]interface{}{
		"session_count": len(sessions),
		"sessions":      sessionList,
	}, startTime)
	resp.Print()
}
