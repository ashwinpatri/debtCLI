package git

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Blame shells out to `git blame --porcelain -- <file>` and returns a map
// from 1-based line number to BlameInfo. Results are cached by filePath.
//
// The porcelain format emits a 40-character commit hash followed by the
// original line number, final line number, and group line count on the
// header line. Author and timestamp fields follow as "key value" pairs.
func (c *realClient) Blame(ctx context.Context, repoPath, filePath string) (map[int]BlameInfo, error) {
	if entry, ok := c.cache.getBlame(filePath); ok {
		return entry.result, entry.err
	}

	out, err := runGit(ctx, repoPath, "blame", "--porcelain", "--", filePath)
	if err != nil {
		c.cache.setBlame(filePath, nil, err)
		return nil, err
	}

	result, err := parsePorcelain(out)
	c.cache.setBlame(filePath, result, err)
	return result, err
}

// parsePorcelain parses the output of `git blame --porcelain`.
// The format alternates between a header line starting with a 40-char hash
// and a series of "key value" metadata lines, ending with a tab-prefixed
// source line. We extract author, author-mail, and author-time per commit hash.
func parsePorcelain(data []byte) (map[int]BlameInfo, error) {
	result := make(map[int]BlameInfo)
	// commitInfo caches blame data for each commit hash seen so far.
	commitInfo := make(map[string]BlameInfo)

	scanner := bufio.NewScanner(bytes.NewReader(data))
	var currentHash string
	var currentLine int

	for scanner.Scan() {
		line := scanner.Text()

		// A porcelain header line starts with a 40-character hex commit hash
		// followed by a space. Nothing else in the output matches this shape.
		if len(line) > 40 && line[40] == ' ' && isHexHash(line[:40]) {
			// Header: "<hash> <orig-line> <final-line> [<count>]"
			fields := strings.Fields(line)
			if len(fields) < 3 {
				continue
			}
			currentHash = fields[0]
			n, err := strconv.Atoi(fields[2])
			if err != nil {
				continue
			}
			currentLine = n
			continue
		}

		if strings.HasPrefix(line, "\t") {
			// Source line — commit block is complete, record the entry.
			if currentHash != "" && currentLine > 0 {
				result[currentLine] = commitInfo[currentHash]
			}
			currentHash = ""
			currentLine = 0
			continue
		}

		// Metadata fields for the current commit hash.
		key, value, ok := strings.Cut(line, " ")
		if !ok {
			continue
		}
		info := commitInfo[currentHash]
		switch key {
		case "author":
			info.Author = value
		case "author-mail":
			info.AuthorEmail = strings.Trim(value, "<>")
		case "author-time":
			ts, err := strconv.ParseInt(value, 10, 64)
			if err == nil {
				info.Timestamp = ts
			}
		}
		commitInfo[currentHash] = info
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("parse blame output: %w", err)
	}

	return result, nil
}

// isHexHash returns true if s is exactly 40 lowercase hex characters.
func isHexHash(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// runGit executes git with the supplied arguments in dir and returns stdout.
func runGit(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return out, nil
}
