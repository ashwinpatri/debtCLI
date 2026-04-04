package models

import "time"

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

type Snapshot struct {
	ID          int64
	RepoPath    string
	Timestamp   time.Time
	HealthScore float64
	ItemCount   int
	Items       []DebtItem
}

type Config struct {
	Tags   map[string]float64
	Ignore IgnoreConfig
}

type IgnoreConfig struct {
	Paths      []string
	Extensions []string
}

type ScanResult struct {
	Snapshot *Snapshot
	Delta    *Delta
}

type Delta struct {
	PreviousScore float64
	ScoreDiff     float64
	NewItems      int
	ResolvedItems int
}
