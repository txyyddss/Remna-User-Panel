package database

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestAutomaticRenewalDefaultsToggleAndSuccessorIdempotency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 48_001)
	combo := saveTestCombo(t, store, "automatic", 100, 30)
	now := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	if _, err := store.AdjustBalance(ctx, user.ID, 500, "automatic-seed", "seed", now); err != nil {
		t.Fatal(err)
	}
	first, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID, IdempotencyKey: "automatic-first"}, now)
	if err != nil || !first.AutoRenewEnabled {
		t.Fatalf("CreatePurchase(immediate) = (%+v, %v), want auto renewal enabled", first, err)
	}
	if err := store.SetAutoRenewal(ctx, user.ID, first.ID, false, now); err != nil {
		t.Fatalf("SetAutoRenewal(off): %v", err)
	}
	if err := store.SetAutoRenewal(ctx, user.ID, first.ID, true, now); err != nil {
		t.Fatalf("SetAutoRenewal(on): %v", err)
	}
	successor, err := store.CommitAutoRenewal(ctx, first.ID, first.ValidUntil)
	if err != nil || successor.Status != "queued" || !successor.AutoRenewEnabled {
		t.Fatalf("CommitAutoRenewal() = (%+v, %v)", successor, err)
	}
	replayed, err := store.CommitAutoRenewal(ctx, first.ID, first.ValidUntil.Add(time.Second))
	if err != nil || replayed.ID != successor.ID {
		t.Fatalf("CommitAutoRenewal(replay) = (%+v, %v), want %s", replayed, err, successor.ID)
	}
	if due, err := store.DueAutoRenewals(ctx, first.ValidUntil.Add(time.Second)); err != nil || len(due) != 0 {
		t.Fatalf("DueAutoRenewals(after successor) = (%+v, %v), want none", due, err)
	}
	if err := store.SetAutoRenewal(ctx, user.ID, first.ID, false, first.ValidUntil); err != nil {
		t.Fatalf("SetAutoRenewal(disable chain): %v", err)
	}
	if queuedSuccessor, err := store.PurchaseByID(ctx, successor.ID); err != nil || queuedSuccessor.AutoRenewEnabled {
		t.Fatalf("PurchaseByID(disabled successor) = (%+v, %v)", queuedSuccessor, err)
	}
	var successors, debits int
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM purchases WHERE auto_renew_source_purchase_id=?`, first.ID).Scan(&successors); err != nil {
		t.Fatal(err)
	}
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM ledger_entries WHERE kind='automatic_renewal'`).Scan(&debits); err != nil {
		t.Fatal(err)
	}
	if successors != 1 || debits != 1 {
		t.Fatalf("automatic renewal effects = successors:%d debits:%d, want 1/1", successors, debits)
	}

	other := createTestUser(t, store, 48_002)
	if _, err := store.AdjustBalance(ctx, other.ID, 500, "queued-seed", "seed", now); err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreatePurchase(ctx, PurchaseInput{UserID: other.ID, ComboID: combo.ID, IdempotencyKey: "queued-first"}, now); err != nil {
		t.Fatal(err)
	}
	queued, err := store.CreatePurchase(ctx, PurchaseInput{UserID: other.ID, ComboID: combo.ID, IdempotencyKey: "queued-next"}, now.Add(time.Hour))
	if err != nil || queued.AutoRenewEnabled {
		t.Fatalf("CreatePurchase(queued) = (%+v, %v), want auto renewal disabled", queued, err)
	}
}

func TestAutomaticRenewalFailureDisablesWithoutCharging(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 48_003)
	combo := saveTestCombo(t, store, "automatic-failure", 100, 30)
	now := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	if _, err := store.AdjustBalance(ctx, user.ID, 100, "failure-seed", "seed", now); err != nil {
		t.Fatal(err)
	}
	source, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID, IdempotencyKey: "failure-source"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CommitAutoRenewal(ctx, source.ID, source.ValidUntil); !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("CommitAutoRenewal(insufficient) = %v, want ErrInsufficientBalance", err)
	}
	if err := store.MarkAutoRenewalFailed(ctx, source.ID, AutoRenewalReasonInsufficientBalance, source.ValidUntil); err != nil {
		t.Fatal(err)
	}
	updated, err := store.PurchaseByID(ctx, source.ID)
	if err != nil || updated.AutoRenewEnabled {
		t.Fatalf("PurchaseByID(failed) = (%+v, %v), want disabled", updated, err)
	}
	failure, err := store.AutoRenewalFailure(ctx, user.ID)
	if err != nil || failure == nil || failure.Reason != AutoRenewalReasonInsufficientBalance {
		t.Fatalf("AutoRenewalFailure() = (%+v, %v)", failure, err)
	}
	if balance, err := store.Balance(ctx, user.ID); err != nil || balance.Minor != "0" {
		t.Fatalf("Balance(after failed renewal) = (%+v, %v), want 0", balance, err)
	}
	if err := store.EnqueueDueEntitlementTransitions(ctx, source.ValidUntil); err != nil {
		t.Fatalf("EnqueueDueEntitlementTransitions(): %v", err)
	}
	var rollovers int
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM purchase_rollovers WHERE purchase_id=?`, source.ID).Scan(&rollovers); err != nil {
		t.Fatal(err)
	}
	if rollovers != 1 {
		t.Fatalf("rollovers after failure = %d, want 1", rollovers)
	}
}
