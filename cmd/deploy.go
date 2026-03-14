package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	shipinternal "ship/internal"
)

var loadDeployConfig = shipinternal.LoadDeployConfig
var loadServerState = shipinternal.LoadServerState
var runDeploy = shipinternal.Run

func newDeployCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			deployConfig, err := loadDeployConfig()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()

			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Minute)
			defer cancel()

			opts := shipinternal.Options{}
			serverIP := ""
			if deployConfig.RequiresServer() {
				state, err := loadServerState()
				if err != nil {
					return err
				}
				opts.ServerIP = state.IP
				opts.User = state.EffectiveSSHUser()
				serverIP = state.IP
			}

			if err := runDeploy(ctx, opts); err != nil {
				return err
			}

			fmt.Fprint(out, "STATUS=DEPLOY_COMPLETE\n")
			if serverIP != "" {
				fmt.Fprintf(out, "SERVER_IP=%s\n", serverIP)
			}
			return nil
		},
	}
}
