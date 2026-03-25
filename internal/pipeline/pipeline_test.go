package pipeline

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ashwinpatri/debtCLI/internal/git"
	"github.com/ashwinpatri/debtCLI/internal/models"
	"github.com/ashwinpatri/debtCLI/internal/scanner"
)

func defaultCfg(repoPath string) Config {
	return Config{
		RepoPath: repoPath,
		Tags: map[string]float64{
			"TODO":  2.0,
			"FIXME": 3.0,
		},
		Workers: 2,
	}
}

func fileChannel(paths ...string) <-chan string {
	ch := make(chan string, len(paths))
	for _, p := range paths {
		ch <- p
	}
	close(ch)
	return ch
}

func TestRun_FindsItems(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	if err := os.WriteFile(f, []byte("// TODO: fix this\n// FIXME: and this\n"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := defaultCfg(dir)
	sc, err := scanner.New(&models.Config{Tags: cfg.Tags})
	if err != nil {
		t.Fatal(err)
	}
	mc := git.NewMockClient()

	result, err := Run(context.Background(), fileChannel(f), cfg, sc, mc)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(result.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(result.Items))
	}
}

func TestRun_BlameEnrichment(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "auth.go")
	if err := os.WriteFile(f, []byte("// TODO: refactor auth\n"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := defaultCfg(dir)
	sc, _ := scanner.New(&models.Config{Tags: cfg.Tags})
	mc := git.NewMockClient()
	mc.BlameData[f] = map[int]git.BlameInfo{
		1: {Author: "ashwin", AuthorEmail: "a@example.com", Timestamp: 1700000000},
	}
	mc.ChurnData[f] = 5

	result, _ := Run(context.Background(), fileChannel(f), cfg, sc, mc)
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
	item := result.Items[0]
	if item.Author != "ashwin" {
		t.Errorf("author: got %q", item.Author)
	}
	if item.Churn != 5 {
		t.Errorf("churn: got %d", item.Churn)
	}
	if item.Score == 0 {
		t.Error("expected non-zero score")
	}
}

func TestRun_BlameErrorKeepsItem(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "broken.go")
	if err := os.WriteFile(f, []byte("// FIXME: broken blame\n"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := defaultCfg(dir)
	sc, _ := scanner.New(&models.Config{Tags: cfg.Tags})
	mc := git.NewMockClient()
	mc.BlameErr[f] = errors.New("git not available")

	result, err := Run(context.Background(), fileChannel(f), cfg, sc, mc)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected item preserved on blame error, got %d", len(result.Items))
	}
	if result.Items[0].Author != "unknown" {
		t.Errorf("expected author=unknown, got %q", result.Items[0].Author)
	}
}

func TestRun_EmptyChannel(t *testing.T) {
	cfg := defaultCfg(t.TempDir())
	sc, _ := scanner.New(&models.Config{Tags: cfg.Tags})
	mc := git.NewMockClient()

	result, err := Run(context.Background(), fileChannel(), cfg, sc, mc)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(result.Items) != 0 {
		t.Errorf("expected no items, got %d", len(result.Items))
	}
}

func TestRun_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := defaultCfg(t.TempDir())
	sc, _ := scanner.New(&models.Config{Tags: cfg.Tags})
	mc := git.NewMockClient()

	_, err := Run(ctx, fileChannel(), cfg, sc, mc)
	// Either nil (empty channel drained before cancel observed) or context error.
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Errorf("unexpected error: %v", err)
	}
}
