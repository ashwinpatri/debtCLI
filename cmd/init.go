package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const defaultDebtTOML = `[tags]
TODO     = 2.0
FIXME    = 3.0
HACK     = 2.5
PERF     = 2.0
SECURITY = 4.0
NOTE     = 1.0

[ignore]
paths      = ["vendor/", "node_modules/", ".git/", "dist/", "build/"]
extensions = [".pb.go", ".gen.go", ".min.js", ".lock", ".sum"]
`

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Create a .debt.toml configuration file in the current repository",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(_ *cobra.Command, args []string) error {
	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	dest := filepath.Join(absRoot, ".debt.toml")
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf(".debt.toml already exists at %s", dest)
	}

	if err := os.WriteFile(dest, []byte(defaultDebtTOML), 0600); err != nil {
		return fmt.Errorf("write .debt.toml: %w", err)
	}

	fmt.Printf("Created %s\n", dest)
	return nil
}
