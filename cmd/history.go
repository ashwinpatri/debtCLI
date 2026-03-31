package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ashwinpatri/debtCLI/internal/db"
	"github.com/ashwinpatri/debtCLI/internal/output"
)

var historyCmd = &cobra.Command{
	Use:   "history [path]",
	Short: "Show health score history for a repository",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runHistory,
}

func init() {
	rootCmd.AddCommand(historyCmd)
}

func runHistory(_ *cobra.Command, args []string) error {
	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	dbPath := filepath.Join(absRoot, ".debt", "history.db")
	database, err := db.Open(dbPath)
	if err != nil {
		return err
	}
	defer database.Close()

	snapshots, err := db.LoadHistory(database, absRoot)
	if err != nil {
		return fmt.Errorf("load history: %w", err)
	}

	r := &output.HistoryTableRenderer{}
	return r.RenderHistory(os.Stdout, snapshots)
}
