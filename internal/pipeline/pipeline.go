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

type Config struct {
	RepoPath string
	Tags     map[string]float64
	Workers  int
}

type Result struct {
	Items []models.DebtItem
}

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

	var all []models.DebtItem
	collectDone := make(chan struct{})
	go func() {
		defer close(collectDone)
		for items := range results {
			all = append(all, items...)
		}
	}()

	// Drain results before calling g.Wait() — workers block if the channel fills.
	err := g.Wait()
	close(results)
	<-collectDone

	if err != nil {
		return nil, err
	}
	return &Result{Items: all}, nil
}

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
		items[i].Score = scorer.ScoreItem(items[i], cfg.Tags[items[i].Tag])
	}

	return items, nil
}
