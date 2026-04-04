package git

import (
	"bufio"
	"bytes"
	"context"
)

func (c *realClient) Churn(ctx context.Context, repoPath, filePath string) (int, error) {
	if entry, ok := c.cache.getChurn(filePath); ok {
		return entry.result, entry.err
	}

	out, err := runGit(ctx, repoPath, "log", "--oneline", "--", filePath)
	if err != nil {
		c.cache.setChurn(filePath, 0, err)
		return 0, err
	}

	count := countLines(out)
	c.cache.setChurn(filePath, count, nil)
	return count, nil
}

func countLines(data []byte) int {
	sc := bufio.NewScanner(bytes.NewReader(data))
	n := 0
	for sc.Scan() {
		if len(sc.Text()) > 0 {
			n++
		}
	}
	return n
}
