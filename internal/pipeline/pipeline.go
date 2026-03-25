// Package pipeline orchestrates the fan-out scan: for each file path received
// from the walker, it runs the scanner, enriches items with git blame and churn
// data, scores them, and collects the results.
package pipeline

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/ashwinpatri/debtCLI/internal/git"
	"github.com/ashwinpatri/debtCLI/internal/models"
	"github.com/ashwinpatri/debtCLI/internal/scorer"
	"github.com/ashwinpatri/debtCLI/internal/scanner"
)

// Config parameterises a pipeline run.
type Config struct {
	// RepoPath is the root of the git repository, used for blame and churn calls.
	RepoPath string
	// Tags maps tag names to their base severity, forwarded to the scorer.
	Tags map[string]float64
	// Workers is the number of concurrent file-processing goroutines.
	// Zero means runtime.NumCPU()*2.
	Workers int
}

// Result holds all debt items found across all files in a single run.
type Result struct {
	Items []models.DebtItem
}

// Run fans out file paths from filePaths across Workers goroutines, processes
// each file through the scanner and git enrichment, and returns the aggregated
// result. A non-nil error is returned only when the context is cancelled or a
// bug causes a worker to return an unexpected error — individual file failures
// are logged and skipped.
func Run(ctx context.Context, filePaths <-chan string, cfg Config, sc *scanner.Scanner, gc git.Client) (*Result, error) {
	workers := cfg.Workers
	if workers <= 0 {
		workers = runtime.NumCPU() * 2
	}

	results := make(chan []models.DebtItem, workers)

	g, ctx := errgroup.WithContext(ctx)

	for i := 0; i < workers; i++ {
		g.Go(func() error {
			return processFiles(ctx, filePaths, cfg, sc, gc, results)
		})
	}

	// Collect runs concurrently with the workers; it must finish draining
	// before g.Wait() returns, otherwise workers block on a full results channel.
	var all []models.DebtItem
	collectDone := make(chan struct{})
	go func() {
		defer close(collectDone)
		for items := range results {
			all = append(all, items...)
		}
	}()

	// Wait for all workers to finish, then close results so the collector exits.
	err := g.Wait()
	close(results)
	<-collectDone

	if err != nil {
		return nil, err
	}
	return &Result{Items: all}, nil
}

// processFiles is the worker body: it reads file paths from the channel,
// processes each one, and sends results downstream.
func processFiles(ctx context.Context, filePaths <-chan string, cfg Config, sc *scanner.Scanner, gc git.Client, results chan<- []models.DebtItem) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case path, ok := <-filePaths:
			if !ok {
				return nil
			}
			items, err := processFile(ctx, path, cfg, sc, gc)
			if err != nil {
				log.Printf("pipeline: skipping %s: %v", path, err)
				continue
			}
			if len(items) > 0 {
				select {
				case results <- items:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}
}

// processFile scans a single file, enriches each item with blame and churn
// data, and scores it. Blame and churn errors are non-fatal: items are kept
// with zero/unknown values rather than discarded.
func processFile(ctx context.Context, path string, cfg Config, sc *scanner.Scanner, gc git.Client) ([]models.DebtItem, error) {
	items, err := sc.Scan(path)
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}
	if len(items) == 0 {
		return nil, nil
	}

	blameMap, blameErr := gc.Blame(ctx, cfg.RepoPath, path)
	if blameErr != nil {
		log.Printf("pipeline: blame failed for %s: %v", path, blameErr)
	}

	churn, churnErr := gc.Churn(ctx, cfg.RepoPath, path)
	if churnErr != nil {
		log.Printf("pipeline: churn failed for %s: %v", path, churnErr)
	}

	for i := range items {
		if blameMap != nil {
			if info, ok := blameMap[items[i].Line]; ok {
				items[i].Author = info.Author
				items[i].AuthorEmail = info.AuthorEmail
				items[i].Date = time.Unix(info.Timestamp, 0)
			}
		}
		if items[i].Author == "" {
			items[i].Author = "unknown"
		}

		items[i].Churn = churn

		severity := cfg.Tags[items[i].Tag]
		items[i].Score = scorer.ScoreItem(items[i], severity)
	}

	return items, nil
}
