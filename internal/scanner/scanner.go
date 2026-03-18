// Package scanner reads source files line by line and identifies technical
// debt markers based on the tag patterns defined in the config.
package scanner

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ashwinpatri/debtCLI/internal/models"
)

// Scanner holds compiled tag patterns and is safe for concurrent use once
// constructed. Create one per scan run with New and reuse across goroutines.
type Scanner struct {
	patterns []tagPattern
}

// New compiles the tag patterns from cfg and returns a ready Scanner.
// Returns an error if any pattern fails to compile, which should not happen
// in practice since tag names are sanitised with regexp.QuoteMeta.
func New(cfg *models.Config) (*Scanner, error) {
	patterns, err := compilePatterns(cfg.Tags)
	if err != nil {
		return nil, fmt.Errorf("scanner: %w", err)
	}
	return &Scanner{patterns: patterns}, nil
}

// Scan reads the file at path and returns one DebtItem per matched line.
// Only the first matching tag on each line is recorded. The blame fields
// (Author, AuthorEmail, Date, Churn, Score) are left at their zero values
// and filled in later by the pipeline.
func (s *Scanner) Scan(path string) ([]models.DebtItem, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var items []models.DebtItem
	sc := bufio.NewScanner(f)
	lineNum := 0

	for sc.Scan() {
		lineNum++
		line := sc.Text()

		for _, tp := range s.patterns {
			m := tp.pattern.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			comment := ""
			if len(m) > 1 {
				comment = strings.TrimSpace(m[1])
			}
			items = append(items, models.DebtItem{
				File:    path,
				Line:    lineNum,
				Tag:     tp.tag,
				Comment: comment,
			})
			break // one match per line
		}
	}

	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", path, err)
	}

	return items, nil
}
