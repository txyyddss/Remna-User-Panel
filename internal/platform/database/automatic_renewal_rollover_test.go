package database

import (
	"context"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

func recordAutomaticRollover(t *testing.T, store *Store, purchaseID string, now time.Time, used, eligible int64) {
	t.Helper()
	if err := store.EnqueueDueEntitlementTransitions(context.Background(), now); err != nil {
		t.Fatal(err)
	}
	if err := store.MarkRolloverProcessing(context.Background(), purchaseID, now); err != nil {
		t.Fatal(err)
	}
	if _, err := store.RecordRolloverCalculation(context.Background(), purchaseID, model.RolloverUsageSummary{
		AllocatedBytes: 1_000, UsedBytes: used, EligibleUnusedBytes: eligible, AlgorithmVersion: "cadence-v3",
	}, now); err != nil {
		t.Fatal(err)
	}
}

func TestCalculatedRolloverFundsAutomaticRenewalAtomically(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 48_010)
	combo := saveTestCombo(t, store, "atomic-rollover", 100, 1)
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	if _, err := store.AdjustBalance(ctx, user.ID, 160, "atomic-rollover-seed", "seed", now); err != nil {
		t.Fatal(err)
	}
	source, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID, IdempotencyKey: "atomic-rollover-source"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SetAutoRenewal(ctx, user.ID, source.ID, true, now); err != nil {
		t.Fatalf("enable automatic renewal: %v", err)
	}
	recordAutomaticRollover(t, store, source.ID, source.ValidUntil, 500, 500)
	successor, err := store.CommitAutoRenewal(ctx, source.ID, source.ValidUntil)
	if err != nil || successor.Status != "activating" {
		t.Fatalf("CommitAutoRenewal() = (%+v, %v)", successor, err)
	}
	if balance, err := store.Balance(ctx, user.ID); err != nil || balance.MinorInt64() != 10 {
		t.Fatalf("Balance() = (%+v, %v), want 10", balance, err)
	}
	if replay, err := store.CommitAutoRenewal(ctx, source.ID, source.ValidUntil); err != nil || replay.ID != successor.ID {
		t.Fatalf("CommitAutoRenewal(replay) = (%+v, %v)", replay, err)
	}
	var credits, debits int
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM ledger_entries WHERE kind='rollover_credit'`).Scan(&credits); err != nil {
		t.Fatal(err)
	}
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM ledger_entries WHERE kind='automatic_renewal'`).Scan(&debits); err != nil {
		t.Fatal(err)
	}
	if credits != 1 || debits != 1 {
		t.Fatalf("credit/debit ledgers = %d/%d, want 1/1", credits, debits)
	}
}

func TestRolloverEligibilityRequiresAutoRenewalWithoutQueuedCombo(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 48_011)
	combo := saveTestCombo(t, store, "rollover-gate", 100, 1)
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	if _, err := store.AdjustBalance(ctx, user.ID, 400, "rollover-gate-seed", "seed", now); err != nil {
		t.Fatal(err)
	}
	source, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID, IdempotencyKey: "rollover-gate-source"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if eligible, err := store.RolloverEligible(ctx, source.ID); err != nil || eligible {
		t.Fatalf("RolloverEligible(auto off) = (%t, %v)", eligible, err)
	}
	if err := store.SetAutoRenewal(ctx, user.ID, source.ID, true, now); err != nil {
		t.Fatal(err)
	}
	if eligible, err := store.RolloverEligible(ctx, source.ID); err != nil || !eligible {
		t.Fatalf("RolloverEligible(auto on) = (%t, %v)", eligible, err)
	}
	if _, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID, IdempotencyKey: "rollover-gate-queued"}, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if eligible, err := store.RolloverEligible(ctx, source.ID); err != nil || eligible {
		t.Fatalf("RolloverEligible(queued combo) = (%t, %v)", eligible, err)
	}
}
