package cmd

import (
	"database/sql"
	"fmt"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/ashwinpatri/debtCLI/internal/config"
	"github.com/ashwinpatri/debtCLI/internal/db"
	"github.com/ashwinpatri/debtCLI/internal/git"
	"github.com/ashwinpatri/debtCLI/internal/models"
	"github.com/ashwinpatri/debtCLI/internal/output"
	"github.com/ashwinpatri/debtCLI/internal/pipeline"
	"github.com/ashwinpatri/debtCLI/internal/scanner"
	"github.com/ashwinpatri/debtCLI/internal/scorer"
	"github.com/ashwinpatri/debtCLI/internal/walker"
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan a repository for technical debt",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

func runScan(cmd *cobra.Command, args []string) error {
	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	cfg, err := config.Load(absRoot)
	if err != nil {
		return err
	}

	gc := git.NewClient()
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := gc.ValidateRepo(ctx, absRoot); err != nil {
		return err
	}

	dbPath := filepath.Join(absRoot, ".debt", "history.db")
	database, err := db.Open(dbPath)
	if err != nil {
		return err
	}
	defer database.Close()

	sc, err := scanner.New(cfg)
	if err != nil {
		return fmt.Errorf("build scanner: %w", err)
	}

	filePaths, err := walker.Walk(ctx, absRoot, cfg)
	if err != nil {
		return fmt.Errorf("walk: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Scanning %s...\n", absRoot)

	pipelineCfg := pipeline.Config{
		RepoPath: absRoot,
		Tags:     cfg.Tags,
	}
	result, err := pipeline.Run(ctx, filePaths, pipelineCfg, sc, gc)
	if err != nil {
		return fmt.Errorf("pipeline: %w", err)
	}

	health := scorer.RepoHealth(result.Items)

	prev, err := db.LoadLastSnapshot(database, absRoot)
	if err != nil {
		return fmt.Errorf("load previous snapshot: %w", err)
	}

	snap := &models.Snapshot{
		RepoPath:    absRoot,
		Timestamp:   time.Now().UTC(),
		HealthScore: health,
		ItemCount:   len(result.Items),
		Items:       result.Items,
	}

	delta := computeDelta(prev, snap)

	if err := db.WriteSnapshot(database, snap); err != nil {
		return fmt.Errorf("save snapshot: %w", err)
	}

	return renderScanResult(cmd, &models.ScanResult{Snapshot: snap, Delta: delta})
}

func computeDelta(prev, current *models.Snapshot) *models.Delta {
	if prev == nil {
		return nil
	}

	prevLines := make(map[string]map[int]bool)
	for _, item := range prev.Items {
		if prevLines[item.File] == nil {
			prevLines[item.File] = make(map[int]bool)
		}
		prevLines[item.File][item.Line] = true
	}

	currLines := make(map[string]map[int]bool)
	for _, item := range current.Items {
		if currLines[item.File] == nil {
			currLines[item.File] = make(map[int]bool)
		}
		currLines[item.File][item.Line] = true
	}

	newItems := 0
	for _, item := range current.Items {
		if prevLines[item.File] == nil || !prevLines[item.File][item.Line] {
			newItems++
		}
	}
	resolved := 0
	for _, item := range prev.Items {
		if currLines[item.File] == nil || !currLines[item.File][item.Line] {
			resolved++
		}
	}

	return &models.Delta{
		PreviousScore: prev.HealthScore,
		ScoreDiff:     math.Round(current.HealthScore - prev.HealthScore),
		NewItems:      newItems,
		ResolvedItems: resolved,
	}
}

func renderScanResult(cmd *cobra.Command, result *models.ScanResult) error {
	format, _ := cmd.Flags().GetString("format")
	if format == "" {
		format, _ = rootCmd.PersistentFlags().GetString("format")
	}

	var r output.Renderer
	switch format {
	case "json":
		r = &output.JSONRenderer{}
	default:
		r = &output.TableRenderer{}
	}
	return r.Render(os.Stdout, result)
}

var _ interface{ Close() error } = (*sql.DB)(nil)
