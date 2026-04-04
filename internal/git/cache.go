package git

import "sync"

// blameEntry caches both the result and any error so workers don't retry failed blame calls.
type blameEntry struct {
	result map[int]BlameInfo
	err    error
}

type churnEntry struct {
	result int
	err    error
}

// cache is goroutine-safe via RWMutex — concurrent reads don't block each other,
// only writes do.
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
