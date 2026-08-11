package database

import (
	"context"
	"math"
	"testing"
	"time"
)

func TestFinalizeRolloverFormulaThresholdCapAndIdempotency(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		limit      int64
		used       int64
		threshold  int
		maximum    int64
		wantCredit int64
	}{
		{name: "strictly above threshold capped", limit: 1000, used: 400, threshold: 5000, maximum: 400, wantCredit: 400},
		{name: "equal threshold is zero", limit: 1000, used: 500, threshold: 5000, maximum: 1000, wantCredit: 0},
		{name: "zero limit is zero", limit: 0, used: 0, threshold: 0, maximum: 1000, wantCredit: 0},
	}
	for index, test := range tests {
		index, test := index, test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			store := newTestStore(t)
			user := createTestUser(t, store, int64(26100+index))
			combo, err := store.SaveCombo(ctx, ComboInput{
				Name: "rollover", PriceTXBMinor: 1000, ValidityDays: 1, TrafficLimitBytes: 1000,
				ResetStrategy: "MONTH", Active: true, RolloverMinRemainingBPS: test.threshold, RolloverMaxTXBMinor: test.maximum,
			})
			if err != nil {
				t.Fatalf("SaveCombo(): %v", err)
			}
			if _, err := store.AdjustBalance(ctx, user.ID, 1000, "seed", "test", time.Now()); err != nil {
				t.Fatalf("AdjustBalance(): %v", err)
			}
			purchase, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID, IdempotencyKey: "rollover-" + test.name}, time.Now())
			if err != nil {
				t.Fatalf("CreatePurchase(): %v", err)
			}
			if err := store.EnqueueDueEntitlementTransitions(ctx, purchase.ValidUntil); err != nil {
				t.Fatalf("EnqueueDueEntitlementTransitions(): %v", err)
			}
			if err := store.MarkRolloverProcessing(ctx, purchase.ID, purchase.ValidUntil); err != nil {
				t.Fatalf("MarkRolloverProcessing(): %v", err)
			}
			settled, err := store.FinalizeRollover(ctx, purchase.ID, test.limit, test.used, "", purchase.ValidUntil)
			if err != nil {
				t.Fatalf("FinalizeRollover(): %v", err)
			}
			if settled.CreditedTXBMinor != test.wantCredit {
				t.Fatalf("credited = %d, want %d", settled.CreditedTXBMinor, test.wantCredit)
			}
			if _, err := store.FinalizeRollover(ctx, purchase.ID, test.limit, test.used, "", purchase.ValidUntil.Add(time.Second)); err != nil {
				t.Fatalf("FinalizeRollover(retry): %v", err)
			}
			var balance, ledgerCount int64
			if err := store.DB().QueryRowContext(ctx, `SELECT txb_minor FROM balances WHERE user_id=?`, user.ID).Scan(&balance); err != nil {
				t.Fatalf("read balance: %v", err)
			}
			if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM ledger_entries WHERE kind='rollover_credit' AND reference_id=?`, purchase.ID).Scan(&ledgerCount); err != nil {
				t.Fatalf("count ledger: %v", err)
			}
			if balance != test.wantCredit {
				t.Fatalf("balance = %d, want %d", balance, test.wantCredit)
			}
			wantLedger := int64(0)
			if test.wantCredit > 0 {
				wantLedger = 1
			}
			if ledgerCount != wantLedger {
				t.Fatalf("rollover ledger count = %d, want %d", ledgerCount, wantLedger)
			}
		})
	}
}

func TestFinalizeRolloverBalanceOverflowRollsBack(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 26_200)
	combo, err := store.SaveCombo(ctx, ComboInput{
		Name: "rollover overflow", PriceTXBMinor: 1000, ValidityDays: 1, TrafficLimitBytes: 1000,
		ResetStrategy: "MONTH", Active: true, RolloverMaxTXBMinor: 1000,
	})
	if err != nil {
		t.Fatalf("SaveCombo(): %v", err)
	}
	now := time.Now().UTC()
	if _, err := store.AdjustBalance(ctx, user.ID, 1000, "overflow-seed", "test", now); err != nil {
		t.Fatalf("AdjustBalance(seed): %v", err)
	}
	purchase, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID, IdempotencyKey: "rollover-overflow"}, now)
	if err != nil {
		t.Fatalf("CreatePurchase(): %v", err)
	}
	if _, err := store.AdjustBalance(ctx, user.ID, math.MaxInt64, "overflow-max", "test", now); err != nil {
		t.Fatalf("AdjustBalance(max): %v", err)
	}
	if err := store.EnqueueDueEntitlementTransitions(ctx, purchase.ValidUntil); err != nil {
		t.Fatalf("EnqueueDueEntitlementTransitions(): %v", err)
	}
	if err := store.MarkRolloverProcessing(ctx, purchase.ID, purchase.ValidUntil); err != nil {
		t.Fatalf("MarkRolloverProcessing(): %v", err)
	}
	if _, err := store.FinalizeRollover(ctx, purchase.ID, 1000, 0, "", purchase.ValidUntil); err == nil {
		t.Fatal("FinalizeRollover() unexpectedly allowed balance overflow")
	}
	rollover, err := store.RolloverByPurchase(ctx, purchase.ID)
	if err != nil || rollover.Status != "processing" || rollover.CreditedTXBMinor != 0 {
		t.Fatalf("rollover after rollback = (%+v, %v)", rollover, err)
	}
	stored, err := store.PurchaseByID(ctx, purchase.ID)
	if err != nil || stored.Status == "expired" {
		t.Fatalf("purchase after rollback = (%+v, %v)", stored, err)
	}
}
