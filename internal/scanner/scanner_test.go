package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ashwinpatri/debtCLI/internal/models"
)

func defaultCfg() *models.Config {
	return &models.Config{
		Tags: map[string]float64{
			"TODO":  2.0,
			"FIXME": 3.0,
			"HACK":  2.5,
		},
	}
}

func TestNew_CompilesWithoutError(t *testing.T) {
	_, err := New(defaultCfg())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
}

func TestScan_FindsItems(t *testing.T) {
	content := `package main

// TODO: refactor this function
func bad() {
	// FIXME: off-by-one error here
	x := 1
	// unrelated comment
	_ = x
}
`
	path := writeTmp(t, content)
	sc, _ := New(defaultCfg())
	items, err := sc.Scan(path)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	if items[0].Tag != "TODO" || items[0].Line != 3 {
		t.Errorf("item[0]: got tag=%s line=%d", items[0].Tag, items[0].Line)
	}
	if items[1].Tag != "FIXME" || items[1].Line != 5 {
		t.Errorf("item[1]: got tag=%s line=%d", items[1].Tag, items[1].Line)
	}
}

func TestScan_OneMatchPerLine(t *testing.T) {
	// A line with both TODO and FIXME should produce exactly one item.
	content := "// TODO: also FIXME this mess\n"
	path := writeTmp(t, content)
	sc, _ := New(defaultCfg())
	items, err := sc.Scan(path)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item per line, got %d", len(items))
	}
}

func TestScan_CapturesComment(t *testing.T) {
	content := "// TODO: fix the auth middleware\n"
	path := writeTmp(t, content)
	sc, _ := New(defaultCfg())
	items, err := sc.Scan(path)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Comment != "fix the auth middleware" {
		t.Errorf("comment: got %q", items[0].Comment)
	}
}

func TestScan_CaseInsensitive(t *testing.T) {
	content := "// todo: lowercase tag\n// Todo: mixed case\n"
	path := writeTmp(t, content)
	sc, _ := New(defaultCfg())
	items, err := sc.Scan(path)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 case-insensitive matches, got %d", len(items))
	}
}

func TestScan_EmptyFile(t *testing.T) {
	path := writeTmp(t, "")
	sc, _ := New(defaultCfg())
	items, err := sc.Scan(path)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected no items for empty file, got %d", len(items))
	}
}

func TestScan_NoMatches(t *testing.T) {
	content := "package main\n\nfunc main() {}\n"
	path := writeTmp(t, content)
	sc, _ := New(defaultCfg())
	items, err := sc.Scan(path)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected no items, got %d", len(items))
	}
}

func writeTmp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.go")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(f.Name())
}
