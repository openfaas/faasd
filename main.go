package main

import (
	"os"

	"github.com/alexellis/faasd/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	return
}
