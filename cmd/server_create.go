package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	shipinternal "ship/internal"
)

var createProvider = shipinternal.New
var waitForCreatedServerSSH = shipinternal.WaitForSSH
var runCreatedServerSetup = shipinternal.RunCommands
var saveCreatedServerState = shipinternal.SaveServerState
var addCreatedServerInventory = shipinternal.AddServerInventoryRecord

func newServerCommand() *cobra.Command {
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Manage server lifecycle",
	}

	serverCmd.AddCommand(newServerCreateCommand())
	serverCmd.AddCommand(newServerDestroyCommand())
	serverCmd.AddCommand(newServerListCommand())
	return serverCmd
}

func newServerCreateCommand() *cobra.Command {
	var providerName string
	var region string
	var size string
	var image string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 20*time.Minute)
			defer cancel()
			projectPath, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve current directory: %w", err)
			}

			provider, err := createProvider(providerName)
			if err != nil {
				return err
			}

			name := fmt.Sprintf("ship-%d", time.Now().Unix())
			state, err := provider.CreateServer(ctx, shipinternal.CreateRequest{
				Name:   name,
				Region: region,
				Size:   size,
				Image:  image,
			})
			if err != nil {
				return err
			}
			if state.SSHUser == "" {
				state.SSHUser = "root"
			}

			if err := saveCreatedServerState(state); err != nil {
				return err
			}
			if err := addCreatedServerInventory(state, projectPath); err != nil {
				return err
			}

			sshClient, err := waitForCreatedServerSSH(ctx, state.SSHUser, state.IP, 10*time.Second)
			if err != nil {
				return err
			}
			if sshClient != nil {
				defer sshClient.Close()
			}

			if err := runCreatedServerSetup(ctx, sshClient, []string{
				"apt update",
				"apt install -y docker.io",
				"systemctl enable docker",
				"systemctl start docker",
			}); err != nil {
				return err
			}

			projectConfig, err := loadProjectConfig()
			if err != nil {
				return err
			}
			if err := applyBootstrap(ctx, sshClient, projectConfig); err != nil {
				return err
			}

			return writeCommandOutput(cmd, fmt.Sprintf("STATUS=SERVER_CREATED\nSERVER_ID=%s\nSERVER_IP=%s\n", state.ServerID, state.IP), map[string]any{
				"status":    "SERVER_CREATED",
				"server_id": state.ServerID,
				"server_ip": state.IP,
			})
		},
	}

	cmd.Flags().StringVar(&providerName, "provider", "digitalocean", "Infrastructure provider: digitalocean, hetzner, vultr")
	cmd.Flags().StringVar(&region, "region", "", "Server region/location")
	cmd.Flags().StringVar(&size, "size", "", "Server size/plan/type")
	cmd.Flags().StringVar(&image, "image", "", "Server image")
	return cmd
}
