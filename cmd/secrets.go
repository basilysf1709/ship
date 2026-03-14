package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	shipinternal "ship/internal"
)

var loadSecrets = shipinternal.LoadSecrets
var saveSecrets = shipinternal.SaveSecrets
var deleteSecrets = shipinternal.DeleteSecrets

func newSecretsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secrets",
		Short: "Manage local and remote secrets",
	}
	cmd.AddCommand(newSecretsSetCommand())
	cmd.AddCommand(newSecretsListCommand())
	cmd.AddCommand(newSecretsRemoveCommand())
	cmd.AddCommand(newSecretsSyncCommand())
	return cmd
}

func newSecretsSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set KEY=VALUE [KEY=VALUE...]",
		Short: "Set secrets locally",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			secrets, err := loadSecrets()
			if err != nil {
				return err
			}
			for _, arg := range args {
				parts := strings.SplitN(arg, "=", 2)
				if len(parts) != 2 || parts[0] == "" {
					return fmt.Errorf("invalid secret %q; expected KEY=VALUE", arg)
				}
				secrets[parts[0]] = parts[1]
			}
			if err := saveSecrets(secrets); err != nil {
				return err
			}
			return writeCommandOutput(cmd, fmt.Sprintf("STATUS=SECRETS_UPDATED\nCOUNT=%d\n", len(secrets)), map[string]any{
				"status": "SECRETS_UPDATED",
				"count":  len(secrets),
			})
		},
	}
}

func newSecretsListCommand() *cobra.Command {
	var showValues bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List local secrets",
		RunE: func(cmd *cobra.Command, args []string) error {
			secrets, err := loadSecrets()
			if err != nil {
				return err
			}
			keys := make([]string, 0, len(secrets))
			for key := range secrets {
				keys = append(keys, key)
			}
			sort.Strings(keys)

			var builder strings.Builder
			if showValues {
				builder.WriteString("KEY\tVALUE\n")
				for _, key := range keys {
					builder.WriteString(fmt.Sprintf("%s\t%s\n", key, secrets[key]))
				}
			} else {
				builder.WriteString("KEY\n")
				for _, key := range keys {
					builder.WriteString(fmt.Sprintf("%s\n", key))
				}
			}
			builder.WriteString(fmt.Sprintf("TOTAL_SECRETS=%d\n", len(keys)))

			payload := map[string]any{
				"count":   len(keys),
				"keys":    keys,
				"secrets": secrets,
			}
			return writeCommandOutput(cmd, builder.String(), payload)
		},
	}
	cmd.Flags().BoolVar(&showValues, "show-values", false, "Show secret values")
	return cmd
}

func newSecretsRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove KEY [KEY...]",
		Short: "Remove local secrets",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			secrets, err := loadSecrets()
			if err != nil {
				return err
			}
			for _, key := range args {
				delete(secrets, key)
			}
			if len(secrets) == 0 {
				if err := deleteSecrets(); err != nil {
					return err
				}
			} else if err := saveSecrets(secrets); err != nil {
				return err
			}
			return writeCommandOutput(cmd, fmt.Sprintf("STATUS=SECRETS_REMOVED\nCOUNT=%d\n", len(secrets)), map[string]any{
				"status": "SECRETS_REMOVED",
				"count":  len(secrets),
			})
		},
	}
}

func newSecretsSyncCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Upload local secrets to the current server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
			defer cancel()

			state, client, err := currentServerClient(ctx, 30*time.Second)
			if err != nil {
				return err
			}
			defer client.Close()

			if err := syncSecretsToServer(ctx, client); err != nil {
				return err
			}
			return writeCommandOutput(cmd, fmt.Sprintf("STATUS=SECRETS_SYNCED\nSERVER_IP=%s\n", state.IP), map[string]any{
				"status":    "SECRETS_SYNCED",
				"server_ip": state.IP,
			})
		},
	}
}
