package main

import (
	"fmt"
	"os"

	"github.com/oa-plugins/webauto/pkg/bootstrap"
	"github.com/oa-plugins/webauto/pkg/cli"
)

func main() {
	// Bootstrap Node.js runtime on first run
	nodePath, err := bootstrap.EnsureRuntime()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to setup runtime: %v\n", err)
		os.Exit(1)
	}

	// Override config with bootstrapped Node.js path
	os.Setenv("PLAYWRIGHT_NODE_PATH", nodePath)

	// Execute CLI commands
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
