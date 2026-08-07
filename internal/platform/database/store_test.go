package database

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestReserveUsernameIsImmutableAndRetryable(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 20001)
	if err := store.AdvanceToMembership(ctx, user.ID); err != nil {
		t.Fatalf("AdvanceToMembership(): %v", err)
	}
	if _, err := store.UpdateMembership(ctx, user.ID, true, true); err != nil {
		t.Fatalf("UpdateMembership(): %v", err)
	}
	if err := store.ReserveUsername(ctx, user.ID, "river"); err != nil {
		t.Fatalf("ReserveUsername(first): %v", err)
	}
	if err := store.ReserveUsername(ctx, user.ID, "river"); err != nil {
		t.Fatalf("ReserveUsername(retry): %v", err)
	}
	if err := store.ReserveUsername(ctx, user.ID, "meadow"); !errors.Is(err, ErrConflict) {
		t.Fatalf("ReserveUsername(rename) error = %v, want ErrConflict", err)
	}

	updated, err := store.UserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("UserByID(): %v", err)
	}
	if updated.Username == nil || *updated.Username != "river" {
		t.Fatalf("username = %v, want river", updated.Username)
	}
}

func TestClaimPurchaseTrafficResetIsAtMostOnce(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 20002)
	combo := saveTestCombo(t, store, "reset-once", 100, 30)
	if _, err := store.AdjustBalance(ctx, user.ID, 100, "reset-seed", "test credit", time.Now()); err != nil {
		t.Fatalf("AdjustBalance(): %v", err)
	}
	purchase, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID}, time.Now())
	if err != nil {
		t.Fatalf("CreatePurchase(): %v", err)
	}
	claimed, err := store.ClaimPurchaseTrafficReset(ctx, purchase.ID, time.Now())
	if err != nil || !claimed {
		t.Fatalf("ClaimPurchaseTrafficReset(first) = %t, %v; want true, nil", claimed, err)
	}
	claimed, err = store.ClaimPurchaseTrafficReset(ctx, purchase.ID, time.Now().Add(time.Second))
	if err != nil || claimed {
		t.Fatalf("ClaimPurchaseTrafficReset(retry) = %t, %v; want false, nil", claimed, err)
	}
}

func TestDueRenewalEnqueuesExactlyOneActivation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 20003)
	combo := saveTestCombo(t, store, "single-activation", 100, 30)
	if _, err := store.AdjustBalance(ctx, user.ID, 200, "activation-seed", "test credit", time.Now()); err != nil {
		t.Fatalf("AdjustBalance(): %v", err)
	}
	start := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	first, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID}, start)
	if err != nil {
		t.Fatalf("CreatePurchase(first): %v", err)
	}
	renewal, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID}, start.Add(time.Hour))
	if err != nil {
		t.Fatalf("CreatePurchase(renewal): %v", err)
	}
	if err := store.EnqueueDueEntitlementTransitions(ctx, first.ValidUntil); err != nil {
		t.Fatalf("EnqueueDueEntitlementTransitions(first): %v", err)
	}
	if err := store.EnqueueDueEntitlementTransitions(ctx, first.ValidUntil.Add(time.Second)); err != nil {
		t.Fatalf("EnqueueDueEntitlementTransitions(retry): %v", err)
	}
	var count int
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM outbox_jobs WHERE kind='remna_apply_entitlement' AND aggregate_id=?`, renewal.ID).Scan(&count); err != nil {
		t.Fatalf("count renewal activation jobs: %v", err)
	}
	if count != 1 {
		t.Fatalf("renewal activation job count = %d, want 1", count)
	}
}

func TestRecoverOutboxReleasesAbandonedLease(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	if err := store.EnqueueOutbox(ctx, "remna_sync_user", "user-1", `{"userId":"user-1"}`, now); err != nil {
		t.Fatalf("EnqueueOutbox(): %v", err)
	}
	claimed, err := store.ClaimOutboxJob(ctx, now)
	if err != nil || claimed == nil || claimed.Status != "processing" {
		t.Fatalf("ClaimOutboxJob(first) = %+v, %v", claimed, err)
	}
	if err := store.RecoverOutbox(ctx, now.Add(time.Second), now.Add(time.Second)); err != nil {
		t.Fatalf("RecoverOutbox(): %v", err)
	}
	reclaimed, err := store.ClaimOutboxJob(ctx, now.Add(time.Second))
	if err != nil || reclaimed == nil || reclaimed.ID != claimed.ID || reclaimed.Attempts != 2 {
		t.Fatalf("ClaimOutboxJob(recovered) = %+v, %v", reclaimed, err)
	}
}
