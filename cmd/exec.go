package cmd

import (
	"context"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newExecCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "exec <command>",
		Short: "Run a command on the current server",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			command := strings.Join(args, " ")
			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Minute)
			defer cancel()

			state, client, err := currentServerClient(ctx, 30*time.Second)
			if err != nil {
				return err
			}
			defer client.Close()

			output, err := runRemoteCommand(ctx, client, command)
			if err != nil {
				return err
			}

			return writeCommandOutput(cmd, output, map[string]any{
				"server_ip": state.IP,
				"command":   command,
				"output":    output,
			})
		},
	}
}
