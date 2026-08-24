package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/abuse"
)

func warningCooldownBlockedTx(ctx context.Context, tx *sql.Tx, userID string, policy abuse.Policy, now time.Time) (bool, error) {
	if policy.WarningCooldownMinutes == 0 {
		return false, nil
	}
	cutoff := stamp(now.Add(-time.Duration(policy.WarningCooldownMinutes) * time.Minute))
	var count int
	err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM abuse_records WHERE user_id=? AND selected_action='warning' AND deleted_at IS NULL AND created_at>=?`, userID, cutoff).Scan(&count)
	return count > 0, err
}
