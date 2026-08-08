package database

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/activity"
)

type fixedActivityRandom struct {
	value int64
	err   error
}

func (random fixedActivityRandom) Int63n(upperBound int64) (int64, error) {
	if random.err != nil {
		return 0, random.err
	}
	return random.value % upperBound, nil
}

func TestActivityBetAtomicOutcomeAndReplay(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		roll        int64
		wantWon     bool
		wantPayout  int64
		wantBalance string
		wantLedger  int
	}{
		{name: "win returns total multiplier", roll: 0, wantWon: true, wantPayout: 200, wantBalance: "1100", wantLedger: 2},
		{name: "loss keeps stake debited", roll: 9_999, wantWon: false, wantPayout: 0, wantBalance: "900", wantLedger: 1},
	}
	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			store := newTestStore(t)
			user := createTestUser(t, store, 31_000+int64(index))
			if _, err := store.AdjustBalance(ctx, user.ID, 1_000, "activity-seed-"+test.name, "seed", time.Now()); err != nil {
				t.Fatalf("AdjustBalance(): %v", err)
			}
			game, err := store.SaveActivityGame(ctx, activity.GameInput{Name: "Coin", Enabled: true, WinChanceBPS: 5_000,
				MinimumStakeMinor: 100, MaximumStakeMinor: 500, ReturnMultiplierBPS: 20_000}, time.Now())
			if err != nil {
				t.Fatalf("SaveActivityGame(): %v", err)
			}
			played, err := store.PlaceActivityBet(ctx, user.ID, game.ID, 100, "bet-key", fixedActivityRandom{value: test.roll}, time.Now())
			if err != nil {
				t.Fatalf("PlaceActivityBet(): %v", err)
			}
			if played.Won != test.wantWon || played.PayoutMinor != test.wantPayout {
				t.Fatalf("bet = won %t payout %d, want %t/%d", played.Won, played.PayoutMinor, test.wantWon, test.wantPayout)
			}
			replayed, err := store.PlaceActivityBet(ctx, user.ID, game.ID, 100, "bet-key", fixedActivityRandom{value: 1}, time.Now())
			if err != nil || !replayed.Replayed || replayed.ID != played.ID {
				t.Fatalf("replay = (%+v, %v)", replayed, err)
			}
			balance, err := store.Balance(ctx, user.ID)
			if err != nil || balance.Minor != test.wantBalance {
				t.Fatalf("Balance() = (%s, %v), want %s", balance.Minor, err, test.wantBalance)
			}
			var ledgerCount int
			if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM ledger_entries WHERE reference_id=?`, played.ID).Scan(&ledgerCount); err != nil {
				t.Fatalf("count bet ledger: %v", err)
			}
			if ledgerCount != test.wantLedger {
				t.Fatalf("bet ledger count = %d, want %d", ledgerCount, test.wantLedger)
			}
		})
	}
}

func TestDailyActivityUsesConfiguredTimezoneDate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 31_100)
	now := time.Date(2026, 8, 8, 16, 30, 0, 0, time.UTC) // 2026-08-09 in Asia/Shanghai.
	claimed, err := store.ClaimDailyActivity(ctx, user.ID, "2026-08-09", "Asia/Shanghai", 125, now)
	if err != nil {
		t.Fatalf("ClaimDailyActivity(): %v", err)
	}
	replayed, err := store.ClaimDailyActivity(ctx, user.ID, "2026-08-09", "Asia/Shanghai", 999, now.Add(time.Hour))
	if err != nil || !replayed.AlreadyClaimed || replayed.ID != claimed.ID || replayed.RewardMinor != 125 {
		t.Fatalf("check-in replay = (%+v, %v)", replayed, err)
	}
	if _, err := store.ClaimDailyActivity(ctx, user.ID, "2026-08-08", "Asia/Shanghai", 125, now); !errors.Is(err, activity.ErrInvalidInput) {
		t.Fatalf("wrong local date error = %v, want invalid input", err)
	}
}

func TestDailyActivityRewardCanReduceExistingDebt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 31_101)
	now := time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)
	if _, err := store.AdjustBalance(ctx, user.ID, -500, "activity-debt", "refund debt", now); err != nil {
		t.Fatalf("AdjustBalance(): %v", err)
	}
	claimed, err := store.ClaimDailyActivity(ctx, user.ID, "2026-08-08", "UTC", 125, now)
	if err != nil {
		t.Fatalf("ClaimDailyActivity(): %v", err)
	}
	if claimed.BalanceAfterMinor != -375 {
		t.Fatalf("BalanceAfterMinor = %d, want -375", claimed.BalanceAfterMinor)
	}
	var ledgerCount int
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM ledger_entries WHERE kind='activity_daily_checkin' AND reference_id=?`, claimed.ID).Scan(&ledgerCount); err != nil {
		t.Fatalf("count ledger: %v", err)
	}
	if ledgerCount != 1 {
		t.Fatalf("ledger count = %d, want 1", ledgerCount)
	}
}

func TestGroupMessageRewardCountsEligibleMessagesAndReplaysAtomically(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 31_300)
	now := time.Date(2026, 8, 8, 16, 30, 0, 0, time.UTC)
	config := activity.GroupMessageRewardConfig{Timezone: "Asia/Shanghai", Threshold: 2, RewardMinor: 125}

	withoutSubscription, err := store.RecordGroupMessage(ctx, user.ID, -100, 1, "2026-08-09", "Asia/Shanghai", config.Threshold, config.RewardMinor, now)
	if err != nil || withoutSubscription.Counted || withoutSubscription.Status.MessageCount != 0 {
		t.Fatalf("message without subscription = (%+v, %v)", withoutSubscription, err)
	}

	combo := saveTestCombo(t, store, "group-reward", 0, 30)
	purchase, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID, IdempotencyKey: "group-reward-purchase"}, now)
	if err != nil {
		t.Fatalf("CreatePurchase(): %v", err)
	}
	if _, err := store.DB().ExecContext(ctx, `UPDATE purchases SET status='active',valid_from=?,valid_until=? WHERE id=?`, stamp(now.Add(-time.Hour)), stamp(now.Add(24*time.Hour)), purchase.ID); err != nil {
		t.Fatalf("activate test purchase: %v", err)
	}

	first, err := store.RecordGroupMessage(ctx, user.ID, -100, 2, "2026-08-09", "Asia/Shanghai", config.Threshold, config.RewardMinor, now)
	if err != nil || !first.Counted || first.Status.MessageCount != 1 || first.Status.Rewarded {
		t.Fatalf("first message = (%+v, %v)", first, err)
	}
	second, err := store.RecordGroupMessage(ctx, user.ID, -100, 3, "2026-08-09", "Asia/Shanghai", config.Threshold, config.RewardMinor, now)
	if err != nil || !second.Counted || second.Status.MessageCount != 2 || !second.Status.Rewarded {
		t.Fatalf("threshold message = (%+v, %v)", second, err)
	}
	replay, err := store.RecordGroupMessage(ctx, user.ID, -100, 3, "2026-08-09", "Asia/Shanghai", config.Threshold, config.RewardMinor, now.Add(time.Minute))
	if err != nil || !replay.Replayed || replay.Status.MessageCount != 2 {
		t.Fatalf("duplicate message = (%+v, %v)", replay, err)
	}
	status, err := store.GroupMessageRewardStatus(ctx, user.ID, "2026-08-09", config.Threshold, config.RewardMinor)
	if err != nil || !status.Rewarded || status.MessageCount != 2 {
		t.Fatalf("status = (%+v, %v)", status, err)
	}
	balance, err := store.Balance(ctx, user.ID)
	if err != nil || balance.Minor != "125" {
		t.Fatalf("reward balance = (%+v, %v), want 125", balance, err)
	}
	var rewardLedgerCount int
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM ledger_entries WHERE user_id=? AND kind='activity_group_message_reward'`, user.ID).Scan(&rewardLedgerCount); err != nil || rewardLedgerCount != 1 {
		t.Fatalf("reward ledger count = (%d, %v), want 1", rewardLedgerCount, err)
	}
}

func TestDeductBalanceRejectsInsufficientFunds(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 31_301)
	if _, err := store.AdjustBalance(ctx, user.ID, 500, "deduct-seed", "seed", time.Now()); err != nil {
		t.Fatalf("AdjustBalance(): %v", err)
	}
	if _, err := store.DeductBalance(ctx, user.ID, 250, "deduct-1", "test deduction", time.Now()); err != nil {
		t.Fatalf("DeductBalance(): %v", err)
	}
	if _, err := store.DeductBalance(ctx, user.ID, 250, "deduct-1", "test deduction", time.Now()); err != nil {
		t.Fatalf("DeductBalance(replay): %v", err)
	}
	if _, err := store.DeductBalance(ctx, user.ID, 300, "deduct-2", "test deduction", time.Now()); !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("insufficient deduction error = %v, want %v", err, ErrInsufficientBalance)
	}
	balance, err := store.Balance(ctx, user.ID)
	if err != nil || balance.Minor != "250" {
		t.Fatalf("balance = (%+v, %v), want 250", balance, err)
	}
}

func TestLuckyDrawRequiresWorstCaseCoverageAndReplays(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 31_200)
	if _, err := store.AdjustBalance(ctx, user.ID, 349, "draw-seed", "seed", time.Now()); err != nil {
		t.Fatalf("AdjustBalance(): %v", err)
	}
	draw, err := store.SaveLuckyDraw(ctx, activity.LuckyDrawInput{Name: "Exact coverage", Enabled: true, FeeMinor: 100,
		Prizes: []activity.PrizeInput{{Name: "Minus", Weight: 1, Reward: activity.Reward{Kind: activity.RewardTXBDelta, TXBDeltaMinor: -250}}}}, time.Now())
	if err != nil {
		t.Fatalf("SaveLuckyDraw(): %v", err)
	}
	if _, err := store.PlayLuckyDraw(ctx, user.ID, draw.ID, "draw-key", fixedActivityRandom{}, time.Now()); !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("PlayLuckyDraw(under-covered) = %v, want insufficient balance", err)
	}
	if _, err := store.AdjustBalance(ctx, user.ID, 1, "draw-coverage", "coverage", time.Now()); err != nil {
		t.Fatalf("AdjustBalance(coverage): %v", err)
	}
	result, err := store.PlayLuckyDraw(ctx, user.ID, draw.ID, "draw-key", fixedActivityRandom{}, time.Now())
	if err != nil || result.BalanceAfterMinor != 0 {
		t.Fatalf("PlayLuckyDraw() = (%+v, %v), want zero balance", result, err)
	}
	replayed, err := store.PlayLuckyDraw(ctx, user.ID, draw.ID, "draw-key", fixedActivityRandom{}, time.Now())
	if err != nil || !replayed.Replayed || replayed.ID != result.ID {
		t.Fatalf("PlayLuckyDraw(replay) = (%+v, %v)", replayed, err)
	}
	var results int
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM activity_draw_results`).Scan(&results); err != nil || results != 1 {
		t.Fatalf("draw result count = %d, %v, want 1", results, err)
	}
}

func TestActivityDescriptionsPersistAndEnterResultSnapshots(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 31_300)
	now := time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)
	if _, err := store.AdjustBalance(ctx, user.ID, 1_000, "activity-description-seed", "seed", now); err != nil {
		t.Fatalf("AdjustBalance(): %v", err)
	}
	game, err := store.SaveActivityGame(ctx, activity.GameInput{Name: "Coin", Icon: "coin", Description: "  Clear game terms  ", Enabled: true,
		WinChanceBPS: 10_000, MinimumStakeMinor: 100, MaximumStakeMinor: 100, ReturnMultiplierBPS: 10_000}, now)
	if err != nil {
		t.Fatalf("SaveActivityGame(): %v", err)
	}
	if game.Description != "Clear game terms" {
		t.Fatalf("game description = %q", game.Description)
	}
	bet, err := store.PlaceActivityBet(ctx, user.ID, game.ID, 100, "description-bet", fixedActivityRandom{}, now)
	if err != nil {
		t.Fatalf("PlaceActivityBet(): %v", err)
	}
	if !strings.Contains(bet.ConfigurationSnapshot, `"description":"Clear game terms"`) {
		t.Fatalf("game snapshot omits description: %s", bet.ConfigurationSnapshot)
	}
	draw, err := store.SaveLuckyDraw(ctx, activity.LuckyDrawInput{Name: "Draw", Description: "  Clear draw terms  ", Enabled: true,
		Prizes: []activity.PrizeInput{{Name: "Nothing", Weight: 1, Reward: activity.Reward{Kind: activity.RewardNone}}}}, now)
	if err != nil {
		t.Fatalf("SaveLuckyDraw(): %v", err)
	}
	if draw.Description != "Clear draw terms" {
		t.Fatalf("draw description = %q", draw.Description)
	}
	result, err := store.PlayLuckyDraw(ctx, user.ID, draw.ID, "description-draw", fixedActivityRandom{}, now)
	if err != nil {
		t.Fatalf("PlayLuckyDraw(): %v", err)
	}
	if !strings.Contains(result.ConfigurationSnapshot, `"description":"Clear draw terms"`) {
		t.Fatalf("draw snapshot omits description: %s", result.ConfigurationSnapshot)
	}
}

func TestStoredExtensionAppliesToExactNextActivationAndDelaysQueue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 31_400)
	base := time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)
	if _, err := store.AdjustBalance(ctx, user.ID, 10_000, "extension-seed", "seed", base); err != nil {
		t.Fatal(err)
	}
	combo := saveTestCombo(t, store, "Extension term", 100, 1)
	first, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID, IdempotencyKey: "extension-first"}, base)
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID, IdempotencyKey: "extension-second"}, base)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.DB().ExecContext(ctx, `UPDATE purchases SET status='expired' WHERE id=?`, first.ID); err != nil {
		t.Fatal(err)
	}
	draw, err := store.SaveLuckyDraw(ctx, activity.LuckyDrawInput{Name: "Extension", Enabled: true, Prizes: []activity.PrizeInput{{
		Name: "Three days", Weight: 1, Reward: activity.Reward{Kind: activity.RewardSubscriptionExtension, ExtensionDays: 3},
	}}}, base)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.PlayLuckyDraw(ctx, user.ID, draw.ID, "extension-draw", fixedActivityRandom{}, base.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	third, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID, IdempotencyKey: "extension-third"}, base.Add(2*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	var consumedBefore int
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM activity_extension_credits WHERE consumed_at IS NOT NULL`).Scan(&consumedBefore); err != nil || consumedBefore != 0 {
		t.Fatalf("credits consumed by later queued purchase = %d, %v", consumedBefore, err)
	}
	if err := store.EnqueueDueEntitlementTransitions(ctx, second.ValidFrom); err != nil {
		t.Fatalf("EnqueueDueEntitlementTransitions(): %v", err)
	}
	activated, err := store.PurchaseByID(ctx, second.ID)
	if err != nil {
		t.Fatal(err)
	}
	delayed, err := store.PurchaseByID(ctx, third.ID)
	if err != nil {
		t.Fatal(err)
	}
	if activated.Status != "activating" || !activated.ValidUntil.Equal(second.ValidUntil.AddDate(0, 0, 3)) {
		t.Fatalf("activated term = status %q until %s, want activating/%s", activated.Status, activated.ValidUntil, second.ValidUntil.AddDate(0, 0, 3))
	}
	if !delayed.ValidFrom.Equal(third.ValidFrom.AddDate(0, 0, 3)) || !delayed.ValidUntil.Equal(third.ValidUntil.AddDate(0, 0, 3)) {
		t.Fatalf("delayed term = %s..%s, want %s..%s", delayed.ValidFrom, delayed.ValidUntil, third.ValidFrom.AddDate(0, 0, 3), third.ValidUntil.AddDate(0, 0, 3))
	}
	var consumedBy string
	if err := store.DB().QueryRowContext(ctx, `SELECT consumed_by_purchase_id FROM activity_extension_credits`).Scan(&consumedBy); err != nil || consumedBy != second.ID {
		t.Fatalf("extension consumed by %q, %v, want %q", consumedBy, err, second.ID)
	}
}
