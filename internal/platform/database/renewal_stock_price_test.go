package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/coupons"
)

func TestStockBlockedRenewalStillQuotesCurrentDiscountedPrice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 51_002)
	addon := saveTestSquad(t, store, "stock-repricing", 300, true)
	combo := saveTestCombo(t, store, "stock-repricing", 1_000, 30)
	now := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC)
	if _, err := store.AdjustBalance(ctx, user.ID, 5_000, "stock-repricing-seed", "seed", now); err != nil {
		t.Fatal(err)
	}
	coupon, err := store.SaveCoupon(ctx, coupons.CouponInput{Code: "REPRICE", Name: "Recurring discount",
		Kind: coupons.KindPurchaseRecurring, DiscountMode: coupons.DiscountPercent, ValueMinorOrBPS: 1_000, Active: true}, now)
	if err != nil {
		t.Fatal(err)
	}
	redemption, err := store.RedeemCoupon(ctx, user.ID, coupon.Code, "stock-repricing-coupon", now)
	if err != nil || redemption.Grant == nil {
		t.Fatalf("RedeemCoupon() = (%+v, %v)", redemption, err)
	}
	source, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID,
		AddonSquadIDs: []string{addon.ID}, CouponGrantID: redemption.Grant.ID, IdempotencyKey: "stock-repricing-source"}, now)
	if err != nil {
		t.Fatal(err)
	}
	zero := 0
	if _, err := store.SaveSquadProduct(ctx, SquadProductInput{RemnaSquadUUID: addon.ID,
		Name: addon.Name, UpstreamPresent: true, Visible: true, PriceTXBMinor: 700, StockLimit: &zero}); err != nil {
		t.Fatal(err)
	}
	plan, err := store.AutoRenewalPlan(ctx, user.ID, source.ID, now)
	if err != nil || plan.IneligibleReason != AutoRenewalReasonPaidAddonUnavailable || plan.GrossMinor != 1_700 || plan.DiscountMinor != 170 || plan.NetMinor != 1_530 {
		t.Fatalf("AutoRenewalPlan(stock blocked) = (%+v, %v), want current 1700 - 170 = 1530", plan, err)
	}
	if _, err := store.CommitAutoRenewal(ctx, source.ID, source.ValidUntil); !errors.Is(err, ErrConflict) {
		t.Fatalf("CommitAutoRenewal(stock blocked) = %v, want ErrConflict", err)
	}
	if balance, err := store.Balance(ctx, user.ID); err != nil || balance.MinorInt64() != 3_830 {
		t.Fatalf("Balance(stock blocked) = (%+v, %v), want no renewal debit", balance, err)
	}
	if _, err := store.SaveSquadProduct(ctx, SquadProductInput{RemnaSquadUUID: addon.ID,
		Name: addon.Name, UpstreamPresent: true, Visible: true, PriceTXBMinor: 700}); err != nil {
		t.Fatal(err)
	}
	renewed, err := store.CommitAutoRenewal(ctx, source.ID, source.ValidUntil)
	if err != nil || renewed.PriceTXBMinor != 1_530 || renewed.CouponDiscountTXBMinor != 170 || !renewed.RecurringDiscountAttached {
		t.Fatalf("CommitAutoRenewal(stock restored) = (%+v, %v)", renewed, err)
	}
}
