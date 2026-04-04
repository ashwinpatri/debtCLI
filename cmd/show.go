package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ashwinpatri/debtCLI/internal/db"
	"github.com/ashwinpatri/debtCLI/internal/models"
	"github.com/ashwinpatri/debtCLI/internal/output"
)

var showCmd = &cobra.Command{
	Use:   "show <file>",
	Short: "Show debt items for a specific file from the last scan",
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

func init() {
	rootCmd.AddCommand(showCmd)
}

func runShow(cmd *cobra.Command, args []string) error {
	targetFile := args[0]

	absFile, err := filepath.Abs(targetFile)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	repoRoot, err := findRepoRoot(filepath.Dir(absFile))
	if err != nil {
		return fmt.Errorf("find repo root: %w", err)
	}

	dbPath := filepath.Join(repoRoot, ".debt", "history.db")
	database, err := db.Open(dbPath)
	if err != nil {
		return err
	}
	defer database.Close()

	snap, err := db.LoadLastSnapshot(database, repoRoot)
	if err != nil {
		return fmt.Errorf("load snapshot: %w", err)
	}
	if snap == nil {
		fmt.Fprintln(os.Stdout, "No scan history found. Run `debt scan` first.")
		return nil
	}

	var filtered []models.DebtItem
	for _, item := range snap.Items {
		if item.File == absFile {
			filtered = append(filtered, item)
		}
	}

	if len(filtered) == 0 {
		fmt.Fprintf(os.Stdout, "No debt items found for %s in the last scan.\n", targetFile)
		return nil
	}

	filteredSnap := &models.Snapshot{
		ID:          snap.ID,
		RepoPath:    snap.RepoPath,
		Timestamp:   snap.Timestamp,
		HealthScore: snap.HealthScore,
		ItemCount:   len(filtered),
		Items:       filtered,
	}

	r := output.Renderer(&output.TableRenderer{})
	format, _ := cmd.Flags().GetString("format")
	if format == "json" {
		r = &output.JSONRenderer{}
	}

	return r.Render(os.Stdout, &models.ScanResult{Snapshot: filteredSnap})
}

func findRepoRoot(dir string) (string, error) {
	for {
		if _, err := os.Stat(filepath.Join(dir, ".debt")); err == nil {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no git repo or .debt directory found")
		}
		dir = parent
	}
}
