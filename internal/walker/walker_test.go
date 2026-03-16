package walker

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/ashwinpatri/debtCLI/internal/models"
)

func collect(t *testing.T, root string, cfg *models.Config) []string {
	t.Helper()
	ch, err := Walk(context.Background(), root, cfg)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	var paths []string
	for p := range ch {
		rel, _ := filepath.Rel(root, p)
		paths = append(paths, rel)
	}
	sort.Strings(paths)
	return paths
}

func defaultCfg() *models.Config {
	return &models.Config{
		Tags: map[string]float64{"TODO": 2.0},
		Ignore: models.IgnoreConfig{
			Paths:      []string{"vendor/", ".git/"},
			Extensions: []string{".pb.go"},
		},
	}
}

func TestWalk_BasicFiles(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "main.go"), "package main\n")
	write(t, filepath.Join(dir, "README.md"), "# hello\n")

	paths := collect(t, dir, defaultCfg())
	if len(paths) != 2 {
		t.Errorf("expected 2 files, got %d: %v", len(paths), paths)
	}
}

func TestWalk_SkipsIgnoredDir(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "main.go"), "package main\n")
	if err := os.MkdirAll(filepath.Join(dir, "vendor", "lib"), 0700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, "vendor", "lib", "lib.go"), "package lib\n")

	paths := collect(t, dir, defaultCfg())
	for _, p := range paths {
		if filepath.HasPrefix(p, "vendor") {
			t.Errorf("expected vendor/ to be skipped, got %s", p)
		}
	}
	if len(paths) != 1 {
		t.Errorf("expected 1 file, got %d: %v", len(paths), paths)
	}
}

func TestWalk_SkipsIgnoredExtension(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "gen.pb.go"), "package foo\n")
	write(t, filepath.Join(dir, "real.go"), "package foo\n")

	paths := collect(t, dir, defaultCfg())
	if len(paths) != 1 || paths[0] != "real.go" {
		t.Errorf("expected only real.go, got %v", paths)
	}
}

func TestWalk_SkipsBinaryFile(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "text.go"), "package main\n")
	// Write a file with a null byte to trigger binary detection.
	if err := os.WriteFile(filepath.Join(dir, "binary.bin"), []byte("ELF\x00data"), 0600); err != nil {
		t.Fatal(err)
	}

	paths := collect(t, dir, defaultCfg())
	if len(paths) != 1 || paths[0] != "text.go" {
		t.Errorf("expected only text.go, got %v", paths)
	}
}

func TestWalk_SkipsEmptyFile(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "real.go"), "package main\n")
	if err := os.WriteFile(filepath.Join(dir, "empty.go"), []byte{}, 0600); err != nil {
		t.Fatal(err)
	}

	paths := collect(t, dir, defaultCfg())
	if len(paths) != 1 || paths[0] != "real.go" {
		t.Errorf("expected only real.go, got %v", paths)
	}
}

func TestWalk_ContextCancel(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 20; i++ {
		write(t, filepath.Join(dir, filepath.Join("file"+string(rune('a'+i))+".go")), "package main\n")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ch, err := Walk(ctx, dir, defaultCfg())
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	// Drain whatever arrives before the goroutine notices cancellation.
	for range ch {
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}
