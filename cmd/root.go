package cmd

import (
	"github.com/spf13/cobra"
)

// WelcomeMessage to introduce ofc-bootstrap
const WelcomeMessage = "Welcome to faasd"

func init() {
	rootCommand.AddCommand(versionCmd)
	rootCommand.AddCommand(upCmd)
	rootCommand.AddCommand(installCmd)
}

func Execute() error {

	if err := rootCommand.Execute(); err != nil {
		return err
	}
	return nil
}

var rootCommand = &cobra.Command{
	Use:   "faasd",
	Short: "Start faasd",
	Long: `
faasd - serverless without Kubernetes
`,
	RunE:         runRootCommand,
	SilenceUsage: true,
}

func runRootCommand(cmd *cobra.Command, args []string) error {

	printLogo()
	cmd.Help()

	return nil
}
