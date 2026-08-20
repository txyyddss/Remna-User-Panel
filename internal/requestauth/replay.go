package requestauth

import (
	"sync"
	"time"
)

type sessionReplays struct {
	nonces  map[string]time.Time
	lastUse time.Time
}

type replayCache struct {
	mu          sync.Mutex
	sessions    map[string]*sessionReplays
	maxNonces   int
	nextCleanup time.Time
}

func newReplayCache(maxNonces int) *replayCache {
	return &replayCache{sessions: make(map[string]*sessionReplays), maxNonces: maxNonces}
}

func (cache *replayCache) add(sessionKey, nonce string, now, expiresAt time.Time) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if !now.Before(cache.nextCleanup) {
		cache.cleanup(now)
	}
	session := cache.sessions[sessionKey]
	if session == nil {
		session = &sessionReplays{nonces: make(map[string]time.Time)}
		cache.sessions[sessionKey] = session
	}
	session.lastUse = now
	if expiry, exists := session.nonces[nonce]; exists && expiry.After(now) {
		return ErrReplay
	}
	if len(session.nonces) >= cache.maxNonces {
		for value, expiry := range session.nonces {
			if !expiry.After(now) {
				delete(session.nonces, value)
			}
		}
	}
	if len(session.nonces) >= cache.maxNonces {
		return ErrRateLimited
	}
	session.nonces[nonce] = expiresAt
	return nil
}

func (cache *replayCache) cleanup(now time.Time) {
	for key, session := range cache.sessions {
		for nonce, expiry := range session.nonces {
			if !expiry.After(now) {
				delete(session.nonces, nonce)
			}
		}
		if len(session.nonces) == 0 {
			delete(cache.sessions, key)
		}
	}
	cache.nextCleanup = now.Add(time.Minute)
}
