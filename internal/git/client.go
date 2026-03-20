// Package git provides a Client interface and a real implementation that
// shells out to the git binary. All exec.Command calls use discrete string
// arguments — never a shell-interpolated string — to prevent injection.
package git

import (
	"context"
)

// BlameInfo holds the authorship metadata for a single line returned by git blame.
type BlameInfo struct {
	Author      string
	AuthorEmail string
	Timestamp   int64 // Unix seconds
}

// Client is the interface the pipeline uses to query git metadata.
// Tests supply a mock; production code uses NewClient().
type Client interface {
	// Blame returns per-line authorship for the given file.
	// Results are cached in memory for the lifetime of the client.
	Blame(ctx context.Context, repoPath, filePath string) (map[int]BlameInfo, error)

	// Churn returns the number of commits that have touched filePath.
	// Results are cached in memory for the lifetime of the client.
	Churn(ctx context.Context, repoPath, filePath string) (int, error)

	// ValidateRepo checks that path is inside a valid git repository.
	ValidateRepo(ctx context.Context, path string) error
}

// realClient implements Client by shelling out to git.
type realClient struct {
	cache *cache
}

// NewClient returns a Client backed by the local git binary.
// The cache is scoped to this client instance and is not shared across runs.
func NewClient() Client {
	return &realClient{cache: newCache()}
}
