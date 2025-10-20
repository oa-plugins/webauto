package playwright

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/oa-plugins/webauto/pkg/bootstrap"
)

//go:embed runner/session-server.js
var sessionRunnerSource string

var (
	sessionRunnerOnce sync.Once
	sessionRunnerPath string
	sessionRunnerErr  error
)

func ensureSessionRunnerScript() (string, error) {
	sessionRunnerOnce.Do(func() {
		cacheDir := bootstrap.GetCacheDir()
		runnerDir := filepath.Join(cacheDir, "runner")

		if err := os.MkdirAll(runnerDir, 0755); err != nil {
			sessionRunnerErr = fmt.Errorf("failed to create runner directory: %w", err)
			return
		}

		targetPath := filepath.Join(runnerDir, "session-server.js")
		if err := os.WriteFile(targetPath, []byte(sessionRunnerSource), 0644); err != nil {
			sessionRunnerErr = fmt.Errorf("failed to write runner script: %w", err)
			return
		}

		sessionRunnerPath = targetPath
	})

	return sessionRunnerPath, sessionRunnerErr
}
