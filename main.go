package main

import (
	"os"

	"github.com/alexellis/faasd/cmd"
)

// These values will be injected into these variables at the build time.
var (
	// GitCommit Git Commit SHA
	GitCommit string
	// Version version of the CLI
	Version string
)

func main() {

	if err := cmd.Execute(Version, GitCommit); err != nil {
		os.Exit(1)
	}

	return
}
