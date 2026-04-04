package git

import (
	"context"
	"fmt"
)

func (c *realClient) ValidateRepo(ctx context.Context, path string) error {
	_, err := runGit(ctx, path, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return fmt.Errorf("%s is not inside a git repository: %w", path, err)
	}
	return nil
}
