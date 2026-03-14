package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	shipinternal "ship/internal"
)

func newLogsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logs",
		Short: "Fetch container logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := shipinternal.LoadServerState()
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
			defer cancel()

			client, err := shipinternal.WaitForSSH(ctx, state.EffectiveSSHUser(), state.IP, 10*time.Second)
			if err != nil {
				return err
			}
			defer client.Close()

			output, err := shipinternal.RunCommand(ctx, client, "docker logs app --tail 100")
			if err != nil {
				return err
			}

			return writeCommandOutput(cmd, output, map[string]any{
				"server_ip": state.IP,
				"output":    output,
			})
		},
	}
}
