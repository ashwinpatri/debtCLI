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
	binarySampleSize = 512
	// chanBuffer keeps the pipeline fed without letting the walker run too far ahead.
	chanBuffer = 64
)

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

func hasIgnoredExtension(path string, extensions []string) bool {
	for _, ext := range extensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

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
