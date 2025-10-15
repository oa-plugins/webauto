package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "webauto",
	Short:   "Playwright Agents-based intelligent browser automation",
	Long:    `webauto is a Playwright Agents-based intelligent browser automation plugin for OA CLI system.
It targets Korean tax/accounting services (Hometax, Wehago) with sophisticated UI automation and anti-bot capabilities.`,
	Version: "1.0.0",
}

func init() {
	// Register commands
	rootCmd.AddCommand(browserLaunchCmd)
	rootCmd.AddCommand(browserCloseCmd)
	rootCmd.AddCommand(pageNavigateCmd)
	rootCmd.AddCommand(elementClickCmd)
	rootCmd.AddCommand(elementTypeCmd)
	rootCmd.AddCommand(pageScreenshotCmd)
	rootCmd.AddCommand(sessionListCmd)
	rootCmd.AddCommand(sessionCloseCmd)

	// Global flags
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose logging")
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
