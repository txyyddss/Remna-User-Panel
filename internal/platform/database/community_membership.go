package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// HasActiveCombo reports whether a user has a current active combo. It is
// deliberately narrower than home summaries: activating, queued, future, and
// expired purchases never grant Telegram community access.
func (s *Store) HasActiveCombo(ctx context.Context, userID string, now time.Time) (bool, error) {
	if strings.TrimSpace(userID) == "" {
		return false, ErrConflict
	}
	var found int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM purchases
		WHERE user_id=? AND status='active' AND valid_from<=? AND valid_until>?
		LIMIT 1`, userID, stamp(now), stamp(now)).Scan(&found)
	if err == nil {
		return true, nil
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	return false, fmt.Errorf("check active community combo: %w", err)
}
