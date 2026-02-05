package main

import (
	"github.com/spf13/cobra"
)

// version is set via ldflags at build time.
var version = "dev"

var rootCmd = &cobra.Command{
	Use:          "csvpp",
	Short:        "CSV++ CLI tool for working with CSV++ files",
	Long:         `csvpp is a CLI tool for viewing, converting, and validating CSV++ files.`,
	Version:      version,
	SilenceUsage: true,
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
