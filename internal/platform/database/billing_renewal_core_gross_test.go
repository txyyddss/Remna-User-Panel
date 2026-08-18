package database

import (
	"context"
	"testing"
	"time"
)

func TestRenewPersistsCoreGrossWithoutAddons(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 50_004)
	addon := saveTestSquad(t, store, "renewal-core-addon", 300, true)
	combo := saveTestCombo(t, store, "renewal-core", 1_000, 30)
	now := time.Date(2026, 8, 18, 11, 0, 0, 0, time.UTC)
	if _, err := store.AdjustBalance(ctx, user.ID, 5_000, "renewal-core-seed", "seed", now); err != nil {
		t.Fatal(err)
	}
	source, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID,
		AddonSquadIDs: []string{addon.ID}, IdempotencyKey: "renewal-core-source"}, now)
	if err != nil {
		t.Fatal(err)
	}
	batch, err := store.Renew(ctx, RenewalInput{UserID: user.ID, PurchaseID: source.ID,
		IdempotencyKey: "renewal-core-batch", TermCount: 2}, now.Add(time.Hour))
	if err != nil {
		t.Fatalf("Renew(): %v", err)
	}
	if len(batch.Purchases) != 2 {
		t.Fatalf("renewal purchase count = %d, want 2", len(batch.Purchases))
	}
	for index, purchase := range batch.Purchases {
		if purchase.CoreGrossTXBMinor != combo.PriceTXBMinor || purchase.GrossPriceTXBMinor != 1_300 {
			t.Fatalf("renewal[%d] pricing = core:%d gross:%d", index, purchase.CoreGrossTXBMinor, purchase.GrossPriceTXBMinor)
		}
	}
}
