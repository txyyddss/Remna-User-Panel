package requestauth

import (
	"errors"
	"sync"
	"time"
)

var errReplayCapacity = errors.New("request replay cache is at capacity")

type replayCache struct {
	mu          sync.Mutex
	entries     map[string]time.Time
	maxEntries  int
	nextCleanup time.Time
}

func newReplayCache(maxEntries int) *replayCache {
	return &replayCache{entries: make(map[string]time.Time), maxEntries: maxEntries}
}

func (cache *replayCache) add(key string, now, expiresAt time.Time) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if !now.Before(cache.nextCleanup) || len(cache.entries) >= cache.maxEntries {
		for nonce, expiry := range cache.entries {
			if !expiry.After(now) {
				delete(cache.entries, nonce)
			}
		}
		cache.nextCleanup = now.Add(time.Minute)
	}
	if expiry, exists := cache.entries[key]; exists && expiry.After(now) {
		return ErrReplay
	}
	if len(cache.entries) >= cache.maxEntries {
		return errReplayCapacity
	}
	cache.entries[key] = expiresAt
	return nil
}
