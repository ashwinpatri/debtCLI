package config

// defaultTags maps each built-in tag name to its base severity weight.
// Severities are in the range (0, 10] as enforced by validate.
var defaultTags = map[string]float64{
	"TODO":     2.0,
	"FIXME":    3.0,
	"HACK":     2.5,
	"PERF":     2.0,
	"SECURITY": 4.0,
	"NOTE":     1.0,
}

// defaultIgnorePaths lists directory prefixes that the walker skips by default.
var defaultIgnorePaths = []string{
	"vendor/",
	"node_modules/",
	".git/",
	"dist/",
	"build/",
}

// defaultIgnoreExtensions lists file suffixes that the walker skips by default.
var defaultIgnoreExtensions = []string{
	".pb.go",
	".gen.go",
	".min.js",
	".lock",
	".sum",
}
