// Package models defines the core value types shared across all packages.
// It has no imports beyond the standard library and no methods or logic.
package models

import "time"

// DebtItem represents a single technical debt marker found in source code,
// enriched with git blame metadata and a computed score.
type DebtItem struct {
	File        string
	Line        int
	Tag         string
	Comment     string
	Author      string
	AuthorEmail string
	Date        time.Time
	Churn       int
	Score       float64
}

// Snapshot captures the full state of a scan run: every debt item found,
// the aggregate health score, and identifying metadata for the repo.
type Snapshot struct {
	ID          int64
	RepoPath    string
	Timestamp   time.Time
	HealthScore float64
	ItemCount   int
	Items       []DebtItem
}

// Config holds the validated, merged configuration for a scan run.
// It is produced by config.Load and consumed by the scanner and walker.
type Config struct {
	Tags   map[string]float64
	Ignore IgnoreConfig
}

// IgnoreConfig lists paths and file extensions the walker should skip.
type IgnoreConfig struct {
	Paths      []string
	Extensions []string
}

// ScanResult is the top-level value returned to the output layer after a
// complete scan: the new snapshot and an optional delta against the previous one.
type ScanResult struct {
	Snapshot *Snapshot
	Delta    *Delta
}

// Delta describes how this scan compares to the most recent prior snapshot.
// It is nil when no prior snapshot exists for the repo.
type Delta struct {
	PreviousScore float64
	ScoreDiff     float64
	NewItems      int
	ResolvedItems int
}
