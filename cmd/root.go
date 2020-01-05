package cmd

import (
	"fmt"

	"github.com/morikuni/aec"
	"github.com/spf13/cobra"
)

// WelcomeMessage to introduce ofc-bootstrap
const WelcomeMessage = "Welcome to faasd"

func init() {
	rootCommand.AddCommand(versionCmd)
	rootCommand.AddCommand(upCmd)
	rootCommand.AddCommand(makeInstallCmd())
}

var (
	// GitCommit Git Commit SHA
	GitCommit string
	// Version version of the CLI
	Version string
)

// Execute faasd
func Execute(version, gitCommit string) error {

	// Get Version and GitCommit values from main.go.
	Version = version
	GitCommit = gitCommit

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

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information.",
	Run:   parseBaseCommand,
}

func parseBaseCommand(_ *cobra.Command, _ []string) {
	printLogo()

	fmt.Printf(
		`faasd
Commit: %s
Version: %s
`, GitCommit, GetVersion())
}

func printLogo() {
	logoText := aec.WhiteF.Apply(Logo)
	fmt.Println(logoText)
}

// GetVersion get latest version
func GetVersion() string {
	if len(Version) == 0 {
		return "dev"
	}
	return Version
}

// Logo for version and root command
const Logo = `  __                     _ 
 / _| __ _  __ _ ___  __| |
| |_ / _` + "`" + ` |/ _` + "`" + ` / __|/ _` + "`" + ` |
|  _| (_| | (_| \__ \ (_| |
|_|  \__,_|\__,_|___/\__,_|
`
