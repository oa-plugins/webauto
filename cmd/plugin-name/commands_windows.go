//go:build windows

package main

import (
	"github.com/spf13/cobra"
)

// registerPlatformCommands registers Windows-specific commands
func registerPlatformCommands(rootCmd *cobra.Command) {
	// Example: Register a Windows-only command
	rootCmd.AddCommand(exampleWindowsCmd)
}

// exampleWindowsCmd is an example command only available on Windows
var exampleWindowsCmd = &cobra.Command{
	Use:   "windows-example",
	Short: "Example command available only on Windows",
	Long:  `This command demonstrates platform-specific functionality for Windows.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement Windows-specific logic
		outputJSON(map[string]string{
			"platform": "windows",
			"message":  "This command is running on Windows",
		})
	},
}
