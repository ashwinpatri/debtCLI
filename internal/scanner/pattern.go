package scanner

import (
	"fmt"
	"regexp"
)

// tagPattern pairs a compiled regular expression with the tag name it matches.
// The pattern matches the tag followed by an optional colon and any comment text.
type tagPattern struct {
	tag     string
	pattern *regexp.Regexp
}

// compilePatterns builds one tagPattern per tag in the provided map.
// Tag names are passed through regexp.QuoteMeta before compilation so that
// arbitrary tag names cannot inject regex syntax.
func compilePatterns(tags map[string]float64) ([]tagPattern, error) {
	patterns := make([]tagPattern, 0, len(tags))
	for tag := range tags {
		// Match the tag as a whole word, optionally followed by a colon,
		// then capture everything that follows as the comment body.
		expr := fmt.Sprintf(`(?i)\b%s\b[:\s]?\s*(.*)`, regexp.QuoteMeta(tag))
		re, err := regexp.Compile(expr)
		if err != nil {
			return nil, fmt.Errorf("compile pattern for tag %q: %w", tag, err)
		}
		patterns = append(patterns, tagPattern{tag: tag, pattern: re})
	}
	return patterns, nil
}
