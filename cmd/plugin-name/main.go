package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

const (
	pluginName    = "plugin-name" // TODO: Change this to your plugin name
	pluginVersion = "0.1.0"       // TODO: Update version
)

var (
	rootCmd = &cobra.Command{
		Use:     pluginName,
		Short:   "TODO: Short description of your plugin",
		Long:    `TODO: Long description of your plugin`,
		Version: pluginVersion,
	}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Register platform-specific commands
	// These are defined in commands_*.go files
	registerPlatformCommands(rootCmd)

	// Global flags
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose logging")
}

// platformNotSupported returns an error for commands not available on current platform
func platformNotSupported(commandName string) error {
	return fmt.Errorf("command '%s' is not available on %s/%s", commandName, runtime.GOOS, runtime.GOARCH)
}

// outputJSON outputs a JSON response
func outputJSON(data interface{}) {
	// TODO: Implement JSON output
	fmt.Printf("%+v\n", data)
}

// outputError outputs an error in JSON format
func outputError(code, message, details string) {
	// TODO: Implement error output
	fmt.Fprintf(os.Stderr, "Error [%s]: %s - %s\n", code, message, details)
}
