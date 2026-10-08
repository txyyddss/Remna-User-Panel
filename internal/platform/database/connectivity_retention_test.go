package database

import (
	"context"
	"testing"
	"time"
)

func TestConnectivityRetentionUsesExactStartBoundaryBeforePruning(t *testing.T) {
	t.Parallel()
	ctx, store := context.Background(), newTestStore(t)
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	cutoff := now.Add(-24 * time.Hour)
	seedConnectivityAttempt(t, store, connectivityTestAttempt(1, connectivityTestHost, cutoff.Add(-time.Nanosecond)))
	seedConnectivityAttempt(t, store, connectivityTestAttempt(2, connectivityTestHost, cutoff))
	seedConnectivityAttempt(t, store, connectivityTestAttempt(3, connectivityTestHost, cutoff.Add(time.Nanosecond)))
	seedConnectivityAttempt(t, store, connectivityTestAttempt(4, connectivityOtherHost, cutoff))
	page, err := store.ConnectivityHistory(ctx, "", "", 50, now)
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != connectivityTestAttempt(3, "", now).ID {
		t.Fatalf("unpruned retention history=%+v,%v", page, err)
	}
	latest, err := store.LatestConnectivityAttempts(ctx, page.Items[0].ConfigHash, now)
	if err != nil || len(latest) != 1 || latest[0].ID != page.Items[0].ID {
		t.Fatalf("unpruned latest=%+v,%v", latest, err)
	}
	// Completion cannot extend an attempt's retention window.
	if err := store.FinishConnectivityAttempt(ctx, connectivityTestAttempt(2, "", now).ID, connectivityOutcome(), now); err != nil {
		t.Fatal(err)
	}
	if err := store.PruneConnectivityAttempts(ctx, now); err != nil {
		t.Fatal(err)
	}
	var remaining int
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM host_connectivity_attempts`).Scan(&remaining); err != nil || remaining != 1 {
		t.Fatalf("retained count=%d,%v", remaining, err)
	}
	if err := store.PruneConnectivityAttempts(ctx, now.Add(time.Nanosecond)); err != nil {
		t.Fatal(err)
	}
	page, err = store.ConnectivityHistory(ctx, "", "", 50, now.Add(time.Nanosecond))
	if err != nil || len(page.Items) != 0 {
		t.Fatalf("exact subsequent boundary=%+v,%v", page, err)
	}
}

func TestConnectivityRecoveryRetainsStartsAndTerminalResults(t *testing.T) {
	t.Parallel()
	ctx, store := context.Background(), newTestStore(t)
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	started := now.Add(-time.Hour)
	running := connectivityTestAttempt(1, connectivityTestHost, started)
	done := connectivityTestAttempt(2, connectivityOtherHost, started)
	seedConnectivityAttempt(t, store, running)
	seedConnectivityAttempt(t, store, done)
	if err := store.FinishConnectivityAttempt(ctx, done.ID, connectivityOutcome(), started.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := store.RecoverConnectivityAttempts(ctx, now); err != nil {
		t.Fatal(err)
	}
	page, err := store.ConnectivityHistory(ctx, "", "", 50, now)
	if err != nil || len(page.Items) != 2 {
		t.Fatalf("recovered history=%+v,%v", page, err)
	}
	for _, item := range page.Items {
		if !item.StartedAt.Equal(started) {
			t.Fatalf("start changed=%+v", item)
		}
		if item.ID == running.ID && (item.Status != "interrupted" || item.ErrorCode != "CONNECTIVITY_INTERRUPTED" || item.FinishedAt == nil || !item.FinishedAt.Equal(now)) {
			t.Fatalf("unfinished recovery=%+v", item)
		}
		if item.ID == done.ID && (item.Status != "connected" || item.FinishedAt == nil || !item.FinishedAt.Equal(started.Add(time.Second))) {
			t.Fatalf("terminal result changed=%+v", item)
		}
	}
	if err := store.RecoverConnectivityAttempts(ctx, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	latest, err := store.LatestConnectivityAttempts(ctx, running.ConfigHash, now.Add(time.Minute))
	if err != nil || len(latest) != 2 || latest[0].FinishedAt == nil || !latest[0].FinishedAt.Equal(now) {
		t.Fatalf("recovery replaced evidence=%+v,%v", latest, err)
	}
}
