package main

import (
	"fmt"
	"os"

	"github.com/openfaas/faasd/cmd"
)

// These values will be injected into these variables at the build time.
var (
	// GitCommit Git Commit SHA
	GitCommit string
	// Version version of the CLI
	Version string
)

func main() {

	if _, ok := os.LookupEnv("CONTAINER_ID"); ok {
		collect := cmd.RootCommand()
		collect.SetArgs([]string{"collect"})
		collect.SilenceUsage = true
		collect.SilenceErrors = true

		err := collect.Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	if err := cmd.Execute(Version, GitCommit); err != nil {
		os.Exit(1)
	}
	return
}
