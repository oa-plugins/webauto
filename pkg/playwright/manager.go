package playwright

import (
	"fmt"
	"sync"
	"time"

	"github.com/oa-plugins/webauto/pkg/config"
)

// Global singleton instance
var (
	globalSessionManager *SessionManager
	managerOnce          sync.Once
)

// GetGlobalSessionManager returns the global singleton SessionManager instance.
// This ensures all commands share the same session manager and in-memory session cache,
// eliminating redundant file I/O operations.
//
// Benefits:
// - Shared session cache across all commands
// - Reduced file I/O (no need to reload sessions from disk)
// - Consistent session state
// - Better performance for session lookups
func GetGlobalSessionManager() *SessionManager {
	managerOnce.Do(func() {
		cfg := config.Load()
		globalSessionManager = NewSessionManager(cfg)

		// Start background session cleanup goroutine
		go globalSessionManager.startBackgroundCleanup()
	})

	return globalSessionManager
}

// ResetGlobalSessionManager resets the global singleton instance.
// This is primarily useful for testing.
func ResetGlobalSessionManager() {
	managerOnce = sync.Once{}
	if globalSessionManager != nil {
		// Close all sessions before resetting
		sessions := globalSessionManager.List()
		for _, session := range sessions {
			globalSessionManager.Close(session.ID)
		}
	}
	globalSessionManager = nil
}

// startBackgroundCleanup starts a background goroutine that periodically
// cleans up expired sessions and syncs session state to disk.
func (sm *SessionManager) startBackgroundCleanup() {
	ticker := time.NewTicker(30 * time.Second) // Flush every 30 seconds
	go func() {
		for range ticker.C {
			sm.flushSessionsToDisk()
			sm.CleanupExpired()
		}
	}()
}

// flushSessionsToDisk persists all in-memory sessions to disk.
// This is called periodically by the background cleanup goroutine.
func (sm *SessionManager) flushSessionsToDisk() {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, session := range sm.sessions {
		if err := session.saveSession(); err != nil {
			// Log error but continue with other sessions
			fmt.Printf("Warning: failed to flush session %s: %v\n", session.ID, err)
		}
	}
}
