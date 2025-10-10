//go:build darwin

package main

import (
	"github.com/spf13/cobra"
)

// registerPlatformCommands registers macOS-specific commands
func registerPlatformCommands(rootCmd *cobra.Command) {
	// Example: Register a macOS-only command
	rootCmd.AddCommand(exampleDarwinCmd)
}

// exampleDarwinCmd is an example command only available on macOS
var exampleDarwinCmd = &cobra.Command{
	Use:   "macos-example",
	Short: "Example command available only on macOS",
	Long:  `This command demonstrates platform-specific functionality for macOS.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement macOS-specific logic (e.g., AppleScript)
		outputJSON(map[string]string{
			"platform": "darwin",
			"message":  "This command is running on macOS",
		})
	},
}
