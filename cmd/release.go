package cmd

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newReleaseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Inspect tracked deploy releases",
	}
	cmd.AddCommand(newReleaseListCommand())
	return cmd
}

func newReleaseListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List tracked releases",
		RunE: func(cmd *cobra.Command, args []string) error {
			records, err := listReleaseHistory()
			if err != nil {
				return err
			}
			if jsonOutput {
				return writeCommandOutput(cmd, "", records)
			}

			var builder strings.Builder
			tw := tabwriter.NewWriter(&builder, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "ID\tCREATED_AT\tGIT_SHA\tSERVER_IP\tSTATUS\tROLLBACK_OF")
			for _, record := range records {
				gitSHA := record.GitSHA
				if len(gitSHA) > 8 {
					gitSHA = gitSHA[:8]
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n", record.ID, record.CreatedAt, gitSHA, record.ServerIP, record.Status, record.RollbackOf)
			}
			_ = tw.Flush()
			builder.WriteString(fmt.Sprintf("TOTAL_RELEASES=%d\n", len(records)))
			return writeCommandOutput(cmd, builder.String(), records)
		},
	}
}
