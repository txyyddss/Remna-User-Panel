package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// PaymentCallbackReplay reports whether a verified provider callback was
// already applied, including after its terminal payment detail was compacted.
func (s *Store) PaymentCallbackReplay(ctx context.Context, provider, dedupeKey, orderID string) (bool, error) {
	var existingOrderID string
	err := s.db.QueryRowContext(ctx, `SELECT order_id FROM payment_callback_tombstones
		WHERE provider=? AND dedupe_key=?`, provider, dedupeKey).Scan(&existingOrderID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("load payment callback tombstone: %w", err)
	}
	if existingOrderID != orderID {
		return false, ErrConflict
	}
	return true, nil
}
