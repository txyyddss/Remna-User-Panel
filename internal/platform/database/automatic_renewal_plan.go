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
	Purchase         model.Purchase
	Combo            model.Combo
	Addons           []model.SquadProduct
	GrossMinor       int64
	DiscountMinor    int64
	NetMinor         int64
	ScheduledAt      time.Time
	NextCycleEndsAt  time.Time
	IneligibleReason string
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
	if purchase.Status != "active" && purchase.Status != "activating" {
		plan.IneligibleReason = AutoRenewalReasonPurchaseUnavailable
		return plan, nil
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
	plan.NextCycleEndsAt = plan.ScheduledAt.AddDate(0, 0, combo.ValidityDays)
	plan.GrossMinor, plan.DiscountMinor, plan.NetMinor = combo.PriceTXBMinor, 0, combo.PriceTXBMinor
	addons, err := automaticRenewalAddonsTx(ctx, tx, purchase.ID)
	if errors.Is(err, ErrNotFound) {
		plan.IneligibleReason = AutoRenewalReasonPaidAddonUnavailable
		return plan, nil
	}
	if err != nil {
		return AutoRenewalPlan{}, err
	}
	plan.Addons = addons
	addonIDs := make([]string, 0, len(addons))
	for _, addon := range addons {
		plan.GrossMinor += addon.PriceTXBMinor
		addonIDs = append(addonIDs, addon.RemnaSquadUUID)
	}
	if err := checkSquadStockTx(ctx, tx, addonIDs, userID); errors.Is(err, ErrStockUnavailable) {
		plan.IneligibleReason = AutoRenewalReasonPaidAddonUnavailable
		return plan, nil
	} else if err != nil {
		return AutoRenewalPlan{}, err
	}
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
	return plan, nil
}

func automaticRenewalAddonsTx(ctx context.Context, tx *sql.Tx, purchaseID string) ([]model.SquadProduct, error) {
	rows, err := tx.QueryContext(ctx, `SELECT remna_squad_uuid FROM purchase_addons WHERE purchase_id=? ORDER BY remna_squad_uuid`, purchaseID)
	if err != nil {
		return nil, fmt.Errorf("load automatic renewal add-ons: %w", err)
	}
	defer func() { _ = rows.Close() }()
	addons := make([]model.SquadProduct, 0)
	for rows.Next() {
		var squadID string
		if err := rows.Scan(&squadID); err != nil {
			return nil, err
		}
		addon, err := squadByIDTx(ctx, tx, squadID, true)
		if err != nil {
			return nil, err
		}
		addons = append(addons, addon)
	}
	return addons, rows.Err()
}
