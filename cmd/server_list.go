package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	shipinternal "ship/internal"
)

var listServerInventory = shipinternal.ListServerInventory
var currentWorkingDirectory = os.Getwd

func newServerListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List locally tracked servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			records, err := listServerInventory()
			if err != nil {
				return err
			}
			currentDir, _ := currentWorkingDirectory()
			if jsonOutput {
				payload := make([]map[string]any, 0, len(records))
				for _, record := range records {
					lastDeploy := ""
					if releases, err := listReleaseHistoryAt(record.ProjectPath); err == nil && len(releases) > 0 {
						lastDeploy = releases[0].CreatedAt
					}
					payload = append(payload, map[string]any{
						"provider":    record.Provider,
						"server_id":   record.ServerID,
						"ip":          record.IP,
						"link":        record.Link(),
						"project":     filepath.Base(record.ProjectPath),
						"current":     currentDir != "" && record.ProjectPath == currentDir,
						"last_deploy": lastDeploy,
						"created_at":  record.CreatedAt,
					})
				}
				return writeCommandOutput(cmd, "", map[string]any{
					"total_servers": len(records),
					"servers":       payload,
				})
			}

			var builder strings.Builder
			tw := tabwriter.NewWriter(&builder, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "PROVIDER\tSERVER_ID\tIP\tLINK\tPROJECT\tCURRENT\tLAST_DEPLOY\tCREATED_AT")
			for _, record := range records {
				lastDeploy := ""
				if releases, err := listReleaseHistoryAt(record.ProjectPath); err == nil && len(releases) > 0 {
					lastDeploy = releases[0].CreatedAt
				}
				fmt.Fprintf(
					tw,
					"%s\t%s\t%s\t%s\t%s\t%t\t%s\t%s\n",
					record.Provider,
					record.ServerID,
					record.IP,
					record.Link(),
					filepath.Base(record.ProjectPath),
					currentDir != "" && record.ProjectPath == currentDir,
					lastDeploy,
					record.CreatedAt,
				)
			}
			if err := tw.Flush(); err != nil {
				return fmt.Errorf("flush server list output: %w", err)
			}

			builder.WriteString(fmt.Sprintf("TOTAL_SERVERS=%d\n", len(records)))
			return writeCommandOutput(cmd, builder.String(), nil)
		},
	}
}
