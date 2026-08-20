package accounts

import (
	"testing"
	"time"
)

func TestAuthIdentityLimiterIsolatesAndRefillsIdentities(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	limiter := newAuthIdentityLimiter()
	limiter.now = func() time.Time { return now }

	for range authIdentityCapacity {
		if !limiter.allow(42) {
			t.Fatal("identity was limited before consuming its capacity")
		}
	}
	if limiter.allow(42) {
		t.Fatal("identity was allowed after consuming its capacity")
	}
	if !limiter.allow(43) {
		t.Fatal("one identity exhausted another identity's capacity")
	}

	now = now.Add(authIdentityRefill)
	if !limiter.allow(42) {
		t.Fatal("identity did not regain a token after the refill interval")
	}
}
