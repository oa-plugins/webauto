package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

// Config holds all configuration for the webauto plugin
type Config struct {
	// Playwright
	PlaywrightNodePath   string
	PlaywrightAgentsPath string
	PlaywrightCachePath  string

	// Browser
	DefaultBrowserType    string
	DefaultHeadless       bool
	DefaultViewportWidth  int
	DefaultViewportHeight int

	// Session
	SessionMaxCount       int
	SessionTimeoutSeconds int

	// Anti-Bot
	EnableStealth        bool
	EnableFingerprint    bool
	EnableBehaviorRandom bool
	TypingDelayMs        int
	MouseMoveJitterPx    int
}

// Load loads configuration from environment variables with sensible defaults
func Load() *Config {
	return &Config{
		PlaywrightNodePath:   getEnvOrDefault("PLAYWRIGHT_NODE_PATH", getDefaultNodePath()),
		PlaywrightAgentsPath: getEnvOrDefault("PLAYWRIGHT_AGENTS_PATH", "@playwright/agents"),
		PlaywrightCachePath:  getEnvOrDefault("PLAYWRIGHT_CACHE_PATH", getDefaultCachePath()),

		DefaultBrowserType:    getEnvOrDefault("DEFAULT_BROWSER_TYPE", "chromium"),
		DefaultHeadless:       getEnvBoolOrDefault("DEFAULT_HEADLESS", true),
		DefaultViewportWidth:  getEnvIntOrDefault("DEFAULT_VIEWPORT_WIDTH", 1920),
		DefaultViewportHeight: getEnvIntOrDefault("DEFAULT_VIEWPORT_HEIGHT", 1080),

		SessionMaxCount:       getEnvIntOrDefault("SESSION_MAX_COUNT", 10),
		SessionTimeoutSeconds: getEnvIntOrDefault("SESSION_TIMEOUT_SECONDS", 3600),

		EnableStealth:        getEnvBoolOrDefault("ENABLE_STEALTH", true),
		EnableFingerprint:    getEnvBoolOrDefault("ENABLE_FINGERPRINT", true),
		EnableBehaviorRandom: getEnvBoolOrDefault("ENABLE_BEHAVIOR_RANDOM", true),
		TypingDelayMs:        getEnvIntOrDefault("TYPING_DELAY_MS", 30),
		MouseMoveJitterPx:    getEnvIntOrDefault("MOUSE_MOVE_JITTER_PX", 10),
	}
}

func getDefaultNodePath() string {
	// Try to use node from PATH
	return "node"
}

func getDefaultCachePath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "oa", "webauto", "cache")
	default:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".cache", "oa", "webauto")
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "true" || value == "1" {
		return true
	}
	if value == "false" || value == "0" {
		return false
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	return defaultValue
}
