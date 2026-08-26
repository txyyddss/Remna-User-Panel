package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
)

// CommitAutoRenewal atomically debits one due source and creates its one queued successor.
func (s *Store) CommitAutoRenewal(ctx context.Context, purchaseID string, now time.Time) (model.Purchase, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	now = now.UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Purchase{}, fmt.Errorf("begin automatic renewal: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if successorID, found, err := automaticRenewalSuccessorIDTx(ctx, tx, purchaseID); err != nil {
		return model.Purchase{}, err
	} else if found {
		if err := tx.Rollback(); err != nil {
			return model.Purchase{}, err
		}
		return s.PurchaseByID(ctx, successorID)
	}
	var userID string
	if err := tx.QueryRowContext(ctx, `SELECT user_id FROM purchases WHERE id=?`, purchaseID).Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Purchase{}, ErrNotFound
		}
		return model.Purchase{}, fmt.Errorf("load automatic renewal owner: %w", err)
	}
	plan, err := automaticRenewalPlanTx(ctx, tx, userID, purchaseID, now)
	if err != nil {
		return model.Purchase{}, err
	}
	if !plan.Purchase.AutoRenewEnabled || plan.ScheduledAt.After(now.Add(EntitlementContinuityLead)) || plan.IneligibleReason != "" {
		return model.Purchase{}, ErrConflict
	}
	newBalance, err := debitBalanceTx(ctx, tx, userID, plan.NetMinor, now)
	if err != nil {
		return model.Purchase{}, err
	}
	successorID, err := ids.New()
	if err != nil {
		return model.Purchase{}, err
	}
	var couponGrantID any
	if plan.Purchase.RecurringDiscountAttached && plan.Purchase.CouponGrantID != nil {
		couponGrantID = *plan.Purchase.CouponGrantID
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO purchases(id,user_id,combo_id,charged_txb_minor,valid_from,valid_until,status,coupon_grant_id,
		gross_price_txb_minor,core_gross_txb_minor,coupon_discount_txb_minor,auto_renew_enabled,recurring_discount_attached,auto_renew_source_purchase_id,request_fingerprint,
		entitlement_traffic_limit_bytes,entitlement_reset_strategy,entitlement_squad_uuids,created_at,updated_at)
		VALUES(?,?,?,?,?,?,'queued',?,?,?,?,?,?,?,?,?,?,?,?,?)`, successorID, userID, plan.Combo.ID, plan.NetMinor,
		stamp(plan.ScheduledAt), stamp(plan.NextCycleEndsAt), couponGrantID, plan.GrossMinor, plan.Combo.PriceTXBMinor, plan.DiscountMinor, 1,
		boolInt(plan.Purchase.RecurringDiscountAttached), purchaseID, "automatic-renewal:"+purchaseID, plan.trafficLimitOverride, plan.resetStrategyOverride,
		plan.squadUUIDsOverride, stamp(now), stamp(now))
	if err != nil {
		if isUniqueConstraint(err) {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return model.Purchase{}, rollbackErr
			}
			return s.autoRenewalSuccessor(ctx, purchaseID)
		}
		return model.Purchase{}, fmt.Errorf("insert automatic renewal successor: %w", err)
	}
	for _, addon := range plan.Addons {
		if _, err := tx.ExecContext(ctx, `INSERT INTO purchase_addons(purchase_id,remna_squad_uuid,charged_txb_minor) VALUES(?,?,?)`, successorID, addon.RemnaSquadUUID, addon.PriceTXBMinor); err != nil {
			return model.Purchase{}, fmt.Errorf("snapshot automatic renewal add-on: %w", err)
		}
	}
	if err := enqueuePurchaseTransitionTx(ctx, tx, successorID, "queued", plan.ScheduledAt, now); err != nil {
		return model.Purchase{}, err
	}
	if _, err := insertLedgerTx(ctx, tx, userID, -plan.NetMinor, newBalance, "automatic_renewal", successorID, plan.Combo.Name, now); err != nil {
		return model.Purchase{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.Purchase{}, fmt.Errorf("commit automatic renewal: %w", err)
	}
	return s.PurchaseByID(ctx, successorID)
}

func automaticRenewalSuccessorIDTx(ctx context.Context, tx *sql.Tx, sourceID string) (string, bool, error) {
	var successorID string
	err := tx.QueryRowContext(ctx, `SELECT id FROM purchases WHERE auto_renew_source_purchase_id=?`, sourceID).Scan(&successorID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("load automatic renewal successor: %w", err)
	}
	return successorID, true, nil
}

func (s *Store) autoRenewalSuccessor(ctx context.Context, sourceID string) (model.Purchase, error) {
	var successorID string
	if err := s.db.QueryRowContext(ctx, `SELECT id FROM purchases WHERE auto_renew_source_purchase_id=?`, sourceID).Scan(&successorID); err != nil {
		return model.Purchase{}, err
	}
	return s.PurchaseByID(ctx, successorID)
}
