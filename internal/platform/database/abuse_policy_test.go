package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/abuse"
)

func TestAbusePolicyStreakBoundsAndRevision(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	now := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
	policy, err := store.Policy(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if policy.StreakSeconds != abuse.DefaultStreakSeconds {
		t.Fatalf("migration streak = %d, want %d", policy.StreakSeconds, abuse.DefaultStreakSeconds)
	}
	for _, value := range []int{abuse.MinStreakSeconds, abuse.MaxStreakSeconds} {
		policy.StreakSeconds = value
		policy, err = store.UpdatePolicy(ctx, "admin", policy, now)
		if err != nil {
			t.Fatalf("UpdatePolicy(%d): %v", value, err)
		}
		if policy.StreakSeconds != value {
			t.Fatalf("stored streak = %d, want %d", policy.StreakSeconds, value)
		}
	}
	stale := policy
	policy.StreakSeconds = 30
	policy, err = store.UpdatePolicy(ctx, "admin", policy, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.UpdatePolicy(ctx, "admin", stale, now); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale update error = %v, want ErrConflict", err)
	}
}

func TestAbusePolicyRejectsOutOfRangeStreak(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	policy, err := store.Policy(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []int{0, abuse.MaxStreakSeconds + 1} {
		policy.StreakSeconds = value
		if _, err = store.UpdatePolicy(ctx, "admin", policy, time.Now().UTC()); !errors.Is(err, abuse.ErrInvalid) {
			t.Fatalf("UpdatePolicy(%d) error = %v, want ErrInvalid", value, err)
		}
	}
}
