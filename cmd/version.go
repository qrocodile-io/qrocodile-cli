package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version is a plain build-time variable rather than debug.ReadBuildInfo() module version,
// since that reports "(devel)" for a locally-built binary and only resolves to something
// useful once this is tagged and built via `go install pkg@version` or a release pipeline.
// A release build injects the real value with -ldflags; "dev" is honest until then.
var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the QRocodile CLI version",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), version)
		return nil
	},
}
