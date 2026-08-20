package accounts

import (
	"sync"
	"time"
)

const (
	authIdentityCapacity = 4
	authIdentityRefill   = 30 * time.Second
	authIdentityMaxKeys  = 8192
)

type authIdentityState struct {
	tokens float64
	at     time.Time
}

type authIdentityLimiter struct {
	mu     sync.Mutex
	states map[int64]authIdentityState
	now    func() time.Time
}

func newAuthIdentityLimiter() *authIdentityLimiter {
	return &authIdentityLimiter{states: make(map[int64]authIdentityState), now: time.Now}
}

func (limiter *authIdentityLimiter) allow(telegramID int64) bool {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	now := limiter.now()
	state, exists := limiter.states[telegramID]
	if !exists {
		if len(limiter.states) >= authIdentityMaxKeys {
			limiter.removeOldest()
		}
		state = authIdentityState{tokens: authIdentityCapacity, at: now}
	}
	state.tokens = min(authIdentityCapacity, state.tokens+float64(now.Sub(state.at))/float64(authIdentityRefill))
	state.at = now
	allowed := state.tokens >= 1
	if allowed {
		state.tokens--
	}
	limiter.states[telegramID] = state
	return allowed
}

func (limiter *authIdentityLimiter) removeOldest() {
	var oldestID int64
	var oldest time.Time
	for telegramID, state := range limiter.states {
		if oldestID == 0 || state.at.Before(oldest) {
			oldestID, oldest = telegramID, state.at
		}
	}
	delete(limiter.states, oldestID)
}
