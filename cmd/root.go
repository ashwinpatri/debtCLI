// Package cmd wires up the CLI commands using cobra.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// outputFormat is the global --format flag value shared across commands.
var outputFormat string

// rootCmd is the parent command. All subcommands are registered on it.
var rootCmd = &cobra.Command{
	Use:   "debt",
	Short: "Track technical debt markers in a Git repository",
	Long: `debt scans source files for technical debt markers (TODO, FIXME, HACK, etc.),
enriches each finding with git blame metadata, scores them by severity and age,
and tracks repo health over time using a local SQLite history store.`,
}

// Execute runs the CLI. Called by main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&outputFormat, "format", "table", "output format: table or json")
}
