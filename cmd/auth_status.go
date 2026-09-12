package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/qrocodile-io/qrocodile-cli/internal/apiclient"
	"github.com/qrocodile-io/qrocodile-cli/internal/config"
)

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show whether an API key is configured",
	Long: `Show whether an API key is configured.

Reports only that a key is stored, not whether it's still valid — the API has no
introspection endpoint to check that against without spending a real request.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if key := os.Getenv(apiclient.EnvAPIKey); key != "" {
			fmt.Fprintf(out, "Using an API key from %s (%s).\n", apiclient.EnvAPIKey, maskKey(key))
			return nil
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		key := cfg.APIKey()
		if key == "" {
			fmt.Fprintln(out, "Not logged in. Run `qrocodile auth login`.")
			return nil
		}
		fmt.Fprintf(out, "API key stored (%s).\n", maskKey(key))
		return nil
	},
}

func maskKey(key string) string {
	if len(key) <= 12 {
		return "…"
	}
	return key[:8] + "…" + key[len(key)-4:]
}
