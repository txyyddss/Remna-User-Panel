package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/coupons"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

const (
	AutoRenewalReasonQueuedPurchase       = "QUEUED_PURCHASE"
	AutoRenewalReasonPurchaseUnavailable  = "PURCHASE_UNAVAILABLE"
	AutoRenewalReasonComboUnavailable     = "COMBO_UNAVAILABLE"
	AutoRenewalReasonPaidAddonUnavailable = "PAID_ADDON_UNAVAILABLE"
	AutoRenewalReasonDiscountUnavailable  = "RECURRING_DISCOUNT_UNAVAILABLE"
	AutoRenewalReasonInsufficientBalance  = "INSUFFICIENT_BALANCE"
	AutoRenewalReasonNoAccessibleNodes    = "NO_ACCESSIBLE_NODES"
	AutoRenewalReasonUnavailable          = "RENEWAL_UNAVAILABLE"
)

// AutoRenewalPlan is the local, current-price source for one next-cycle quote.
type AutoRenewalPlan struct {
	Purchase              model.Purchase
	Combo                 model.Combo
	Addons                []model.SquadProduct
	GrossMinor            int64
	DiscountMinor         int64
	NetMinor              int64
	ScheduledAt           time.Time
	NextCycleEndsAt       time.Time
	IneligibleReason      string
	trafficLimitOverride  *int64
	resetStrategyOverride *string
	squadUUIDsOverride    *string
}

// AutoRenewalPlan returns current local pricing and availability for an owned term.
func (s *Store) AutoRenewalPlan(ctx context.Context, userID, purchaseID string, now time.Time) (AutoRenewalPlan, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return AutoRenewalPlan{}, fmt.Errorf("begin automatic renewal quote: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	return automaticRenewalPlanTx(ctx, tx, userID, purchaseID, now.UTC())
}

func automaticRenewalPlanTx(ctx context.Context, tx *sql.Tx, userID, purchaseID string, now time.Time) (AutoRenewalPlan, error) {
	purchase, err := scanPurchase(tx.QueryRowContext(ctx, purchaseSelect+` WHERE purchases.id=? AND purchases.user_id=?`, purchaseID, userID))
	if err != nil {
		return AutoRenewalPlan{}, err
	}
	plan := AutoRenewalPlan{Purchase: purchase, GrossMinor: purchase.GrossPriceTXBMinor,
		DiscountMinor: purchase.CouponDiscountTXBMinor, NetMinor: purchase.PriceTXBMinor, ScheduledAt: purchase.ValidUntil}
	if duration := purchase.ValidUntil.Sub(purchase.ValidFrom); duration > 0 {
		plan.NextCycleEndsAt = plan.ScheduledAt.Add(duration)
	} else {
		plan.NextCycleEndsAt = plan.ScheduledAt
	}
	var trafficLimit sql.NullInt64
	var resetStrategy, squadUUIDs sql.NullString
	if err = tx.QueryRowContext(ctx, `SELECT entitlement_traffic_limit_bytes,entitlement_reset_strategy,entitlement_squad_uuids FROM purchases WHERE id=?`, purchaseID).Scan(
		&trafficLimit, &resetStrategy, &squadUUIDs); err != nil {
		return AutoRenewalPlan{}, fmt.Errorf("load renewal entitlement overrides: %w", err)
	}
	if trafficLimit.Valid {
		value := trafficLimit.Int64
		plan.trafficLimitOverride = &value
	}
	if resetStrategy.Valid {
		value := resetStrategy.String
		plan.resetStrategyOverride = &value
	}
	if squadUUIDs.Valid {
		value := squadUUIDs.String
		plan.squadUUIDsOverride = &value
	}
	if purchase.Status != "active" && purchase.Status != "activating" {
		if purchase.Status != "expired" {
			plan.IneligibleReason = AutoRenewalReasonPurchaseUnavailable
			return plan, nil
		}
		var settled int
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM purchase_rollovers WHERE purchase_id=? AND status IN ('credited','zero'))`, purchaseID).Scan(&settled); err != nil {
			return AutoRenewalPlan{}, fmt.Errorf("check settled rollover: %w", err)
		}
		if settled != 1 {
			plan.IneligibleReason = AutoRenewalReasonPurchaseUnavailable
			return plan, nil
		}
	}
	var queued int
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM purchases WHERE user_id=? AND status='queued' AND id<>?)`, userID, purchaseID).Scan(&queued); err != nil {
		return AutoRenewalPlan{}, fmt.Errorf("check queued purchase: %w", err)
	}
	if queued == 1 {
		plan.IneligibleReason = AutoRenewalReasonQueuedPurchase
	}
	combo, err := comboByIDTx(ctx, tx, purchase.ComboID, true)
	if errors.Is(err, ErrNotFound) {
		plan.IneligibleReason = AutoRenewalReasonComboUnavailable
		return plan, nil
	}
	if err != nil {
		return AutoRenewalPlan{}, err
	}
	plan.Combo = combo
	plan.GrossMinor, plan.DiscountMinor, plan.NetMinor = combo.PriceTXBMinor, 0, combo.PriceTXBMinor
	addons, err := renewalAddonsTx(ctx, tx, purchase.ID)
	if err != nil {
		return AutoRenewalPlan{}, err
	}
	plan.Addons = addons
	addonIDs := make([]string, 0, len(addons))
	for _, addon := range addons {
		plan.GrossMinor += addon.PriceTXBMinor
		addonIDs = append(addonIDs, addon.RemnaSquadUUID)
	}
	plan.NetMinor = plan.GrossMinor
	discount := coupons.Discount{GrossMinor: plan.GrossMinor, NetMinor: plan.GrossMinor}
	if purchase.RecurringDiscountAttached {
		if purchase.CouponGrantID == nil {
			plan.IneligibleReason = AutoRenewalReasonDiscountUnavailable
			return plan, nil
		}
		discount, err = quoteAttachedRecurringDiscountTx(ctx, tx, userID, *purchase.CouponGrantID, plan.GrossMinor)
		if errors.Is(err, ErrNotFound) || errors.Is(err, ErrConflict) {
			plan.IneligibleReason = AutoRenewalReasonDiscountUnavailable
			return plan, nil
		}
		if err != nil {
			return AutoRenewalPlan{}, err
		}
	}
	plan.DiscountMinor, plan.NetMinor = discount.DiscountMinor, discount.NetMinor
	if err := checkSquadStockTx(ctx, tx, addonIDs, userID); errors.Is(err, ErrStockUnavailable) {
		plan.IneligibleReason = AutoRenewalReasonPaidAddonUnavailable
	} else if err != nil {
		return AutoRenewalPlan{}, err
	}
	return plan, nil
}
