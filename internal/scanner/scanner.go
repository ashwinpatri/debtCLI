package scanner

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ashwinpatri/debtCLI/internal/models"
)

type Scanner struct {
	patterns []tagPattern
}

func New(cfg *models.Config) (*Scanner, error) {
	patterns, err := compilePatterns(cfg.Tags)
	if err != nil {
		return nil, fmt.Errorf("scanner: %w", err)
	}
	return &Scanner{patterns: patterns}, nil
}

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
