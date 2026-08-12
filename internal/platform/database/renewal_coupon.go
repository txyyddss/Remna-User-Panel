package database

import (
	"context"
	"database/sql"
	"github.com/txyyddss/Remna-User-Panel/internal/coupons"
	"time"
)

// quoteRecurringRenewalCouponTx reuses a recurring grant without recording a
// coupon use. A renewal is not a new coupon redemption and therefore must not
// consume global or per-user coupon limits.
func quoteRecurringRenewalCouponTx(ctx context.Context, tx *sql.Tx, input coupons.PurchaseContext, now time.Time) (coupons.Discount, error) {
	base := coupons.Discount{GrossMinor: input.GrossPriceMinor, NetMinor: input.GrossPriceMinor}
	grant, err := grantByIDTx(ctx, tx, input.GrantID)
	if err != nil {
		return coupons.Discount{}, err
	}
	if grant.Coupon.Kind != coupons.KindPurchaseRecurring {
		return base, nil
	}
	if grant.UserID != input.UserID || grant.Status != "active" {
		return coupons.Discount{}, ErrConflict
	}
	discarded, err := grantDiscardedTx(ctx, tx, grant.ID)
	if err != nil {
		return coupons.Discount{}, err
	}
	if discarded || !grant.Coupon.EligibleFor(input.ComboID, input.AddonSquadIDs) {
		return coupons.Discount{}, ErrConflict
	}
	if err := couponAvailable(grant.Coupon, now); err != nil {
		return coupons.Discount{}, err
	}
	value, err := coupons.CalculateDiscount(grant.Coupon, input.GrossPriceMinor)
	if err != nil {
		return coupons.Discount{}, err
	}
	return coupons.Discount{GrantID: grant.ID, CouponID: grant.Coupon.ID, CouponCode: grant.Coupon.Code,
		GrossMinor: input.GrossPriceMinor, DiscountMinor: value, NetMinor: input.GrossPriceMinor - value, Recurring: true}, nil
}
