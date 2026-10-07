package database

import (
	"context"
	"testing"
	"time"
)

func TestBoostMigrationResetsOnlyUnpaidProgressAndKeepsDedupe(t *testing.T) {
	ctx, store := context.Background(), newTestStore(t)
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	paid := createBoostRewardMember(t, store, 31910, now)
	unpaid := createBoostRewardMember(t, store, 31911, now)
	if _, err := store.RecordGroupMessage(ctx, paid.ID, -100123, 1, "2026-10-07", "UTC", 1, 125, 2, now); err != nil {
		t.Fatal(err)
	}
	if _, err := store.RecordGroupMessage(ctx, unpaid.ID, -100123, 2, "2026-10-07", "UTC", 2, 125, 2, now); err != nil {
		t.Fatal(err)
	}
	reapplyRecoveryMigration(t, store, "058_require_boosted_group_messages.sql")
	paidStatus, err := store.GroupMessageRewardStatus(ctx, paid.ID, "2026-10-07", 1, 999)
	if err != nil || !paidStatus.Rewarded || paidStatus.MessageCount != 1 || paidStatus.RewardMinor != 125 {
		t.Fatalf("paid status = %+v, %v", paidStatus, err)
	}
	unpaidStatus, err := store.GroupMessageRewardStatus(ctx, unpaid.ID, "2026-10-07", 2, 125)
	if err != nil || unpaidStatus.Rewarded || unpaidStatus.MessageCount != 0 {
		t.Fatalf("unpaid status = %+v, %v", unpaidStatus, err)
	}
	replay, err := store.RecordGroupMessage(ctx, unpaid.ID, -100123, 2, "2026-10-07", "UTC", 2, 125, 3, now)
	if err != nil || !replay.Replayed || replay.Counted || replay.Status.MessageCount != 0 {
		t.Fatalf("legacy replay = %+v, %v", replay, err)
	}
	var events, ledgers int
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM activity_group_message_events`).Scan(&events); err != nil || events != 2 {
		t.Fatalf("events = %d, %v", events, err)
	}
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM ledger_entries WHERE kind='activity_group_message_reward'`).Scan(&ledgers); err != nil || ledgers != 1 {
		t.Fatalf("ledger rows = %d, %v", ledgers, err)
	}
	assertNotificationCounts(t, store, 1, 1)
}
