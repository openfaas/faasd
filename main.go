package main

import (
	"os"

	"github.com/alexellis/faasd/cmd"
)

var (
	// GitCommit Git Commit SHA
	GitCommit string
	// Version version of the CLI
	Version string
)

func main() {
	if err := cmd.Execute(GitCommit, Version); err != nil {
		os.Exit(1)
	}
	return
}
