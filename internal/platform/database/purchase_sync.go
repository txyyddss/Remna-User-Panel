package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"time"
)

// MarkPurchaseSyncResult advances an activating purchase after upstream work.
func (s *Store) MarkPurchaseSyncResult(ctx context.Context, purchaseID string, success bool, now time.Time) error {
	status := "failed"
	predicate := "status='activating'"
	if success {
		status = "active"
		predicate = "status IN ('activating','failed')"
	}
	_, err := s.db.ExecContext(ctx, `UPDATE purchases SET status=?,updated_at=? WHERE id=? AND `+predicate, status, stamp(now), purchaseID)
	return err
}

// PurchaseTrafficResetPhase returns the durable phase of a new-term reset.
func (s *Store) PurchaseTrafficResetPhase(ctx context.Context, purchaseID string) (string, error) {
	var phase string
	err := s.db.QueryRowContext(ctx, `SELECT traffic_reset_phase FROM purchases WHERE id=?`, purchaseID).Scan(&phase)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("load purchase traffic reset phase: %w", err)
	}
	return phase, nil
}

// AdvancePurchaseTrafficReset atomically records a completed external phase.
// Repeating the same transition is accepted so a DB response loss is harmless.
func (s *Store) AdvancePurchaseTrafficReset(ctx context.Context, purchaseID, from, to string, now time.Time) error {
	valid := (from == "pending" && to == "quiesced") || (from == "quiesced" && to == "reset")
	if !valid {
		return ErrConflict
	}
	result, err := s.db.ExecContext(ctx, `UPDATE purchases SET traffic_reset_phase=?,updated_at=?
		WHERE id=? AND status IN ('activating','failed') AND traffic_reset_phase=?`, to, stamp(now), purchaseID, from)
	if err != nil {
		return fmt.Errorf("advance purchase traffic reset: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect purchase traffic reset transition: %w", err)
	}
	if affected == 1 {
		return nil
	}
	phase, loadErr := s.PurchaseTrafficResetPhase(ctx, purchaseID)
	if loadErr != nil {
		return loadErr
	}
	if phase == to {
		return nil
	}
	return ErrConflict
}

// UserForPurchase returns the local user owning a purchase.
func (s *Store) UserForPurchase(ctx context.Context, purchaseID string) (model.User, error) {
	return scanUser(s.db.QueryRowContext(ctx, userSelect+` JOIN purchases ON purchases.user_id=users.id WHERE purchases.id=?`, purchaseID))
}

// DesiredEntitlement returns the currently effective purchase, if any.
func (s *Store) DesiredEntitlement(ctx context.Context, userID string, now time.Time) (*model.Purchase, error) {
	purchase, err := scanPurchase(s.db.QueryRowContext(ctx, purchaseSelect+` WHERE purchases.user_id=? AND purchases.status IN ('activating','active') AND purchases.valid_from<=? AND purchases.valid_until>? ORDER BY purchases.valid_from DESC LIMIT 1`, userID, stamp(now), stamp(now)))
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	purchase.SquadUUIDs, err = s.purchaseSquads(ctx, purchase.ID)
	return &purchase, err
}
