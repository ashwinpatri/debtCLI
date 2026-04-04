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

func parsePorcelain(data []byte) (map[int]BlameInfo, error) {
	result := make(map[int]BlameInfo)
	commitInfo := make(map[string]BlameInfo)

	scanner := bufio.NewScanner(bytes.NewReader(data))
	var currentHash string
	var currentLine int

	for scanner.Scan() {
		line := scanner.Text()

		// Header lines start with a 40-char hex hash followed by a space.
		// Nothing else in the porcelain output matches this shape.
		if len(line) > 40 && line[40] == ' ' && isHexHash(line[:40]) {
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
			// Tab-prefixed line is the source content — commit block is done.
			if currentHash != "" && currentLine > 0 {
				result[currentLine] = commitInfo[currentHash]
			}
			currentHash = ""
			currentLine = 0
			continue
		}

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

func runGit(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return out, nil
}
