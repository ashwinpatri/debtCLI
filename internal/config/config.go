// Package config loads and validates the .debt.toml configuration file.
// If no config file is found, defaults are used silently.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/ashwinpatri/debtCLI/internal/models"
)

const configFileName = ".debt.toml"

// rawConfig mirrors the TOML structure before validation and merging.
type rawConfig struct {
	Tags   map[string]float64 `toml:"tags"`
	Ignore struct {
		Paths      []string `toml:"paths"`
		Extensions []string `toml:"extensions"`
	} `toml:"ignore"`
}

// Load searches for .debt.toml by walking up from repoRoot.
// If no file is found, defaults are returned. If a file is found but invalid,
// an error is returned — the caller should treat this as a hard failure.
func Load(repoRoot string) (*models.Config, error) {
	path, err := findConfig(repoRoot)
	if err != nil {
		return defaultConfig(), nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	var raw rawConfig
	if _, err := toml.NewDecoder(f).Decode(&raw); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	cfg := defaultConfig()

	if len(raw.Tags) > 0 {
		cfg.Tags = raw.Tags
	}
	if len(raw.Ignore.Paths) > 0 {
		cfg.Ignore.Paths = raw.Ignore.Paths
	}
	if len(raw.Ignore.Extensions) > 0 {
		cfg.Ignore.Extensions = raw.Ignore.Extensions
	}

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// findConfig walks up the directory tree from start looking for configFileName.
// It returns fs.ErrNotExist if no file is found before reaching the filesystem root.
func findConfig(start string) (string, error) {
	dir := start
	for {
		candidate := filepath.Join(dir, configFileName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fs.ErrNotExist
		}
		dir = parent
	}
}

// validate checks that all config values are within acceptable bounds.
func validate(cfg *models.Config) error {
	for tag, severity := range cfg.Tags {
		if severity <= 0 || severity > 10 {
			return fmt.Errorf("tag %q severity %.1f out of range (0, 10]", tag, severity)
		}
	}

	for _, p := range cfg.Ignore.Paths {
		if filepath.IsAbs(p) {
			return errors.New("ignore paths must be relative")
		}
		if strings.Contains(p, "..") {
			return errors.New("ignore paths must not contain ..")
		}
	}

	return nil
}

// defaultConfig returns a deep copy of the built-in defaults.
// Each call returns an independent copy so callers cannot mutate shared state.
func defaultConfig() *models.Config {
	tags := make(map[string]float64, len(defaultTags))
	for k, v := range defaultTags {
		tags[k] = v
	}

	paths := make([]string, len(defaultIgnorePaths))
	copy(paths, defaultIgnorePaths)

	exts := make([]string, len(defaultIgnoreExtensions))
	copy(exts, defaultIgnoreExtensions)

	return &models.Config{
		Tags: tags,
		Ignore: models.IgnoreConfig{
			Paths:      paths,
			Extensions: exts,
		},
	}
}
