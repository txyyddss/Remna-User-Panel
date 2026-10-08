package database

import (
	"context"
	"testing"
	"time"
)

func TestOneBoostSettlesFullBaseForDailyAndGroupRewards(t *testing.T) {
	t.Parallel()
	store, ctx := newTestStore(t), context.Background()
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	user := createBoostRewardMember(t, store, 31904, now)
	claim, err := store.ClaimBoostedDailyActivityRange(ctx, user.ID, "2026-10-07", "UTC", 125, 125, 1, nil, now)
	if err != nil || claim.RewardMinor != 125 {
		t.Fatalf("daily floor = %+v, %v", claim, err)
	}
	message, err := store.RecordGroupMessage(ctx, user.ID, -100123, 1, "2026-10-07", "UTC", 1, 125, 1, now)
	if err != nil || !message.Status.Rewarded || message.Status.RewardMinor != 125 {
		t.Fatalf("group floor = %+v, %v", message, err)
	}
	balance, err := store.Balance(ctx, user.ID)
	if err != nil || balance.Minor != "250" {
		t.Fatalf("balance = %+v, %v", balance, err)
	}
}
