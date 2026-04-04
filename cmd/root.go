package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var outputFormat string

var rootCmd = &cobra.Command{
	Use:   "debt",
	Short: "Track technical debt markers in a Git repository",
	Long: `debt scans source files for technical debt markers (TODO, FIXME, HACK, etc.),
enriches each finding with git blame metadata, scores them by severity and age,
and tracks repo health over time using a local SQLite history store.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&outputFormat, "format", "table", "output format: table or json")
}
