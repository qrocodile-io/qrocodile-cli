// Package cmd holds the QRocodile CLI's command tree.
package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "qrocodile",
	Short: "Generate QR codes with QRocodile",
	// Errors are printed once, by main.go — not here too.
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       version,
}

// Execute runs the CLI, returning any error from the command that ran.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(versionCmd)
}
