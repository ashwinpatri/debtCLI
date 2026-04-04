package git

import (
	"context"
)

type BlameInfo struct {
	Author      string
	AuthorEmail string
	Timestamp   int64 // Unix seconds
}

type Client interface {
	Blame(ctx context.Context, repoPath, filePath string) (map[int]BlameInfo, error)
	Churn(ctx context.Context, repoPath, filePath string) (int, error)
	ValidateRepo(ctx context.Context, path string) error
}

type realClient struct {
	cache *cache
}

func NewClient() Client {
	return &realClient{cache: newCache()}
}
