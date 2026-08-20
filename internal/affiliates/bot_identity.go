package affiliates

import (
	"context"
	"strings"
	"sync"
	"time"
)

type IdentityCache struct {
	mu       sync.RWMutex
	source   BotSource
	ttl      time.Duration
	identity BotIdentity
}

func NewIdentityCache(source BotSource, ttl time.Duration) *IdentityCache {
	return &IdentityCache{source: source, ttl: ttl, identity: BotIdentity{Status: "unresolved"}}
}

func (c *IdentityCache) Snapshot() BotIdentity {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.identity
}

func (c *IdentityCache) Refresh(ctx context.Context, now time.Time) error {
	c.mu.RLock()
	checked := c.identity.CheckedAt
	c.mu.RUnlock()
	if checked != nil && now.Sub(*checked) < c.ttl {
		return nil
	}
	user, err := c.source.GetMe(ctx)
	c.mu.Lock()
	defer c.mu.Unlock()
	if err != nil {
		if c.identity.Username != "" {
			c.identity.Status = "stale"
		} else {
			c.identity.Status = "unavailable"
		}
		return err
	}
	username := strings.TrimLeft(strings.TrimSpace(user.Username), "@")
	if !user.IsBot || !validBotUsername(username) {
		c.identity.Status = "unavailable"
		return ErrInvalidInput
	}
	checkedAt := now.UTC()
	c.identity = BotIdentity{Username: username, Status: "ready", CheckedAt: &checkedAt}
	return nil
}

func validBotUsername(username string) bool {
	if len(username) < 5 || len(username) > 32 || !strings.HasSuffix(strings.ToLower(username), "bot") {
		return false
	}
	for _, character := range username {
		if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '_' {
			continue
		}
		return false
	}
	return true
}
