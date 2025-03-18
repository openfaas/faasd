package cmd

import "github.com/spf13/cobra"

func makeServiceCmd() *cobra.Command {
	var command = &cobra.Command{
		Use:   "service",
		Short: "Manage services",
		Long:  `Manage services created by faasd from the docker-compose.yml file`,
	}

	command.RunE = runServiceE

	command.AddCommand(makeServiceLogsCmd())
	return command
}

func runServiceE(cmd *cobra.Command, args []string) error {

	return cmd.Help()

}
