//go:build linux

package main

import (
	"github.com/spf13/cobra"
)

// registerPlatformCommands registers Linux-specific commands
func registerPlatformCommands(rootCmd *cobra.Command) {
	// Example: Register a Linux-only command
	rootCmd.AddCommand(exampleLinuxCmd)
}

// exampleLinuxCmd is an example command only available on Linux
var exampleLinuxCmd = &cobra.Command{
	Use:   "linux-example",
	Short: "Example command available only on Linux",
	Long:  `This command demonstrates platform-specific functionality for Linux.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement Linux-specific logic (e.g., DBus, CLI tools)
		outputJSON(map[string]string{
			"platform": "linux",
			"message":  "This command is running on Linux",
		})
	},
}
