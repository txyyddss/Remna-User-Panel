package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/txyyddss/Remna-User-Panel/internal/activity"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
	"math"
	"math/big"
	"time"
)

func applySubscriptionExtensionTx(ctx context.Context, tx *sql.Tx, userID string, days int, sourceType, sourceID string, now time.Time) error {
	if days < 1 || days > 3650 {
		return activity.ErrInvalidInput
	}
	var purchaseID, validUntilRaw string
	err := tx.QueryRowContext(ctx, `SELECT id,valid_until FROM purchases WHERE user_id=? AND status IN ('activating','active') AND valid_from<=? AND valid_until>?
		ORDER BY valid_from DESC LIMIT 1`, userID, stamp(now), stamp(now)).Scan(&purchaseID, &validUntilRaw)
	if errors.Is(err, sql.ErrNoRows) {
		creditID, idErr := ids.New()
		if idErr != nil {
			return idErr
		}
		_, insertErr := tx.ExecContext(ctx, `INSERT INTO activity_extension_credits(id,user_id,days,source_type,source_id,created_at) VALUES(?,?,?,?,?,?)`,
			creditID, userID, days, sourceType, sourceID, stamp(now))
		return insertErr
	}
	if err != nil {
		return err
	}
	validUntil, err := parseStamp(validUntilRaw)
	if err != nil {
		return err
	}
	shiftedUntil, err := addSubscriptionDays(validUntil, days)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE purchases SET valid_until=?,updated_at=? WHERE id=?`, stamp(shiftedUntil), stamp(now), purchaseID); err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id,valid_from,valid_until FROM purchases WHERE user_id=? AND status='queued' AND valid_from>=? ORDER BY valid_from`, userID, validUntilRaw)
	if err != nil {
		return err
	}
	type queuedTerm struct{ id, from, until string }
	queued := make([]queuedTerm, 0)
	for rows.Next() {
		var term queuedTerm
		if err := rows.Scan(&term.id, &term.from, &term.until); err != nil {
			_ = rows.Close()
			return err
		}
		queued = append(queued, term)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	_ = rows.Close()
	for _, term := range queued {
		from, parseErr := parseStamp(term.from)
		if parseErr != nil {
			return parseErr
		}
		until, parseErr := parseStamp(term.until)
		if parseErr != nil {
			return parseErr
		}
		shiftedFrom, shiftErr := addSubscriptionDays(from, days)
		if shiftErr != nil {
			return shiftErr
		}
		shiftedUntil, shiftErr := addSubscriptionDays(until, days)
		if shiftErr != nil {
			return shiftErr
		}
		if _, err := tx.ExecContext(ctx, `UPDATE purchases SET valid_from=?,valid_until=?,updated_at=? WHERE id=?`, stamp(shiftedFrom), stamp(shiftedUntil), stamp(now), term.id); err != nil {
			return err
		}
	}
	return nil
}

// consumePendingExtensionsTx marks stored extension credits used by the term
// that is actually becoming active. Queued purchases must not consume credits
// early because an older queued term remains the member's next activation.
func consumePendingExtensionsTx(ctx context.Context, tx *sql.Tx, userID, purchaseID string, now time.Time) (int, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id,days FROM activity_extension_credits WHERE user_id=? AND consumed_at IS NULL ORDER BY created_at,id`, userID)
	if err != nil {
		return 0, err
	}
	idsToConsume := make([]string, 0)
	total := int64(0)
	for rows.Next() {
		var id string
		var days int
		if err := rows.Scan(&id, &days); err != nil {
			_ = rows.Close()
			return 0, err
		}
		if total > math.MaxInt64-int64(days) {
			_ = rows.Close()
			return 0, activity.ErrInvalidInput
		}
		total += int64(days)
		idsToConsume = append(idsToConsume, id)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return 0, err
	}
	_ = rows.Close()
	for _, creditID := range idsToConsume {
		if _, err := tx.ExecContext(ctx, `UPDATE activity_extension_credits SET consumed_at=?,consumed_by_purchase_id=? WHERE id=? AND consumed_at IS NULL`, stamp(now), purchaseID, creditID); err != nil {
			return 0, err
		}
	}
	if total > int64(^uint(0)>>1) {
		return 0, activity.ErrInvalidInput
	}
	return int(total), nil
}

// applyPendingExtensionsToActivationTx applies credits saved while no term was
// active to the exact next activating term, then delays every following queued
// term so subscription periods remain contiguous and non-overlapping.
func applyPendingExtensionsToActivationTx(ctx context.Context, tx *sql.Tx, purchaseID string, now time.Time) error {
	var userID, validUntilRaw string
	if err := tx.QueryRowContext(ctx, `SELECT user_id,valid_until FROM purchases WHERE id=? AND status='activating'`, purchaseID).Scan(&userID, &validUntilRaw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrConflict
		}
		return err
	}
	days, err := consumePendingExtensionsTx(ctx, tx, userID, purchaseID, now)
	if err != nil || days == 0 {
		return err
	}
	validUntil, err := parseStamp(validUntilRaw)
	if err != nil {
		return err
	}
	shiftedUntil, err := addSubscriptionDays(validUntil, days)
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `UPDATE purchases SET valid_until=?,updated_at=? WHERE id=? AND status='activating'`, stamp(shiftedUntil), stamp(now), purchaseID)
	if err != nil {
		return err
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
		return rowsErr
	} else if affected != 1 {
		return ErrConflict
	}
	rows, err := tx.QueryContext(ctx, `SELECT id,valid_from,valid_until FROM purchases
		WHERE user_id=? AND id<>? AND status='queued' AND valid_from>=? ORDER BY valid_from,id`, userID, purchaseID, validUntilRaw)
	if err != nil {
		return err
	}
	type queuedTerm struct{ id, from, until string }
	queued := make([]queuedTerm, 0)
	for rows.Next() {
		var term queuedTerm
		if err := rows.Scan(&term.id, &term.from, &term.until); err != nil {
			_ = rows.Close()
			return err
		}
		queued = append(queued, term)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	_ = rows.Close()
	for _, term := range queued {
		from, parseErr := parseStamp(term.from)
		if parseErr != nil {
			return parseErr
		}
		until, parseErr := parseStamp(term.until)
		if parseErr != nil {
			return parseErr
		}
		shiftedFrom, shiftErr := addSubscriptionDays(from, days)
		if shiftErr != nil {
			return shiftErr
		}
		shiftedUntil, shiftErr := addSubscriptionDays(until, days)
		if shiftErr != nil {
			return shiftErr
		}
		result, updateErr := tx.ExecContext(ctx, `UPDATE purchases SET valid_from=?,valid_until=?,updated_at=? WHERE id=? AND status='queued'`, stamp(shiftedFrom), stamp(shiftedUntil), stamp(now), term.id)
		if updateErr != nil {
			return updateErr
		}
		if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
			return rowsErr
		} else if affected != 1 {
			return ErrConflict
		}
	}
	return nil
}

func addSubscriptionDays(value time.Time, days int) (time.Time, error) {
	shifted := value.AddDate(0, 0, days)
	if shifted.Year() < 1 || shifted.Year() > 9999 || !shifted.After(value) {
		return time.Time{}, activity.ErrInvalidInput
	}
	return shifted, nil
}

func fixedMultiplyFloor(left, right, divisor int64) (int64, error) {
	if left < 0 || right < 0 || divisor <= 0 {
		return 0, activity.ErrInvalidInput
	}
	value := new(big.Int).Mul(big.NewInt(left), big.NewInt(right))
	value.Quo(value, big.NewInt(divisor))
	if !value.IsInt64() {
		return 0, activity.ErrInvalidInput
	}
	return value.Int64(), nil
}
