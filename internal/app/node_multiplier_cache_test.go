package app

import (
	"testing"
	"time"
)

func TestNodeMultiplierCacheCopiesValuesAndExpiresAfterFiveMinutes(t *testing.T) {
	now := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	cache := newNodeMultiplierCache()
	cache.clock = func() time.Time { return now }
	cache.set("node-1", 1_250_000)
	if got, ok := cache.get("node-1"); !ok || got != 1_250_000 {
		t.Fatalf("cache.get() = (%d, %t), want 1250000/true", got, ok)
	}
	now = now.Add(5 * time.Minute)
	if _, ok := cache.get("node-1"); ok {
		t.Fatal("expired node multiplier remained cached")
	}
}
