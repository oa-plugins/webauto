package playwright

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/oa-plugins/webauto/pkg/bootstrap"
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
	PID         int         `json:"pid"`  // Process ID for reconnection
	Port        int         `json:"port"` // TCP port for IPC
	Browser     interface{} `json:"-"`    // WebSocket endpoint (string) for browser reconnection
	Page        interface{} `json:"-"`    // Page reference (for future use)
	Process     interface{} `json:"-"`    // Node.js process reference (for cleanup)
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
	sessions map[string]*managedSession
	mu       sync.RWMutex
}

// NewSessionManager creates a new SessionManager instance
func NewSessionManager(cfg *config.Config) *SessionManager {
	return &SessionManager{
		cfg:      cfg,
		sessions: make(map[string]*managedSession),
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

	// Ensure session runner script is available and configure launch parameters
	scriptPath, err := ensureSessionRunnerScript()
	if err != nil {
		return nil, err
	}

	runnerConfig := map[string]interface{}{
		"browserType": browserType,
		"headless":    headless,
	}

	configJSON, err := json.Marshal(runnerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to encode runner config: %w", err)
	}

	// Create command to run Node.js script
	cmd := exec.CommandContext(ctx, sm.cfg.PlaywrightNodePath, scriptPath)

	// Set working directory to cache dir so Node.js can find playwright module
	cmd.Dir = bootstrap.GetCacheDir()

	// Set environment variables required by the Playwright runner
	browsersDir := bootstrap.GetBrowsersDir()
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PLAYWRIGHT_BROWSERS_PATH=%s", browsersDir),
		fmt.Sprintf("WEBAUTO_RUNNER_CONFIG=%s", string(configJSON)),
	)

	// Get stdout pipe for reading launch response
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Get stderr pipe for error messages
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process (non-blocking)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start browser process: %w", err)
	}

	// Read first line of output (launch response) with timeout
	type scanResult struct {
		data []byte
		err  error
	}
	scanChan := make(chan scanResult, 1)

	go func() {
		scanner := bufio.NewScanner(stdout)
		if scanner.Scan() {
			scanChan <- scanResult{data: scanner.Bytes(), err: nil}
		} else {
			scanChan <- scanResult{data: nil, err: scanner.Err()}
		}
	}()

	// Wait for scan result or timeout
	var scanData []byte
	select {
	case result := <-scanChan:
		if result.err != nil || result.data == nil {
			// Read stderr for error details
			stderrData, _ := io.ReadAll(stderr)
			cmd.Process.Kill()
			if len(stderrData) > 0 {
				return nil, fmt.Errorf("failed to read browser launch response, stderr: %s", string(stderrData))
			}
			return nil, fmt.Errorf("failed to read browser launch response")
		}
		scanData = result.data
	case <-time.After(30 * time.Second):
		// Timeout waiting for response
		stderrData, _ := io.ReadAll(stderr)
		cmd.Process.Kill()
		if len(stderrData) > 0 {
			return nil, fmt.Errorf("timeout waiting for browser launch response, stderr: %s", string(stderrData))
		}
		return nil, fmt.Errorf("timeout waiting for browser launch response")
	}

	// Parse launch response
	var response ipc.NodeResponse
	if err := json.Unmarshal(scanData, &response); err != nil {
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
	port, _ := response.Data["port"].(float64) // JSON numbers are float64

	// Log successful launch (for debugging)
	if !isConnected {
		cmd.Process.Kill()
		return nil, fmt.Errorf("browser launched but is not connected")
	}

	if port == 0 {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to get TCP port from browser launch response")
	}

	// Create session with browser info
	session := &Session{
		ID:          sessionID,
		BrowserType: browserType,
		Headless:    headless,
		CreatedAt:   time.Now(),
		LastUsedAt:  time.Now(),
		PID:         cmd.Process.Pid, // Store PID for process reconnection
		Port:        int(port),       // Store TCP port for IPC
		Browser:     browserVersion,  // Store browser version for info
		Process:     cmd,             // Store process for cleanup
	}

	// Save session to file
	if err := session.saveSession(); err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	worker, err := newSessionWorker(ctx, session)
	if err != nil {
		cmd.Process.Kill()
		_ = deleteSession(sessionID)
		return nil, fmt.Errorf("failed to establish session worker: %w", err)
	}

	// Store session in memory
	sm.sessions[sessionID] = &managedSession{
		session: session,
		worker:  worker,
	}

	return session, nil
}

// Get retrieves a session by ID
func (sm *SessionManager) Get(sessionID string) (*Session, error) {
	sm.mu.RLock()
	managed, ok := sm.sessions[sessionID]
	sm.mu.RUnlock()

	if ok {
		sm.mu.Lock()
		managed.session.LastUsedAt = time.Now()
		sm.mu.Unlock()
		return managed.session, nil
	}

	session, err := loadSession(sessionID)
	if err != nil {
		return nil, err
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if existing, exists := sm.sessions[sessionID]; exists {
		existing.session.LastUsedAt = time.Now()
		return existing.session, nil
	}

	session.LastUsedAt = time.Now()
	sm.sessions[sessionID] = &managedSession{session: session}

	return session, nil
}

// Close closes a session and releases resources
func (sm *SessionManager) Close(sessionID string) error {
	var managed *managedSession

	sm.mu.Lock()
	if ms, ok := sm.sessions[sessionID]; ok {
		managed = ms
		delete(sm.sessions, sessionID)
	}
	sm.mu.Unlock()

	var session *Session

	if managed != nil {
		if managed.worker != nil {
			managed.worker.Close()
		}
		session = managed.session
	} else {
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
				if err := cmd.Process.Kill(); err != nil {
					fmt.Printf("Warning: failed to kill browser process: %v\n", err)
				}
			}
		} else if proc, ok := session.Process.(*os.Process); ok {
			if err := proc.Kill(); err != nil {
				fmt.Printf("Warning: failed to kill browser process: %v\n", err)
			}
		}
	}

	// Delete session file
	if err := deleteSession(sessionID); err != nil {
		fmt.Printf("Warning: failed to delete session file: %v\n", err)
	}

	return nil
}

// List returns all active sessions
func (sm *SessionManager) List() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*Session, 0, len(sm.sessions))
	for _, managed := range sm.sessions {
		sessions = append(sessions, managed.session)
	}

	return sessions
}

// ListAll returns all sessions (memory + file system)
func (sm *SessionManager) ListAll() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Start with sessions in memory
	sessionMap := make(map[string]*Session)
	for id, managed := range sm.sessions {
		sessionMap[id] = managed.session
	}

	// Scan session directory for session files
	dirPath := sessionDir()
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		// Directory doesn't exist or can't be read, return memory sessions only
		sessions := make([]*Session, 0, len(sessionMap))
		for _, session := range sessionMap {
			sessions = append(sessions, session)
		}
		return sessions
	}

	// Load sessions from files (skip if already in memory)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		sessionID := entry.Name()[:len(entry.Name())-5] // Remove .json extension
		if _, exists := sessionMap[sessionID]; exists {
			continue // Already in memory
		}

		// Load session from file
		session, err := loadSession(sessionID)
		if err != nil {
			continue // Skip invalid sessions
		}
		sessionMap[sessionID] = session
	}

	// Convert map to slice
	sessions := make([]*Session, 0, len(sessionMap))
	for _, session := range sessionMap {
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

	for sessionID, managed := range sm.sessions {
		if now.Sub(managed.session.LastUsedAt) > timeout {
			if managed.worker != nil {
				managed.worker.Close()
			}

			if managed.session.Process != nil {
				if cmd, ok := managed.session.Process.(*exec.Cmd); ok {
					if cmd.Process != nil {
						_ = cmd.Process.Kill()
					}
				} else if proc, ok := managed.session.Process.(*os.Process); ok {
					_ = proc.Kill()
				}
			}

			if err := deleteSession(sessionID); err != nil {
				fmt.Printf("Warning: failed to delete session file: %v\n", err)
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

// SendCommand sends a command to a browser session via the session worker queue
func (sm *SessionManager) SendCommand(ctx context.Context, sessionID string, command map[string]interface{}) (*ipc.NodeResponse, error) {
	managed, err := sm.getOrCreateManagedSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	resp, err := managed.worker.Send(ctx, command)
	if err != nil {
		if errors.Is(err, errSessionClosed) {
			sm.mu.Lock()
			if managed.worker != nil && managed.worker.isClosed() {
				managed.worker = nil
			}
			sm.mu.Unlock()
		}
		return nil, err
	}

	sm.mu.Lock()
	managed.session.LastUsedAt = time.Now()
	sm.mu.Unlock()

	return resp, nil
}

func (sm *SessionManager) getOrCreateManagedSession(ctx context.Context, sessionID string) (*managedSession, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	sm.mu.RLock()
	managed, ok := sm.sessions[sessionID]
	sm.mu.RUnlock()

	if ok && managed.worker != nil && !managed.worker.isClosed() {
		return managed, nil
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	managed, ok = sm.sessions[sessionID]
	if !ok {
		session, err := loadSession(sessionID)
		if err != nil {
			return nil, err
		}
		managed = &managedSession{session: session}
		sm.sessions[sessionID] = managed
	}

	if managed.worker == nil || managed.worker.isClosed() {
		worker, err := newSessionWorker(ctx, managed.session)
		if err != nil {
			return nil, err
		}
		managed.worker = worker
	}

	return managed, nil
}
