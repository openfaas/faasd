package main

import (
	"os"

	"github.com/alexellis/faasd/cmd"
	"github.com/alexellis/faasd/pkg"
)

func main() {
	if err := cmd.Execute(pkg.Version, pkg.GitCommit); err != nil {
		os.Exit(1)
	}
	return

}
