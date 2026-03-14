package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var jsonOutput bool

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output structured JSON")
}

func writeCommandOutput(cmd *cobra.Command, text string, payload any) error {
	out := cmd.OutOrStdout()
	if jsonOutput {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(payload)
	}
	_, err := fmt.Fprint(out, text)
	return err
}
