package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	if len(cfg.Tags) == 0 {
		t.Fatal("expected default tags to be non-empty")
	}
	if _, ok := cfg.Tags["FIXME"]; !ok {
		t.Error("expected FIXME in default tags")
	}
	if len(cfg.Ignore.Paths) == 0 {
		t.Error("expected default ignore paths to be non-empty")
	}
	if len(cfg.Ignore.Extensions) == 0 {
		t.Error("expected default ignore extensions to be non-empty")
	}

	// Mutating one copy must not affect another.
	cfg2 := defaultConfig()
	cfg.Tags["INJECTED"] = 9.9
	if _, ok := cfg2.Tags["INJECTED"]; ok {
		t.Error("defaultConfig returned shared state")
	}
}

func TestLoad_NoFile(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load with no config file: %v", err)
	}
	if len(cfg.Tags) == 0 {
		t.Error("expected defaults when no config file present")
	}
}

func TestLoad_ValidFile(t *testing.T) {
	dir := t.TempDir()
	content := `
[tags]
TODO = 1.5
FIXME = 5.0

[ignore]
paths = ["vendor/"]
extensions = [".gen.go"]
`
	if err := os.WriteFile(filepath.Join(dir, ".debt.toml"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Tags["TODO"] != 1.5 {
		t.Errorf("TODO severity: got %.1f, want 1.5", cfg.Tags["TODO"])
	}
	if cfg.Tags["FIXME"] != 5.0 {
		t.Errorf("FIXME severity: got %.1f, want 5.0", cfg.Tags["FIXME"])
	}
	if len(cfg.Ignore.Paths) != 1 || cfg.Ignore.Paths[0] != "vendor/" {
		t.Errorf("ignore paths: got %v", cfg.Ignore.Paths)
	}
}

func TestLoad_WalksUp(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(sub, 0700); err != nil {
		t.Fatal(err)
	}

	content := "[tags]\nTODO = 2.0\n"
	if err := os.WriteFile(filepath.Join(root, ".debt.toml"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(sub)
	if err != nil {
		t.Fatalf("Load from subdirectory: %v", err)
	}
	if cfg.Tags["TODO"] != 2.0 {
		t.Errorf("expected to pick up config from parent: got %v", cfg.Tags)
	}
}

func TestLoad_InvalidSeverity(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{"zero severity", "[tags]\nFIXME = 0.0\n"},
		{"negative severity", "[tags]\nFIXME = -1.0\n"},
		{"over cap", "[tags]\nFIXME = 10.1\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, ".debt.toml"), []byte(tc.content), 0600); err != nil {
				t.Fatal(err)
			}
			_, err := Load(dir)
			if err == nil {
				t.Error("expected error for invalid severity, got nil")
			}
		})
	}
}

func TestLoad_InvalidIgnorePaths(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{"absolute path", "[ignore]\npaths = [\"/etc/passwd\"]\n"},
		{"dotdot traversal", "[ignore]\npaths = [\"../secret\"]\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, ".debt.toml"), []byte(tc.content), 0600); err != nil {
				t.Fatal(err)
			}
			_, err := Load(dir)
			if err == nil {
				t.Error("expected error for invalid ignore path, got nil")
			}
		})
	}
}
