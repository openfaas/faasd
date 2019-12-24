package cmd

import (
	"fmt"

	"github.com/alexellis/faasd/pkg"
	"github.com/morikuni/aec"
	"github.com/spf13/cobra"
)

var (
	// Version as per git repo
	Version string

	// GitCommit as per git repo
	GitCommit string
)

// WelcomeMessage to introduce ofc-bootstrap
const WelcomeMessage = "Welcome to faasd"

func init() {
	rootCommand.AddCommand(versionCmd)
	rootCommand.AddCommand(upCmd)
	rootCommand.AddCommand(installCmd)
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

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information.",
	Run:   parseBaseCommand,
}

func getVersion() string {
	if len(Version) != 0 {
		return Version
	}
	return "dev"
}

func parseBaseCommand(_ *cobra.Command, _ []string) {
	printLogo()

	fmt.Printf(
		`faasd
Commit: %s
Version: %s
`, pkg.GitCommit, pkg.GetVersion())
}

func Execute(version, gitCommit string) error {

	// Get Version and GitCommit values from main.go.
	Version = version
	GitCommit = gitCommit

	if err := rootCommand.Execute(); err != nil {
		return err
	}
	return nil
}

func runRootCommand(cmd *cobra.Command, args []string) error {

	printLogo()
	cmd.Help()

	return nil
}

func printLogo() {
	logoText := aec.WhiteF.Apply(pkg.Logo)
	fmt.Println(logoText)
}
