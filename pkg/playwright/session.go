package playwright

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/oa-plugins/webauto/pkg/config"
	"github.com/oa-plugins/webauto/pkg/ipc"
)

// Session represents a browser session
type Session struct {
	ID          string      `json:"id"`
	BrowserType string      `json:"browser_type"`
	Headless    bool        `json:"headless"`
	CreatedAt   time.Time   `json:"created_at"`
	LastUsedAt  time.Time   `json:"last_used_at"`
	PID         int         `json:"pid"`          // Process ID for reconnection
	Browser     interface{} `json:"-"`            // WebSocket endpoint (string) for browser reconnection
	Page        interface{} `json:"-"`            // Page reference (for future use)
	Process     interface{} `json:"-"`            // Node.js process reference (for cleanup)
}

// sessionDir returns the directory path for session files
func sessionDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".cache", "oa", "webauto", "sessions")
}

// sessionFile returns the file path for a specific session
func sessionFile(sessionID string) string {
	return filepath.Join(sessionDir(), sessionID+".json")
}

// saveSession saves session metadata to a file
func (s *Session) saveSession() error {
	// Create sessions directory if it doesn't exist
	if err := os.MkdirAll(sessionDir(), 0755); err != nil {
		return fmt.Errorf("failed to create sessions directory: %w", err)
	}

	// Marshal session to JSON
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Write to file
	filePath := sessionFile(s.ID)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// loadSession loads session metadata from a file and reattaches to process
func loadSession(sessionID string) (*Session, error) {
	filePath := sessionFile(sessionID)

	// Read session file
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	// Unmarshal session
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Find and attach to existing process
	// Note: On macOS/Linux, os.FindProcess always succeeds
	// Process verification happens when we try to kill it
	process, err := os.FindProcess(session.PID)
	if err != nil {
		return nil, fmt.Errorf("failed to find process %d: %w", session.PID, err)
	}

	session.Process = process

	return &session, nil
}

// deleteSession removes the session file
func deleteSession(sessionID string) error {
	filePath := sessionFile(sessionID)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete session file: %w", err)
	}
	return nil
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
		PID:         cmd.Process.Pid, // Store PID for process reconnection
		Browser:     browserVersion,  // Store browser version for info
		Process:     cmd,              // Store process for cleanup
	}

	// Save session to file
	if err := session.saveSession(); err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// Store session in memory
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

	// Try to load session from file if not in memory
	session, ok := sm.sessions[sessionID]
	if !ok {
		// Load from file
		loadedSession, err := loadSession(sessionID)
		if err != nil {
			return err
		}
		session = loadedSession
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
		} else if proc, ok := session.Process.(*os.Process); ok {
			// Process loaded from file
			if err := proc.Kill(); err != nil {
				fmt.Printf("Warning: failed to kill browser process: %v\n", err)
			}
		}
	}

	// Delete session file
	if err := deleteSession(sessionID); err != nil {
		fmt.Printf("Warning: failed to delete session file: %v\n", err)
	}

	// Remove session from memory map
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
