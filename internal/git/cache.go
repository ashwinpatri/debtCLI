package git

import "sync"

// blameEntry stores the result of a single git blame call, including any error.
// Storing the error allows workers to discover cached failures without re-running git.
type blameEntry struct {
	result map[int]BlameInfo
	err    error
}

// churnEntry stores the result of a single git log churn call.
type churnEntry struct {
	result int
	err    error
}

// cache holds in-memory results for blame and churn lookups within a single
// scan run. It is goroutine-safe for concurrent reads and serialised writes.
//
// The RWMutex allows many workers to read concurrently without blocking each
// other; a write lock is only held for the brief moment an entry is inserted.
type cache struct {
	mu    sync.RWMutex
	blame map[string]blameEntry
	churn map[string]churnEntry
}

func newCache() *cache {
	return &cache{
		blame: make(map[string]blameEntry),
		churn: make(map[string]churnEntry),
	}
}

func (c *cache) getBlame(key string) (blameEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.blame[key]
	return e, ok
}

func (c *cache) setBlame(key string, result map[int]BlameInfo, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.blame[key] = blameEntry{result: result, err: err}
}

func (c *cache) getChurn(key string) (churnEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.churn[key]
	return e, ok
}

func (c *cache) setChurn(key string, result int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.churn[key] = churnEntry{result: result, err: err}
}
