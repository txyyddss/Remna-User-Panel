package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/providerops"
)

// CompensateTrafficReset credits one failed reset debit and closes its receipt atomically.
func (s *Store) CompensateTrafficReset(ctx context.Context, operationID, errorCode string, now time.Time) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	operation, err := scanProviderOperation(tx.QueryRowContext(ctx, providerOperationSelect+` WHERE id=?`, operationID))
	if err != nil {
		return err
	}
	if operation.Receipt.Status == string(providerops.StatusCompensated) {
		return nil
	}
	if operation.Receipt.Status != string(providerops.StatusProcessing) {
		return ErrConflict
	}
	var userID string
	var debit int64
	err = tx.QueryRowContext(ctx, `SELECT user_id,delta_txb_minor FROM ledger_entries
		WHERE kind='traffic_reset_debit' AND reference_id=?`, operationID).Scan(&userID, &debit)
	if err != nil {
		return fmt.Errorf("load traffic reset debit: %w", err)
	}
	if debit >= 0 {
		return errors.New("traffic reset debit is invalid")
	}
	var existing string
	err = tx.QueryRowContext(ctx, `SELECT id FROM ledger_entries WHERE kind='traffic_reset_compensation' AND reference_id=?`, operationID).Scan(&existing)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if errors.Is(err, sql.ErrNoRows) {
		balance, changeErr := changeBalanceTx(ctx, tx, userID, -debit, now)
		if changeErr != nil {
			return changeErr
		}
		if _, changeErr = insertLedgerTx(ctx, tx, userID, -debit, balance, "traffic_reset_compensation", operationID, "traffic reset compensation", now); changeErr != nil {
			return changeErr
		}
	}
	code := operationCode(errorCode)
	if _, err = tx.ExecContext(ctx, `UPDATE provider_operation_items SET status='compensated',error_code=?,result_json='{}',
		completed_at=?,updated_at=? WHERE operation_id=? AND status='processing'`, code, stamp(now), stamp(now), operationID); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `UPDATE provider_operations SET status='compensated',error_code=?,result_json='{}',
		completed_at=?,updated_at=? WHERE id=? AND status='processing'`, code, stamp(now), stamp(now), operationID)
	if err != nil {
		return err
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
		return ErrConflict
	}
	return tx.Commit()
}
