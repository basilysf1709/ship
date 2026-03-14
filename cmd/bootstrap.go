package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func newBootstrapCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "bootstrap",
		Short: "Apply bootstrap and proxy config to the current server",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectConfig, err := loadProjectConfig()
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 20*time.Minute)
			defer cancel()
			state, client, err := currentServerClient(ctx, 30*time.Second)
			if err != nil {
				return err
			}
			defer client.Close()

			if err := applyBootstrap(ctx, client, projectConfig); err != nil {
				return err
			}
			return writeCommandOutput(cmd, fmt.Sprintf("STATUS=BOOTSTRAP_COMPLETE\nSERVER_IP=%s\n", state.IP), map[string]any{
				"status":    "BOOTSTRAP_COMPLETE",
				"server_ip": state.IP,
			})
		},
	}
}
