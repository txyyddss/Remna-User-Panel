package database

import (
	"context"
	"testing"
	"time"
)

func TestBufferedGroupMessageFactsFlushAsDeduplicatedBatch(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	now := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	for _, messageID := range []int64{1, 1, 2} {
		if err := store.BufferGroupMessageFact(-100, messageID, "2026-08-20", now); err != nil {
			t.Fatalf("BufferGroupMessageFact(%d): %v", messageID, err)
		}
	}
	var before int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM activity_group_message_raw_events`).Scan(&before); err != nil || before != 0 {
		t.Fatalf("facts before flush = %d, error %v", before, err)
	}
	if err := store.FlushGroupMessageFacts(context.Background()); err != nil {
		t.Fatalf("FlushGroupMessageFacts(): %v", err)
	}
	var after int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM activity_group_message_raw_events`).Scan(&after); err != nil || after != 2 {
		t.Fatalf("facts after flush = %d, error %v", after, err)
	}
	if err := store.FlushGroupMessageFacts(context.Background()); err != nil {
		t.Fatalf("empty FlushGroupMessageFacts(): %v", err)
	}
}
