package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

// DueAutoRenewal identifies a currently enabled source term due for renewal.
type DueAutoRenewal struct {
	PurchaseID string
	UserID     string
}

// HasEnabledAutoRenewal reports whether an active entitlement chain blocks catalog checkout.
func (s *Store) HasEnabledAutoRenewal(ctx context.Context, userID string, _ time.Time) (bool, error) {
	var enabled int
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM purchases
		WHERE user_id=? AND auto_renew_enabled=1 AND status IN ('active','activating','queued'))`, userID).Scan(&enabled)
	if err != nil {
		return false, fmt.Errorf("check automatic renewal: %w", err)
	}
	return enabled == 1, nil
}

// SetAutoRenewal stores the member's explicit choice for the active term.
func (s *Store) SetAutoRenewal(ctx context.Context, userID, purchaseID string, enabled bool, now time.Time) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin automatic renewal update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(ctx, `UPDATE purchases SET auto_renew_enabled=?,auto_renew_failure_reason=CASE WHEN ?=1 THEN '' ELSE auto_renew_failure_reason END,
		auto_renew_failed_at=CASE WHEN ?=1 THEN NULL ELSE auto_renew_failed_at END,updated_at=?
		WHERE id=? AND user_id=? AND status IN ('active','activating')`, boolInt(enabled), boolInt(enabled), boolInt(enabled), stamp(now), purchaseID, userID)
	if err != nil {
		return fmt.Errorf("set automatic renewal: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect automatic renewal update: %w", err)
	}
	if affected == 1 {
		if !enabled {
			if _, err := tx.ExecContext(ctx, `UPDATE purchases SET auto_renew_enabled=0,updated_at=?
				WHERE auto_renew_source_purchase_id=? AND status='queued'`, stamp(now), purchaseID); err != nil {
				return fmt.Errorf("disable queued automatic renewal successor: %w", err)
			}
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit automatic renewal update: %w", err)
		}
		return nil
	}
	if err := tx.Rollback(); err != nil {
		return fmt.Errorf("close automatic renewal update: %w", err)
	}
	var exists int
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM purchases WHERE id=? AND user_id=?)`, purchaseID, userID).Scan(&exists); err != nil {
		return fmt.Errorf("find automatic renewal purchase: %w", err)
	}
	if exists == 0 {
		return ErrNotFound
	}
	return ErrConflict
}

// DueAutoRenewals lists terms that must be revalidated before expiry work begins.
func (s *Store) DueAutoRenewals(ctx context.Context, now time.Time) ([]DueAutoRenewal, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,user_id FROM purchases WHERE auto_renew_enabled=1
		AND status IN ('active','activating') AND valid_until<=?
		AND NOT EXISTS (SELECT 1 FROM purchases successor WHERE successor.auto_renew_source_purchase_id=purchases.id
			AND successor.status IN ('queued','activating','active'))
		ORDER BY valid_until,id`, stamp(now))
	if err != nil {
		return nil, fmt.Errorf("list due automatic renewals: %w", err)
	}
	defer func() { _ = rows.Close() }()
	result := make([]DueAutoRenewal, 0)
	for rows.Next() {
		var item DueAutoRenewal
		if err := rows.Scan(&item.PurchaseID, &item.UserID); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

// MarkAutoRenewalFailed disables a due source without charging its owner.
func (s *Store) MarkAutoRenewalFailed(ctx context.Context, purchaseID, reason string, now time.Time) error {
	if reason == "" {
		return ErrConflict
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_, err := s.db.ExecContext(ctx, `UPDATE purchases SET auto_renew_enabled=0,auto_renew_failure_reason=?,auto_renew_failed_at=?,updated_at=?
		WHERE id=? AND auto_renew_enabled=1 AND status IN ('active','activating') AND valid_until<=?`, reason, stamp(now), stamp(now), purchaseID, stamp(now))
	if err != nil {
		return fmt.Errorf("record automatic renewal failure: %w", err)
	}
	return nil
}

// AutoRenewalFailure returns the latest member-visible automatic-renewal failure.
func (s *Store) AutoRenewalFailure(ctx context.Context, userID string) (*model.AutoRenewalFailure, error) {
	var failure model.AutoRenewalFailure
	var failedAt sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT id,auto_renew_failure_reason,auto_renew_failed_at FROM purchases
		WHERE user_id=? AND auto_renew_failure_reason<>'' AND auto_renew_failed_at IS NOT NULL ORDER BY auto_renew_failed_at DESC LIMIT 1`, userID).Scan(&failure.PurchaseID, &failure.Reason, &failedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load automatic renewal failure: %w", err)
	}
	failed, err := parseStamp(failedAt.String)
	if err != nil {
		return nil, err
	}
	failure.FailedAt = failed
	return &failure, nil
}
