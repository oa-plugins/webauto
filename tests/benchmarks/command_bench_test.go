package benchmarks

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/oa-plugins/webauto/pkg/config"
	"github.com/oa-plugins/webauto/pkg/playwright"
)

// BenchmarkBrowserLaunch measures browser launch performance
// Target: < 500ms
func BenchmarkBrowserLaunch(b *testing.B) {
	cfg := config.Load()
	sessionMgr := playwright.NewSessionManager(cfg)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session, err := sessionMgr.Create(ctx, "chromium", true)
		if err != nil {
			b.Fatalf("Failed to launch browser: %v", err)
		}

		// Clean up session
		sessionMgr.Close(session.ID)
	}
}

// BenchmarkBrowserClose measures browser close performance
// Target: < 500ms
func BenchmarkBrowserClose(b *testing.B) {
	cfg := config.Load()
	sessionMgr := playwright.NewSessionManager(cfg)
	ctx := context.Background()

	// Pre-create sessions
	sessions := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		session, err := sessionMgr.Create(ctx, "chromium", true)
		if err != nil {
			b.Fatalf("Failed to launch browser: %v", err)
		}
		sessions[i] = session.ID
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := sessionMgr.Close(sessions[i])
		if err != nil {
			b.Fatalf("Failed to close browser: %v", err)
		}
	}
}

// BenchmarkSessionList measures session list performance
// Target: < 100ms
func BenchmarkSessionList(b *testing.B) {
	cfg := config.Load()
	sessionMgr := playwright.NewSessionManager(cfg)
	ctx := context.Background()

	// Create 5 sessions
	for i := 0; i < 5; i++ {
		_, err := sessionMgr.Create(ctx, "chromium", true)
		if err != nil {
			b.Fatalf("Failed to create session: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sessions := sessionMgr.ListAll()
		if len(sessions) == 0 {
			b.Fatal("No sessions found")
		}
	}
}

// BenchmarkPageNavigate measures page navigation performance
// Target: < 1000ms (network-bound, may vary)
func BenchmarkPageNavigate(b *testing.B) {
	cfg := config.Load()
	sessionMgr := playwright.NewSessionManager(cfg)
	ctx := context.Background()

	// Create a single session for all navigations
	session, err := sessionMgr.Create(ctx, "chromium", true)
	if err != nil {
		b.Fatalf("Failed to launch browser: %v", err)
	}
	defer sessionMgr.Close(session.ID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := map[string]interface{}{
			"command":   "navigate",
			"url":       "data:text/html,<html><body><h1>Test</h1></body></html>", // Use data URL to avoid network
			"waitUntil": "load",
			"timeout":   30000,
		}

		_, err := sessionMgr.SendCommand(ctx, session.ID, cmd)
		if err != nil {
			b.Fatalf("Failed to navigate: %v", err)
		}
	}
}

// BenchmarkElementClick measures element click performance
// Target: < 300ms
func BenchmarkElementClick(b *testing.B) {
	cfg := config.Load()
	sessionMgr := playwright.NewSessionManager(cfg)
	ctx := context.Background()

	// Create session and navigate to test page
	session, err := sessionMgr.Create(ctx, "chromium", true)
	if err != nil {
		b.Fatalf("Failed to launch browser: %v", err)
	}
	defer sessionMgr.Close(session.ID)

	// Navigate to a simple test page with a button
	navCmd := map[string]interface{}{
		"command":   "navigate",
		"url":       "data:text/html,<html><body><button id=\"test-btn\">Click Me</button></body></html>",
		"waitUntil": "load",
		"timeout":   30000,
	}
	_, err = sessionMgr.SendCommand(ctx, session.ID, navCmd)
	if err != nil {
		b.Fatalf("Failed to navigate: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clickCmd := map[string]interface{}{
			"command":  "click",
			"selector": "#test-btn",
			"timeout":  30000,
		}

		_, err := sessionMgr.SendCommand(ctx, session.ID, clickCmd)
		if err != nil {
			b.Fatalf("Failed to click element: %v", err)
		}
	}
}

// BenchmarkElementType measures element typing performance
// Target: < 300ms
func BenchmarkElementType(b *testing.B) {
	cfg := config.Load()
	sessionMgr := playwright.NewSessionManager(cfg)
	ctx := context.Background()

	// Create session and navigate to test page
	session, err := sessionMgr.Create(ctx, "chromium", true)
	if err != nil {
		b.Fatalf("Failed to launch browser: %v", err)
	}
	defer sessionMgr.Close(session.ID)

	// Navigate to a simple test page with an input
	navCmd := map[string]interface{}{
		"command":   "navigate",
		"url":       "data:text/html,<html><body><input id=\"test-input\" /></body></html>",
		"waitUntil": "load",
		"timeout":   30000,
	}
	_, err = sessionMgr.SendCommand(ctx, session.ID, navCmd)
	if err != nil {
		b.Fatalf("Failed to navigate: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		typeCmd := map[string]interface{}{
			"command":  "type",
			"selector": "#test-input",
			"text":     "test text",
			"timeout":  30000,
		}

		_, err := sessionMgr.SendCommand(ctx, session.ID, typeCmd)
		if err != nil {
			b.Fatalf("Failed to type: %v", err)
		}
	}
}

// BenchmarkPageScreenshot measures screenshot capture performance
// Target: < 1000ms
func BenchmarkPageScreenshot(b *testing.B) {
	cfg := config.Load()
	sessionMgr := playwright.NewSessionManager(cfg)
	ctx := context.Background()

	// Create session and navigate to test page
	session, err := sessionMgr.Create(ctx, "chromium", true)
	if err != nil {
		b.Fatalf("Failed to launch browser: %v", err)
	}
	defer sessionMgr.Close(session.ID)

	// Navigate to a simple test page
	navCmd := map[string]interface{}{
		"command":   "navigate",
		"url":       "data:text/html,<html><body><h1>Screenshot Test</h1></body></html>",
		"waitUntil": "load",
		"timeout":   30000,
	}
	_, err = sessionMgr.SendCommand(ctx, session.ID, navCmd)
	if err != nil {
		b.Fatalf("Failed to navigate: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		screenshotCmd := map[string]interface{}{
			"command":  "screenshot",
			"type":     "png",
			"fullPage": false,
			"timeout":  30000,
		}

		_, err := sessionMgr.SendCommand(ctx, session.ID, screenshotCmd)
		if err != nil {
			b.Fatalf("Failed to take screenshot: %v", err)
		}
	}
}

// BenchmarkSendCommand measures raw IPC overhead
func BenchmarkSendCommand(b *testing.B) {
	cfg := config.Load()
	sessionMgr := playwright.NewSessionManager(cfg)
	ctx := context.Background()

	// Create session
	session, err := sessionMgr.Create(ctx, "chromium", true)
	if err != nil {
		b.Fatalf("Failed to launch browser: %v", err)
	}
	defer sessionMgr.Close(session.ID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pingCmd := map[string]interface{}{
			"command": "ping",
		}

		_, err := sessionMgr.SendCommand(ctx, session.ID, pingCmd)
		if err != nil {
			b.Fatalf("Failed to ping: %v", err)
		}
	}
}

// TestMain cleans up any leftover sessions before/after tests
func TestMain(m *testing.M) {
	// Clean up any existing sessions
	cfg := config.Load()
	sessionMgr := playwright.NewSessionManager(cfg)

	// Close all sessions
	sessions := sessionMgr.ListAll()
	for _, session := range sessions {
		sessionMgr.Close(session.ID)
	}

	// Run tests
	code := m.Run()

	// Clean up after tests
	sessions = sessionMgr.ListAll()
	for _, session := range sessions {
		sessionMgr.Close(session.ID)
	}

	os.Exit(code)
}

// Helper function to measure single operation
func MeasureSingleOp(name string, op func() error) {
	start := time.Now()
	err := op()
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ %s: FAILED - %v (took %v)\n", name, err, duration)
	} else {
		fmt.Printf("✅ %s: %v\n", name, duration)
	}
}
