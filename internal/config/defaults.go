package config

var defaultTags = map[string]float64{
	"TODO":     2.0,
	"FIXME":    3.0,
	"HACK":     2.5,
	"PERF":     2.0,
	"SECURITY": 4.0,
	"NOTE":     1.0,
}

var defaultIgnorePaths = []string{
	"vendor/",
	"node_modules/",
	".git/",
	"dist/",
	"build/",
}

var defaultIgnoreExtensions = []string{
	".pb.go",
	".gen.go",
	".min.js",
	".lock",
	".sum",
}
