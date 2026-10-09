package database

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestConnectivityTimelineConfigurationRetentionAndCompletion(t *testing.T) {
	t.Parallel()
	ctx, store := context.Background(), newTestStore(t)
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	for index := 1; index <= 205; index++ {
		item := connectivityTestAttempt(index, connectivityTestHost, now.Add(-time.Minute))
		seedConnectivityAttempt(t, store, item)
		if err := store.FinishConnectivityAttempt(ctx, item.ID, connectivityOutcome(), now.Add(-time.Second)); err != nil {
			t.Fatal(err)
		}
	}
	other := connectivityTestAttempt(300, connectivityOtherHost, now.Add(-time.Minute))
	other.ConfigHash = strings.Repeat("b", 64)
	seedConnectivityAttempt(t, store, other)
	if err := store.FinishConnectivityAttempt(ctx, other.ID, connectivityOutcome(), now.Add(-time.Second)); err != nil {
		t.Fatal(err)
	}
	seedConnectivityAttempt(t, store, connectivityTestAttempt(301, connectivityTestHost, now.Add(-time.Minute)))
	expired := connectivityTestAttempt(302, connectivityTestHost, now.Add(-24*time.Hour))
	seedConnectivityAttempt(t, store, expired)
	if err := store.FinishConnectivityAttempt(ctx, expired.ID, connectivityOutcome(), expired.StartedAt.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	items, err := store.ConnectivityTimeline(ctx, strings.Repeat("a", 64), now)
	if err != nil || len(items) != 205 {
		t.Fatalf("timeline=%d error=%v", len(items), err)
	}
	for _, item := range items {
		if item.FinishedAt == nil || item.ConfigHash != strings.Repeat("a", 64) || !item.StartedAt.After(now.Add(-24*time.Hour)) {
			t.Fatal("running, expired or unrelated attempt included")
		}
	}
}
