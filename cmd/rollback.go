package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"

	shipinternal "ship/internal"
)

func newRollbackCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rollback [release-id]",
		Short: "Rollback to a previous release",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var target *shipinternal.ReleaseRecord
			var err error
			if len(args) == 1 {
				target, err = findReleaseRecord(args[0])
			} else {
				target, err = defaultRollbackTargetRecord()
			}
			if err != nil {
				return err
			}
			if target == nil {
				return fmt.Errorf("no rollback target available")
			}
			if !target.RollbackEligible {
				reason := target.RollbackReason
				if reason == "" {
					reason = "release is not rollback eligible"
				}
				return fmt.Errorf("release %s cannot be rolled back automatically: %s", target.ID, reason)
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 20*time.Minute)
			defer cancel()
			state, client, err := currentServerClient(ctx, 30*time.Second)
			if err != nil {
				return err
			}
			defer client.Close()

			restoreCommands := make([]string, 0, len(target.Uploads))
			for _, upload := range target.Uploads {
				restoreCommands = append(restoreCommands,
					fmt.Sprintf("mkdir -p %s", shellQuote(path.Dir(upload.Destination))),
					fmt.Sprintf("cp %s %s", shellQuote(upload.BackupPath), shellQuote(upload.Destination)),
				)
			}
			if err := runRemoteCommands(ctx, client, restoreCommands); err != nil {
				return err
			}
			if err := syncSecretsToServer(ctx, client); err != nil {
				return err
			}
			if err := runRemoteCommands(ctx, client, target.RemoteCommands); err != nil {
				return err
			}

			record, err := shipinternal.NewReleaseRecord(&state, shipinternal.DeployConfig{
				RemoteCommands: append([]string(nil), target.RemoteCommands...),
			})
			if err != nil {
				return err
			}
			record.Status = "rollback"
			record.Stage = "complete"
			record.RollbackOf = target.ID
			record.DeployHash = target.DeployHash
			record.SecretsChecksum = target.SecretsChecksum
			record.Uploads = append([]shipinternal.ReleaseUpload(nil), target.Uploads...)
			record.ErrorMessage = ""
			record.UpdateRollbackEligibility()
			if err := saveReleaseRecord(record); err != nil {
				return err
			}

			return writeCommandOutput(cmd, fmt.Sprintf("STATUS=ROLLBACK_COMPLETE\nRELEASE_ID=%s\nSERVER_IP=%s\n", target.ID, state.IP), map[string]any{
				"status":     "ROLLBACK_COMPLETE",
				"release_id": target.ID,
				"server_ip":  state.IP,
			})
		},
	}
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
