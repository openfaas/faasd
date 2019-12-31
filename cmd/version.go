package cmd

import (
	"fmt"

	"github.com/alexellis/faasd/pkg"
	"github.com/morikuni/aec"
	"github.com/spf13/cobra"
)

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
	logoText := aec.WhiteF.Apply(pkg.Logo)
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
