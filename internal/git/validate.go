package git

import (
	"context"
	"fmt"
)

// ValidateRepo confirms that path is inside a valid git repository by running
// `git rev-parse --is-inside-work-tree`. Any non-zero exit code is treated as
// a hard failure — the caller should propagate this as a fatal error.
func (c *realClient) ValidateRepo(ctx context.Context, path string) error {
	_, err := runGit(ctx, path, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return fmt.Errorf("%s is not inside a git repository: %w", path, err)
	}
	return nil
}
