package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	shipinternal "ship/internal"
)

func newServerDestroyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "destroy",
		Short: "Destroy the current server",
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := shipinternal.LoadServerState()
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
			defer cancel()

			provider, err := shipinternal.New(state.Provider)
			if err != nil {
				return err
			}

			if err := provider.DestroyServer(ctx, state); err != nil {
				return err
			}
			if err := shipinternal.RemoveServerInventoryRecord(state); err != nil {
				return err
			}
			if err := shipinternal.DeleteServerState(); err != nil {
				return err
			}

			return writeCommandOutput(cmd, "STATUS=SERVER_DESTROYED\n", map[string]any{
				"status": "SERVER_DESTROYED",
			})
		},
	}
}
