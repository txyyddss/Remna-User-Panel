package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/coupons"
)

func TestCouponImmediateEffectsAndIdempotency(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 32_000)
	if _, err := store.AdjustBalance(ctx, user.ID, 1_000, "coupon-seed", "seed", time.Now()); err != nil {
		t.Fatalf("AdjustBalance(): %v", err)
	}
	limit := int64(1)
	add, err := store.SaveCoupon(ctx, coupons.CouponInput{Code: "ADD500", Name: "Add", Kind: coupons.KindBalanceAdd, ValueMinorOrBPS: 500,
		PerUserUseLimit: &limit, Active: true}, time.Now())
	if err != nil {
		t.Fatalf("SaveCoupon(add): %v", err)
	}
	first, err := store.RedeemCoupon(ctx, user.ID, add.Code, "add-key", time.Now())
	if err != nil || first.BalanceDeltaMinor != 500 || first.BalanceAfterMinor != 1_500 {
		t.Fatalf("RedeemCoupon(add) = (%+v, %v)", first, err)
	}
	replayed, err := store.RedeemCoupon(ctx, user.ID, add.Code, "add-key", time.Now())
	if err != nil || !replayed.Replayed || replayed.ID != first.ID {
		t.Fatalf("RedeemCoupon(replay) = (%+v, %v)", replayed, err)
	}
	if _, err := store.RedeemCoupon(ctx, user.ID, add.Code, "different-key", time.Now()); !errors.Is(err, ErrConflict) {
		t.Fatalf("RedeemCoupon(limit) = %v, want conflict", err)
	}
	multiply, err := store.SaveCoupon(ctx, coupons.CouponInput{Code: "MULT150", Name: "Multiply", Kind: coupons.KindBalanceMultiply,
		ValueMinorOrBPS: 15_000, Active: true}, time.Now())
	if err != nil {
		t.Fatalf("SaveCoupon(multiply): %v", err)
	}
	multiplied, err := store.RedeemCoupon(ctx, user.ID, multiply.Code, "multiply-key", time.Now())
	if err != nil || multiplied.BalanceDeltaMinor != 750 || multiplied.BalanceAfterMinor != 2_250 {
		t.Fatalf("RedeemCoupon(multiply) = (%+v, %v)", multiplied, err)
	}
}

func TestCouponPurchaseGrantQuoteAndConsumption(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 32_100)
	oneTime, err := store.SaveCoupon(ctx, coupons.CouponInput{Code: "ONCE300", Name: "Once", Kind: coupons.KindPurchaseOnce,
		DiscountMode: coupons.DiscountFixed, ValueMinorOrBPS: 300, EligibleComboIDs: []string{"combo-a"}, Active: true}, time.Now())
	if err != nil {
		t.Fatalf("SaveCoupon(once): %v", err)
	}
	redeemed, err := store.RedeemCoupon(ctx, user.ID, oneTime.Code, "redeem-once", time.Now())
	if err != nil || redeemed.Grant == nil {
		t.Fatalf("RedeemCoupon(once) = (%+v, %v)", redeemed, err)
	}
	input := coupons.PurchaseContext{UserID: user.ID, GrantID: redeemed.Grant.ID, ComboID: "combo-a", GrossPriceMinor: 1_000}
	quote, err := store.QuotePurchaseCoupon(ctx, input, time.Now())
	if err != nil || quote.DiscountMinor != 300 || quote.NetMinor != 700 {
		t.Fatalf("QuotePurchaseCoupon() = (%+v, %v)", quote, err)
	}
	tx, err := store.DB().BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx(): %v", err)
	}
	if _, err := applyPurchaseCouponTx(ctx, tx, input, "purchase-coupon-test", time.Now()); err != nil {
		_ = tx.Rollback()
		t.Fatalf("applyPurchaseCouponTx(): %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit coupon use: %v", err)
	}
	if _, err := store.QuotePurchaseCoupon(ctx, input, time.Now()); !errors.Is(err, ErrConflict) {
		t.Fatalf("QuotePurchaseCoupon(consumed) = %v, want conflict", err)
	}

	capMinor := int64(225)
	recurring, err := store.SaveCoupon(ctx, coupons.CouponInput{Code: "RECUR50", Name: "Recurring", Kind: coupons.KindPurchaseRecurring,
		DiscountMode: coupons.DiscountPercent, ValueMinorOrBPS: 5_000, PercentCapMinor: &capMinor, Active: true}, time.Now())
	if err != nil {
		t.Fatalf("SaveCoupon(recurring): %v", err)
	}
	recurringRedemption, err := store.RedeemCoupon(ctx, user.ID, recurring.Code, "redeem-recurring", time.Now())
	if err != nil || recurringRedemption.Grant == nil {
		t.Fatalf("RedeemCoupon(recurring) = (%+v, %v)", recurringRedemption, err)
	}
	recurringQuote, err := store.QuotePurchaseCoupon(ctx, coupons.PurchaseContext{UserID: user.ID, GrantID: recurringRedemption.Grant.ID,
		ComboID: "any-combo", GrossPriceMinor: 1_000}, time.Now())
	if err != nil || recurringQuote.DiscountMinor != 225 || !recurringQuote.Recurring {
		t.Fatalf("recurring quote = (%+v, %v)", recurringQuote, err)
	}
}
