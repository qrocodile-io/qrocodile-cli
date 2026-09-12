package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/qrocodile-io/qrocodile-cli/internal/config"
)

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove the stored API key",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if cfg.APIKey() == "" {
			fmt.Fprintln(cmd.OutOrStdout(), "No API key stored.")
			return nil
		}
		cfg.Credential = nil
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "API key removed.")
		return nil
	},
}
