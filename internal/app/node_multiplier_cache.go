package app

import (
	"sync"
	"time"
)

const nodeMultiplierCacheTTL = 5 * time.Minute

type cachedNodeMultiplier struct {
	value     int64
	fetchedAt time.Time
}

type nodeMultiplierCache struct {
	mu     sync.RWMutex
	values map[string]cachedNodeMultiplier
	clock  func() time.Time
}

func newNodeMultiplierCache() *nodeMultiplierCache {
	return &nodeMultiplierCache{values: make(map[string]cachedNodeMultiplier), clock: time.Now}
}

func (c *nodeMultiplierCache) get(uuid string) (int64, bool) {
	if c == nil {
		return 0, false
	}
	c.mu.RLock()
	entry, ok := c.values[uuid]
	now := c.clock()
	c.mu.RUnlock()
	if !ok || now.Sub(entry.fetchedAt) >= nodeMultiplierCacheTTL {
		return 0, false
	}
	return entry.value, true
}

func (c *nodeMultiplierCache) set(uuid string, value int64) {
	if c == nil || uuid == "" || value < 0 {
		return
	}
	c.mu.Lock()
	c.values[uuid] = cachedNodeMultiplier{value: value, fetchedAt: c.clock()}
	c.mu.Unlock()
}
