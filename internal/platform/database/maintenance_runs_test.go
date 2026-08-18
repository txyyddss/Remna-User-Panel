package database

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMaintenanceRunLeaseUsesRunIDAsFencingToken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)
	now := time.Date(2026, 8, 18, 2, 0, 0, 0, time.UTC)
	firstID, acquired, err := store.ClaimMaintenanceRun(ctx, "2026-08-18", "worker-a", now.Add(5*time.Minute), now)
	if err != nil || !acquired || firstID == "" {
		t.Fatalf("ClaimMaintenanceRun(first) = (%q, %t, %v)", firstID, acquired, err)
	}
	heldID, acquired, err := store.ClaimMaintenanceRun(ctx, "2026-08-18", "worker-b", now.Add(6*time.Minute), now.Add(time.Minute))
	if err != nil || acquired || heldID != firstID {
		t.Fatalf("ClaimMaintenanceRun(held) = (%q, %t, %v), want %q false", heldID, acquired, err, firstID)
	}

	takeoverAt := now.Add(6 * time.Minute)
	secondID, acquired, err := store.ClaimMaintenanceRun(ctx, "2026-08-18", "worker-b", takeoverAt.Add(5*time.Minute), takeoverAt)
	if err != nil || !acquired || secondID == "" || secondID == firstID {
		t.Fatalf("ClaimMaintenanceRun(takeover) = (%q, %t, %v), first %q", secondID, acquired, err, firstID)
	}
	if err := store.CompleteMaintenanceRun(ctx, firstID, "", map[string]int64{"sessions": 1}, nil, takeoverAt); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale CompleteMaintenanceRun() error = %v, want ErrConflict", err)
	}
	if err := store.CompleteMaintenanceRun(ctx, secondID, "", map[string]int64{"sessions": 1}, nil, takeoverAt); err != nil {
		t.Fatalf("CompleteMaintenanceRun(active): %v", err)
	}
	completedID, acquired, err := store.ClaimMaintenanceRun(ctx, "2026-08-18", "worker-c", takeoverAt.Add(time.Hour), takeoverAt.Add(time.Minute))
	if err != nil || acquired || completedID != secondID {
		t.Fatalf("ClaimMaintenanceRun(completed) = (%q, %t, %v), want %q false", completedID, acquired, err, secondID)
	}
}
