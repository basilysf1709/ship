package cmd

import (
	"fmt"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	shipinternal "ship/internal"
)

var listServerInventory = shipinternal.ListServerInventory

func newServerListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List locally tracked servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			records, err := listServerInventory()
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "PROVIDER\tSERVER_ID\tIP\tLINK\tPROJECT\tCREATED_AT")
			for _, record := range records {
				fmt.Fprintf(
					tw,
					"%s\t%s\t%s\t%s\t%s\t%s\n",
					record.Provider,
					record.ServerID,
					record.IP,
					record.Link(),
					filepath.Base(record.ProjectPath),
					record.CreatedAt,
				)
			}
			if err := tw.Flush(); err != nil {
				return fmt.Errorf("flush server list output: %w", err)
			}

			fmt.Fprintf(out, "TOTAL_SERVERS=%d\n", len(records))
			return nil
		},
	}
}
