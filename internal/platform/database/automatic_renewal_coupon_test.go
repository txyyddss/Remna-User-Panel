package database

import (
	"context"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/coupons"
)

func TestAutomaticRenewalAttachedCouponUsesLatestDiscountOnly(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 48_004)
	combo := saveTestCombo(t, store, "automatic-coupon", 100, 30)
	now := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	if _, err := store.AdjustBalance(ctx, user.ID, 500, "coupon-seed", "seed", now); err != nil {
		t.Fatal(err)
	}
	limit := int64(1)
	coupon, err := store.SaveCoupon(ctx, coupons.CouponInput{Code: "AUTORECUR", Name: "Automatic", Kind: coupons.KindPurchaseRecurring,
		DiscountMode: coupons.DiscountFixed, ValueMinorOrBPS: 20, GlobalUseLimit: &limit, Active: true}, now)
	if err != nil {
		t.Fatal(err)
	}
	redeemed, err := store.RedeemCoupon(ctx, user.ID, coupon.Code, "automatic-coupon-redeem", now)
	if err != nil || redeemed.Grant == nil {
		t.Fatalf("RedeemCoupon() = (%+v, %v)", redeemed, err)
	}
	source, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID, CouponGrantID: redeemed.Grant.ID, IdempotencyKey: "automatic-coupon-source"}, now)
	if err != nil || !source.RecurringDiscountAttached {
		t.Fatalf("CreatePurchase() = (%+v, %v), want recurring attachment", source, err)
	}
	if err := store.DiscardCouponGrant(ctx, user.ID, redeemed.Grant.ID, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	past := stamp(now.Add(-time.Hour))
	if _, err := store.DB().ExecContext(ctx, `UPDATE coupon_definitions SET active=0,kind='balance_add',discount_mode='fixed',value_minor_or_bps=30,expires_at=? WHERE id=?`, past, coupon.ID); err != nil {
		t.Fatal(err)
	}
	plan, err := store.AutoRenewalPlan(ctx, user.ID, source.ID, now.Add(time.Hour))
	if err != nil || plan.DiscountMinor != 30 || plan.NetMinor != 70 {
		t.Fatalf("AutoRenewalPlan() = (%+v, %v), want latest attached 30 discount", plan, err)
	}
}
