package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestMigrationClearsLegacySubscriptionCache(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 20000)
	if _, err := store.DB().ExecContext(ctx, `UPDATE users SET remna_subscription_url=? WHERE id=?`, "https://subscription.example/bearer", user.ID); err != nil {
		t.Fatalf("seed legacy subscription URL: %v", err)
	}
	if _, err := store.DB().ExecContext(ctx, `DELETE FROM schema_migrations WHERE version=?`, "010_clear_subscription_cache.sql"); err != nil {
		t.Fatalf("reset scrub migration: %v", err)
	}
	if err := migrate(ctx, store.DB()); err != nil {
		t.Fatalf("migrate(): %v", err)
	}
	var cached sql.NullString
	if err := store.DB().QueryRowContext(ctx, `SELECT remna_subscription_url FROM users WHERE id=?`, user.ID).Scan(&cached); err != nil {
		t.Fatalf("read scrubbed subscription URL: %v", err)
	}
	if cached.Valid {
		t.Fatalf("legacy subscription URL still persisted: %q", cached.String)
	}
}

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

func TestBeginRemnawaveRecoveryIsIdempotentAcrossConcurrentAuthentication(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 20005)
	now := time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)
	if _, err := store.DB().ExecContext(ctx, `UPDATE users SET username='river',onboarding_state='complete',group_joined=1,channel_joined=1,
		policy_accepted_at=?,remna_user_id='remote-user',remna_subscription_url='https://subscription.example/token' WHERE id=?`, stamp(now), user.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.AdjustBalance(ctx, user.ID, 500, "recovery-balance", "preserved", now); err != nil {
		t.Fatal(err)
	}

	start := make(chan struct{})
	results := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			_, err := store.BeginRemnawaveRecovery(ctx, user.ID, "remnawave_user_missing", now.Add(time.Second))
			results <- err
		}()
	}
	close(start)
	for range 2 {
		if err := <-results; err != nil {
			t.Fatalf("BeginRemnawaveRecovery() error = %v", err)
		}
	}
	recovered, err := store.UserByID(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if recovered.Username == nil || *recovered.Username != "river" || recovered.OnboardingState != "membership" ||
		recovered.RecoveryReason != "remnawave_user_missing" || recovered.RemnaUserID != nil || recovered.PolicyAcceptedAt != nil {
		t.Fatalf("recovered user = %+v", recovered)
	}
	if balance, err := store.Balance(ctx, user.ID); err != nil || balance.Minor != "500" {
		t.Fatalf("preserved balance = (%+v, %v)", balance, err)
	}
	if _, err := store.BeginRemnawaveRecovery(ctx, user.ID, "different_reason", now.Add(2*time.Second)); !errors.Is(err, ErrConflict) {
		t.Fatalf("different recovery reason error = %v, want conflict", err)
	}
}

func TestPurchaseTrafficResetPhasesAreDurableAndIdempotent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 20002)
	combo := saveTestCombo(t, store, "reset-once", 100, 30)
	if _, err := store.AdjustBalance(ctx, user.ID, 100, "reset-seed", "test credit", time.Now()); err != nil {
		t.Fatalf("AdjustBalance(): %v", err)
	}
	purchase, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID, IdempotencyKey: "reset-phase"}, time.Now())
	if err != nil {
		t.Fatalf("CreatePurchase(): %v", err)
	}
	phase, err := store.PurchaseTrafficResetPhase(ctx, purchase.ID)
	if err != nil || phase != "pending" {
		t.Fatalf("PurchaseTrafficResetPhase(initial) = %q, %v", phase, err)
	}
	if err := store.AdvancePurchaseTrafficReset(ctx, purchase.ID, "pending", "quiesced", time.Now()); err != nil {
		t.Fatalf("AdvancePurchaseTrafficReset(quiesced): %v", err)
	}
	if err := store.AdvancePurchaseTrafficReset(ctx, purchase.ID, "pending", "quiesced", time.Now().Add(time.Second)); err != nil {
		t.Fatalf("AdvancePurchaseTrafficReset(replay): %v", err)
	}
	if err := store.AdvancePurchaseTrafficReset(ctx, purchase.ID, "quiesced", "reset", time.Now().Add(2*time.Second)); err != nil {
		t.Fatalf("AdvancePurchaseTrafficReset(reset): %v", err)
	}
	phase, err = store.PurchaseTrafficResetPhase(ctx, purchase.ID)
	if err != nil || phase != "reset" {
		t.Fatalf("PurchaseTrafficResetPhase(final) = %q, %v", phase, err)
	}
	if err := store.AdvancePurchaseTrafficReset(ctx, purchase.ID, "pending", "reset", time.Now()); !errors.Is(err, ErrConflict) {
		t.Fatalf("AdvancePurchaseTrafficReset(invalid) = %v, want ErrConflict", err)
	}
}

func TestCatalogUpdatesCannotCreateUnknownRecords(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	comboInput := ComboInput{Name: "Known", PriceTXBMinor: 100, ValidityDays: 30, TrafficLimitBytes: 1024, ResetStrategy: "MONTH", Active: true}
	combo, err := store.SaveCombo(ctx, comboInput)
	if err != nil {
		t.Fatalf("SaveCombo(create): %v", err)
	}
	comboInput.ID = combo.ID
	comboInput.Name = "Updated"
	if updated, err := store.SaveCombo(ctx, comboInput); err != nil || updated.Name != "Updated" {
		t.Fatalf("SaveCombo(update) = (%+v, %v)", updated, err)
	}
	comboInput.ID = "client-selected-missing-id"
	if _, err := store.SaveCombo(ctx, comboInput); !errors.Is(err, ErrNotFound) {
		t.Fatalf("SaveCombo(unknown update) = %v, want ErrNotFound", err)
	}

	if _, err := store.SaveSquadProduct(ctx, SquadProductInput{ID: "invented", RemnaSquadUUID: "invented", Name: "Invented"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("SaveSquadProduct(unimported) = %v, want ErrNotFound", err)
	}
	if err := store.RefreshImportedSquads(ctx, []ImportedSquad{{UUID: "remote-1", Name: "Remote"}}); err != nil {
		t.Fatalf("RefreshImportedSquads(): %v", err)
	}
	if _, err := store.SquadProductByRemnaUUID(ctx, "remote-1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("SquadProductByRemnaUUID(unedited upstream) = %v, want ErrNotFound", err)
	}
	productInput := SquadProductInput{ID: "remote-1", RemnaSquadUUID: "remote-1", Name: "Merchandised", Description: "Local copy", PriceTXBMinor: 25, Visible: true, UpstreamPresent: true}
	product, err := store.SaveSquadProduct(ctx, productInput)
	if err != nil {
		t.Fatalf("SaveSquadProduct(imported create): %v", err)
	}
	if updated, err := store.SaveSquadProduct(ctx, productInput); err != nil || updated.Name != "Merchandised" || !updated.UpstreamPresent {
		t.Fatalf("SaveSquadProduct(imported update) = (%+v, %v)", updated, err)
	}
	comboInput.ID = ""
	comboInput.SquadProductIDs = []string{product.ID}
	squadCombo, err := store.SaveCombo(ctx, comboInput)
	if err != nil {
		t.Fatalf("SaveCombo(imported squad): %v", err)
	}
	if err := store.RefreshImportedSquads(ctx, nil); err != nil {
		t.Fatalf("RefreshImportedSquads(empty): %v", err)
	}
	if combos, err := store.ListCombos(ctx, true); err != nil || len(combos) != 2 {
		t.Fatalf("ListCombos(after compatibility refresh) = (%+v, %v)", combos, err)
	}
	if loaded, err := store.ComboByID(ctx, squadCombo.ID, true); err != nil || len(loaded.IncludedSquads) != 1 {
		t.Fatalf("ComboByID(sparse squad identity) = (%+v, %v)", loaded, err)
	}

	comboInput.ID = ""
	comboInput.SquadProductIDs = []string{"sparse-upstream-squad"}
	if _, err := store.SaveCombo(ctx, comboInput); err != nil {
		t.Fatalf("SaveCombo(sparse upstream identity): %v", err)
	}
}

func TestDueRenewalWaitsForRolloverThenEnqueuesExactlyOneActivation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 20003)
	combo := saveTestCombo(t, store, "single-activation", 100, 30)
	if _, err := store.AdjustBalance(ctx, user.ID, 200, "activation-seed", "test credit", time.Now()); err != nil {
		t.Fatalf("AdjustBalance(): %v", err)
	}
	start := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	first, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID, IdempotencyKey: "activation-first"}, start)
	if err != nil {
		t.Fatalf("CreatePurchase(first): %v", err)
	}
	renewal, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID, IdempotencyKey: "activation-renewal"}, start.Add(time.Hour))
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
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM outbox_jobs WHERE kind='rollover_finalize' AND payload=?`, `{"purchaseId":"`+first.ID+`"}`).Scan(&count); err != nil {
		t.Fatalf("count rollover jobs: %v", err)
	}
	if count != 1 {
		t.Fatalf("rollover job count = %d, want 1", count)
	}
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM outbox_jobs WHERE kind='remna_apply_entitlement' AND payload=?`, `{"purchaseId":"`+renewal.ID+`"}`).Scan(&count); err != nil {
		t.Fatalf("count renewal activation jobs: %v", err)
	}
	if count != 0 {
		t.Fatalf("renewal activated before rollover: count = %d", count)
	}
	if err := store.MarkRolloverProcessing(ctx, first.ID, first.ValidUntil); err != nil {
		t.Fatalf("MarkRolloverProcessing(): %v", err)
	}
	if _, err := store.FinalizeRollover(ctx, first.ID, 1000, 500, "", first.ValidUntil); err != nil {
		t.Fatalf("FinalizeRollover(): %v", err)
	}
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM outbox_jobs WHERE kind='remna_apply_entitlement' AND payload=?`, `{"purchaseId":"`+renewal.ID+`"}`).Scan(&count); err != nil {
		t.Fatalf("count renewal activation jobs after rollover: %v", err)
	}
	if count != 1 {
		t.Fatalf("renewal activation job count after rollover = %d, want 1", count)
	}
}

func TestRecoverOutboxReleasesAbandonedLease(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	if err := store.EnqueueOutbox(ctx, "remna_sync_user", `{"userId":"user-1"}`, now); err != nil {
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
