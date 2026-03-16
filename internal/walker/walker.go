// Package walker traverses a directory tree and emits file paths that are
// eligible for scanning. It skips binary files, non-UTF-8 files, and any
// paths or extensions listed in the config ignore lists.
package walker

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/ashwinpatri/debtCLI/internal/models"
)

const (
	// binarySampleSize is the number of bytes read to detect binary files.
	binarySampleSize = 512

	// chanBuffer controls how many paths can be queued before the walker blocks.
	// Sized to match the typical worker count so the pipeline stays fed.
	chanBuffer = 64
)

// Walk traverses root and sends eligible file paths to the returned channel.
// The channel is closed when the walk completes or ctx is cancelled.
// Individual file errors are logged and skipped; they do not abort the walk.
func Walk(ctx context.Context, root string, cfg *models.Config) (<-chan string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	paths := make(chan string, chanBuffer)

	go func() {
		defer close(paths)

		err := filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				log.Printf("walker: skipping %s: %v", path, err)
				return nil
			}

			if ctx.Err() != nil {
				return ctx.Err()
			}

			if d.IsDir() {
				if shouldSkipDir(absRoot, path, cfg.Ignore.Paths) {
					return filepath.SkipDir
				}
				return nil
			}

			if !d.Type().IsRegular() {
				return nil
			}

			rel, err := filepath.Rel(absRoot, path)
			if err != nil {
				log.Printf("walker: rel path error for %s: %v", path, err)
				return nil
			}

			if hasIgnoredExtension(path, cfg.Ignore.Extensions) {
				return nil
			}

			if ignored, err := isIgnoredPath(absRoot, rel, cfg.Ignore.Paths); err != nil || ignored {
				return nil
			}

			ok, err := isReadableText(path)
			if err != nil {
				log.Printf("walker: skipping %s: %v", path, err)
				return nil
			}
			if !ok {
				return nil
			}

			select {
			case paths <- path:
			case <-ctx.Done():
				return ctx.Err()
			}

			return nil
		})

		if err != nil && ctx.Err() == nil {
			log.Printf("walker: walk error: %v", err)
		}
	}()

	return paths, nil
}

// shouldSkipDir returns true if dir matches any of the configured ignore prefixes.
// The .git directory is always skipped regardless of config.
func shouldSkipDir(root, dir string, ignorePaths []string) bool {
	base := filepath.Base(dir)
	if base == ".git" {
		return true
	}

	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return false
	}

	for _, p := range ignorePaths {
		normalized := filepath.ToSlash(rel)
		prefix := strings.TrimSuffix(p, "/")
		if normalized == prefix || strings.HasPrefix(normalized, prefix+"/") {
			return true
		}
	}
	return false
}

// isIgnoredPath checks whether a relative file path falls under any ignored prefix.
func isIgnoredPath(root, rel string, ignorePaths []string) (bool, error) {
	for _, p := range ignorePaths {
		prefix := filepath.ToSlash(strings.TrimSuffix(p, "/"))
		normalized := filepath.ToSlash(rel)
		if strings.HasPrefix(normalized, prefix+"/") || normalized == prefix {
			return true, nil
		}
	}
	_ = root
	return false, nil
}

// hasIgnoredExtension returns true if path ends with any of the ignored extensions.
func hasIgnoredExtension(path string, extensions []string) bool {
	for _, ext := range extensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

// isReadableText returns true if the file is a non-empty, valid UTF-8 text file.
// It reads up to binarySampleSize bytes to check for null bytes (binary indicator)
// and validates that the sample is valid UTF-8.
func isReadableText(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	buf := make([]byte, binarySampleSize)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}
	if n == 0 {
		return false, nil
	}

	sample := buf[:n]
	if bytes.IndexByte(sample, 0) >= 0 {
		return false, nil
	}
	if !utf8.Valid(sample) {
		return false, nil
	}

	return true, nil
}
