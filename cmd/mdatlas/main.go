package main

import (
	"os"

	"github.com/mosaan/mdatlas/internal/cli"
)

var (
	version   string = "dev"
	buildDate string = "unknown"
)

func main() {
	// Set version information for CLI
	// This will be set during build time via ldflags
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}