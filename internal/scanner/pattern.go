package scanner

import (
	"fmt"
	"regexp"
)

type tagPattern struct {
	tag     string
	pattern *regexp.Regexp
}

func compilePatterns(tags map[string]float64) ([]tagPattern, error) {
	patterns := make([]tagPattern, 0, len(tags))
	for tag := range tags {
		expr := fmt.Sprintf(`(?i)\b%s\b[:\s]?\s*(.*)`, regexp.QuoteMeta(tag))
		re, err := regexp.Compile(expr)
		if err != nil {
			return nil, fmt.Errorf("compile pattern for tag %q: %w", tag, err)
		}
		patterns = append(patterns, tagPattern{tag: tag, pattern: re})
	}
	return patterns, nil
}
