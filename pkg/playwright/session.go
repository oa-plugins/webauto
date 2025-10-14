package playwright

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/oa-plugins/webauto/pkg/config"
	"github.com/oa-plugins/webauto/pkg/ipc"
)

// Session represents a browser session
type Session struct {
	ID          string
	BrowserType string
	Headless    bool
	CreatedAt   time.Time
	LastUsedAt  time.Time
	Browser     interface{} // WebSocket endpoint (string) for browser reconnection
	Page        interface{} // Page reference (for future use)
	Process     interface{} // Node.js process reference (for cleanup)
}

// SessionManager manages browser sessions
type SessionManager struct {
	cfg      *config.Config
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewSessionManager creates a new SessionManager instance
func NewSessionManager(cfg *config.Config) *SessionManager {
	return &SessionManager{
		cfg:      cfg,
		sessions: make(map[string]*Session),
	}
}

// Create creates a new browser session
func (sm *SessionManager) Create(ctx context.Context, browserType string, headless bool) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check session limit
	if len(sm.sessions) >= sm.cfg.SessionMaxCount {
		return nil, fmt.Errorf("max sessions reached (%d)", sm.cfg.SessionMaxCount)
	}

	// Generate unique session ID
	sessionID := "ses_" + uuid.New().String()[:8]

	// Build Playwright launch script that keeps browser alive
	script := fmt.Sprintf(`
		const { chromium, firefox, webkit } = require('playwright');

		(async () => {
			try {
				let browser;
				const browserType = '%s';
				const headless = %t;

				// Select browser based on type
				if (browserType === 'chromium') {
					browser = await chromium.launch({ headless });
				} else if (browserType === 'firefox') {
					browser = await firefox.launch({ headless });
				} else if (browserType === 'webkit') {
					browser = await webkit.launch({ headless });
				} else {
					throw new Error('Invalid browser type: ' + browserType);
				}

				// Create a new page
				const page = await browser.newPage();

				// Get browser info
				const version = await browser.version();
				const isConnected = browser.isConnected();

				// Output session info to stdout (will be read by Go)
				console.log(JSON.stringify({
					success: true,
					data: {
						browserType: browserType,
						headless: headless,
						version: version,
						isConnected: isConnected
					}
				}));

				// Keep process alive to maintain browser session
				// This process will be killed when the session is closed
				process.stdin.resume();

			} catch (error) {
				console.log(JSON.stringify({
					success: false,
					error: error.message
				}));
				process.exit(1);
			}
		})();
	`, browserType, headless)

	// Create command to run Node.js script
	cmd := exec.CommandContext(ctx, sm.cfg.PlaywrightNodePath, "-e", script)

	// Get stdout pipe for reading launch response
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Start the process (non-blocking)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start browser process: %w", err)
	}

	// Read first line of output (launch response)
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to read browser launch response")
	}

	// Parse launch response
	var response ipc.NodeResponse
	if err := json.Unmarshal(scanner.Bytes(), &response); err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to parse launch response: %w", err)
	}

	if !response.Success {
		cmd.Process.Kill()
		errMsg := "unknown error"
		if response.Error != "" {
			errMsg = response.Error
		}
		return nil, fmt.Errorf("browser launch failed: %s", errMsg)
	}

	// Extract browser info from response
	browserVersion, _ := response.Data["version"].(string)
	isConnected, _ := response.Data["isConnected"].(bool)

	// Log successful launch (for debugging)
	if !isConnected {
		cmd.Process.Kill()
		return nil, fmt.Errorf("browser launched but is not connected")
	}

	// Create session with browser info
	session := &Session{
		ID:          sessionID,
		BrowserType: browserType,
		Headless:    headless,
		CreatedAt:   time.Now(),
		LastUsedAt:  time.Now(),
		Browser:     browserVersion, // Store browser version for info
		Process:     cmd,             // Store process for cleanup
	}

	// Store session
	sm.sessions[sessionID] = session

	return session, nil
}

// Get retrieves a session by ID
func (sm *SessionManager) Get(sessionID string) (*Session, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// Update last used time
	session.LastUsedAt = time.Now()

	return session, nil
}

// Close closes a session and releases resources
func (sm *SessionManager) Close(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Kill the browser process if it exists
	if session.Process != nil {
		if cmd, ok := session.Process.(*exec.Cmd); ok {
			if cmd.Process != nil {
				// Kill the process (this will close the browser)
				if err := cmd.Process.Kill(); err != nil {
					// Log error but don't fail the close operation
					fmt.Printf("Warning: failed to kill browser process: %v\n", err)
				}
			}
		}
	}

	// Remove session from map
	delete(sm.sessions, sessionID)

	return nil
}

// List returns all active sessions
func (sm *SessionManager) List() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// CleanupExpired removes expired sessions based on timeout
func (sm *SessionManager) CleanupExpired() int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	timeout := time.Duration(sm.cfg.SessionTimeoutSeconds) * time.Second
	now := time.Now()
	cleaned := 0

	for sessionID, session := range sm.sessions {
		if now.Sub(session.LastUsedAt) > timeout {
			// Kill the browser process if it exists
			if session.Process != nil {
				if cmd, ok := session.Process.(*exec.Cmd); ok {
					if cmd.Process != nil {
						cmd.Process.Kill()
					}
				}
			}

			delete(sm.sessions, sessionID)
			cleaned++
		}
	}

	return cleaned
}

// Count returns the number of active sessions
func (sm *SessionManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.sessions)
}
