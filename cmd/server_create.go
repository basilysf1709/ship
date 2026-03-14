package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	shipinternal "ship/internal"
)

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

			provider, err := shipinternal.New(providerName)
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

			sshClient, err := shipinternal.WaitForSSH(ctx, state.SSHUser, state.IP, 10*time.Second)
			if err != nil {
				return err
			}
			defer sshClient.Close()

			if err := shipinternal.RunCommands(ctx, sshClient, []string{
				"apt update",
				"apt install -y docker.io",
				"systemctl enable docker",
				"systemctl start docker",
			}); err != nil {
				return err
			}

			if err := shipinternal.SaveServerState(state); err != nil {
				return err
			}
			if err := shipinternal.AddServerInventoryRecord(state, projectPath); err != nil {
				return err
			}

			fmt.Printf("STATUS=SERVER_CREATED\nSERVER_ID=%s\nSERVER_IP=%s\n", state.ServerID, state.IP)
			return nil
		},
	}

	cmd.Flags().StringVar(&providerName, "provider", "digitalocean", "Infrastructure provider: digitalocean, hetzner, vultr")
	cmd.Flags().StringVar(&region, "region", "", "Server region/location")
	cmd.Flags().StringVar(&size, "size", "", "Server size/plan/type")
	cmd.Flags().StringVar(&image, "image", "", "Server image")
	return cmd
}
