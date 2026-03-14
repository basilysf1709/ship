package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:           "ship",
	Short:         "Minimal infrastructure CLI for DigitalOcean, Hetzner, and Vultr",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(newBootstrapCommand())
	rootCmd.AddCommand(newDeployCommand())
	rootCmd.AddCommand(newDomainCommand())
	rootCmd.AddCommand(newExecCommand())
	rootCmd.AddCommand(newInitCommand())
	rootCmd.AddCommand(newLogsCommand())
	rootCmd.AddCommand(newReleaseCommand())
	rootCmd.AddCommand(newRollbackCommand())
	rootCmd.AddCommand(newSecretsCommand())
	rootCmd.AddCommand(newServerCommand())
	rootCmd.AddCommand(newStatusCommand())
}
