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
`, pkg.GitCommit, pkg.GetVersion())
}

func printLogo() {
	logoText := aec.WhiteF.Apply(pkg.Logo)
	fmt.Println(logoText)
}
