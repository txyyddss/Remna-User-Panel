package database

import (
	"context"
	"testing"
	"time"
)

func TestProvisioningConflictRefundsCurrentAndQueuedPurchasesOnce(t *testing.T) {
	t.Parallel()
	ctx, store := context.Background(), newTestStore(t)
	user := createTestUser(t, store, 55_100)
	combo := saveTestCombo(t, store, "conflicted", 100, 30)
	now := time.Date(2026, 9, 24, 0, 0, 0, 0, time.UTC)
	if _, err := store.DB().ExecContext(ctx, `UPDATE users SET username='reserved',onboarding_state='complete',
		policy_accepted_at=?,accepted_agreement_revision=1,group_joined=1,channel_joined=1 WHERE id=?`, stamp(now), user.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.AdjustBalance(ctx, user.ID, 500, "conflict-seed", "seed", now); err != nil {
		t.Fatal(err)
	}
	current, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID, IdempotencyKey: "current"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.MarkPurchaseSyncResult(ctx, current.ID, true, now); err != nil {
		t.Fatal(err)
	}
	queued, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID, IdempotencyKey: "queued"}, now.Add(time.Minute))
	if err != nil || queued.Status != "queued" {
		t.Fatalf("queued purchase = %+v, %v", queued, err)
	}
	// A later paid add-on increases the purchase's charged total.
	if _, err := store.AdjustBalance(ctx, user.ID, -25, "addon-seed", "add-on", now); err != nil {
		t.Fatal(err)
	}
	if _, err := store.DB().ExecContext(ctx, `UPDATE purchases SET charged_txb_minor=charged_txb_minor+25 WHERE id=?`, current.ID); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := store.ResolveProvisioningConflict(ctx, user.ID, "reserved", now.Add(2*time.Minute)); err != nil {
			t.Fatal(err)
		}
	}
	recovered, err := store.UserByID(ctx, user.ID)
	if err != nil || recovered.Username != nil || recovered.RemnaUserID != nil || recovered.OnboardingState != "username" ||
		recovered.RecoveryReason != "remnawave_username_conflict" || recovered.PolicyAcceptedAt != nil ||
		!recovered.GroupJoined || !recovered.ChannelJoined {
		t.Fatalf("recovered user = %+v, err %v", recovered, err)
	}
	if balance, err := store.Balance(ctx, user.ID); err != nil || balance.MinorInt64() != 500 {
		t.Fatalf("refunded balance = %+v, err %v", balance, err)
	}
	var cancelled, refunds, notices int
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM purchases WHERE user_id=? AND status='cancelled'`, user.ID).Scan(&cancelled); err != nil {
		t.Fatal(err)
	}
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM ledger_entries WHERE user_id=? AND kind='remna_identity_conflict_refund'`, user.ID).Scan(&refunds); err != nil {
		t.Fatal(err)
	}
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM user_notification_events WHERE user_id=? AND kind='remnawave_username_conflict_refund'`, user.ID).Scan(&notices); err != nil {
		t.Fatal(err)
	}
	if cancelled != 2 || refunds != 2 || notices != 1 {
		t.Fatalf("cancelled/refunds/notices = %d/%d/%d", cancelled, refunds, notices)
	}
}

func TestUnlinkedPaidBacklogQueuesOneRepair(t *testing.T) {
	t.Parallel()
	ctx, store := context.Background(), newTestStore(t)
	user := createTestUser(t, store, 55_101)
	combo := saveTestCombo(t, store, "repair", 100, 30)
	now := time.Date(2026, 9, 24, 0, 0, 0, 0, time.UTC)
	if _, err := store.DB().ExecContext(ctx, `UPDATE users SET username='repair_name',onboarding_state='complete',
		policy_accepted_at=? WHERE id=?`, stamp(now), user.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.AdjustBalance(ctx, user.ID, 100, "repair-seed", "seed", now); err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID, IdempotencyKey: "repair"}, now); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := store.EnqueueUnlinkedPaidUsers(ctx, now); err != nil {
			t.Fatal(err)
		}
	}
	var jobs int
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM outbox_jobs WHERE kind='remna_sync_user'
		AND json_extract(payload,'$.userId')=? AND status='pending'`, user.ID).Scan(&jobs); err != nil || jobs != 1 {
		t.Fatalf("repair jobs = %d, err %v", jobs, err)
	}
}
